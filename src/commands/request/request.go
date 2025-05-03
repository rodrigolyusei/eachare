package request

// Pacotes nativos de go e pacotes internos
import (
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"EACHare/src/clock"
	"EACHare/src/commands/message"
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

	// Envia mensagem GET_PEERS para cada peer conhecido
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

			// Para cada peer na mensagem adiciona nos peers conhecidos
			peersCount, _ := strconv.Atoi(receivedMessage.Arguments[0])
			for i := range peersCount {
				// Divide a string do peer em partes e salva cada parte
				peerArgs := strings.Split(receivedMessage.Arguments[i+1], ":")
				peerAddress := peerArgs[0] + ":" + peerArgs[1]
				peerStatus := peers.GetStatus(peerArgs[2])
				peerClock, _ := strconv.Atoi(peerArgs[3])

				// Verifica as condições para atualizar ou adicionar o peer recebido
				neighbor, exists := knownPeers.Load(peerAddress)
				if exists {
					// Atualiza o status para online e o clock com o que tiver maior valor
					neighborClock := neighbor.(peers.Peer).Clock
					if peerClock > neighborClock {
						knownPeers.Store(peerAddress, peers.Peer{Status: peerStatus, Clock: peerClock})
					} else {
						knownPeers.Store(peerAddress, peers.Peer{Status: peerStatus, Clock: neighborClock})
					}
					logger.Info("\tAtualizando peer " + peerAddress + " status " + peerArgs[2])
				} else {
					knownPeers.Store(peerAddress, peers.Peer{Status: peerStatus, Clock: peerClock})
					logger.Info("\tAdicionando novo peer " + peerAddress + " status " + peerArgs[2])
				}
			}
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
