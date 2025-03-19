package main

import (
    "fmt"
    "os"
	"strings"
)

func main() {
    x := strings.Split(os.Args[1], ":")

	addr := x[0]
	port := x[1]
	neighbors := os.Args[2]
	shared := os.Args[3]

	fmt.Println("Endereço:", addr)
	fmt.Println("Porta:", port)
	fmt.Println("Vizinhos:", neighbors)
	fmt.Println("Diretório Compartilhado:", shared)
}