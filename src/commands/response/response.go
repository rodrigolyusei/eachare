package response

// Pacotes nativos de go e pacotes internos
import (
	"net"
	"strconv"
	"strings"
	"sync"

	"EACHare/src/commands/message"
	"EACHare/src/commands/request"
	"EACHare/src/logger"
	"EACHare/src/peers"
)

// Função para lidar com o GET_PEERS recebido
func GetPeersResponse(conn net.Conn, receivedMessage message.BaseMessage, knownPeers *sync.Map, requestClient request.IRequest) {
	requestClient.PeersListRequest(conn, receivedMessage, knownPeers)
}

// Função para lidar com o PEERS_LIST recebido
func PeersListResponse(receivedMessage message.BaseMessage, knownPeers *sync.Map) {
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
			neighborStatus := neighbor.(peers.Peer).Status
			neighborClock := neighbor.(peers.Peer).Clock

			// Atualiza o status para online e o clock com o que tiver maior valor
			if peerClock > neighborClock {
				knownPeers.Store(peerAddress, peers.Peer{Status: peerStatus, Clock: peerClock})
			} else {
				knownPeers.Store(peerAddress, peers.Peer{Status: peerStatus, Clock: neighborClock})
			}

			// Mensagem de atualização apenas se o status for diferente do conhecido
			if peerStatus != neighborStatus {
				logger.Info("\tAtualizando peer " + peerAddress + " status " + peerArgs[2])
			}
		} else {
			knownPeers.Store(peerAddress, peers.Peer{Status: peerStatus, Clock: peerClock})
			logger.Info("\tAdicionando novo peer " + peerAddress + " status " + peerArgs[2])
		}
	}
}

// Função para lidar com o BYE recebido
func ByeResponse(receivedMessage message.BaseMessage, knownPeers *sync.Map, neighborClock int) {
	knownPeers.Store(receivedMessage.Origin, peers.Peer{Status: peers.OFFLINE, Clock: neighborClock})
	logger.Info("\tAtualizando peer " + receivedMessage.Origin + " status " + peers.OFFLINE.String())
}
