package clock

// Pacotes nativos de go e pacote interno
import (
	"strconv"
	"sync"

	"EACHare/src/logger"
)

// Estrutura com o valor do relógio e um mutex para controle de concorrência
type SafeClock struct {
	Clock int
	Mutex sync.Mutex
}

// Instância do relógio com mutex
var safeClock = SafeClock{Clock: 0}

// Função para incrementar o relógio e imprimir mensagem de atualização
func UpdateClock() int {
	// Bloqueia o mutex para garantir acesso exclusivo ao relógio
	safeClock.Mutex.Lock()
	defer safeClock.Mutex.Unlock()

	// Incrementa o relógio e imprime a mensagem de atualização
	safeClock.Clock++
	logger.Info("\t=> Atualizando relogio para " + strconv.Itoa(safeClock.Clock))
	return safeClock.Clock
}

// Função para reiniciar o relógio
func ResetClock() {
	safeClock.Clock = 0
}

// Função para obter o valor atual do relógio
func GetClock() int {
	return safeClock.Clock
}
