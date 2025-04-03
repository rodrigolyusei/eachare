package commands

// Pacotes nativos de go e pacotes internos
import (
	"fmt"
	"io/fs"
	"net"
	"os"
	"strconv"
	"strings"

	"EACHare/src/clock"
	"EACHare/src/commands/message"
	"EACHare/src/commands/request"
	"EACHare/src/logger"
	"EACHare/src/peers"
)

// Função para verificar e imprimir mensagem de erro
func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Função para receber mensagem
func ReceiveMessage(receivedMessage string) message.BaseMessage {
	// Recupera as partes da mensagem
	messageParts := strings.Split(receivedMessage, " ")

	// Guarda o valor do clock da mensagem recebida
	receivedClock, err := strconv.Atoi(messageParts[1])
	if err != nil {
		panic(err)
	}

	answer := []string{messageParts[0], strconv.Itoa(receivedClock), message.GetMessageType(messageParts[2]).String()}
	answer = append(answer, messageParts[3:]...)
	receive := strings.Join(answer, " ")

	if message.GetMessageType(messageParts[2]) == message.HELLO {
		logger.Info("\tMensagem recebida: \"" + receive + "\"")
	} else {
		logger.Info("\tResposta recebida: \"" + receive + "\"")
	}

	clock.UpdateClock()

	return message.BaseMessage{
		Origin:    messageParts[0],
		Clock:     receivedClock,
		Type:      message.GetMessageType(messageParts[2]),
		Arguments: messageParts[3:],
	}
}

func GetSharedDirectory(sharedPath string) []fs.DirEntry {
	entries, err := os.ReadDir(sharedPath)
	check(err)

	return entries
}

func GetPeersResponse(conn net.Conn, receivedMessage message.BaseMessage,
	knownPeers map[string]peers.PeerStatus, requestClient request.IRequest) {
	requestClient.PeerListRequest(conn, receivedMessage, knownPeers)
}

func PeerListResponse(baseMessage message.BaseMessage) []peers.Peer {
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

func ListPeers(knownPeers map[string]peers.PeerStatus, requestClient request.IRequest) {
	fmt.Println("Lista de peers: ")
	fmt.Println("\t[0] voltar para o menu anterior")

	// Listar todos os peers conhecidos enquanto conta e armazena os endereços
	var addrList []string
	counter := 0
	for addressPort, peerStatus := range knownPeers {
		counter++
		addrList = append(addrList, addressPort)
		fmt.Println("\t[" + strconv.Itoa(counter) + "] " + addressPort + " " + peerStatus.String())
	}

	var input string
	exit := false
	for !exit {
		// Lê a entrada do usuário
		fmt.Print("> ")
		fmt.Scanln(&input)
		fmt.Println()
		number, err := strconv.Atoi(input)
		check(err)

		// Envio de mensagem para o destino escolhido
		if number == 0 {
			exit = true
		} else if number > 0 && number <= counter {
			// Enviar mensagem HELLO
			peerStatus := requestClient.HelloRequest(addrList[number-1])
			if knownPeers[addrList[number-1]] != peerStatus {
				logger.Info("\tAtualizando peer " + addrList[number-1] + " status " + peerStatus.String())
			}
			exit = true
		} else {
			fmt.Println("Opção inválida")
		}
	}
}

func UpdatePeersMap(knownPeers map[string]peers.PeerStatus, newPeers []peers.Peer) {
	for _, newPeer := range newPeers {
		_, exists := knownPeers[newPeer.FullAddress()]
		if !exists {
			logger.Info("\tAdicionando novo peer " + newPeer.FullAddress() + " status " + newPeer.Status.String())
			knownPeers[newPeer.FullAddress()] = newPeer.Status
		}
	}
}
