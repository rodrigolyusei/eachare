package commands

import (
	"fmt"
	"io/fs"
	"net"
	"os"
	"strings"

	"EACHare/src/peers"
)

type BaseMessage struct {
	Clock     int
	Type      string
	Arguments []string
}

var Address string = "localhost"

func sendMessage(connection net.Conn, message BaseMessage) error {
	arguments := ""
	if message.Arguments != nil {
		arguments = " " + strings.Join(message.Arguments, " ")
	}
	messageStr := fmt.Sprintf("%s %d %s%s", Address, message.Clock, message.Type, arguments)
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

func GetPeers(knowPeers []peers.Peer) {
	//baseMessage := BaseMessage{Clock: 0, Type: "GET_PEERS", Arguments: nil}

}
