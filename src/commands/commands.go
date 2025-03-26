package commands

import (
	"EACHare/src/peers"
	"fmt"
	"io/fs"
	"net"
	"os"
	"strconv"
	"strings"

	"EACHare/src/clock"
)

type BaseMessage struct {
	Origin    string
	Clock     int
	Type      CommandType
	Arguments []string
}

var Address string = "localhost"

func sendMessage(connection net.Conn, message BaseMessage, receiverAddress string) error {
	message.Clock = clock.UpdateClock()
	arguments := ""
	if message.Arguments != nil {
		arguments = " " + strings.Join(message.Arguments, " ")
	}
	messageStr := fmt.Sprintf("%s %d %s%s", Address, message.Clock, message.Type.String(), arguments)
	fmt.Printf("\tEncaminhando mensagem \"%s\" para %s\n", messageStr, receiverAddress)
	_, err := connection.Write([]byte(messageStr))
	return err
}

func receiveMessage(message string) BaseMessage {
	messageParts := strings.Split(message, " ")
	receivedClock, err := strconv.Atoi(messageParts[1])
	if err != nil {
		panic(err)
	}

	fmt.Println("\t Resposta recebida: " + message)

	clock.UpdateClock()

	return BaseMessage{
		Origin:    messageParts[0],
		Clock:     receivedClock,
		Type:      GetCommandType(messageParts[2]),
		Arguments: messageParts[3:],
	}
}

func check(e error) {
	if e != nil {
		_ = fmt.Errorf("Error: %s", e)
		panic(e)
	}
}

func GetSharedDirectory(sharedPath string) []fs.DirEntry {
	entries, err := os.ReadDir(sharedPath)
	check(err)

	return entries
}

func GetCommands() string {
	fmt.Println("Escolha um comando:\n\t[1] Listar peers\n\t[2] Obter peers\n\t[3] Listar arquivos locais\n\t[4] Buscar arquivos\n\t[5] Exibir estatisticas\n\t[6] Alterar tamanho de chunk\n\t[9] Sair")
	var x string
	fmt.Scanln(&x)
	return x
}

func GetPeersSend(knowPeers []peers.Peer) {
	baseMessage := BaseMessage{Clock: 0, Type: GET_PEERS, Arguments: nil}
	for i, peer := range knowPeers {
		conn, err := net.Dial("tcp", peer.FullAddress())
		err = sendMessage(conn, baseMessage, peer.FullAddress())
		if err != nil {
			knowPeers[i].Status = false
		}
	}
}

func GetPeersResponse(conn net.Conn, receivedMessage BaseMessage, knowPeers []peers.Peer) {

	peers := []string{}
	for _, peer := range knowPeers {
		peers = append(peers, peer.Address+":"+peer.Port+":"+peer.Status.String()+":"+"0")
	}

	arguments := append([]string{strconv.Itoa(len(knowPeers))}, peers...)

	dropMessage := BaseMessage{Origin: Address, Clock: 0, Type: PEER_LIST, Arguments: arguments}

	sendMessage(conn, dropMessage, receivedMessage.Origin)
}

func PeerListReceive(baseMessage BaseMessage) []peers.Peer {
	peersCount, err := strconv.Atoi(baseMessage.Arguments[0])
	check(err)

	newPeers := make([]peers.Peer, peersCount)

	for i := range peersCount {
		subMessage := strings.Split(baseMessage.Arguments[1+i], ":")
		peer := peers.Peer{Address: subMessage[0], Port: subMessage[1], Status: peers.GetPeerStatus(subMessage[2])}
		newPeers[i] = peer
	}

	return newPeers
}

func UpdatePeersList(knowPeers []peers.Peer, newPeers []peers.Peer) []peers.Peer {
	for _, newPeer := range newPeers {
		peerFound := false
		for i, knowPeer := range knowPeers {
			if newPeer.FullAddress() == knowPeer.FullAddress() {
				knowPeers[i].Status = newPeer.Status
				peerFound = true
				break
			}
		}
		if !peerFound {
			knowPeers = append(knowPeers, newPeer)
		}
	}
	return knowPeers
}
