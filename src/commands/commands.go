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
	"slices"
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

const MAX_CONCURRENT_PER_MANAGER = 50
const MAX_FAILURES_PER_ORIGIN = 15
const MAX_RETRIES_PER_CHUNK = 15

type HealthyOrigins struct {
	mu         sync.Mutex
	origins    []string
	failCounts map[string]int
}

func NewHealthyOrigins(initialOrigins []string) *HealthyOrigins {
	return &HealthyOrigins{
		origins:    initialOrigins,
		failCounts: make(map[string]int),
	}
}

func (h *HealthyOrigins) Remove(originToRemove string) {
	if originToRemove == "" {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.failCounts[originToRemove]++
	if h.failCounts[originToRemove] >= MAX_FAILURES_PER_ORIGIN {
		// remove origin from h.origins
		for i, origin := range h.origins {
			if origin == originToRemove {
				h.origins = slices.Delete(h.origins, i, i+1)
				return
			}
		}
	}
}

func (h *HealthyOrigins) ErrorSummary() string {
	h.mu.Lock()
	defer h.mu.Unlock()

	var sb strings.Builder
	for origin, count := range h.failCounts {
		sb.WriteString(fmt.Sprintf("Origin %s: %d errors\n", origin, count))
	}
	return sb.String()
}

func (h *HealthyOrigins) UnsafeErrorSummary() string {
	var sb strings.Builder
	for origin, count := range h.failCounts {
		sb.WriteString(fmt.Sprintf("Origin %s: %d errors\n", origin, count))
	}
	return sb.String()
}

func (h *HealthyOrigins) GetNext() (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.origins) == 0 {
		return "", errors.New("não era pra dar errado: " + h.UnsafeErrorSummary())
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

// Estrutura para estatísticas do download
type Statistic struct {
	chunckSize int
	peersQty   int
	fileSize   int
	times      []float64
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
		var addrList []string
		for i, peer := range knownPeers.GetAll() {
			addrList = append(addrList, peer.Address)
			logger.Std("\t[" + strconv.Itoa(i+1) + "] " + peer.Address + " " + peer.Status.String() + " (clock: " + strconv.Itoa(peer.Clock) + ")" + "\n")
		}

		// Lê a entrada do usuário
		logger.Std("> ")
		fmt.Scanln(&comm)
		number, err := strconv.Atoi(comm)
		if err != nil {
			logger.Std("\nOpção inválida, tente novamente.\n\n")
			continue
		}

		// Envia HELLO para o destino escolhido
		if number == 0 {
			break
		} else if number > 0 && number <= len(addrList) {
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
	entries, err := os.ReadDir(sharedPath)
	check(err)
	for _, entry := range entries {
		logger.Std("\t" + entry.Name() + "\n")
	}
}

// Função para mensagem LS, solicita para os vizinhos onlines os seus arquivos
func LsRequest(knownPeers *peers.SafePeers, senderAddress string, sharedPath string, chunkSize int, statistics *[]Statistic) {
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
		logger.Std("Não havia nenhum peer online na busca\n")
	} else if files.Empty() {
		logger.Std("Não havia nenhum arquivo disponível na busca\n")
	} else {
		DlMenu(knownPeers, senderAddress, sharedPath, files, chunkSize, statistics)
	}
}

// Função para mensagem DL, escolhe um arquivo dentre os buscados para baixar
func DlMenu(knownPeers *peers.SafePeers, senderAddress string, sharedPath string, fileList *FileList, chunkSize int, statistics *[]Statistic) {
	// Declara variável para o comando e inicia o loop do menu de arquivos
	var comm string
	for {
		// Encontra o nome e o tamanho com maior quantidade de caracteres
		biggestName := len("<Cancelar>")
		biggestSize := len("Tamanho")
		for _, file := range fileList.files {
			if len(file.name) > biggestName {
				biggestName = len(file.name)
			}
			if len(strconv.Itoa(file.size)) > biggestSize {
				biggestSize = len(strconv.Itoa(file.size))
			}
		}

		// Formata o menu de opções
		header := fmt.Sprintf("\t     %%-%ds | %%-%ds | %%s\n", biggestName, biggestSize)
		row := fmt.Sprintf("\t[%%2d] %%-%ds | %%-%ds | %%s\n", biggestName, biggestSize)

		// Imprime o menu de opções
		logger.Std("\nArquivos encontrados na rede:\n")
		logger.Std(fmt.Sprintf(header, "Nome", "Tamanho", "Peer"))
		logger.Std(fmt.Sprintf(row, 0, "<Cancelar>", "", ""))
		for i, file := range fileList.files {
			logger.Std(fmt.Sprintf(row, i+1, file.name, strconv.Itoa(file.size), file.OriginsString()))
		}

		// Lê a entrada do usuário
		logger.Std("\nDigite o numero do arquivo para fazer o download:\n> ")
		fmt.Scanln(&comm)
		number, err := strconv.Atoi(comm)
		if err != nil {
			logger.Std("\nOpção inválida, tente novamente.\n")
			continue
		}

		// Solicitação de download para o arquivo escolhido
		if number == 0 {
			break
		} else if number > 0 && number <= fileList.Len() {
			DlRequest(knownPeers, fileList.files[number-1], senderAddress, sharedPath, chunkSize, statistics)
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

	// Constrói a mensagem a ser enviada.
	arguments := []string{cfg.file.name, strconv.Itoa(cfg.chunkSize), strconv.Itoa(index)}
	sendMessage := message.BaseMessage{Origin: cfg.senderAddress, Clock: 0, Type: message.DL, Arguments: arguments}

	// Nessa mensagem em específico, enviamos com o contexto. Se ele for cancelado, as mensagens
	// Param de ser enviadas mais rapidamente.
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", origin)
	connection.SendMessage(cfg.knownPeers, conn, sendMessage, origin)
	if err != nil {
		// Se há algum erro, enviamos para o retry e matamos o manager que chamou esse requestChunk.
		// Isso vale para todos os erros dessa função.
		defer cancel()
		cfg.retryCh <- &DlResponse{index: index, err: err, origin: origin}
		return
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(10 * time.Second))

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

	// Envia o resultado de sucesso para o canal de resultados.
	cfg.resultCh <- &DlResponse{index: receivedIdx, hash: receivedMessage.Arguments[3], err: nil, origin: receivedMessage.Origin}
}

// Função principal do manager. Envia requisições considerando um intervalo de índices.
func (om *OriginManager) createRequests(initialIndex, finalIndex int) {
	defer om.cfg.mainWg.Done()

	// Semáforo para limitar a quantidade de requisições enviadas.
	sem := make(chan struct{}, MAX_CONCURRENT_PER_MANAGER)

	var lastCreatedIndex int
	// O loop está nomeado para caso haja alguma falha seja fácil de sair dele.
mainloop:
	for indexNum := initialIndex; indexNum < finalIndex; indexNum++ {
		// select está checando sempre se houve alguma falha.
		select {
		case <-om.ctx.Done():
			// Caso haja alguma falha, envia o último index enviado e o final para o rebalance.
			lastCreatedIndex = indexNum
			om.cfg.rebalanceCh <- &ManagerError{
				lastCreatedIndex: lastCreatedIndex,
				finalIndex:       finalIndex,
				origin:           om.origin,
			}
			om.cfg.healthyOrigins.Remove(om.origin) // remove o peer dos peers saudáveis.
			break mainloop
		default:
			sem <- struct{}{} // adiciona +1 no semáforo.
			om.wg.Add(1)      // adiciona +1 no waitgroup.
			// Goroutine que envia a requisição de determinado index.
			go func(index int) {
				defer func() { <-sem }() // essa função libera o semáforo no fim da execução da request.
				requestChunk(om.ctx, om.cancel, om.cfg, om.wg, index, om.origin)
			}(indexNum)
		}
	}
	// Espera todas as requisições terminem.
	om.wg.Wait()
}

// Goroutine responsável para executar retries em peers que ainda estão vivos.
func retryManager(cfg *OriginManagerConfig, retryWg *sync.WaitGroup) {
	defer retryWg.Done()

	// Conta quantas vezes um peer falhou.
	retryCounts := make(map[int]int)

	for failedReq := range cfg.retryCh {
		chunkIndex := failedReq.index
		failedOrigin := failedReq.origin

		// Tenta remover a origem.
		cfg.healthyOrigins.Remove(failedOrigin)

		// Se um chunk falha MAX_RETRIES_PER_CHUNK, o download é cancelado.
		retryCounts[chunkIndex]++
		if retryCounts[chunkIndex] > MAX_RETRIES_PER_CHUNK {
			logger.Debug(fmt.Sprintf("Chunk %d failed more than %d times. Aborting download. Last faling origin: %s", chunkIndex, MAX_RETRIES_PER_CHUNK, failedReq.origin))
			return
		}

		newOrigin, err := cfg.healthyOrigins.GetNext() // Escolhemos uma origem aleatória.
		if err != nil {
			// Função que adiciona um tempo antes de reenviar a requisição para o retry.
			go func(resp *DlResponse) {
				time.Sleep(time.Second)
				cfg.retryCh <- resp
			}(failedReq)
		}

		// Aumenta em 1 o WaitGroup e envia reenvia a requisição para determinado chunk.
		retryWg.Add(1)
		ctx, cancel := context.WithCancel(context.Background())
		go requestChunk(ctx, cancel, cfg, retryWg, chunkIndex, newOrigin)
	}
}

// Goroutine responsável pelo rebalanceamento das requisições caso um manager falhe.
func rebalanceManager(cfg *OriginManagerConfig, rebalanceWg *sync.WaitGroup) {
	defer rebalanceWg.Done()

	// semáforo para limitar o número de requisições concorrentes enviadas a cada peer.
	sem := make(chan struct{}, MAX_CONCURRENT_PER_MANAGER)

	for job := range cfg.rebalanceCh {
		if job.lastCreatedIndex-job.finalIndex == 0 {
			continue
		}

		// Caso haja alguma requisição para rebalancear, captura todos os peers ainda saudáveis
		healthyPeers, err := cfg.healthyOrigins.GetAll()
		if err != nil {
			return
		}

		peerCount := len(healthyPeers)
		counter := 0
		// lógica de redistribuição de carga round robin.
		for i := job.lastCreatedIndex; i < job.finalIndex; i++ {
			sem <- struct{}{} // envia uma struct para o semáforo. Se ele estiver cheio, a rotina espera um espaço.
			originIdx := counter % peerCount
			newOrigin := healthyPeers[originIdx]
			rebalanceWg.Add(1) // Avisa para o WaitGroup que está chegando mais uma requisição.

			// Função que vai enviar 1 requisição para alguma origem disponível.
			go func(idx int, origin string) {
				chunkReqCtx, chunkReqCancel := context.WithCancel(context.Background())

				// Essa função vai ser executada no final da operação da atual goroutine.
				// Ela finaliza o contexto corretamente e libera um espaço do semáforo, consumindo-o.
				defer func() {
					chunkReqCancel()
					<-sem
				}()

				// requisição sendo enviada.
				requestChunk(chunkReqCtx, chunkReqCancel, cfg, rebalanceWg, idx, origin)
			}(i, newOrigin)

			counter++
		}
	}
}

// Função para mensagem DL, solicita o download do arquivo escolhido em chunks.
// Funcionamento: criamos 3 tipos de goroutines: n managers, 1 RetryManager e 1 RebalanceManager.
// Caso uma goroutine manager falhe em alguma requisição, envia a falha para RetryManager, depois
// é morta e envia as requisições que ainda fez para o RebalanceManager. O RebalanceManager divide
// todas as requisições de quem teve um erro entre as origens que ainda estão ativas.
// Para o RebalanceManager e o RetryManager, um peer só é dado como morto mesmo depois de um certo
// número de falhas. Caso todos os peers morram durante o download, ele é cancelado.
func DlRequest(knownPeers *peers.SafePeers, file File, senderAddress string, sharedPath string, chunkSize int, statistics *[]Statistic) error {
	logger.Std("\nArquivo escolhido " + file.name + "\n")
	startTime := time.Now()

	// Calcula a quantidade de requisições necessárias e cria o canal de respostas
	totalRequests := int(math.Ceil(float64(file.size) / float64(chunkSize)))

	// cria os canais de comunicação para o valor resultado retornado, resiliência (retry) e rebalanceamento
	resultCh := make(chan *DlResponse, totalRequests)
	retryCh := make(chan *DlResponse, totalRequests)
	rebalanceCh := make(chan *ManagerError, len(file.origin))

	// grupos de espera para sincronizar chamadas que foram para os canais de resiliência e rebalanceamento
	var retryWg sync.WaitGroup
	var rebalanceWg sync.WaitGroup

	// configuração comum para os gerenciadores de origem
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

	// criamos o array de gerentes e populamos
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

	nManagers := len(managers)
	cfg.mainWg.Add(nManagers)

	// A partir daqui ocorre a lógica de distribuição inicial das requisições entre os peers.
	base := totalRequests / nManagers
	remainder := totalRequests % nManagers
	start := 0
	for i := range managers {
		count := base
		if i < remainder {
			count++
		}
		end := start + count
		if start < end {
			go managers[i].createRequests(start, end)
		}
		start = end
	}

	// Iniciando goroutines responsáveis para resiliência
	retryWg.Add(1)
	go retryManager(&cfg, &retryWg)
	rebalanceWg.Add(1)
	go rebalanceManager(&cfg, &rebalanceWg)

	// Rotina para fechar os canais de comunicação na ordem correta.
	// Primeiro, esperamos os Managers terminarem.
	// Depois, fechamos o canal de rebalanceamento e esperamos todas as operações do rebalanceManager terminarem.
	// Por fim, fechamos o canal de retry e esperamos todas as chamadas de retry restantes terminarem e fechamos o canal de resultados.
	go func() {
		cfg.mainWg.Wait()
		close(rebalanceCh)
		rebalanceWg.Wait()
		close(cfg.retryCh) // aqui pode ter um bug. O RetryManager em alguns momentos coloca na própria fila, então tem chance de colocar nela já fechada
		retryWg.Wait()
		close(resultCh)
	}()

	// Array para guardar os hashs recebidos.
	receivedHashes := make([]string, totalRequests)

	// Nesse loop, como iteramos em cima de um go channel, ele espera mensagens chegarem nele até que o canal se feche.
	for dlResponse := range resultCh {
		receivedHashes[dlResponse.index] = dlResponse.hash
	}

	// Salva a nova estatística do download
	finalTime := time.Since(startTime).Seconds()
	found := false
	for i, stat := range *statistics {
		if stat.chunckSize == chunkSize && stat.peersQty == len(file.origin) && stat.fileSize == file.size {
			(*statistics)[i].times = append((*statistics)[i].times, finalTime)
			found = true
			break
		}
	}
	if !found {
		*statistics = append(*statistics, Statistic{
			chunckSize: chunkSize,
			peersQty:   len(file.origin),
			fileSize:   file.size,
			times:      []float64{finalTime},
		})
	}

	// Loop em que decodificamos os hashs recebidos e tratamos erros
	var decodedChunks []byte
	for i, r := range receivedHashes {
		if r == "" {
			logger.Std("Não foi possível fazer o download.")
			return fmt.Errorf("chunk %d está vazio. próximo chunk: %s. Total chunks: %d", i, receivedHashes[i+1], totalRequests)
		}
		dec, err := base64.StdEncoding.DecodeString(r)
		if err != nil {
			logger.Std("Não foi possível fazer o download.")
			return fmt.Errorf("erro ao decodificar chunk %d: %v", i, err)
		}
		decodedChunks = append(decodedChunks, dec...)
	}

	// Cria/substitui o arquivo e escreve o conteúdo decodificado
	createdFile, err := os.Create(sharedPath + file.name)
	if err != nil {
		return err
	}
	defer createdFile.Close()
	_, err = createdFile.Write(decodedChunks)
	if err != nil {
		return err
	}
	logger.Std("\nDownload do arquivo " + file.name + " finalizado.\n")
	//logger.Std("\nErros de peer: " + cfg.healthyOrigins.ErrorSummary())
	return nil
}

// Função para mostrar as estatísticas do download
func ShowStatistics(statistics *[]Statistic) {
	// Encontra os maiores tamanhos de cada coluna
	biggestChunkSize := len("Tam. chunk")
	biggestPeersQty := len("N peers")
	biggestFileSize := len("Tam. arquivo")
	biggestAttemps := len("N")
	biggestMeanTime := len("Tempo [s]")
	for _, stat := range *statistics {
		if len(strconv.Itoa(stat.chunckSize)) > biggestChunkSize {
			biggestChunkSize = len(strconv.Itoa(stat.chunckSize))
		}
		if len(strconv.Itoa(stat.peersQty)) > biggestPeersQty {
			biggestPeersQty = len(strconv.Itoa(stat.peersQty))
		}
		if len(strconv.Itoa(stat.fileSize)) > biggestFileSize {
			biggestFileSize = len(strconv.Itoa(stat.fileSize))
		}
		if len(strconv.Itoa(len(stat.times))) > biggestAttemps {
			biggestAttemps = len(strconv.Itoa(len(stat.times)))
		}
	}

	// Formata a tabela de estatísticas
	header := fmt.Sprintf("%%-%ds | %%-%ds | %%-%ds | %%-%ds | %%-%ds | %%s\n", biggestChunkSize, biggestPeersQty, biggestFileSize, biggestAttemps, biggestMeanTime)
	row := fmt.Sprintf("%%-%ds | %%-%ds | %%-%ds | %%-%ds | %%-%ds | %%s\n", biggestChunkSize, biggestPeersQty, biggestFileSize, biggestAttemps, biggestMeanTime)

	// Imprime a tabela de estatísticas
	logger.Std(fmt.Sprintf(header, "Tam. chunk", "N peers", "Tam. arquivo", "N", "Tempo [s]", "Desvio"))
	for _, stat := range *statistics {
		stdDeviation, meanTime := 0.0, 0.0
		for _, t := range stat.times {
			meanTime += t
		}
		meanTime /= float64(len(stat.times))
		for _, t := range stat.times {
			stdDeviation += (t - meanTime) * (t - meanTime)
		}
		stdDeviation = math.Sqrt(stdDeviation / float64(len(stat.times)))
		logger.Std(fmt.Sprintf(row, strconv.Itoa(stat.chunckSize), strconv.Itoa(stat.peersQty), strconv.Itoa(stat.fileSize), strconv.Itoa(len(stat.times)), fmt.Sprintf("%.5f", meanTime), fmt.Sprintf("%.5f", stdDeviation)))
	}
}

// Função para alterar o tamanho do chunk
func ChangeChunk(chunkSize *int) {
	var chunk string
	logger.Std("Digite novo tamanho de chunk:\n> ")
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
