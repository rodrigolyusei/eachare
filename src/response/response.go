package response

// Pacotes nativos de go e pacotes internos
import (
	"net"
	"strconv"

	"EACHare/src/connection"
	"EACHare/src/logger"
	"EACHare/src/message"
	"EACHare/src/peers"
)

// Função para lidar com o GET_PEERS recebido
func GetPeersResponse(knownPeers *peers.SafePeers, receivedMessage message.BaseMessage, conn net.Conn, senderAddress string) {
	// Cria uma lista de strings para os peers conhecidos
	myPeers := make([]string, 0)

	// Adiciona cada peer conhecido na lista, exceto quem pediu a lista
	for _, peer := range knownPeers.GetAll() {
		if peer.Address == receivedMessage.Origin {
			continue
		}
		myPeers = append(myPeers, peer.Address+":"+peer.Status.String()+":"+strconv.Itoa(peer.Clock))
	}

	// Cria uma única string da lista inteira e depois cria e envia a mensagem
	arguments := append([]string{strconv.Itoa(len(myPeers))}, myPeers...)
	sendMessage := message.BaseMessage{Origin: senderAddress, Clock: 0, Type: message.PEERS_LIST, Arguments: arguments}
	connection.SendMessage(knownPeers, conn, sendMessage, receivedMessage.Origin)
}

// Função para lidar com o BYE recebido
func ByeResponse(knownPeers *peers.SafePeers, receivedMessage message.BaseMessage, neighborClock int) {
	knownPeers.Add(peers.Peer{Address: receivedMessage.Origin, Status: peers.OFFLINE, Clock: neighborClock})
	logger.Info("\tAtualizando peer " + receivedMessage.Origin + " status " + peers.OFFLINE.String())
}
