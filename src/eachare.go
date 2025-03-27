package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"EACHare/src/commands"
	"EACHare/src/number"
	"EACHare/src/peers"
)

type SelfArgs struct {
	Address   string
	Port      string
	Neighbors string
	Shared    string
}

var knowPeers = make(map[string]peers.PeerStatus)

func (args SelfArgs) FullAddress() string {
	return args.Address + ":" + args.Port
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func listen(args SelfArgs) {
	ln, err := net.Listen("tcp", args.FullAddress())
	check(err)
	go cliInterface(args)

	fmt.Println("Server running on port " + args.Port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	check(err)

	message := commands.ReceiveMessage(string(buf))

	_, exists := knowPeers[message.Origin]

	if (exists && knowPeers[message.Origin] == peers.OFFLINE) || knowPeers[message.Origin] == peers.OFFLINE {
		knowPeers[message.Origin] = peers.ONLINE
		fmt.Println("\tAtualizando peer " + message.Origin + " status ONLINE")
	}

	switch message.Type {
	case commands.GET_PEERS:
		commands.GetPeersResponse(conn, message, knowPeers)
	case commands.PEER_LIST:
		newPeers := commands.PeerListReceive(message)
		commands.UpdatePeersMap(knowPeers, newPeers)
	}
}

func cliInterface(args SelfArgs) {
	for {
		comm := commands.GetCommands()
		if comm == "1" {
			commands.ListPeers(knowPeers)
		} else if comm == "2" {
			// var input string
			// fmt.Scanln(&input)

			// number, err := strconv.Atoi(input)
			// if err != nil {
			// 	fmt.Println("Error casting port to int! Did you write a number?")
			// 	continue
			// }
			// go client(number)
			commands.GetPeersRequest(knowPeers)
		} else if comm == "3" {
			shared := commands.GetSharedDirectory(args.Shared)
			fmt.Println(shared)
		} else if comm == "9" {
			fmt.Println("Saindo...")
			return
		}
		fmt.Println()
	}
}

func client(port int) {
	conn, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		fmt.Println("Inválid port!")
		return
	}
	defer conn.Close()
	_, err = conn.Write([]byte("Hello, friend! From port " + strconv.Itoa(port) + "\n"))
	if err != nil {
		fmt.Println(err)
		return
	}
}

func getArgs(args []string) SelfArgs {
	if len(args) != 4 {
		fmt.Println("Parâmetros de entrada inválidos, por favor, siga o formato abaixo:")
		fmt.Println("./eachare <endereço>:<porta> <vizinhos> <diretório compartilhado>")
		os.Exit(1)
	} else if !strings.Contains(args[1], ":") {
		fmt.Println("Endereço e porta inválidos, por favor, siga o formato abaixo:")
		fmt.Println("./eachare <endereço>:<porta> <vizinhos> <diretório compartilhado>")
		os.Exit(1)
	}

	x := strings.Split(args[1], ":")
	return SelfArgs{Address: x[0], Port: x[1], Neighbors: args[2], Shared: args[3]}
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
