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
)

type SelfArgs struct {
	Address   string
	Port      string
	Neighbors string
	Shared    string
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func listen(args SelfArgs) {
	port, err := number.GetNextPort()
	check(err)

	ln, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
	check(err)

	fmt.Println("Server running on port " + strconv.Itoa(port))
	for {
		go cliInterface()
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

	fmt.Printf("Received: %s", buf)
}

func cliInterface() {
	for {
		comm := commands.GetCommand()
		if comm == "2" {
			var input string
			fmt.Scanln(&input)

			number, err := strconv.Atoi(input)
			if err != nil {
				fmt.Println("Error casting port to int! Did you write a number?")
				continue
			}
			go client(number)
		}
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
	all_args := getArgs(os.Args)

	// Imprime os parâmetros de entrada
	fmt.Println("Endereço:", all_args.Address)
	fmt.Println("Porta:", all_args.Port)
	fmt.Println("Vizinhos:", all_args.Neighbors)
	fmt.Println("Diretório Compartilhado:", all_args.Shared)

	// Inicializa o servidor
	// comm := commands.GetCommand()
	// fmt.Println("Valor escolhido: " + comm)
	listen(all_args)
}
