package request

// Pacotes nativos de go e pacotes internos
import (
	"errors"
	"net"
	"strconv"
	"time"

	"EACHare/src/clock"
	"EACHare/src/commands/message"
	"EACHare/src/logger"
	"EACHare/src/peers"
)

// Interface para definir os métodos de requisição
type IRequest interface {
	GetPeersRequest(knownPeers map[string]peers.PeerStatus) []net.Conn
	ByeRequest(knownPeers map[string]peers.PeerStatus) bool
	PeersListRequest(conn net.Conn, receivedMessage message.BaseMessage, knownPeers map[string]peers.PeerStatus)
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
	logger.Info("Encaminhando mensagem \"" + message.String() + "\" para " + receiverAddress)

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
	baseMessage := message.BaseMessage{Origin: r.Address, Clock: 0, Type: message.HELLO, Arguments: nil}
	conn, _ := net.Dial("tcp", receiverAddress)
	if conn != nil {
		defer conn.Close()
	}
	err := r.sendMessage(conn, baseMessage, receiverAddress)
	if err != nil {
		return peers.OFFLINE
	}
	return peers.ONLINE
}

// Função para mensagem GET_PEERS, solicita para os vizinhos sobre quem eles conhecem
func (r RequestClient) GetPeersRequest(knownPeers map[string]peers.PeerStatus) []net.Conn {
	// Cria um slice de conexões e a estrutura da mensagem GET_PEERS
	connections := make([]net.Conn, 0, len(knownPeers))
	baseMessage := message.BaseMessage{Origin: r.Address, Clock: 0, Type: message.GET_PEERS, Arguments: nil}

	// Itera sobre os peers conhecidos
	for address := range knownPeers {
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
			if knownPeers[address] == peers.ONLINE {
				logger.Info("Atualizando peer " + address + " status " + peers.OFFLINE.String())
				knownPeers[address] = peers.OFFLINE
			}
		} else {
			// Se a conexão for bem-sucedida e o peer estiver OFFLINE, atualiza o status para ONLINE
			if knownPeers[address] == peers.OFFLINE {
				logger.Info("Atualizando peer " + address + " status " + peers.ONLINE.String())
				knownPeers[address] = peers.ONLINE
			}
		}
	}

	// Retorna as conexões estabelecidas para criar um receiver para cada um
	return connections
}

// Função para mensagem PEER_LIST, envia os meus peers conhecidos para quem solicitou
func (r RequestClient) PeersListRequest(conn net.Conn, receivedMessage message.BaseMessage, knownPeers map[string]peers.PeerStatus) {
	// Cria uma lista de strings para os peers conhecidos
	peers := make([]string, 0, len(knownPeers))

	// Adicioona cada peer que conhece na lista, exceto quem pediu a lista
	for addressPort, peerStatus := range knownPeers {
		if addressPort == receivedMessage.Origin {
			continue
		}
		peers = append(peers, addressPort+":"+peerStatus.String()+":"+"0")
	}

	// Cria uma única string da lista inteira e depois cria e envia a mensagem
	arguments := append([]string{strconv.Itoa(len(peers))}, peers...)
	dropMessage := message.BaseMessage{Origin: r.Address, Clock: 0, Type: message.PEERS_LIST, Arguments: arguments}
	r.sendMessage(conn, dropMessage, receivedMessage.Origin)
}

// Função para mensagem BYE, avisando os peers sobre a saída
func (r RequestClient) ByeRequest(knownPeers map[string]peers.PeerStatus) bool {
	// Imprime mensagem de saída e cria a mensagem BYE
	logger.Info("Saindo...")
	baseMessage := message.BaseMessage{Origin: r.Address, Clock: 0, Type: message.BYE, Arguments: nil}

	// Itera sobre os peers conhecidos
	for addressPort := range knownPeers {
		// Tenta conectar e se conectar define o deadline de 2 segundos
		conn, _ := net.Dial("tcp", addressPort)
		if conn != nil {
			conn.SetDeadline(time.Now().Add(2 * time.Second))
			defer conn.Close()
		}
		r.sendMessage(conn, baseMessage, addressPort)
	}

	return true
}
