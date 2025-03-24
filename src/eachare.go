package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	addr, port, neighbors, shared := getArgs(os.Args)

	fmt.Println("Endereço:", addr)
	fmt.Println("Porta:", port)
	fmt.Println("Vizinhos:", neighbors)
	fmt.Println("Diretório Compartilhado:", shared)
}

func getArgs(args []string) (string, string, string, string) {
	x := strings.Split(args[1], ":")
	return x[0], x[1], os.Args[2], os.Args[3]
}
