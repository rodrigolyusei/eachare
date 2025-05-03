package commands

// Pacotes nativos de go e pacotes internos
import (
	"fmt"
	"os"
	"strconv"

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
func ListPeers(knownPeers *peers.SafePeers, senderAddress string) {
	// Declara variável para o comando e inicia o loop do menu
	var comm string
	for {
		// Imprime o menu de opções
		fmt.Println("Lista de peers: ")
		fmt.Println("\t[0] voltar para o menu anterior")

		// Lista os peers e cria uma lisat dos endereços para enviar o HELLO
		var addrList []string
		for i, peer := range knownPeers.GetAll() {
			addrList = append(addrList, peer.Address)
			fmt.Println("\t[" + strconv.Itoa(i+1) + "] " + peer.Address + " " + peer.Status.String())
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
			break
		} else if number > 0 && number <= len(addrList) {
			request.HelloRequest(knownPeers, senderAddress, addrList[number-1])
			break
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
