package request

// Pacotes nativos de go e pacotes internos
import (
	"net"
	"sync"
	"time"

	"EACHare/src/commands/message"
	"EACHare/src/commands/response"
	"EACHare/src/connection"
	"EACHare/src/logger"
)

// Função para mensagem HELLO, avisa o peer que estou online
func HelloRequest(knownPeers *sync.Map, senderAddress string, receiverAddress string) {
	// Cria uma mensagem HELLO
	baseMessage := message.BaseMessage{Origin: senderAddress, Clock: 0, Type: message.HELLO, Arguments: nil}

	// Tenta conectar e se conectar, define o deadline de 2 segundos
	conn, _ := net.Dial("tcp", receiverAddress)
	if conn != nil {
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(2 * time.Second))
	}

	// Envia a mensagem HELLO
	connection.SendMessage(conn, baseMessage, receiverAddress, knownPeers)
}

// Função para mensagem GET_PEERS, solicita para os vizinhos sobre quem eles conhecem
func GetPeersRequest(knownPeers *sync.Map, senderAddress string) []net.Conn {
	// Cria uma lista de conexões e a estrutura da mensagem GET_PEERS
	peerCount := 0
	knownPeers.Range(func(_, _ any) bool {
		peerCount++
		return true
	})
	connections := make([]net.Conn, 0, peerCount)
	baseMessage := message.BaseMessage{Origin: senderAddress, Clock: 0, Type: message.GET_PEERS, Arguments: nil}

	// Envia mensagem GET_PEERS para cada peer conhecido e adiciona a conexão na lista
	knownPeers.Range(func(key, value any) bool {
		address := key.(string)
		conn, _ := net.Dial("tcp", address)
		if conn != nil {
			connections = append(connections, conn)
			conn.SetDeadline(time.Now().Add(2 * time.Second))
		}
		connection.SendMessage(conn, baseMessage, address, knownPeers)

		receivedMessage := connection.ReceiveMessage(conn, knownPeers, false)
		logger.Info("\tResposta recebida: \"" + receivedMessage.String() + "\"")
		response.PeersListResponse(knownPeers, receivedMessage)
		return true
	})

	// Retorna as conexões estabelecidas para reutilizar
	return connections
}

// Função para mensagem BYE, avisando os peers sobre a saída
func ByeRequest(knownPeers *sync.Map, senderAddress string) {
	// Imprime mensagem de saída e cria a mensagem BYE
	logger.Info("Saindo...")
	baseMessage := message.BaseMessage{Origin: senderAddress, Clock: 0, Type: message.BYE, Arguments: nil}

	// Envia mensagem BYE para cada peer conhecido
	knownPeers.Range(func(key, _ any) bool {
		address := key.(string)
		conn, _ := net.Dial("tcp", address)
		if conn != nil {
			defer conn.Close()
			conn.SetDeadline(time.Now().Add(2 * time.Second))
		}
		connection.SendMessage(conn, baseMessage, address, knownPeers)
		return true
	})
}
