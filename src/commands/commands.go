package commands

// Pacotes nativos de go e pacotes internos
import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

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

// Função para construir a mensagem a partir da string recebida
func ReceiveMessage(receivedMessage string) message.BaseMessage {
	// Recupera as partes da mensagem
	receivedMessage = strings.TrimSuffix(receivedMessage, "\n")
	messageParts := strings.Split(receivedMessage, " ")

	// Imprime a mensagem/resposta recebida e atualiza o clock
	if messageParts[2] == "PEERS_LIST" {
		logger.Info("\tResposta recebida: \"" + receivedMessage + "\"")
	} else {
		logger.Info("\tMensagem recebida: \"" + receivedMessage + "\"")
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
func ListPeers(knownPeers *sync.Map, requestClient request.IRequest) {
	fmt.Println("Lista de peers: ")
	fmt.Println("\t[0] voltar para o menu anterior")

	// Listar todos os peers conhecidos enquanto conta e armazena os endereços
	var addrList []string
	var counter int = 0
	knownPeers.Range(func(key, value interface{}) bool {
		addressPort := key.(string)
		peerStatus := value.(peers.PeerStatus)

		counter++
		addrList = append(addrList, addressPort)
		fmt.Println("\t[" + strconv.Itoa(counter) + "] " + addressPort + " " + peerStatus.String())
		return true
	})

	var comm string
	var exit bool = false
	for !exit {
		// Lê a entrada do usuário
		fmt.Print("> ")
		fmt.Scanln(&comm)
		fmt.Println()

		// Converte a entrada para inteiro
		number, err := strconv.Atoi(comm)
		check(err)

		// Envio de mensagem para o destino escolhido
		if number == 0 {
			exit = true
		} else if number > 0 && number <= counter {
			// Envia mensagem HELLO e atualiza o status do peer
			peerStatus := requestClient.HelloRequest(addrList[number-1])
			if value, _ := knownPeers.Load(addrList[number-1]); value != peerStatus {
				logger.Info("\tAtualizando peer " + addrList[number-1] + " status " + peerStatus.String())
			}
			knownPeers.Store(addrList[number-1], peerStatus)
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
func GetPeersResponse(receivedMessage message.BaseMessage, knownPeers *sync.Map, conn net.Conn, requestClient request.IRequest) {
	requestClient.PeersListRequest(conn, receivedMessage, knownPeers)
}

// Função para responder ao peers list recebido
func PeersListResponse(receivedMessage message.BaseMessage, knownPeers *sync.Map) {
	// Conta quantos peers foram recebidos na mensagem
	peersCount, err := strconv.Atoi(receivedMessage.Arguments[0])
	check(err)

	// Para cada peer na mensagem adiciona nos peers conhecidos
	for i := range peersCount {
		peerInfos := strings.Split(receivedMessage.Arguments[1+i], ":")
		newPeer := peers.Peer{Address: peerInfos[0], Port: peerInfos[1], Status: peers.GetStatus(peerInfos[2])}
		_, exists := knownPeers.Load(newPeer.FullAddress())
		if !exists {
			logger.Info("\tAdicionando novo peer " + newPeer.FullAddress() + " status " + newPeer.Status.String())
			knownPeers.Store(newPeer.FullAddress(), newPeer.Status)
		}
	}
}

// Função para lidar com o BYE recebido
func ByeResponse(receivedMessage message.BaseMessage, knownPeers *sync.Map) {
	knownPeers.Store(receivedMessage.Origin, peers.OFFLINE)
	logger.Info("\tAtualizando peer " + receivedMessage.Origin + " status " + peers.OFFLINE.String())
}
