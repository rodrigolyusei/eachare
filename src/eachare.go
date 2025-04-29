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
	"sync"
	"time"

	"EACHare/src/commands"
	"EACHare/src/commands/request"
	"EACHare/src/connection"
	"EACHare/src/logger"
	"EACHare/src/peers"
)

// Struct para os argumentos de entrada, sendo as informações do Peer próprio
type SelfArgs struct {
	Address   string
	Neighbors string
	Shared    string
}

// Variáveis globais
var knownPeers sync.Map // Hashmap syncronizado para os peers conhecidos
var myArgs SelfArgs     // Armazena os parâmetros de si mesmo
var waitingCli = false  // Variável para controlar o estado do CLI

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
		knownPeers.Store("127.0.0.1:"+strconv.Itoa(port+1), peers.Peer{Status: peers.ONLINE, Clock: 0})
		knownPeers.Store("127.0.0.1:"+strconv.Itoa(port+2), peers.Peer{Status: peers.OFFLINE, Clock: 0})
	} else {
		knownPeers.Store("127.0.0.1:"+strconv.Itoa(port+1), peers.Peer{Status: peers.ONLINE, Clock: 0})
		knownPeers.Store("127.0.0.1:"+strconv.Itoa(port+3), peers.Peer{Status: peers.OFFLINE, Clock: 0})
	}

	// Cria o SelfArgs com os argumentos de teste
	myArgs = SelfArgs{Address: "127.0.0.1:" + strconv.Itoa(port), Neighbors: args[2], Shared: args[3]}

	// Imprime os parâmetros de entrada
	fmt.Println("Modo de teste")
	fmt.Println("Endereço:", myArgs.Address)
	fmt.Println("Vizinhos:", myArgs.Neighbors)
	fmt.Println("Diretório Compartilhado:", myArgs.Shared)
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

	// Se os parâmetros estiverem corretos, retorna a struct preenchida
	myArgs = SelfArgs{Address: args[1], Neighbors: args[2], Shared: args[3]}
}

// Função para adicionar vizinhos conhecidos a partir de um arquivo
func addNeighbors() {
	// Abre o arquivo de vizinhos
	file, err := os.Open(myArgs.Neighbors)
	check(err)
	defer file.Close()

	// Lê o arquivo linha por linha
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		knownPeers.Store(scanner.Text(), peers.Peer{Status: peers.OFFLINE, Clock: 0})
		logger.Info("Adicionando novo peer " + scanner.Text() + " status " + peers.OFFLINE.String())
	}
}

// Verifica se o diretório compartilhado existe e está acessível
func verifySharedDirectory() {
	_, err := os.ReadDir(myArgs.Shared)
	check(err)
}

// Função para iniciar o peer e escutar conexões
func listener(requestClient request.RequestClient) {
	// Cria um listener TCP no endereço e porta especificado
	listener, err := net.Listen("tcp", myArgs.Address)
	check(err)
	defer listener.Close()

	// Loop para receber mensagens de outros peers
	for {
		// Accept trava o programa até receber uma conexão
		conn, err := listener.Accept()
		check(err)

		// Cria uma goroutine/thread para lidar com a conexão recebida
		go connection.ReceiveMessage(conn, &knownPeers, requestClient, waitingCli)
	}
}

// Função para a CLI/menu de interação com o usuário
func cliInterface(requestClient request.RequestClient) {
	// Variável para o comando digitado e a saída
	var comm string
	var exit bool = false

	// Loop para manter a CLI ativa
	for !exit {
		// Indica que a CLI está esperando por uma entrada
		waitingCli = true

		// Imprime o menu de opções
		fmt.Println("\nEscolha um comando:")
		fmt.Println("\t[1] Listar peers")
		fmt.Println("\t[2] Obter peers")
		fmt.Println("\t[3] Listar arquivos locais")
		fmt.Println("\t[4] Buscar arquivos")
		fmt.Println("\t[5] Exibir estatisticas")
		fmt.Println("\t[6] Alterar tamanho de chunk")
		fmt.Println("\t[9] Sair")

		// Lê a entrada do usuário
		fmt.Print("> ")
		fmt.Scanln(&comm)
		fmt.Println()

		// Executa o comando correspondente
		switch comm {
		case "1":
			commands.ListPeers(&knownPeers, requestClient)
		case "2":
			connections := requestClient.GetPeersRequest(&knownPeers)
			for _, conn := range connections {
				go connection.ReceiveMessage(conn, &knownPeers, requestClient, waitingCli)
			}
		case "3":
			commands.ListLocalFiles(myArgs.Shared)
		case "4":
			fmt.Println("Comando ainda não implementado")
		case "5":
			fmt.Println("Comando ainda não implementado")
		case "6":
			fmt.Println("Comando ainda não implementado")
		case "9":
			requestClient.ByeRequest(&knownPeers, &exit)
		default:
			fmt.Println("Comando inválido, tente novamente.")
		}

		// Indica que a CLI não está mais esperando por uma entrada
		waitingCli = false
		time.Sleep(500 * time.Millisecond)
	}

	// Encerra o programa
	os.Exit(0)
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

	// Cria o cliente de requisições que será usado para enviar mensagens
	requestClient := request.RequestClient{Address: myArgs.Address}

	// Verifica o diretório compartilhado
	verifySharedDirectory()

	// Cria uma goroutine/thread para a CLI
	go cliInterface(requestClient)

	// Inicializa o peer
	listener(requestClient)
}
