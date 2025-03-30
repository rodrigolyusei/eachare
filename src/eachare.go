package main

// Pacotes nativos de go e pacotes internos
import (
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
var knowPeers = make(map[string]peers.PeerStatus)

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
			commands.ListPeers(knowPeers)
		case "2":
			commands.GetPeersRequest(knowPeers)
		case "3":
			shared := commands.GetSharedDirectory(args.Shared)
			fmt.Println(shared)
		case "4":
			fmt.Println("Comando ainda não implementado")
		case "5":
			fmt.Println("Comando ainda não implementado")
		case "6":
			fmt.Println("Comando ainda não implementado")
		case "9":
			fmt.Println("Saindo...")
			return
		default:
			fmt.Println("Comando inválido, tente novamente.")
		}

		// Indica que a CLI não está mais esperando por uma entrada
		waiting_cli = false
		time.Sleep(500 * time.Millisecond)
	}
}

// Função para obter os argumentos de entrada
func getArgs(args []string) SelfArgs {
	// Verifica se o número de argumentos é válido e se o formato do endereço e porta está correto
	if len(args) != 4 {
		fmt.Println("Parâmetros de entrada inválidos, por favor, siga o formato abaixo:")
		fmt.Println("./eachare <endereço>:<porta> <vizinhos> <diretório compartilhado>")
		os.Exit(1)
	} else if !strings.Contains(args[1], ":") {
		fmt.Println("Endereço e porta inválidos, por favor, siga o formato abaixo:")
		fmt.Println("./eachare <endereço>:<porta> <vizinhos> <diretório compartilhado>")
		os.Exit(1)
	}

	// Se os parâmetros estiverem corretos, retorna cada uma separadamente
	x := strings.Split(args[1], ":")
	return SelfArgs{Address: x[0], Port: x[1], Neighbors: args[2], Shared: args[3]}
}

// Função para lidar com a conexão recebida
func handleConnection(conn net.Conn) {
	// defer(adia) a função de fechamento da conexão quando as operações terminarem
	defer conn.Close()

	// Verifica se a CLI está esperando por uma entrada
	if waiting_cli {
		// Se sim, imprime uma nova linha para manter a formatação
		fmt.Println()
	}

	// Buffer para armazenar os dados recebidos da conexão
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	check(err)

	// Decodifica os dados recebidos em string
	message := commands.ReceiveMessage(string(buf))

	// Verifica se a mensagem recebida é de um peer conhecido
	_, exists := knowPeers[message.Origin]

	// Se o peer não for conhecido, adiciona-o ao mapa de peers conhecidos
	// Se o peer for conhecido e estiver OFFLINE, atualiza seu status para ONLINE
	if !exists {
		fmt.Println("\tAdicionando novo peer ", message.Origin, "status", peers.ONLINE)
	} else if knowPeers[message.Origin] == peers.OFFLINE {
		knowPeers[message.Origin] = peers.ONLINE
		fmt.Println("\tAtualizando peer "+message.Origin+" status ", peers.ONLINE)
	}

	// Lida o comando recebido de acordo com o tipo de mensagem
	switch message.Type {
	case commands.HELLO:
	case commands.GET_PEERS:
		commands.GetPeersResponse(conn, message, knowPeers)
	case commands.PEER_LIST:
		newPeers := commands.PeerListReceive(message)
		commands.UpdatePeersMap(knowPeers, newPeers)
	}

	// Verifica se a CLI está esperando por uma entrada
	if waiting_cli {
		fmt.Println()
		fmt.Print("> ")
	}
}

// Função para iniciar o peer e escutar conexões
func listen(args SelfArgs) {
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
		go handleConnection(conn)
	}
}

func main() {
	// Pega os argumentos de entrada
	all_args := getArgs(os.Args)

	var test = true
	if test {
		port, err := number.GetNextPort()
		check(err)
		all_args.Address = "127.0.0.1"
		all_args.Port = strconv.Itoa(port)
	}
	port, err := strconv.Atoi(all_args.Port)
	check(err)
	if port%2 == 0 {
		knowPeers["127.0.0.1:"+strconv.Itoa(port+1)] = peers.ONLINE
		knowPeers["127.0.0.1:"+strconv.Itoa(port+2)] = peers.OFFLINE
	} else {
		knowPeers["127.0.0.1:"+strconv.Itoa(port+1)] = peers.ONLINE
		knowPeers["127.0.0.1:"+strconv.Itoa(port+3)] = peers.OFFLINE
	}

	commands.Address = all_args.Address + ":" + all_args.Port
	// Imprime os parâmetros de entrada
	fmt.Println("Endereço:", all_args.Address)
	fmt.Println("Porta:", all_args.Port)
	fmt.Println("Vizinhos:", all_args.Neighbors)
	fmt.Println("Diretório Compartilhado:", all_args.Shared)

	// Inicializa o servidor
	listen(all_args)
}
