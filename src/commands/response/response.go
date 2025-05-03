package response

// Pacotes nativos de go e pacotes internos
import (
	"net"
	"strconv"
	"strings"
	"sync"

	"EACHare/src/commands/message"
	"EACHare/src/connection"
	"EACHare/src/logger"
	"EACHare/src/peers"
)

// Função para lidar com o GET_PEERS recebido
func GetPeersResponse(knownPeers *sync.Map, receivedMessage message.BaseMessage, conn net.Conn, senderAddress string) {
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
	sendMessage := message.BaseMessage{Origin: senderAddress, Clock: 0, Type: message.PEERS_LIST, Arguments: arguments}
	connection.SendMessage(conn, sendMessage, receivedMessage.Origin, knownPeers)
}

// Função para lidar com o PEERS_LIST recebido
func PeersListResponse(knownPeers *sync.Map, receivedMessage message.BaseMessage) {
	// Conta quantos peers foram recebidos na mensagem
	peersCount, err := strconv.Atoi(receivedMessage.Arguments[0])
	if err != nil {
		logger.Error("Erro ao converter o número de peers: " + err.Error())
		return
	}

	// Para cada peer na mensagem adiciona nos peers conhecidos
	for i := range peersCount {
		// Divide a string do peer em partes e salva cada parte
		peerArgs := strings.Split(receivedMessage.Arguments[i+1], ":")
		peerAddress := peerArgs[0] + ":" + peerArgs[1]
		peerStatus := peers.GetStatus(peerArgs[2])
		peerClock, _ := strconv.Atoi(peerArgs[3])

		// Verifica as condições para atualizar ou adicionar o peer recebido
		neighbor, exists := knownPeers.Load(peerAddress)
		if exists {
			neighborClock := neighbor.(peers.Peer).Clock

			// Atualiza o status para online e o clock com o que tiver maior valor
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

// Função para lidar com o BYE recebido
func ByeResponse(knownPeers *sync.Map, receivedMessage message.BaseMessage, neighborClock int) {
	knownPeers.Store(receivedMessage.Origin, peers.Peer{Status: peers.OFFLINE, Clock: neighborClock})
	logger.Info("\tAtualizando peer " + receivedMessage.Origin + " status " + peers.OFFLINE.String())
}
