package commands

// Pacotes nativos de go e pacotes internos
import (
	"context"
	"encoding/base64"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"eachare/src/clock"
	"eachare/src/connection"
	"eachare/src/logger"
	"eachare/src/message"
	"eachare/src/peers"
)

// Estrutura para um arquivo do download
type File struct {
	name   string
	size   int
	origin []string
}

func (f *File) OriginsString() string {
	return strings.Join(f.origin, ", ")
}

func (f *File) AppendOrigin(origin string) {
	f.origin = append(f.origin, origin)
}

// Estrutura para lista de arquivos do download
type FileList struct {
	files []File
}

func (fl *FileList) Empty() bool {
	return len(fl.files) == 0
}

func (fl *FileList) Len() int {
	return len(fl.files)
}

func (fl *FileList) AppendFile(filename string, size int, origin string) {
	for idx, file := range fl.files {
		if file.name == filename && file.size == size {
			fl.files[idx].AppendOrigin(origin)
			return
		}
	}
	fl.files = append(fl.files, File{filename, size, []string{origin}})
}

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
func LsRequest(knownPeers *peers.SafePeers, senderAddress string, sharedPath string, chunkSize int) {
	// Cria a estrutura da mensagem LS
	sendMessage := message.BaseMessage{Origin: senderAddress, Clock: 0, Type: message.LS, Arguments: nil}

	// Envia mensagem LS para cada peer conhecido online
	var noPeers bool = true
	var files *FileList = &FileList{files: []File{}}
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
				nameSize := strings.Split(file, ":")
				size, err := strconv.Atoi(nameSize[1])
				if err != nil {
					continue
				}
				files.AppendFile(nameSize[0], size, receivedMessage.Origin)
			}
		}
	}

	// Chama a função para download apenas se havia arquivos disponíveis na busca
	if noPeers {
		logger.Std("\nNão havia nenhum peer online na busca\n")
	} else if files.Empty() {
		logger.Std("\nNão havia nenhum arquivo disponível na busca\n")
	} else {
		DlMenu(knownPeers, senderAddress, sharedPath, files, chunkSize)
	}
}

// Função para mensagem DL, escolhe um arquivo dentre os buscados para baixar
func DlMenu(knownPeers *peers.SafePeers, senderAddress string, sharedPath string, fileList *FileList, chunkSize int) {
	// Declara variável para o comando e inicia o loop do menu de arquivos
	var comm string
	for {
		// Encontra o nome e o tamanho com maior quantidade de caracteres
		maxName := len("<Cancelar>")
		maxSize := len("Tamanho")
		for _, file := range fileList.files {
			if len(file.name) > maxName {
				maxName = len(file.name)
			}
			if len(strconv.Itoa(file.size)) > maxSize {
				maxSize = len(strconv.Itoa(file.size))
			}
		}

		// Formata o menu de opções
		header := fmt.Sprintf("\t     %%-%ds | %%-%ds | %%s\n", maxName, maxSize)
		row := fmt.Sprintf("\t[%%2d] %%-%ds | %%-%ds | %%s\n", maxName, maxSize)

		// Imprime o menu de opções
		logger.Std("\nArquivos encontrados na rede:\n")
		logger.Std(fmt.Sprintf(header, "Nome", "Tamanho", "Peer"))
		logger.Std(fmt.Sprintf(row, 0, "<Cancelar>", "", ""))

		// Lista os peers e cria uma lista dos endereços para enviar o HELLO
		for i, file := range fileList.files {
			logger.Std(fmt.Sprintf(row, i+1, file.name, strconv.Itoa(file.size), file.OriginsString()))
		}

		// Lê a entrada do usuário
		logger.Std("\nDigite o numero do arquivo para fazer o download:\n> ")
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
		} else if number > 0 && number <= fileList.Len() {
			DlRequest(knownPeers, fileList.files[number-1], senderAddress, sharedPath, chunkSize)
			break
		} else {
			logger.Std("\nOpção inválida, tente novamente.\n")
		}
	}
}

// Estrutura para resposta do download
type DlResponse struct {
	index  int
	hash   string
	origin string
	err    error
}

func (dr DlResponse) String() string {
	errMsg := "nil"
	if dr.err != nil {
		errMsg = dr.err.Error()
	}
	return fmt.Sprintf("index: %d, hash: %s, origin: %s, error: %s", dr.index, dr.hash, dr.origin, errMsg)
}

type OriginManagerConfig struct {
	knownPeers    *peers.SafePeers
	file          *File
	senderAddress string
	chunkSize     int
	resultCh      chan *DlResponse
	retryCh       chan *DlResponse
}

type OriginManager struct {
	origin   string
	requests []func(index int)
	wg       *sync.WaitGroup
	cfg      *OriginManagerConfig
	ctx      *context.Context
}

func (om *OriginManager) cancel() {
	cancel := <-om.cancellationCh
	if cancel {

	}
}

func (om *OriginManager) createRequests(initialIndex, finalIndex int) {
	k := func(index int) {
		defer om.wg.Done()

		arguments := []string{om.cfg.file.name, strconv.Itoa(om.cfg.chunkSize), strconv.Itoa(index)}
		sendMessage := message.BaseMessage{Origin: om.cfg.senderAddress, Clock: 0, Type: message.DL, Arguments: arguments}

		conn, err := net.Dial("tcp", om.origin)
		connection.SendMessage(om.cfg.knownPeers, conn, sendMessage, om.origin)
		if err != nil {
			om.cfg.retryCh <- &DlResponse{index: index, err: err, origin: om.origin}
			return
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(2 * time.Second))

		receivedMessage := connection.ReceiveMessage(om.cfg.knownPeers, conn)
		if receivedMessage.Origin == "" {
			om.cfg.retryCh <- &DlResponse{index: index, err: err, origin: om.origin}
			return
		}
		logger.Info("Resposta recebida: \"" + receivedMessage.String() + "\"")
		clock.UpdateMaxClock(receivedMessage.Clock)
		logger.Info("Atualizando peer " + receivedMessage.Origin + " status " + peers.ONLINE.String())

		receivedIdx, err := strconv.Atoi(receivedMessage.Arguments[2])
		if err != nil {
			om.cfg.retryCh <- &DlResponse{index: index, err: err, origin: om.origin}
			return
		}
		om.cfg.resultCh <- &DlResponse{index: receivedIdx, hash: receivedMessage.Arguments[3], err: nil, origin: om.origin}
	}

	indexNum := initialIndex
mainloop:
	for {
		select {
		case <-om.ctx.Done():
			lastCreatedIndex := indexNum
			break mainloop
		default:
			go k(indexNum)
			indexNum++
		}
	}
}

// Função para mensagem DL, solicita o download do arquivo escolhido em chunks
func DlRequest(knownPeers *peers.SafePeers, file File, senderAddress string, sharedPath string, chunkSize int) {
	logger.Std("\nArquivo escolhido " + file.name + "\n")

	// Calcula a quantidade de requisições necessárias e cria o canal de respostas
	totalRequests := int(math.Ceil(float64(file.size) / float64(chunkSize)))

	requestsPerOrigin := totalRequests / len(file.origin)
	for _, origin := range file.origin {

	}

	// Função para enviar requisições de download para o chunk especificado

	// Envia requisições de download para cada chunk do arquivo
	responses := make(chan *DlResponse, totalRequests)
	var wg sync.WaitGroup
	wg.Add(totalRequests)
	for index := range totalRequests {
		senderIdx := index % len(file.origin)
		go sendRequest(file.origin[senderIdx], index, &wg, responses)
	}
	go func() {
		wg.Wait()
		close(responses)
	}()
	wg.Wait()

	// Processa as respostas recebidas para verificar as falhas
	receivedHashes := make([]string, totalRequests)
	failedIndexes := make([]int, 0)
	verifyPeers := make(map[string]bool)
	for dlResponse := range responses {
		if dlResponse.err != nil {
			failedIndexes = append(failedIndexes, dlResponse.index)
			verifyPeers[dlResponse.origin] = true
		} else {
			receivedHashes[dlResponse.index] = dlResponse.hash
		}
	}
	if len(failedIndexes) == len(file.origin) {
		logger.Std("Todos os peers falharam ao enviar o arquivo. Processo cancelado.\n")
		return
	}
	successfulPeers := make([]string, 0)
	for _, origin := range file.origin {
		if !verifyPeers[origin] {
			successfulPeers = append(successfulPeers, origin)
		}
	}

	// Se houver falhas, tenta reenviar as requisições dos chunks que falharam para outros peers
	retryChan := make(chan *DlResponse, len(failedIndexes))
	var retryWg sync.WaitGroup
	retryWg.Add(len(failedIndexes))
	for _, index := range failedIndexes {
		senderIdx := index % len(successfulPeers)
		go sendRequest(successfulPeers[senderIdx], index, &retryWg, retryChan)
	}
	go func() {
		retryWg.Wait()
		close(retryChan)
	}()
	retryWg.Wait()

	// Processa as respostas recebidas para verificar se houve falhas
	for dlResponse := range retryChan {
		if dlResponse.err != nil {
			logger.Std("Retry do download falhado. Processo cancelado. Falhei em refazer \n" + dlResponse.String())
			return
		}
		receivedHashes[dlResponse.index] = dlResponse.hash
	}

	// Decodifica os chunks recebidos e junta
	var decodedChunks []byte
	for i, r := range receivedHashes {
		if r == "" {
			var wgh sync.WaitGroup
			tempChan := make(chan *DlResponse)
			wgh.Add(1)
			senderIdx := i % len(successfulPeers)
			sendRequest(successfulPeers[senderIdx], i, &wgh, tempChan)
			k := <-tempChan
			receivedHashes[i] = k.hash
			//panic(fmt.Sprintf("Chunk %d está vazio. próximo chunk: %s", i, receivedHashes[i]))
		}
		dec, err := base64.StdEncoding.DecodeString(r)
		if err != nil {
			panic(fmt.Sprintf("Erro ao decodificar chunk %d: %v", i, err))
		}
		decodedChunks = append(decodedChunks, dec...)
	}

	// Cria/substitui o arquivo e escreve o conteúdo decodificado
	createdFile, err := os.Create(sharedPath + file.name)
	check(err)
	defer createdFile.Close()
	_, err = createdFile.Write(decodedChunks)
	check(err)
	logger.Std("\nDownload do arquivo " + file.name + " finalizado.\n")
}

// Função para alterar o tamanho do chunk
func ChangeChunk(chunkSize *int) {
	var chunk string
	logger.Std("\nDigite novo tamanho de chunk:\n> ")
	for {
		fmt.Scanln(&chunk)
		number, err := strconv.Atoi(chunk)
		if err == nil && number > 0 {
			*chunkSize = number
			logger.Info("Tamanho de chunk alterado: " + strconv.Itoa(number))
			return
		}
		logger.Std("\nValor inválido. Precisa ser um inteiro maior que 0.\n> ")
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
