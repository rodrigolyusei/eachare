package commands

// Pacotes nativos de go e pacotes internos
import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"EACHare/src/commands/request"
	"EACHare/src/peers"
)

// Função para verificar e imprimir mensagem de erro
func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Função para listar os peers conhecidos e enviar HELLO para o peer escolhido
func ListPeers(knownPeers *sync.Map, requestClient request.IRequest) {
	// Conta os peers conhecidos e armazena os endereços
	var addrList []string
	var statusList []peers.PeerStatus
	knownPeers.Range(func(key, value any) bool {
		addr := key.(string)
		neighbor := value.(peers.Peer)
		neighborStatus := neighbor.Status

		addrList = append(addrList, addr)
		statusList = append(statusList, neighborStatus)
		return true
	})

	var comm string
	var exit bool = false
	for !exit {
		// Imprime o menu de opções
		fmt.Println("Lista de peers: ")
		fmt.Println("\t[0] voltar para o menu anterior")
		for i, addr := range addrList {
			fmt.Println("\t[" + strconv.Itoa(i+1) + "] " + addr + " " + statusList[i].String())
		}

		// Lê a entrada do usuário
		fmt.Print("> ")
		fmt.Scanln(&comm)
		fmt.Println()

		// Converte a entrada para inteiro
		number, err := strconv.Atoi(comm)
		if err != nil {
			fmt.Println("Comando inválido, tente novamente.")
			continue
		}

		// Envio de mensagem para o destino escolhido
		if number == 0 {
			exit = true
		} else if number > 0 && number <= len(addrList) {
			requestClient.HelloRequest(addrList[number-1], knownPeers)
			exit = true
		} else {
			fmt.Println("Opção inválida, tente novamente.")
		}
	}
}

// Função para listar os arquivos do diretório compartilhado
func ListLocalFiles(sharedPath string) {
	// Lê o diretório e imprime os arquivos
	entries, err := os.ReadDir(sharedPath)
	check(err)
	for _, entry := range entries {
		fmt.Println("\t" + entry.Name())
	}
}
