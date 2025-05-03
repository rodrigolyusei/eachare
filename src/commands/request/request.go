package request

// Pacotes nativos de go e pacotes internos
import (
	"net"
	"sync"
	"time"

	"EACHare/src/clock"
	"EACHare/src/commands/message"
	"EACHare/src/commands/response"
	"EACHare/src/connection"
	"EACHare/src/logger"
	"EACHare/src/peers"
)

// Função para mensagem HELLO, avisa o peer que estou online
func HelloRequest(knownPeers *sync.Map, senderAddress string, receiverAddress string) {
	// Cria uma mensagem HELLO
	sendMessage := message.BaseMessage{Origin: senderAddress, Clock: 0, Type: message.HELLO, Arguments: nil}

	// Envia mensagem HELLO para o peer escolhido
	conn, _ := net.Dial("tcp", receiverAddress)
	connection.SendMessage(conn, sendMessage, receiverAddress, knownPeers)
	if conn != nil {
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(2 * time.Second))
	}
}

// Função para mensagem GET_PEERS, solicita para os vizinhos sobre quem eles conhecem
func GetPeersRequest(knownPeers *sync.Map, senderAddress string) {
	// Cria uma lista de conexões e a estrutura da mensagem GET_PEERS
	peerCount := 0
	knownPeers.Range(func(_, _ any) bool {
		peerCount++
		return true
	})
	sendMessage := message.BaseMessage{Origin: senderAddress, Clock: 0, Type: message.GET_PEERS, Arguments: nil}

	// Envia mensagem GET_PEERS para cada peer conhecido e espera a resposta
	knownPeers.Range(func(key, value any) bool {
		address := key.(string)
		conn, _ := net.Dial("tcp", address)
		connection.SendMessage(conn, sendMessage, address, knownPeers)
		if conn != nil {
			defer conn.Close()
			conn.SetDeadline(time.Now().Add(2 * time.Second))

			// Recebe a resposta apenas se a conexão for bem-sucedida
			receivedMessage := connection.ReceiveMessage(conn, knownPeers)
			logger.Info("\tResposta recebida: \"" + receivedMessage.String() + "\"")
			clock.UpdateMaxClock(receivedMessage.Clock)
			logger.Info("\tAtualizando peer " + receivedMessage.Origin + " status " + peers.ONLINE.String())
			response.PeersListResponse(knownPeers, receivedMessage)
		}
		return true
	})
}

// Função para mensagem BYE, avisando os peers sobre a saída
func ByeRequest(knownPeers *sync.Map, senderAddress string) {
	// Imprime mensagem de saída e cria a mensagem BYE
	logger.Info("Saindo...")
	sendMessage := message.BaseMessage{Origin: senderAddress, Clock: 0, Type: message.BYE, Arguments: nil}

	// Envia mensagem BYE para cada peer conhecido
	knownPeers.Range(func(key, _ any) bool {
		address := key.(string)
		conn, _ := net.Dial("tcp", address)
		connection.SendMessage(conn, sendMessage, address, knownPeers)
		if conn != nil {
			defer conn.Close()
			conn.SetDeadline(time.Now().Add(2 * time.Second))
		}
		return true
	})
}
