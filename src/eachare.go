package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/rodrigolyusei/EACHare/src/commands"
)

func main() {
		// Recebe os parâmetros de entrada
	addr, port, neighbors, shared := getArgs(os.Args)

	// Imprime os parâmetros de entrada
	fmt.Println("Endereço:", addr)
	fmt.Println("Porta:", port)
	fmt.Println("Vizinhos:", neighbors)
	fmt.Println("Diretório Compartilhado:", shared)
	
	// Inicializa o servidor
	comm := commands.GetCommand()
	fmt.Println("Valor escolhido: " + comm)
}

func getArgs(args []string) (string, string, string, string) {
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
	return x[0], x[1], args[2], args[3]
}
