package commands

// Pacotes nativos de go e pacotes internos
import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"math/rand"
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

type HealthyOrigins struct {
	mu      sync.Mutex
	origins []string
}

func NewHealthyOrigins(initialOrigins []string) *HealthyOrigins {
	return &HealthyOrigins{
		origins: initialOrigins,
	}
}

func (h *HealthyOrigins) Remove(originToRemove string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i, origin := range h.origins {
		if origin == originToRemove {
			h.origins = append(h.origins[:i], h.origins[i+1:]...)
			return
		}
	}
}

func (h *HealthyOrigins) GetNext() (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.origins) == 0 {
		return "", errors.New("no healthy origins available")
	}

	randomIndex := rand.Intn(len(h.origins))
	return h.origins[randomIndex], nil
}

func (h *HealthyOrigins) GetAll() ([]string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.origins) == 0 {
		return nil, errors.New("no healthy origins available")
	}
	listCopy := make([]string, len(h.origins))
	copy(listCopy, h.origins)
	return listCopy, nil
}

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
	knownPeers     *peers.SafePeers
	file           *File
	senderAddress  string
	chunkSize      int
	resultCh       chan *DlResponse
	retryCh        chan *DlResponse
	rebalanceCh    chan *ManagerError
	mainWg         *sync.WaitGroup
	healthyOrigins *HealthyOrigins
}

type OriginManager struct {
	origin string
	wg     *sync.WaitGroup
	cfg    *OriginManagerConfig
	ctx    context.Context
	cancel context.CancelFunc
}

type ManagerError struct {
	lastCreatedIndex int
	finalIndex       int
	origin           string
}

func requestChunk(ctx context.Context, cancel context.CancelFunc, cfg *OriginManagerConfig, wg *sync.WaitGroup, index int, origin string) {
	defer wg.Done()

	arguments := []string{cfg.file.name, strconv.Itoa(cfg.chunkSize), strconv.Itoa(index)}
	sendMessage := message.BaseMessage{Origin: cfg.senderAddress, Clock: 0, Type: message.DL, Arguments: arguments}

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", origin)
	connection.SendMessage(cfg.knownPeers, conn, sendMessage, origin)
	if err != nil {
		defer cancel()
		cfg.retryCh <- &DlResponse{index: index, err: err, origin: origin}
		return
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// Receive the response
	receivedMessage := connection.ReceiveMessage(cfg.knownPeers, conn)
	if receivedMessage.Origin == "" {
		err := fmt.Errorf("empty response from origin %s for chunk %d", origin, index)
		cfg.retryCh <- &DlResponse{index: index, err: err, origin: origin}
		defer cancel()
		return
	}

	logger.Info("Resposta recebida: \"" + receivedMessage.String() + "\"")
	clock.UpdateMaxClock(receivedMessage.Clock)
	logger.Info("Atualizando peer " + receivedMessage.Origin + " status " + peers.ONLINE.String())

	receivedIdx, err := strconv.Atoi(receivedMessage.Arguments[2])
	if err != nil {
		defer cancel()
		cfg.retryCh <- &DlResponse{index: index, err: err, origin: origin}
		return
	}

	cfg.resultCh <- &DlResponse{index: receivedIdx, hash: receivedMessage.Arguments[3], err: nil, origin: receivedMessage.Origin}
}

func (om *OriginManager) createRequests(initialIndex, finalIndex int) {
	defer om.cfg.mainWg.Done()

	var lastCreatedIndex int
mainloop:
	for indexNum := initialIndex; indexNum < finalIndex; indexNum++ {
		select {
		case <-om.ctx.Done():
			lastCreatedIndex = indexNum
			om.cfg.rebalanceCh <- &ManagerError{lastCreatedIndex: lastCreatedIndex, finalIndex: finalIndex, origin: om.origin}
			om.cfg.healthyOrigins.Remove(om.origin)
			break mainloop
		default:
			om.wg.Add(1)
			go requestChunk(om.ctx, om.cancel, om.cfg, om.wg, indexNum, om.origin)
		}
	}
	om.wg.Wait()
}

const MAX_RETRIES_PER_CHUNK = 3

func retryManager(cfg *OriginManagerConfig, retryWg *sync.WaitGroup) {
	defer retryWg.Done()

	retryCounts := make(map[int]int)

	for failedReq := range cfg.retryCh {
		chunkIndex := failedReq.index
		failedOrigin := failedReq.origin

		cfg.healthyOrigins.Remove(failedOrigin)

		retryCounts[chunkIndex]++
		if retryCounts[chunkIndex] > MAX_RETRIES_PER_CHUNK {
			panic(fmt.Sprintf("Chunk %d failed more than %d times. Aborting download.", chunkIndex, MAX_RETRIES_PER_CHUNK))
		}

		newOrigin, err := cfg.healthyOrigins.GetNext()
		if err != nil {
			panic(fmt.Sprintf("No healthy peers available to retry chunk %d. Aborting download.", chunkIndex))
		}

		retryWg.Add(1)
		ctx, cancel := context.WithCancel(context.Background())
		go requestChunk(ctx, cancel, cfg, retryWg, chunkIndex, newOrigin)
	}
}

func rebalanceManager(cfg *OriginManagerConfig, rebalanceWg *sync.WaitGroup) {
	defer rebalanceWg.Done()

	for job := range cfg.rebalanceCh {
		chunksToRebalance := make([]int, 0, job.finalIndex-job.lastCreatedIndex)
		for i := job.lastCreatedIndex; i < job.finalIndex; i++ {
			chunksToRebalance = append(chunksToRebalance, i)
		}

		if len(chunksToRebalance) == 0 {
			continue
		}

		healthyPeers, err := cfg.healthyOrigins.GetAll()
		if err != nil {
			panic(fmt.Sprintf("Cannot rebalance chunks: %v. Aborting download.", err))
		}

		peerCount := len(healthyPeers)
		for i, chunkIndex := range chunksToRebalance {
			peerIndex := i % peerCount
			newOrigin := healthyPeers[peerIndex]

			rebalanceWg.Add(1)
			context, cancel := context.WithCancel(context.Background())
			go requestChunk(
				context,
				cancel,
				cfg,
				rebalanceWg,
				chunkIndex,
				newOrigin,
			)
		}
	}
}

// Função para mensagem DL, solicita o download do arquivo escolhido em chunks
func DlRequest(knownPeers *peers.SafePeers, file File, senderAddress string, sharedPath string, chunkSize int) {
	logger.Std("\nArquivo escolhido " + file.name + "\n")

	// Calcula a quantidade de requisições necessárias e cria o canal de respostas
	totalRequests := int(math.Ceil(float64(file.size) / float64(chunkSize)))

	resultCh := make(chan *DlResponse, totalRequests)
	retryCh := make(chan *DlResponse, totalRequests)
	rebalanceCh := make(chan *ManagerError, len(file.origin))

	var retryWg sync.WaitGroup
	var rebalanceWg sync.WaitGroup

	cfg := OriginManagerConfig{
		knownPeers:     knownPeers,
		file:           &file,
		senderAddress:  senderAddress,
		chunkSize:      chunkSize,
		resultCh:       resultCh,
		retryCh:        retryCh,
		rebalanceCh:    rebalanceCh,
		mainWg:         &sync.WaitGroup{},
		healthyOrigins: NewHealthyOrigins(file.origin),
	}

	cfg.mainWg.Add(len(file.origin))

	managers := make([]OriginManager, 0, len(file.origin))

	for _, origin := range file.origin {
		ctx, cancel := context.WithCancel(context.Background())
		manager := OriginManager{
			origin: origin,
			wg:     &sync.WaitGroup{},
			cfg:    &cfg,
			ctx:    ctx,
			cancel: cancel,
		}

		managers = append(managers, manager)
	}

	nOrigins := len(managers)

	base := totalRequests / nOrigins
	remainder := totalRequests % nOrigins

	start := 0
	for i := range managers {
		count := base
		if i < remainder {
			count++
		}
		end := start + count
		if start < end {
			go managers[i].createRequests(start, end) // [start, end)
		}
		start = end
	}
	retryWg.Add(1)
	go retryManager(&cfg, &retryWg)

	rebalanceWg.Add(1)
	go rebalanceManager(&cfg, &rebalanceWg)

	go func() {
		cfg.mainWg.Wait()
		close(rebalanceCh)
		rebalanceWg.Wait()
		close(cfg.retryCh)
		retryWg.Wait()
		close(resultCh)
	}()

	//cfg.mainWg.Wait()

	receivedHashes := make([]string, totalRequests)

	for dlResponse := range resultCh {
		receivedHashes[dlResponse.index] = dlResponse.hash
	}

	// Decodifica os chunks recebidos e junta
	var decodedChunks []byte
	for i, r := range receivedHashes {
		if r == "" {
			panic(fmt.Sprintf("Chunk %d está vazio. próximo chunk: %s", i, receivedHashes[i]))
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
