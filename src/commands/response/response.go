package response

// Pacotes nativos de go e pacotes internos
import (
	"net"
	"strconv"
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

// Função para lidar com o BYE recebido
func ByeResponse(knownPeers *sync.Map, receivedMessage message.BaseMessage, neighborClock int) {
	knownPeers.Store(receivedMessage.Origin, peers.Peer{Status: peers.OFFLINE, Clock: neighborClock})
	logger.Info("\tAtualizando peer " + receivedMessage.Origin + " status " + peers.OFFLINE.String())
}
