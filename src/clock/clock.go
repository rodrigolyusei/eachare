package clock

// Pacote nativo de go
import "fmt"

// Variável para relógio começando com 0
var clock int = 0

// Função para incrementar o relógio e imprimir mensagem de atualização
func UpdateClock() int {
	clock++
	fmt.Println("\t=> Atualizando relogio para ", clock)
	return clock
}

// Função para reiniciar o relógio
func ResetClock() {
	clock = 0
}

// Função para obter o valor atual do relógio
func GetClock() int {
	return clock
}
