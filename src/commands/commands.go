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

func sendMessage(connection net.Conn, message BaseMessage, peer peers.Peer) error {
	message.Clock = clock.UpdateClock()
	arguments := ""
	if message.Arguments != nil {
		arguments = " " + strings.Join(message.Arguments, " ")
	}
	messageStr := fmt.Sprintf("%s %d %s%s", Address, message.Clock, message.Type.String(), arguments)
	fmt.Printf("\tEncaminhando mensagem \"%s\" para %s\n", messageStr, peer.FullAddress())
	_, err := connection.Write([]byte(messageStr))
	return err
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
		err = sendMessage(conn, baseMessage, peer)
		if err != nil {
			knowPeers[i].Status = false
		}
	}
}

func GetPeersReceive(connection net.Conn) {

}

func receiveMessage(message string) BaseMessage {
	messageParts := strings.Split(message, " ")
	clock, err := strconv.Atoi(messageParts[1])
	if err != nil {
		panic(err)
	}

	return BaseMessage{
		Origin:    messageParts[0],
		Clock:     clock,
		Type:      GetCommandType(messageParts[2]),
		Arguments: messageParts[3:],
	}
}
