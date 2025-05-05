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

	"EACHare/src/clock"
	"EACHare/src/commands"
	"EACHare/src/connection"
	"EACHare/src/logger"
	"EACHare/src/message"
	"EACHare/src/peers"
	"EACHare/src/response"
)

// Variáveis globais
var myAddress string           // Endereço do peer
var myNeighbors string         // Vizinhos do peer
var myShared string            // Diretório compartilhado do peer
var knownPeers peers.SafePeers // Lista dos peers conhecidos seguro para concorrência
var waitingCli = false         // Variável para controlar o estado do CLI

// Função para verificar e imprimir mensagem de erro
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Função para modo de teste, simulando a execução do programa com argumentos específicos
func testArgs(args []string) {
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

	// Cria um mapa de peers dinamicamente
	if port%2 == 0 {
		knownPeers.Add(peers.Peer{Address: "127.0.0.1:" + strconv.Itoa(port+1), Status: peers.ONLINE, Clock: 0})
		knownPeers.Add(peers.Peer{Address: "127.0.0.1:" + strconv.Itoa(port+2), Status: peers.OFFLINE, Clock: 0})
	} else {
		knownPeers.Add(peers.Peer{Address: "127.0.0.1:" + strconv.Itoa(port+1), Status: peers.ONLINE, Clock: 0})
		knownPeers.Add(peers.Peer{Address: "127.0.0.1:" + strconv.Itoa(port+3), Status: peers.OFFLINE, Clock: 0})
	}

	myAddress = "127.0.0.1:" + strconv.Itoa(port)
	myNeighbors = args[2]
	myShared = args[3]

	// Imprime os parâmetros de entrada
	logger.Std("Modo de teste\n")
	logger.Std("Endereço: " + myAddress + "\n")
	logger.Std("Vizinhos: " + myNeighbors + "\n")
	logger.Std("Diretório Compartilhado: " + myShared + "\n")
}

// Função para obter os argumentos de entrada
func getArgs(args []string) {
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
	myAddress = args[1]
	myNeighbors = args[2]
	myShared = args[3]
}

// Função para adicionar vizinhos conhecidos a partir de um arquivo
func addNeighbors() {
	// Abre o arquivo de vizinhos
	file, err := os.Open(myNeighbors)
	check(err)
	defer file.Close()

	// Lê o arquivo linha por linha
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		knownPeers.Add(peers.Peer{Address: scanner.Text(), Status: peers.OFFLINE, Clock: 0})
		logger.Std("Adicionando novo peer " + scanner.Text() + " status " + peers.OFFLINE.String() + "\n")
	}
}

// Verifica se o diretório compartilhado existe e está acessível
func verifySharedDirectory() {
	_, err := os.ReadDir(myShared)
	check(err)
}

// Função para a CLI/menu de interação com o usuário
func cliInterface() {
	// Declara variável para o comando e saída, depois inicia o loop do menu
	var comm string
	var exit bool = false
	for !exit {
		// Indica que a CLI está esperando por uma entrada
		waitingCli = true

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
			commands.ListPeers(&knownPeers, myAddress)
		case "2":
			commands.GetPeersRequest(&knownPeers, myAddress)
		case "3":
			commands.ListLocalFiles(myShared)
		case "4":
			commands.LsRequest(&knownPeers, myAddress)
		case "5":
			logger.Std("Comando ainda não implementado.\n")
		case "6":
			logger.Std("Comando ainda não implementado.\n")
		case "9":
			commands.ByeRequest(&knownPeers, myAddress)
			exit = true
		default:
			logger.Std("Comando inválido, tente novamente.\n")
		}

		// Indica que a CLI não está mais esperando por uma entrada
		waitingCli = false
		time.Sleep(500 * time.Millisecond)
	}

	// Encerra o programa
	os.Exit(0)
}

// Função para iniciar o peer e escutar conexões
func listener() {
	// Cria um listener TCP no endereço e porta especificado
	listener, err := net.Listen("tcp", myAddress)
	check(err)
	defer listener.Close()

	// Loop para receber mensagens de outros peers
	for {
		// Accept trava o programa até receber uma conexão
		conn, err := listener.Accept()
		check(err)

		// Cria uma goroutine/thread para lidar com a conexão recebida
		go receiver(conn, &knownPeers, waitingCli)
	}
}

// Função para lidar com a conexão recebida
func receiver(conn net.Conn, knownPeers *peers.SafePeers, waitingCli bool) {
	// defer (adia) o fechamento da conexão até o final da função
	defer conn.Close()

	// Recebe a mensagem da conexão recebida
	receivedMessage := connection.ReceiveMessage(knownPeers, conn)

	// Se a CLI está esperando por uma entrada e não é um PEERS_LIST, formata
	if waitingCli {
		logger.Std("\n\n")
	}
	logger.Info("Mensagem recebida: \"" + receivedMessage.String() + "\"")

	// Atualiza o relógio local comparando o valor local e recebido
	clock.UpdateMaxClock(receivedMessage.Clock)

	// Mostra mensagem de atualização apenas se for de peer OFFLINE e não for uma mensagem de BYE
	neighbor, exists := knownPeers.Get(receivedMessage.Origin)
	if !exists {
		logger.Info("Adicionando novo peer " + receivedMessage.Origin + " status " + peers.ONLINE.String())
	} else if receivedMessage.Type != message.BYE {
		logger.Info("Atualizando peer " + receivedMessage.Origin + " status " + peers.ONLINE.String())
	}

	// Lida o comando recebido de acordo com o tipo de mensagem
	switch receivedMessage.Type {
	case message.GET_PEERS:
		response.GetPeersResponse(knownPeers, receivedMessage.Origin, myAddress, conn)
	case message.LS:
		response.LsResponse(knownPeers, receivedMessage.Origin, myAddress, myShared, conn)
	case message.BYE:
		response.ByeResponse(knownPeers, receivedMessage.Origin, neighbor.Clock)
	}

	// Verifica se a CLI está esperando por uma entrada
	if waitingCli {
		logger.Std("\n> ")
	}
}

// Função principal do programa
func main() {
	// Verifica se o programa está sendo executado em modo de teste ou não
	if len(os.Args) == 5 && os.Args[4] == "--test" {
		// Cria os argumentos de teste
		testArgs(os.Args)
	} else {
		// Pega os argumentos de entrada
		getArgs(os.Args)

		// Adiciona os vizinhos conhecidos no arquivo de vizinhos
		addNeighbors()
	}

	// Verifica o diretório compartilhado
	verifySharedDirectory()

	// Cria uma goroutine/thread para a CLI
	go cliInterface()

	// Inicializa o peer
	listener()
}
