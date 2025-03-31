package main

// Pacotes nativos de go e pacotes internos
import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"EACHare/src/commands"
	"EACHare/src/number"
	"EACHare/src/peers"
)

// Struct para os argumentos de entrada, sendo as informações do Peer próprio
type SelfArgs struct {
	Address   string
	Port      string
	Neighbors string
	Shared    string
}

// Método do SelfArgs para retornar o endereço completo (endereço + porta)
func (args SelfArgs) FullAddress() string {
	return args.Address + ":" + args.Port
}

// Cria um hashmap para armazenar os peers conhecidos e seus status
var knownPeers = make(map[string]peers.PeerStatus)

// Variável para controlar o estado do CLI
var waiting_cli = false

// Função para verificar e imprimir mensagem de erro
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Função para a CLI/menu de interação com o usuário
func cliInterface(args SelfArgs) {
	for {
		// Indica que a CLI está esperando por uma entrada
		waiting_cli = true

		// Imprime o menu de opções
		fmt.Println("\nEscolha um comando:")
		fmt.Println("\t[1] Listar peers")
		fmt.Println("\t[2] Obter peers")
		fmt.Println("\t[3] Listar arquivos locais")
		fmt.Println("\t[4] Buscar arquivos")
		fmt.Println("\t[5] Exibir estatisticas")
		fmt.Println("\t[6] Alterar tamanho de chunk")
		fmt.Println("\t[9] Sair")
		fmt.Print("> ")

		// Lê a entrada do usuário
		var comm string
		fmt.Scanln(&comm)
		fmt.Println()

		// Executa o comando correspondente
		switch comm {
		case "1":
			commands.ListPeers(knownPeers)
		case "2":
			connections := commands.GetPeersRequest(knownPeers)
			for _, conn := range connections {
				if conn != nil {
					go receiver(conn)
				}
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
			fmt.Println("Saindo...")
			commands.ByeRequest(knownPeers)
			os.Exit(0)
		default:
			fmt.Println("Comando inválido, tente novamente.")
		}

		// Indica que a CLI não está mais esperando por uma entrada
		waiting_cli = false
		time.Sleep(500 * time.Millisecond)
	}
}

// Função de teste para simular a execução do programa com argumentos específicos
func testArgs(args []string) (SelfArgs, error) {
	// Obtém a porta a partir do número no data.txt
	port, err := number.GetNextPort()
	check(err)

	// Cria um mapa de peers dinamicamente
	if port%2 == 0 {
		knownPeers["127.0.0.1:"+strconv.Itoa(port+1)] = peers.ONLINE
		knownPeers["127.0.0.1:"+strconv.Itoa(port+2)] = peers.OFFLINE
	} else {
		knownPeers["127.0.0.1:"+strconv.Itoa(port+1)] = peers.ONLINE
		knownPeers["127.0.0.1:"+strconv.Itoa(port+3)] = peers.OFFLINE
	}

	// Cria o SelfArgs com os argumentos de teste
	myargs := SelfArgs{Address: "127.0.0.1", Port: strconv.Itoa(port), Neighbors: args[2], Shared: args[3]}

	// Imprime os parâmetros de entrada
	fmt.Println("Endereço:", myargs.Address)
	fmt.Println("Porta:", myargs.Port)
	fmt.Println("Vizinhos:", myargs.Neighbors)
	fmt.Println("Diretório Compartilhado:", myargs.Shared)

	return myargs, nil
}

// Função para obter os argumentos de entrada
func getArgs(args []string) (SelfArgs, error) {
	// Verifica se o número de argumentos é válido e se o formato do endereço e porta está correto
	if len(args) != 4 {
		str1 := "\nParâmetros de entrada inválidos, por favor, siga o formato abaixo:"
		str2 := "\n./eachare <endereço>:<porta> <vizinhos> <diretório compartilhado>"
		return SelfArgs{}, errors.New(str1 + str2)
	} else if !strings.Contains(args[1], ":") {
		str1 := "\nEndereço e porta inválidos, por favor, siga o formato abaixo:"
		str2 := "\n./eachare <endereço>:<porta> <vizinhos> <diretório compartilhado>"
		return SelfArgs{}, errors.New(str1 + str2)
	}

	// Se os parâmetros estiverem corretos, retorna cada uma separadamente
	x := strings.Split(args[1], ":")
	return SelfArgs{Address: x[0], Port: x[1], Neighbors: args[2], Shared: args[3]}, nil
}

// Função para lidar com a conexão recebida
func receiver(conn net.Conn) {
	if conn == nil {
		fmt.Println("Conexão inválida")
		return
	}
	// defer(adia) a função de fechamento da conexão quando as operações terminarem
	defer conn.Close()

	for {
		// Verifica se a CLI está esperando por uma entrada
		if waiting_cli {
			// Se sim, imprime uma nova linha para manter a formatação
			fmt.Println()
		}

		// Buffer para armazenar os dados recebidos da conexão
		buf := make([]byte, 1024)
		_, err := conn.Read(buf)
		if err != nil {
			//fmt.Println("Erro ao ler dados da conexão:", err)
			break
		}
		check(err)

		// Decodifica os dados recebidos em string
		message := commands.ReceiveMessage(string(buf))

		// Verifica se a mensagem recebida é de um peer conhecido
		_, exists := knownPeers[message.Origin]

		// Mensagem para o caso do peer não ser conhecido ou não estar online
		if !exists {
			fmt.Println("\tAdicionando novo peer", message.Origin, "status", peers.ONLINE)
		} else if knownPeers[message.Origin] == peers.OFFLINE {
			fmt.Println("\tAtualizando peer", message.Origin, "status", peers.ONLINE)
		}

		// Adiciona o peer para conhecidos com status ONLINE
		knownPeers[message.Origin] = peers.ONLINE

		// Lida o comando recebido de acordo com o tipo de mensagem
		switch message.Type {
		case commands.HELLO:
		case commands.GET_PEERS:
			commands.GetPeersResponse(conn, message, knownPeers)
		case commands.PEER_LIST:
			newPeers := commands.PeerListResponse(message)
			commands.UpdatePeersMap(knownPeers, newPeers)
		case commands.BYE:
			knownPeers[message.Origin] = peers.OFFLINE
			fmt.Println("\tAtualizando peer", message.Origin, "status", peers.OFFLINE)
		}

		// Verifica se a CLI está esperando por uma entrada
		if waiting_cli {
			fmt.Println()
			fmt.Print("> ")
		}
	}
}

// Função para iniciar o peer e escutar conexões
func listener(args SelfArgs) {
	// Cria um listener TCP no endereço e porta especificado
	listener, err := net.Listen("tcp", args.FullAddress())
	check(err)
	defer listener.Close()

	// Cria uma goroutine/thread para a CLI
	go cliInterface(args)

	fmt.Println("Server running on port " + args.Port)
	// Loop para receber mensagens de outros peers
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		// Cria uma goroutine/thread para lidar com a conexão recebida
		go receiver(conn)
	}
}

// Verifica se o diretório compartilhado existe e está acessível
func verifySharedDirectory(sharedPath string) error {
	_, err := os.ReadDir(sharedPath)
	return err
}

func main() {
	var myargs SelfArgs
	var err error

	// Verifica se o programa está sendo executado em modo de teste ou não
	if len(os.Args) == 5 && os.Args[4] == "--test" {
		myargs, err = testArgs(os.Args)
		check(err)
	} else {
		// Pega os argumentos de entrada
		myargs, err = getArgs(os.Args)
		check(err)

		// Adiciona os vizinhos conhecidos pelo arquivo de vizinhos
		// err = addNeighbors(myargs.Neighbors)
		// check(err)
	}

	// Verifica o diretório compartilhado
	err = verifySharedDirectory(myargs.Shared)
	check(err)

	commands.Address = myargs.Address + ":" + myargs.Port

	// Inicializa o servidor
	listener(myargs)
}
