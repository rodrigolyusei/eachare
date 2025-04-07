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
	GetPeersRequest(knownPeers *sync.Map) []net.Conn
	ByeRequest(knownPeers *sync.Map, exit *bool)
	PeersListRequest(conn net.Conn, receivedMessage message.BaseMessage, knownPeers *sync.Map)
	HelloRequest(receiverAddress string) peers.PeerStatus
}

// Estrutura para o cliente que faz requisições
type RequestClient struct {
	Address string
}

// Função para enviar mensagem
func (r RequestClient) sendMessage(connection net.Conn, message message.BaseMessage, receiverAddress string) error {
	// Atualiza o clock e mostra o encaminhamento
	message.Clock = clock.UpdateClock()
	logger.Info("\tEncaminhando mensagem \"" + message.String() + "\" para " + receiverAddress)

	// Se a conexão é nula retorna um erro
	if connection == nil {
		return errors.New("connection is nil")
	}

	// Envia a mensagem pela conexão
	_, err := connection.Write([]byte(message.String()))
	return err
}

// Função para mensagem HELLO, avisa o peer que estou online
func (r RequestClient) HelloRequest(receiverAddress string) peers.PeerStatus {
	// Cria uma mensagem HELLO
	baseMessage := message.BaseMessage{Origin: r.Address, Clock: 0, Type: message.HELLO, Arguments: nil}

	// Tenta conectar e se conectar, define o deadline de 2 segundos
	conn, _ := net.Dial("tcp", receiverAddress)
	if conn != nil {
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(2 * time.Second))
	}

	// Envia a mensagem HELLO e retorna o status do peer
	err := r.sendMessage(conn, baseMessage, receiverAddress)
	if err != nil {
		return peers.OFFLINE
	}
	return peers.ONLINE
}

// Função para mensagem GET_PEERS, solicita para os vizinhos sobre quem eles conhecem
func (r RequestClient) GetPeersRequest(knownPeers *sync.Map) []net.Conn {
	// Cria um slice de conexões e a estrutura da mensagem GET_PEERS
	peerCount := 0
	knownPeers.Range(func(_, _ interface{}) bool {
		peerCount++
		return true
	})
	connections := make([]net.Conn, 0, peerCount)
	baseMessage := message.BaseMessage{Origin: r.Address, Clock: 0, Type: message.GET_PEERS, Arguments: nil}

	// Itera sobre os peers conhecidos
	knownPeers.Range(func(key, value interface{}) bool {
		address := key.(string)

		// Tenta conectar e se conectar, adiciona a conexão à lista e define o deadline de 2 segundos
		conn, _ := net.Dial("tcp", address)
		if conn != nil {
			connections = append(connections, conn)
			conn.SetDeadline(time.Now().Add(2 * time.Second))
		}

		// Envia a mensagem GET_PEERS para e considera diferentes casos
		err := r.sendMessage(conn, baseMessage, address)
		if err != nil {
			// Se a conexão falhar e o peer estiver ONLINE, atualiza o status para OFFLINE
			if value == peers.ONLINE {
				logger.Info("\tAtualizando peer " + address + " status " + peers.OFFLINE.String())
				knownPeers.Store(address, peers.OFFLINE)
			}
		} else {
			// Se a conexão for bem-sucedida e o peer estiver OFFLINE, atualiza o status para ONLINE
			if value == peers.OFFLINE {
				logger.Info("\tAtualizando peer " + address + " status " + peers.ONLINE.String())
				knownPeers.Store(address, peers.ONLINE)
			}
		}
		return true
	})

	// Retorna as conexões estabelecidas para criar um receiver para cada um
	return connections
}

// Função para mensagem PEER_LIST, envia os meus peers conhecidos para quem solicitou
func (r RequestClient) PeersListRequest(conn net.Conn, receivedMessage message.BaseMessage, knownPeers *sync.Map) {
	// Cria uma lista de strings para os peers conhecidos
	peerCount := 0
	knownPeers.Range(func(_, _ interface{}) bool {
		peerCount++
		return true
	})
	myPeers := make([]string, 0, peerCount)

	// Adicioona cada peer que conhece na lista, exceto quem pediu a lista
	knownPeers.Range(func(key, value interface{}) bool {
		addressPort := key.(string)
		peerStatus := value.(peers.PeerStatus)

		if addressPort == receivedMessage.Origin {
			return true
		}
		myPeers = append(myPeers, addressPort+":"+peerStatus.String()+":"+"0")
		return true
	})

	// Cria uma única string da lista inteira e depois cria e envia a mensagem
	arguments := append([]string{strconv.Itoa(len(myPeers))}, myPeers...)
	dropMessage := message.BaseMessage{Origin: r.Address, Clock: 0, Type: message.PEERS_LIST, Arguments: arguments}
	r.sendMessage(conn, dropMessage, receivedMessage.Origin)
}

// Função para mensagem BYE, avisando os peers sobre a saída
func (r RequestClient) ByeRequest(knownPeers *sync.Map, exit *bool) {
	// Imprime mensagem de saída e cria a mensagem BYE
	logger.Info("Saindo...")
	baseMessage := message.BaseMessage{Origin: r.Address, Clock: 0, Type: message.BYE, Arguments: nil}

	// Itera sobre os peers conhecidos
	knownPeers.Range(func(key, value interface{}) bool {
		addressPort := key.(string)
		// Tenta conectar e se conectar define o deadline de 2 segundos
		conn, _ := net.Dial("tcp", addressPort)
		if conn != nil {
			defer conn.Close()
			conn.SetDeadline(time.Now().Add(2 * time.Second))
		}
		r.sendMessage(conn, baseMessage, addressPort)
		return true
	})

	// Altera a saída como verdadeiro para finalizar o programa
	*exit = true
}
