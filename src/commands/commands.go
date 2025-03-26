package commands

import (
	"EACHare/src/peers"
	"errors"
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
	conn, err := net.Dial("tcp", receiverAddress)

	message.Clock = clock.UpdateClock()
	arguments := ""
	if message.Arguments != nil {
		arguments = " " + strings.Join(message.Arguments, " ")
	}
	messageStr := fmt.Sprintf("%s %d %s%s", Address, message.Clock, message.Type.String(), arguments)
	fmt.Printf("\tEncaminhando mensagem \"%s\" para %s\n", messageStr, receiverAddress)

	if conn == nil {
		return errors.New("Connection is nil")
	}
	_, err = conn.Write([]byte(messageStr))
	return err
}

func ReceiveMessage(message string) BaseMessage {
	messageParts := strings.Split(message, " ")
	receivedClock, err := strconv.Atoi(messageParts[1])
	if err != nil {
		panic(err)
	}

	answer := []string{messageParts[0], strconv.Itoa(receivedClock), GetCommandType(messageParts[2]).String()}
	answer = append(answer, messageParts[3:]...)
	receive := strings.Join(answer, " ")

	fmt.Println("\t Resposta recebida: \"" + receive + "\"")

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
	fmt.Print("> ")
	fmt.Scanln(&x)
	return x
}

func GetPeersRequest(knowPeers map[string]peers.PeerStatus) {
	baseMessage := BaseMessage{Clock: 0, Type: GET_PEERS, Arguments: nil}
	for addressPort, _ := range knowPeers {
		fmt.Println("Enviando mensagem para ", addressPort)
		conn, err := net.Dial("tcp", addressPort)
		err = sendMessage(conn, baseMessage, addressPort)
		if err != nil {
			knowPeers[addressPort] = peers.OFFLINE
		} else {
			knowPeers[addressPort] = peers.ONLINE
		}
	}
}

func GetPeersResponse(conn net.Conn, receivedMessage BaseMessage, knowPeers map[string]peers.PeerStatus) {
	//fmt.Print("Preparando get Peersresponse...")
	peers := []string{}
	for addressPort, peerStatus := range knowPeers {
		peers = append(peers, addressPort+":"+peerStatus.String()+":"+"0")
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

func ListPeers(knowPeers map[string]peers.PeerStatus) {
	fmt.Println("Lista de peers: ")
	fmt.Println("\t[0] voltar para o menu anterior")
	counter := 0
	for addressPort, peerStatus := range knowPeers {
		counter++
		fmt.Println("\t[" + strconv.Itoa(counter) + "] " + addressPort + " " + peerStatus.String())
	}

	var input string

	for {
		fmt.Print("> ")
		fmt.Scanln(&input)
		number, err := strconv.Atoi(input)
		if err != nil || number != 0 {
			fmt.Print("Inválido!\n")
			continue
		} else {
			break
		}

	}

}

func UpdatePeersMap(knowPeers map[string]peers.PeerStatus, newPeers []peers.Peer) {
	for _, newPeer := range newPeers {
		_, exists := knowPeers[newPeer.FullAddress()]
		if !exists {
			knowPeers[newPeer.FullAddress()] = newPeer.Status
		}
	}
}
