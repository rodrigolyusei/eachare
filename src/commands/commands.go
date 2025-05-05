package commands

// Pacotes nativos de go e pacotes internos
import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"EACHare/src/clock"
	"EACHare/src/connection"
	"EACHare/src/logger"
	"EACHare/src/message"
	"EACHare/src/peers"
)

// Função para verificar e imprimir mensagem de erro
func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Função para listar os peers conhecidos e enviar HELLO para o peer escolhido
func ListPeers(knownPeers *peers.SafePeers, senderAddress string) {
	// Declara variável para o comando e inicia o loop do menu
	var comm string
	for {
		// Imprime o menu de opções
		logger.Std("Lista de peers: \n")
		logger.Std("\t[0] voltar para o menu anterior\n")

		// Lista os peers e cria uma lista dos endereços para enviar o HELLO
		var addrList []string
		for i, peer := range knownPeers.GetAll() {
			addrList = append(addrList, peer.Address)
			logger.Std("\t[" + strconv.Itoa(i+1) + "] " + peer.Address + " " + peer.Status.String() + "\n")
		}

		// Lê a entrada do usuário
		logger.Std("> ")
		fmt.Scanln(&comm)
		logger.Std("\n")

		// Converte a entrada para inteiro
		number, err := strconv.Atoi(comm)
		if err != nil {
			logger.Std("Opção inválida, tente novamente.\n\n")
			continue
		}

		// Envio de mensagem para o destino escolhido
		if number == 0 {
			break
		} else if number > 0 && number <= len(addrList) {
			// Cria uma mensagem HELLO
			sendMessage := message.BaseMessage{Origin: senderAddress, Clock: 0, Type: message.HELLO, Arguments: nil}

			// Envia mensagem HELLO para o peer escolhido
			conn, _ := net.Dial("tcp", addrList[number-1])
			connection.SendMessage(knownPeers, conn, sendMessage, addrList[number-1])
			if conn != nil {
				defer conn.Close()
				conn.SetDeadline(time.Now().Add(2 * time.Second))
			}
			break
		} else {
			logger.Std("Opção inválida, tente novamente.\n\n")
		}
	}
}

// Função para mensagem GET_PEERS, solicita para os vizinhos sobre quem eles conhecem
func GetPeersRequest(knownPeers *peers.SafePeers, senderAddress string) {
	// Cria a estrutura da mensagem GET_PEERS
	sendMessage := message.BaseMessage{Origin: senderAddress, Clock: 0, Type: message.GET_PEERS, Arguments: nil}

	// Envia mensagem GET_PEERS para cada peer conhecido
	for _, peer := range knownPeers.GetAll() {
		conn, _ := net.Dial("tcp", peer.Address)
		connection.SendMessage(knownPeers, conn, sendMessage, peer.Address)
		if conn != nil {
			defer conn.Close()
			conn.SetDeadline(time.Now().Add(2 * time.Second))

			// Recebe a resposta apenas se a conexão for bem-sucedida
			receivedMessage := connection.ReceiveMessage(knownPeers, conn)
			logger.Info("Resposta recebida: \"" + receivedMessage.String() + "\"")
			clock.UpdateMaxClock(receivedMessage.Clock)
			logger.Info("Atualizando peer " + receivedMessage.Origin + " status " + peers.ONLINE.String())

			// Para cada peer na mensagem adiciona nos peers conhecidos
			peersCount, _ := strconv.Atoi(receivedMessage.Arguments[0])
			for i := range peersCount {
				// Divide a string do peer em partes e salva cada parte
				peerArgs := strings.Split(receivedMessage.Arguments[i+1], ":")
				peerAddress := peerArgs[0] + ":" + peerArgs[1]
				peerStatus := peers.GetStatus(peerArgs[2])
				peerClock, _ := strconv.Atoi(peerArgs[3])

				// Verifica as condições para atualizar ou adicionar o peer recebido
				neighbor, exists := knownPeers.Get(peerAddress)
				if exists {
					// Atualiza o status para online e o clock com o que tiver maior valor
					if peerClock > neighbor.Clock {
						knownPeers.Add(peers.Peer{Address: peerAddress, Status: peerStatus, Clock: peerClock})
					} else {
						knownPeers.Add(peers.Peer{Address: peerAddress, Status: peerStatus, Clock: neighbor.Clock})
					}
					logger.Info("Atualizando peer " + peerAddress + " status " + peerArgs[2])
				} else {
					knownPeers.Add(peers.Peer{Address: peerAddress, Status: peerStatus, Clock: peerClock})
					logger.Info("Adicionando novo peer " + peerAddress + " status " + peerArgs[2])
				}
			}
		}
	}
}

// Função para listar os arquivos do diretório compartilhado
func ListLocalFiles(sharedPath string) {
	// Lê o diretório e imprime os arquivos
	entries, err := os.ReadDir(sharedPath)
	check(err)
	for _, entry := range entries {
		logger.Std("\t" + entry.Name() + "\n")
	}
}

// Função para mensagem BYE, avisando os peers sobre a saída
func ByeRequest(knownPeers *peers.SafePeers, senderAddress string) {
	// Imprime mensagem de saída e cria a mensagem BYE
	logger.Std("Saindo...\n")
	sendMessage := message.BaseMessage{Origin: senderAddress, Clock: 0, Type: message.BYE, Arguments: nil}

	// Envia mensagem BYE para cada peer conhecido
	for _, peer := range knownPeers.GetAll() {
		conn, _ := net.Dial("tcp", peer.Address)
		connection.SendMessage(knownPeers, conn, sendMessage, peer.Address)
		if conn != nil {
			defer conn.Close()
			conn.SetDeadline(time.Now().Add(2 * time.Second))
		}
	}
}
