package request

// Pacotes nativos de go e pacotes internos
import (
	"errors"
	"net"
	"strconv"
	"sync"
	"time"

	"EACHare/src/clock"
	"EACHare/src/commands/message"
	"EACHare/src/logger"
	"EACHare/src/peers"
)

// Interface para definir os métodos de requisição
type IRequest interface {
	HelloRequest(receiverAddress string, knownPeers *sync.Map)
	GetPeersRequest(knownPeers *sync.Map) []net.Conn
	PeersListRequest(conn net.Conn, receivedMessage message.BaseMessage, knownPeers *sync.Map)
	ByeRequest(knownPeers *sync.Map)
}

// Estrutura para o cliente que faz requisições
type RequestClient struct {
	Address string
}

// Função para enviar mensagem
func (r RequestClient) sendMessage(connection net.Conn, message message.BaseMessage, receiverAddress string, knownPeers *sync.Map) {
	// Atualiza o clock e mostra o encaminhamento
	message.Clock = clock.UpdateClock()
	logger.Info("\tEncaminhando mensagem \"" + message.String() + "\" para " + receiverAddress)

	// Tenta enviar a mensagem e verificar se há um erro
	var err error
	if connection == nil {
		err = errors.New("connection is nil")
	} else {
		_, err = connection.Write([]byte(message.String() + "\n"))
	}

	// Atualiza o peer e imprime mensagem apenas quando o status muda
	neighbor, _ := knownPeers.Load(receiverAddress)
	neighborStatus := neighbor.(peers.Peer).Status
	neighborClock := neighbor.(peers.Peer).Clock
	if err != nil && neighborStatus == peers.ONLINE {
		logger.Info("\tAtualizando peer " + receiverAddress + " status " + peers.OFFLINE.String())
		knownPeers.Store(receiverAddress, peers.Peer{Status: peers.OFFLINE, Clock: neighborClock})
	} else if err == nil && neighborStatus == peers.OFFLINE {
		logger.Info("\tAtualizando peer " + receiverAddress + " status " + peers.ONLINE.String())
		knownPeers.Store(receiverAddress, peers.Peer{Status: peers.ONLINE, Clock: neighborClock})
	}
}

// Função para mensagem HELLO, avisa o peer que estou online
func (r RequestClient) HelloRequest(receiverAddress string, knownPeers *sync.Map) {
	// Cria uma mensagem HELLO
	baseMessage := message.BaseMessage{Origin: r.Address, Clock: 0, Type: message.HELLO, Arguments: nil}

	// Tenta conectar e se conectar, define o deadline de 2 segundos
	conn, _ := net.Dial("tcp", receiverAddress)
	if conn != nil {
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(2 * time.Second))
	}

	// Envia a mensagem HELLO
	r.sendMessage(conn, baseMessage, receiverAddress, knownPeers)
}

// Função para mensagem GET_PEERS, solicita para os vizinhos sobre quem eles conhecem
func (r RequestClient) GetPeersRequest(knownPeers *sync.Map) []net.Conn {
	// Cria uma lista de conexões e a estrutura da mensagem GET_PEERS
	peerCount := 0
	knownPeers.Range(func(_, _ any) bool {
		peerCount++
		return true
	})
	connections := make([]net.Conn, 0, peerCount)
	baseMessage := message.BaseMessage{Origin: r.Address, Clock: 0, Type: message.GET_PEERS, Arguments: nil}

	// Envia mensagem GET_PEERS para cada peer conhecido e adiciona a conexão na lista
	knownPeers.Range(func(key, value any) bool {
		address := key.(string)
		conn, _ := net.Dial("tcp", address)
		if conn != nil {
			connections = append(connections, conn)
			conn.SetDeadline(time.Now().Add(2 * time.Second))
		}
		r.sendMessage(conn, baseMessage, address, knownPeers)
		return true
	})

	// Retorna as conexões estabelecidas para reutilizar
	return connections
}

// Função para mensagem PEER_LIST, envia os meus peers conhecidos para quem solicitou
func (r RequestClient) PeersListRequest(conn net.Conn, receivedMessage message.BaseMessage, knownPeers *sync.Map) {
	// Cria uma lista de strings para os peers conhecidos
	myPeers := make([]string, 0)

	// Adiciona cada peer conhecido na lista, exceto quem pediu a lista
	knownPeers.Range(func(key, value any) bool {
		address := key.(string)
		neighbor := value.(peers.Peer)
		if address == receivedMessage.Origin {
			return true
		}
		myPeers = append(myPeers, address+":"+neighbor.Status.String()+":"+strconv.Itoa(neighbor.Clock))
		return true
	})

	// Cria uma única string da lista inteira e depois cria e envia a mensagem
	arguments := append([]string{strconv.Itoa(len(myPeers))}, myPeers...)
	dropMessage := message.BaseMessage{Origin: r.Address, Clock: 0, Type: message.PEERS_LIST, Arguments: arguments}
	r.sendMessage(conn, dropMessage, receivedMessage.Origin, knownPeers)
}

// Função para mensagem BYE, avisando os peers sobre a saída
func (r RequestClient) ByeRequest(knownPeers *sync.Map) {
	// Imprime mensagem de saída e cria a mensagem BYE
	logger.Info("Saindo...")
	baseMessage := message.BaseMessage{Origin: r.Address, Clock: 0, Type: message.BYE, Arguments: nil}

	// Envia mensagem BYE para cada peer conhecido
	knownPeers.Range(func(key, _ any) bool {
		address := key.(string)
		conn, _ := net.Dial("tcp", address)
		if conn != nil {
			defer conn.Close()
			conn.SetDeadline(time.Now().Add(2 * time.Second))
		}
		r.sendMessage(conn, baseMessage, address, knownPeers)
		return true
	})
}
