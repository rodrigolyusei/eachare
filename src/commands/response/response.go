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

// Função para responder ao get peers recebido
func GetPeersResponse(receivedMessage message.BaseMessage, knownPeers *sync.Map, conn net.Conn, requestClient request.IRequest) {
	requestClient.PeersListRequest(conn, receivedMessage, knownPeers)
}

// Função para responder ao peers list recebido
func PeersListResponse(receivedMessage message.BaseMessage, knownPeers *sync.Map) {
	// Conta quantos peers foram recebidos na mensagem
	peersCount, err := strconv.Atoi(receivedMessage.Arguments[0])
	if err != nil {
		logger.Error("Erro ao converter o número de peers: " + err.Error())
		return
	}

	// Para cada peer na mensagem adiciona nos peers conhecidos
	for i := range peersCount {
		peerInfos := strings.Split(receivedMessage.Arguments[1+i], ":")
		newPeer := peers.Peer{Address: peerInfos[0], Port: peerInfos[1], Status: peers.GetStatus(peerInfos[2])}
		_, exists := knownPeers.Load(newPeer.FullAddress())
		if !exists {
			logger.Info("\tAdicionando novo peer " + newPeer.FullAddress() + " status " + newPeer.Status.String())
			knownPeers.Store(newPeer.FullAddress(), newPeer.Status)
		}
	}
}

// Função para lidar com o BYE recebido
func ByeResponse(receivedMessage message.BaseMessage, knownPeers *sync.Map) {
	knownPeers.Store(receivedMessage.Origin, peers.OFFLINE)
	logger.Info("\tAtualizando peer " + receivedMessage.Origin + " status " + peers.OFFLINE.String())
}
