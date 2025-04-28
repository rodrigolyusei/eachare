package commands

// Pacotes nativos de go e pacotes internos
import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"EACHare/src/commands/request"
	"EACHare/src/logger"
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
	var counter int = 0
	knownPeers.Range(func(key, value any) bool {
		addressPort := key.(string)
		addrList = append(addrList, addressPort)
		counter++
		return true
	})

	var comm string
	var exit bool = false
	for !exit {
		// Imprime o menu de opções
		fmt.Println("Lista de peers: ")
		fmt.Println("\t[0] voltar para o menu anterior")
		for counter, addr := range addrList {
			fmt.Println("\t[" + strconv.Itoa(counter+1) + "] " + addr + " " + peers.GetStatus(addr).String())
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
		} else if number > 0 && number <= counter {
			// Envia mensagem HELLO e atualiza o status do peer
			peerStatus := requestClient.HelloRequest(addrList[number-1])
			if value, _ := knownPeers.Load(addrList[number-1]); value != peerStatus {
				logger.Info("\tAtualizando peer " + addrList[number-1] + " status " + peerStatus.String())
			}
			knownPeers.Store(addrList[number-1], peerStatus)
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
