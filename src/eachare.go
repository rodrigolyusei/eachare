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

	"EACHare/src/clock"
	"EACHare/src/commands"
	"EACHare/src/commands/message"
	"EACHare/src/commands/request"
	"EACHare/src/commands/response"
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
var myargs SelfArgs     // Armazena os parâmetros de si mesmo
var waitingCli = false  // Variável para controlar o estado do CLI

// Função para verificar e imprimir mensagem de erro
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Função para modo de teste, simulando a execução do programa com argumentos específicos
func testArgs(args []string) SelfArgs {
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
	myargs := SelfArgs{Address: "127.0.0.1" + strconv.Itoa(port), Neighbors: args[2], Shared: args[3]}

	// Imprime os parâmetros de entrada
	fmt.Println("Endereço:", myargs.Address)
	fmt.Println("Vizinhos:", myargs.Neighbors)
	fmt.Println("Diretório Compartilhado:", myargs.Shared)

	return myargs
}

// Função para obter os argumentos de entrada
func getArgs(args []string) SelfArgs {
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
	return SelfArgs{Address: args[1], Neighbors: args[2], Shared: args[3]}
}

// Função para adicionar vizinhos conhecidos a partir de um arquivo
func addNeighbors(neighborsPath string) {
	// Abre o arquivo de vizinhos
	file, err := os.Open(neighborsPath)
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
func verifySharedDirectory(sharedPath string) {
	_, err := os.ReadDir(sharedPath)
	check(err)
}

// Função para iniciar o peer e escutar conexões
func listener(args SelfArgs, requestClient request.RequestClient) {
	// Cria um listener TCP no endereço e porta especificado
	listener, err := net.Listen("tcp", args.Address)
	check(err)
	defer listener.Close()

	// Loop para receber mensagens de outros peers
	for {
		// Accept trava o programa até receber uma conexão
		conn, err := listener.Accept()
		check(err)

		// Cria uma goroutine/thread para lidar com a conexão recebida
		go receiveMessage(conn, requestClient)
	}
}

// Função para lidar com a conexão recebida
func receiveMessage(conn net.Conn, requestClient request.RequestClient) {
	// defer(adia) a função de fechamento da conexão quando as operações terminarem
	defer conn.Close()

	// Se a CLI está esperando por uma entrada, imprime nova linha para formatação
	if waitingCli {
		logger.Std("\n\n")
	}

	// Lê a mensagem recebida no buffer até encontrar \n
	msg, err := bufio.NewReader(conn).ReadString('\n')
	check(err)

	// Recupera as partes da mensagem
	msg = strings.TrimSuffix(msg, "\n")
	msgParts := strings.Split(msg, " ")

	// Imprime a mensagem/resposta recebida e atualiza o clock
	if msgParts[2] == "PEERS_LIST" {
		logger.Info("\tResposta recebida: \"" + msg + "\"")
	} else {
		logger.Info("\tMensagem recebida: \"" + msg + "\"")
	}

	// Guarda o valor do clock da mensagem recebida
	receivedClock, err := strconv.Atoi(msgParts[1])
	check(err)

	// Atualiza o relógio local comparando o valor local e recebido
	clock.UpdateMaxClock(receivedClock)

	// Monta a mensagem recebida
	receivedMessage := message.BaseMessage{
		Origin:    msgParts[0],
		Clock:     receivedClock,
		Type:      message.GetMessageType(msgParts[2]),
		Arguments: msgParts[3:],
	}

	// Verifica se a mensagem recebida é de um peer conhecido
	neighbor, exists := knownPeers.Load(receivedMessage.Origin)
	if exists {
		neighborStatus := neighbor.(peers.Peer).Status
		neighborClock := neighbor.(peers.Peer).Clock

		// Atualiza o status do peer conhecido com o maior clock
		if receivedClock > neighborClock {
			knownPeers.Store(receivedMessage.Origin, peers.Peer{Status: peers.ONLINE, Clock: receivedClock})
		} else {
			knownPeers.Store(receivedMessage.Origin, peers.Peer{Status: peers.ONLINE, Clock: neighborClock})
		}

		// Mostra mensagem de atualização apenas se for de peer OFFLINE e não for uma mensagem de BYE
		if neighborStatus == peers.OFFLINE && receivedMessage.Type != message.BYE {
			logger.Info("\tAtualizando peer " + receivedMessage.Origin + " status " + peers.ONLINE.String())
		}
	} else {
		knownPeers.Store(receivedMessage.Origin, peers.Peer{Status: peers.ONLINE, Clock: receivedClock})
		logger.Info("\tAdicionando novo peer " + receivedMessage.Origin + " status " + peers.ONLINE.String())
	}

	// Lida o comando recebido de acordo com o tipo de mensagem
	neighbor, _ = knownPeers.Load(receivedMessage.Origin)
	switch receivedMessage.Type {
	case message.GET_PEERS:
		response.GetPeersResponse(receivedMessage, &knownPeers, conn, requestClient)
	case message.PEERS_LIST:
		response.PeersListResponse(receivedMessage, &knownPeers)
	case message.BYE:
		response.ByeResponse(receivedMessage, &knownPeers, neighbor.(peers.Peer).Clock)
	}

	// Verifica se a CLI está esperando por uma entrada
	if waitingCli {
		logger.Std("\n> ")
	}
}

// Função para a CLI/menu de interação com o usuário
func cliInterface(args SelfArgs, requestClient request.RequestClient) {
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
				go receiveMessage(conn, requestClient)
			}
		case "3":
			commands.ListLocalFiles(args.Shared)
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
		myargs = testArgs(os.Args)
	} else {
		// Pega os argumentos de entrada
		myargs = getArgs(os.Args)

		// Adiciona os vizinhos conhecidos no arquivo de vizinhos
		addNeighbors(myargs.Neighbors)
	}

	// Cria o cliente de requisições que será usado para enviar mensagens
	requestClient := request.RequestClient{Address: myargs.Address}

	// Verifica o diretório compartilhado
	verifySharedDirectory(myargs.Shared)

	// Cria uma goroutine/thread para a CLI
	go cliInterface(myargs, requestClient)

	// Inicializa o peer
	listener(myargs, requestClient)
}
