package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	// Verifica se a quantidade de parâmetros, endereço e porta de entrada é válida
	if len(os.Args) < 4 || len(os.Args) > 4 {
		fmt.Println("Parâmetros de entrada inválidos, por favor, siga o formato abaixo:")
		fmt.Println("./eachare <endereço>:<porta> <vizinhos> <diretório compartilhado>")
		os.Exit(1)
	} else if !strings.Contains(os.Args[1], ":") {
		fmt.Println("Endereço e porta inválidos, por favor, siga o formato abaixo:")
		fmt.Println("./eachare <endereço>:<porta> <vizinhos> <diretório compartilhado>")
		os.Exit(1)
	}

	// Recebe os parâmetros de entrada
	addr, port, neighbors, shared := getArgs(os.Args)

	// Imprime os parâmetros de entrada
	fmt.Println("Endereço:", addr)
	fmt.Println("Porta:", port)
	fmt.Println("Vizinhos:", neighbors)
	fmt.Println("Diretório Compartilhado:", shared)
}

func getArgs(args []string) (string, string, string, string) {
	x := strings.Split(args[1], ":")
	return x[0], x[1], args[2], args[3]
}
