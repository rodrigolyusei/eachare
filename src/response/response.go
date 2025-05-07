package response

// Pacotes nativos de go e pacotes internos
import (
	"encoding/base64"
	"net"
	"os"
	"strconv"

	"eachare/src/connection"
	"eachare/src/logger"
	"eachare/src/message"
	"eachare/src/peers"
)

// Função para verificar e imprimir mensagem de erro
func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Função para lidar com o GET_PEERS recebido
func GetPeersResponse(knownPeers *peers.SafePeers, receiverAddress string, senderAddress string, conn net.Conn) {
	// Cria uma lista de strings para os peers conhecidos
	myPeers := make([]string, 0)

	// Adiciona cada peer conhecido na lista, exceto quem pediu a lista
	for _, peer := range knownPeers.GetAll() {
		if peer.Address == receiverAddress {
			continue
		}
		myPeers = append(myPeers, peer.Address+":"+peer.Status.String()+":"+strconv.Itoa(peer.Clock))
	}

	// Cria uma única string da lista inteira e envia a mensagem
	arguments := append([]string{strconv.Itoa(len(myPeers))}, myPeers...)
	sendMessage := message.BaseMessage{Origin: senderAddress, Clock: 0, Type: message.PEERS_LIST, Arguments: arguments}
	connection.SendMessage(knownPeers, conn, sendMessage, receiverAddress)
}

// Função para lidar com o LS recebido
func LsResponse(knownPeers *peers.SafePeers, receiverAddress string, senderAddress string, sharedPath string, conn net.Conn) {
	// Cria uma lista de strings para os peers conhecidos
	myFiles := make([]string, 0)

	// Lê o diretório e imprime os arquivos
	entries, err := os.ReadDir(sharedPath)
	check(err)
	for _, entry := range entries {
		stat, _ := entry.Info()
		myFiles = append(myFiles, entry.Name()+":"+strconv.Itoa(int(stat.Size())))
	}

	// Cria uma única string da lista inteira e envia a mensagem
	arguments := append([]string{strconv.Itoa(len(myFiles))}, myFiles...)
	sendMessage := message.BaseMessage{Origin: senderAddress, Clock: 0, Type: message.LS_LIST, Arguments: arguments}
	connection.SendMessage(knownPeers, conn, sendMessage, receiverAddress)
}

// Função para lidar com o LS recebido
func DlResponse(knownPeers *peers.SafePeers, receivedMessage message.BaseMessage, senderAddress string, sharedPath string, conn net.Conn) {
	// Lê o arquivo escolhido e codifica em base64
	chosenFile := receivedMessage.Arguments[0]
	data, err := os.ReadFile(sharedPath + "/" + chosenFile)
	check(err)
	encoded := base64.StdEncoding.EncodeToString(data)

	// Cria o argumento sobre o arquivo e envia a mensagem
	arguments := []string{receivedMessage.Arguments[0], "0", "0", encoded}
	sendMessage := message.BaseMessage{Origin: senderAddress, Clock: 0, Type: message.FILE, Arguments: arguments}
	connection.SendMessage(knownPeers, conn, sendMessage, receivedMessage.Origin)
}

// Função para lidar com o BYE recebido
func ByeResponse(knownPeers *peers.SafePeers, receiverAddress string, neighborClock int) {
	knownPeers.Add(peers.Peer{Address: receiverAddress, Status: peers.OFFLINE, Clock: neighborClock})
	logger.Info("Atualizando peer " + receiverAddress + " status " + peers.OFFLINE.String())
}
