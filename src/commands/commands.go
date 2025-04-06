package commands

// Pacotes nativos de go e pacotes internos
import (
	"fmt"
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

	// Imprime a mensagem/resposta recebida e atualiza o clock
	if messageParts[2] == "PEERS_LIST" {
		logger.Info("Resposta recebida: \"" + receivedMessage + "\"")
	} else {
		logger.Info("Mensagem recebida: \"" + receivedMessage + "\"")
	}
	clock.UpdateClock()

	// Guarda o valor do clock da mensagem recebida
	receivedClock, err := strconv.Atoi(messageParts[1])
	check(err)

	// Monta a mensagem e retorna ela
	return message.BaseMessage{
		Origin:    messageParts[0],
		Clock:     receivedClock,
		Type:      message.GetMessageType(messageParts[2]),
		Arguments: messageParts[3:],
	}
}

// Função para listar os peers conhecidos e enviar HELLO para o peer escolhido
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
				logger.Info("Atualizando peer " + addrList[number-1] + " status " + peerStatus.String())
			}
			exit = true
		} else {
			fmt.Println("Opção inválida")
		}
	}
}

// Função para listar os arquivos do diretório compartilhado
func ListLocalFiles(sharedPath string) {
	// Lê o diretório e imprime os arquivos
	entries, err := os.ReadDir(sharedPath)
	check(err)
	for _, entry := range entries {
		fmt.Println("\t" + entry.Name())
	}
}

// Função para responder ao get peers recebido
func GetPeersResponse(receivedMessage message.BaseMessage, knownPeers map[string]peers.PeerStatus, conn net.Conn, requestClient request.IRequest) {
	requestClient.PeersListRequest(conn, receivedMessage, knownPeers)
}

// Função para responder ao peers list recebido
func PeersListResponse(receivedMessage message.BaseMessage, knownPeers map[string]peers.PeerStatus) {
	// Conta quantos peers foram recebidos na mensagem
	peersCount, err := strconv.Atoi(receivedMessage.Arguments[0])
	check(err)

	// Para cada peer na mensagem adiciona nos peers conhecidos
	for i := range peersCount {
		peerInfos := strings.Split(receivedMessage.Arguments[1+i], ":")
		newPeer := peers.Peer{Address: peerInfos[0], Port: peerInfos[1], Status: peers.GetStatus(peerInfos[2])}
		_, exists := knownPeers[newPeer.FullAddress()]
		if !exists {
			logger.Info("Adicionando novo peer " + newPeer.FullAddress() + " status " + newPeer.Status.String())
			knownPeers[newPeer.FullAddress()] = newPeer.Status
		}
	}
}

func ByeResponse(receivedMessage message.BaseMessage, knownPeers map[string]peers.PeerStatus) {
	knownPeers[receivedMessage.Origin] = peers.OFFLINE
	logger.Info("Atualizando peer " + receivedMessage.Origin + " status " + peers.OFFLINE.String())
}
