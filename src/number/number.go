package number

// Pacotes nativos de go
import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Função para obter o próximo número de porta
func GetNextPort() (int, error) {
	// Atualiza e obtém o número do arquivo
	number, err := UpdateAndGetNumberFromFile()
	if err != nil {
		return 0, err
	}

	// Concatena com 100 para selecionar a porta não privilegiada
	port, err := strconv.Atoi("100" + strconv.Itoa(number))
	if err != nil {
		return 0, err
	}
	return port, nil
}

// Função para atualizar o número no arquivo e retornar o número atualizado
func UpdateAndGetNumberFromFile() (int, error) {
	// Define o caminho do arquivo
	filepath := filepath.Join("number", "data.txt")

	// Atualiza antes de ler o número
	err := updateNumberInFile(filepath)
	if err != nil {
		return 0, err
	}

	// Lê o número do arquivo e retorna se não tiver erro
	number, err := getNumberFromFile(filepath)
	if err != nil {
		return 0, err
	}

	return number, nil
}

// Função para atualizar o número no arquivo
func updateNumberInFile(filename string) error {
	// Lê o arquivo
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Converte o dado para um inteiro
	number, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return err
	}

	// Incrementa e escreve o número no arquivo
	number++
	return os.WriteFile(filename, []byte(strconv.Itoa(number)), 0644)
}

// Função para obter o número do arquivo
func getNumberFromFile(filename string) (int, error) {
	// Lê o arquivo
	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, err
	}

	// Converte o dado para um inteiro
	number, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, err
	}

	// Calcula o módulo para obter valor entre 0 e 90
	number = number % 90

	// O valor retornado será entre 10 e 100
	return number + 10, nil
}
