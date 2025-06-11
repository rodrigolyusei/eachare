package commands

// Pacotes nativos de go e pacotes internos
import (
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"eachare/src/clock"
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

// Função para listar os peers conhecidos e enviar HELLO para o peer escolhido
func ListPeers(knownPeers *peers.SafePeers, senderAddress string) {
	// Declara variável para o comando e inicia o loop do menu de peers
	var comm string
	for {
		// Imprime o menu de opções
		logger.Std("Lista de peers:\n")
		logger.Std("\t[0] voltar para o menu anterior\n")

		// Lista os peers e cria uma lista dos endereços para enviar o HELLO
		var addrList []string
		for i, peer := range knownPeers.GetAll() {
			addrList = append(addrList, peer.Address)
			logger.Std("\t[" + strconv.Itoa(i+1) + "] " + peer.Address + " " + peer.Status.String() + " (clock: " + strconv.Itoa(peer.Clock) + ")" + "\n")
		}

		// Lê a entrada do usuário
		logger.Std("> ")
		fmt.Scanln(&comm)

		// Converte a entrada para inteiro
		number, err := strconv.Atoi(comm)
		if err != nil {
			logger.Std("\nOpção inválida, tente novamente.\n\n")
			continue
		}

		// Envio de mensagem para o destino escolhido
		if number == 0 {
			break
		} else if number > 0 && number <= len(addrList) {
			// Imprime uma quebra de linha
			logger.Std("\n")

			// Cria e envia a mensagem HELLO para o peer escolhido
			sendMessage := message.BaseMessage{Origin: senderAddress, Clock: 0, Type: message.HELLO, Arguments: nil}
			conn, _ := net.Dial("tcp", addrList[number-1])
			connection.SendMessage(knownPeers, conn, sendMessage, addrList[number-1])
			if conn != nil {
				logger.Info("Atualizando peer " + addrList[number-1] + " status " + peers.ONLINE.String())
				defer conn.Close()
				conn.SetDeadline(time.Now().Add(2 * time.Second))
			}
			break
		} else {
			logger.Std("\nOpção inválida, tente novamente.\n\n")
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

			// Itera sobre os peers no argumento da mensagem recebida
			for _, peer := range receivedMessage.Arguments[1:] {
				// Salva as partes do peer
				peerParts := strings.Split(peer, ":")
				peerAddress := peerParts[0] + ":" + peerParts[1]
				peerStatus := peers.GetStatus(peerParts[2])
				peerClock, _ := strconv.Atoi(peerParts[3])

				// Verifica as condições para atualizar ou adicionar o peer recebido
				neighbor, exists := knownPeers.Get(peerAddress)
				if exists {
					// Atualiza o status e o clock apenas se for mais recente
					if peerClock >= neighbor.Clock {
						knownPeers.Add(peers.Peer{Address: peerAddress, Status: peerStatus, Clock: peerClock})
						logger.Info("Atualizando peer " + peerAddress + " status " + peerParts[2])
					} else {
						logger.Info("Continuando peer " + peerAddress + " status " + neighbor.Status.String() + " (informação desatualizada recebida)")
					}
				} else {
					knownPeers.Add(peers.Peer{Address: peerAddress, Status: peerStatus, Clock: peerClock})
					logger.Info("Adicionando novo peer " + peerAddress + " status " + peerParts[2])
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

// Função para mensagem LS, solicita para os vizinhos onlines os seus arquivos
func LsRequest(knownPeers *peers.SafePeers, senderAddress string, sharedPath string) {
	// Cria a estrutura da mensagem LS
	sendMessage := message.BaseMessage{Origin: senderAddress, Clock: 0, Type: message.LS, Arguments: nil}

	// Envia mensagem LS para cada peer conhecido online
	var noPeers bool = true
	var files []string
	for _, peer := range knownPeers.GetAll() {
		if !peer.Status {
			continue
		}
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
			noPeers = false

			// Itera sobre os arquivos no argumento da mensagem recebida
			for _, file := range receivedMessage.Arguments[1:] {
				file = file + ":" + receivedMessage.Origin
				files = append(files, file)
			}
		}
	}

	// Chama a função para download apenas se havia arquivos disponíveis na busca
	if noPeers {
		logger.Std("Não havia nenhum peer online na busca\n")
	} else if files == nil {
		logger.Std("\nNão havia nenhum arquivo disponível na busca\n")
	} else {
		DlRequest(knownPeers, senderAddress, sharedPath, files)
	}
}

// Função para mensagem DL, escolhe um arquivo dentre os buscados para baixar
func DlRequest(knownPeers *peers.SafePeers, senderAddress string, sharedPath string, files []string) {
	// Declara variável para o comando e inicia o loop do menu de arquivos
	var comm string
	for {
		// Encontra o nome e o tamanho com maior quantidade de caracteres
		maxName := len("<Cancelar>")
		maxSize := len("Tamanho")
		for _, file := range files {
			fileParts := strings.SplitN(file, ":", 3)
			if len(fileParts[0]) > maxName {
				maxName = len(fileParts[0])
			}
			if len(fileParts[1]) > maxSize {
				maxSize = len(fileParts[1])
			}
		}

		// Formata o cabeçalho e a linha do menu
		header := fmt.Sprintf("\t     %%-%ds | %%-%ds | %%s\n", maxName, maxSize)
		row := fmt.Sprintf("\t[%%2d] %%-%ds | %%-%ds | %%s\n", maxName, maxSize)

		// Imprime o menu de opções
		logger.Std("\nArquivos encontrados na rede:\n")
		logger.Std(fmt.Sprintf(header, "Nome", "Tamanho", "Peer"))
		logger.Std(fmt.Sprintf(row, 0, "<Cancelar>", "", ""))

		// Lista os peers e cria uma lista dos endereços para enviar o HELLO
		for i, file := range files {
			fileParts := strings.SplitN(file, ":", 3)
			logger.Std(fmt.Sprintf(row, i+1, fileParts[0], fileParts[1], fileParts[2]))
		}

		// Lê a entrada do usuário
		logger.Std("\nDigite o numero do arquivo para fazer o download:\n")
		logger.Std("> ")
		fmt.Scanln(&comm)

		// Converte a entrada para inteiro
		number, err := strconv.Atoi(comm)
		if err != nil {
			logger.Std("\nOpção inválida, tente novamente.\n")
			continue
		}

		// Solicitação de download para o arquivo escolhido
		if number == 0 {
			break
		} else if number > 0 && number <= len(files) {
			// Cria uma mensagem DL
			chosenParts := strings.SplitN(files[number-1], ":", 3)
			argument := []string{chosenParts[0], "0", "0"}
			sendMessage := message.BaseMessage{Origin: senderAddress, Clock: 0, Type: message.DL, Arguments: argument}

			logger.Std("\nArquivo escolhido " + chosenParts[0] + "\n")

			// Envia mensagem DL para o peer escolhido
			conn, _ := net.Dial("tcp", chosenParts[2])
			connection.SendMessage(knownPeers, conn, sendMessage, chosenParts[2])
			if conn != nil {
				defer conn.Close()
				conn.SetDeadline(time.Now().Add(2 * time.Second))

				// Recebe a resposta apenas se a conexão for bem-sucedida
				receivedMessage := connection.ReceiveMessage(knownPeers, conn)
				logger.Info("Resposta recebida: \"" + receivedMessage.String() + "\"")
				clock.UpdateMaxClock(receivedMessage.Clock)
				logger.Info("Atualizando peer " + receivedMessage.Origin + " status " + peers.ONLINE.String())

				// Decodifica o conteúdo do arquivo recebido
				decoded, err := base64.StdEncoding.DecodeString(receivedMessage.Arguments[3])
				check(err)

				// Cria/substitui o arquivo e escreve o conteúdo decodificado
				file, err := os.Create(sharedPath + receivedMessage.Arguments[0])
				check(err)
				defer file.Close()
				_, err = file.Write(decoded)
				check(err)
				logger.Std("\nDownload do arquivo " + receivedMessage.Arguments[0] + " finalizado.\n")
			}
			break
		} else {
			logger.Std("\nOpção inválida, tente novamente.\n")
		}
	}
}

// Função para mensagem BYE, avisando os peers sobre a saída
func ByeRequest(knownPeers *peers.SafePeers, senderAddress string) {
	// Imprime mensagem de saída e cria a mensagem BYE
	logger.Std("Saindo...\n")
	sendMessage := message.BaseMessage{Origin: senderAddress, Clock: 0, Type: message.BYE, Arguments: nil}

	// Envia mensagem BYE para cada peer conhecido
	for _, peer := range knownPeers.GetAll() {
		if !peer.Status {
			continue
		}
		conn, _ := net.Dial("tcp", peer.Address)
		connection.SendMessage(knownPeers, conn, sendMessage, peer.Address)
		if conn != nil {
			defer conn.Close()
			conn.SetDeadline(time.Now().Add(2 * time.Second))
		}
	}
}

func ChangeChunk() int {
	var chunk string
	logger.Std("\nDigite novo tamanho de chunk:\n>")
	for {
		fmt.Scanln(&chunk)
		number, err := strconv.Atoi(chunk)

		if err == nil && number > 0 {
			return number
		}
		logger.Std("\nValor inválido. Precisa ser um inteiro maior que 0.\n>")
	}
}
