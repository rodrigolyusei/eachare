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
		peerArgs := strings.Split(receivedMessage.Arguments[i+1], ":")
		peerAddress := peerArgs[0] + ":" + peerArgs[1]
		_, exists := knownPeers.Load(peerAddress)
		if !exists {
			logger.Info("\tAdicionando novo peer " + peerAddress + " status " + peerArgs[2])
			knownPeers.Store(peerAddress, peers.Peer{Status: peers.GetStatus(peerArgs[2]), Clock: 0})
		}
	}
}

// Função para lidar com o BYE recebido
func ByeResponse(receivedMessage message.BaseMessage, knownPeers *sync.Map, neighborClock int) {
	knownPeers.Store(receivedMessage.Origin, peers.Peer{Status: peers.OFFLINE, Clock: neighborClock})
	logger.Info("\tAtualizando peer " + receivedMessage.Origin + " status " + peers.OFFLINE.String())
}
