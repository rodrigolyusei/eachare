package main

// Pacotes nativos de go e pacotes internos
import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"eachare/src/clock"
	"eachare/src/commands"
	"eachare/src/connection"
	"eachare/src/logger"
	"eachare/src/message"
	"eachare/src/peers"
	"eachare/src/response"
)

type Client struct {
	address    string           // Endereço do peer
	neighbors  string           // Vizinhos do peer
	shared     string           // Diretório compartilhado do peer
	knownPeers *peers.SafePeers // Lista dos peers conhecidos seguro para concorrência
	waitingCli bool             // Variável para controlar o estado do CLI
	chunkSize  int              // Tamanho do chunk
}

func NewClient(address string, neighbors string, shared string) Client {
	return Client{
		address:    address,
		neighbors:  neighbors,
		shared:     shared,
		knownPeers: &peers.SafePeers{},
		waitingCli: false,
		chunkSize:  256,
	}
}

// Função para verificar e imprimir mensagem de erro
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Função para modo de teste, simulando a execução do programa com argumentos específicos
func testArgs(args []string) *Client {
	port := 10000
	// Vai criando um listener TCP em portas diferentes até encontrar uma porta livre
	for {
		port++
		listener, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
		if err != nil {
			continue
		}
		listener.Close()
		break
	}

	client := NewClient("127.0.0.1:"+strconv.Itoa(port), args[2], args[3])

	// Cria um mapa de peers dinamicamente
	if port%2 == 0 {
		client.knownPeers.Add(peers.Peer{Address: "127.0.0.1:" + strconv.Itoa(port+1), Status: peers.ONLINE, Clock: 0})
		client.knownPeers.Add(peers.Peer{Address: "127.0.0.1:" + strconv.Itoa(port+2), Status: peers.OFFLINE, Clock: 0})
	} else {
		client.knownPeers.Add(peers.Peer{Address: "127.0.0.1:" + strconv.Itoa(port+1), Status: peers.ONLINE, Clock: 0})
		client.knownPeers.Add(peers.Peer{Address: "127.0.0.1:" + strconv.Itoa(port+3), Status: peers.OFFLINE, Clock: 0})
	}

	// Imprime os parâmetros de entrada
	logger.Std("Modo de teste\n")
	logger.Std("Endereço: " + client.address + "\n")
	logger.Std("Vizinhos: " + client.neighbors + "\n")
	logger.Std("Diretório Compartilhado: " + client.shared + "\n")
	return &client
}

// Função para obter os argumentos de entrada
func getArgs(args []string) *Client {
	// Verifica a quantidade de parâmetros e o formato do endereço
	if len(args) != 4 {
		str1 := "\nParâmetros de entrada inválidos, por favor, siga o formato abaixo:"
		str2 := "\n./eachare <endereço>:<porta> <vizinhos> <diretório compartilhado>"
		check(errors.New(str1 + str2))
	} else if !strings.Contains(args[1], ":") {
		str1 := "\nEndereço e porta inválidos, por favor, siga o formato abaixo:"
		str2 := "\n./eachare <endereço>:<porta> <vizinhos> <diretório compartilhado>"
		check(errors.New(str1 + str2))
	}

	// Define os parâmetros se estiverem corretos
	client := NewClient(args[1], args[2], args[3])
	return &client
}

// Função para adicionar vizinhos conhecidos a partir de um arquivo
func (c *Client) addNeighbors() {
	// Abre o arquivo de vizinhos
	file, err := os.Open(c.neighbors)
	check(err)
	defer file.Close()

	// Lê o arquivo linha por linha
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		c.knownPeers.Add(peers.Peer{Address: scanner.Text(), Status: peers.OFFLINE, Clock: 0})
		logger.Std("Adicionando novo peer " + scanner.Text() + " status " + peers.OFFLINE.String() + "\n")
	}
}

// Verifica se o diretório compartilhado existe e está acessível
func (c *Client) verifySharedDirectory() {
	_, err := os.ReadDir(c.shared)
	check(err)
	if c.shared[len(c.shared)-1:] != "/" {
		c.shared += "/"
	}
}

// Função para a CLI/menu de interação com o usuário
func cliInterface(client *Client) {
	// Declara variável para o comando e saída, depois inicia o loop do menu
	var comm string
	var exit bool = false
	for !exit {
		// Indica que a CLI está esperando por uma entrada
		client.waitingCli = true

		// Imprime o menu de opções
		logger.Std("\nEscolha um comando:\n")
		logger.Std("\t[1] Listar peers\n")
		logger.Std("\t[2] Obter peers\n")
		logger.Std("\t[3] Listar arquivos locais\n")
		logger.Std("\t[4] Buscar arquivos\n")
		logger.Std("\t[5] Exibir estatisticas\n")
		logger.Std("\t[6] Alterar tamanho de chunk\n")
		logger.Std("\t[9] Sair\n")

		// Lê a entrada do usuário
		logger.Std("> ")
		fmt.Scanln(&comm)
		logger.Std("\n")

		// Executa o comando correspondente
		switch comm {
		case "1":
			commands.ListPeers(client.knownPeers, client.address)
		case "2":
			commands.GetPeersRequest(client.knownPeers, client.address)
		case "3":
			commands.ListLocalFiles(client.shared)
		case "4":
			commands.LsRequest(client.knownPeers, client.address, client.shared)
		case "5":
			logger.Std("Comando ainda não implementado.\n")
		case "6":
			client.chunkSize = commands.ChangeChunk()
			logger.Info("Tamanho de chunk alterado: " + strconv.Itoa(client.chunkSize))
			//logger.Std("Comando ainda não implementado.\n")
		case "9":
			commands.ByeRequest(client.knownPeers, client.address)
			exit = true
		default:
			logger.Std("Comando inválido, tente novamente.\n")
		}

		// Indica que a CLI não está mais esperando por uma entrada
		client.waitingCli = false
		time.Sleep(500 * time.Millisecond)
	}

	// Encerra o programa
	os.Exit(0)
}

// Função para iniciar o peer e escutar conexões
func listener(client *Client) {
	// Cria um listener TCP no endereço e porta especificado
	listener, err := net.Listen("tcp", client.address)
	check(err)
	defer listener.Close()

	// Loop para receber mensagens de outros peers
	for {
		// Accept trava o programa até receber uma conexão
		conn, err := listener.Accept()
		check(err)

		// Cria uma goroutine/thread para lidar com a conexão recebida
		go receiver(conn, client)
	}
}

// Função para lidar com a conexão recebida
func receiver(conn net.Conn, client *Client) {
	// defer (adia) o fechamento da conexão até o final da função
	defer conn.Close()

	// Recebe a mensagem da conexão recebida
	receivedMessage := connection.ReceiveMessage(client.knownPeers, conn)

	// Se a CLI está esperando por uma entrada formata
	if client.waitingCli {
		logger.Std("\n\n")
	}
	logger.Info("Mensagem recebida: \"" + receivedMessage.String() + "\"")

	// Atualiza o relógio local comparando o valor local e recebido
	clock.UpdateMaxClock(receivedMessage.Clock)

	// Mostra mensagem de adição se não tinha o peer e atualização se tinha não é BYE
	neighbor, exists := client.knownPeers.Get(receivedMessage.Origin)
	if !exists {
		logger.Info("Adicionando novo peer " + receivedMessage.Origin + " status " + peers.ONLINE.String())
	} else if receivedMessage.Type != message.BYE {
		logger.Info("Atualizando peer " + receivedMessage.Origin + " status " + peers.ONLINE.String())
	}

	// Lida o comando recebido de acordo com o tipo de mensagem
	switch receivedMessage.Type {
	case message.GET_PEERS:
		response.GetPeersResponse(client.knownPeers, receivedMessage.Origin, client.address, conn)
	case message.LS:
		response.LsResponse(client.knownPeers, receivedMessage.Origin, client.address, client.shared, conn)
	case message.DL:
		response.DlResponse(client.knownPeers, receivedMessage, client.address, client.shared, conn)
	case message.BYE:
		response.ByeResponse(client.knownPeers, receivedMessage.Origin, neighbor.Clock)
	}

	// Verifica se a CLI está esperando por uma entrada
	if client.waitingCli {
		logger.Std("\n> ")
	}
}

// Função principal do programa
func main() {
	// Verifica se o programa está sendo executado em modo de teste ou não
	var client *Client

	if len(os.Args) == 5 && os.Args[4] == "--test" {
		// Cria os argumentos de teste
		client = testArgs(os.Args)
	} else {
		// Pega os argumentos de entrada
		client = getArgs(os.Args)

		// Adiciona os vizinhos conhecidos no arquivo de vizinhos
		client.addNeighbors()
	}

	// Verifica o diretório compartilhado
	client.verifySharedDirectory()

	// Cria uma goroutine/thread para a CLI
	go cliInterface(client)

	// Inicializa o peer
	listener(client)
}
