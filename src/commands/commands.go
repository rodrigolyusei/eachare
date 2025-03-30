package commands

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"EACHare/src/clock"
	"EACHare/src/peers"
)

type BaseMessage struct {
	Origin    string
	Clock     int
	Type      CommandType
	Arguments []string
}

var Address string = "localhost"

func check(e error) {
	if e != nil {
		_ = fmt.Errorf("error: %s", e)
		panic(e)
	}
}

func sendMessage(connection net.Conn, message BaseMessage, receiverAddress string) error {
	//conn, _ := net.Dial("tcp", receiverAddress)
	message.Clock = clock.UpdateClock()
	arguments := ""
	if message.Arguments != nil {
		arguments = " " + strings.Join(message.Arguments, " ")
	}
	messageStr := fmt.Sprintf("%s %d %s%s", Address, message.Clock, message.Type.String(), arguments)
	fmt.Printf("\tEncaminhando mensagem \"%s\" para %s\n", messageStr, receiverAddress)

	if connection == nil {
		return errors.New("connection is nil")
	}
	//defer conn.Close()
	_, err := connection.Write([]byte(messageStr))
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

	if strings.Trim(messageParts[2], "\x00") == "HELLO" {
		fmt.Println("\tMensagem recebida: \"" + receive + "\"")
	} else {
		fmt.Println("\tResposta recebida: \"" + receive + "\"")
	}

	clock.UpdateClock()

	return BaseMessage{
		Origin:    messageParts[0],
		Clock:     receivedClock,
		Type:      GetCommandType(messageParts[2]),
		Arguments: messageParts[3:],
	}
}

func GetSharedDirectory(sharedPath string) []fs.DirEntry {
	entries, err := os.ReadDir(sharedPath)
	check(err)

	return entries
}

func GetPeersRequest(knowPeers map[string]peers.PeerStatus) []net.Conn {
	connections := make([]net.Conn, 0)
	baseMessage := BaseMessage{Clock: 0, Type: GET_PEERS, Arguments: nil}
	for addressPort := range knowPeers {
		//fmt.Println("Enviando mensagem para ", addressPort)
		conn, _ := net.Dial("tcp", addressPort)
		if conn != nil {
			connections = append(connections, conn)
			conn.SetDeadline(time.Now().Add(2 * time.Second))
			//go connectionT(conn)
		}
		err := sendMessage(conn, baseMessage, addressPort)
		if err != nil {
			if knowPeers[addressPort] == peers.ONLINE {
				fmt.Println("\tAtualizando peer " + addressPort + " status OFFLINE")
				knowPeers[addressPort] = peers.OFFLINE
			}
		} else {
			if knowPeers[addressPort] == peers.OFFLINE {
				fmt.Println("\tAtualizando peer " + addressPort + " status ONLINE")
				knowPeers[addressPort] = peers.ONLINE
			}
		}
	}
	return connections
}

// func connectionT(con net.Conn) {
// 	for {
// 		buf := make([]byte, 1024)
// 		_, err := con.Read(buf)
// 		if err != nil {
// 			break
// 			//fmt.Println("Erro ao ler conexão:", err)
// 		}
// 		fmt.Println("\tMensagem recebida pela thread errada: \"" + string(buf) + "\"")
// 		//message := ReceiveMessage(string(buf))
// 	}
// }

func GetPeersResponse(conn net.Conn, receivedMessage BaseMessage, knowPeers map[string]peers.PeerStatus) {
	//fmt.Print("Preparando get Peersresponse...")
	peers := []string{}

	size := 0
	for addressPort, peerStatus := range knowPeers {
		if addressPort == receivedMessage.Origin {
			continue
		}
		size++
		peers = append(peers, addressPort+":"+peerStatus.String()+":"+"0")
	}

	arguments := append([]string{strconv.Itoa(size)}, peers...)

	dropMessage := BaseMessage{Origin: Address, Clock: 0, Type: PEER_LIST, Arguments: arguments}

	sendMessage(conn, dropMessage, receivedMessage.Origin)
}

func PeerListResponse(baseMessage BaseMessage) []peers.Peer {
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

func ListLocalFiles(sharedPath string) {
	entries, err := os.ReadDir(sharedPath)
	check(err)
	for _, entry := range entries {
		fmt.Println("\t" + entry.Name())
	}
}

func ListPeers(knowPeers map[string]peers.PeerStatus) {
	fmt.Println("Lista de peers: ")
	fmt.Println("\t[0] voltar para o menu anterior")

	// Listar todos os peers conhecidos enquanto conta e armazena os endereços
	var addrList []string
	counter := 0
	for addressPort, peerStatus := range knowPeers {
		counter++
		addrList = append(addrList, addressPort)
		fmt.Println("\t[" + strconv.Itoa(counter) + "] " + addressPort + " " + peerStatus.String())
	}

	var input string
	for {
		// Lê a entrada do usuário
		fmt.Print("> ")
		fmt.Scanln(&input)
		number, err := strconv.Atoi(input)
		check(err)

		// Envio de mensagem para o destino escolhido
		if number == 0 {
			return
		} else if number > 0 && number <= counter {
			// Enviar mensagem HELLO
			baseMessage := BaseMessage{Clock: 0, Type: HELLO, Arguments: nil}
			conn, _ := net.Dial("tcp", addrList[number-1])
			if conn != nil {
				defer conn.Close()
			}
			err := sendMessage(conn, baseMessage, addrList[number-1])
			if err != nil {
				knowPeers[addrList[number-1]] = peers.OFFLINE
				fmt.Println("\tAtualizando peer " + addrList[number-1] + " status OFFLINE")
			} else {
				knowPeers[addrList[number-1]] = peers.ONLINE
				fmt.Println("\tAtualizando peer " + addrList[number-1] + " status ONLINE")
			}
			return
		} else {
			fmt.Println("Opção inválida")
		}
	}
}

func UpdatePeersMap(knowPeers map[string]peers.PeerStatus, newPeers []peers.Peer) {
	for _, newPeer := range newPeers {
		_, exists := knowPeers[newPeer.FullAddress()]
		if !exists {
			fmt.Println("\tAdicionando novo peer", newPeer.FullAddress(), "status", newPeer.Status)
			knowPeers[newPeer.FullAddress()] = newPeer.Status
		}
	}
}
