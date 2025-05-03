package clock

// Pacotes nativos de go e pacote interno
import (
	"strconv"
	"sync"

	"EACHare/src/logger"
)

// Estrutura com o valor do relógio e um mutex para controle de concorrência
type SafeClock struct {
	mutex sync.Mutex
	clock int
}

// Instância do relógio com mutex
var safeClock = SafeClock{clock: 0}

// Função para incrementar o relógio e imprimir mensagem de atualização
func UpdateClock() int {
	// Bloqueia o mutex para garantir acesso exclusivo ao relógio
	safeClock.mutex.Lock()
	defer safeClock.mutex.Unlock()

	// Incrementa o relógio e imprime a mensagem de atualização
	safeClock.clock++
	logger.Info("\t=> Atualizando relogio para " + strconv.Itoa(safeClock.clock))
	return safeClock.clock
}

// Função para atualizar o relógio entre o valor local e o recebido
func UpdateMaxClock(clockRecebido int) int {
	// Bloqueia o mutex para garantir acesso exclusivo ao relógio
	safeClock.mutex.Lock()
	defer safeClock.mutex.Unlock()

	// Incrementa o relógio e imprime a mensagem de atualização
	safeClock.clock = max(safeClock.clock, clockRecebido)
	safeClock.clock++
	logger.Info("\t=> Atualizando relogio para " + strconv.Itoa(safeClock.clock))
	return safeClock.clock
}

// Função para obter o valor atual do relógio
func GetClock() int {
	// Bloqueia o mutex para garantir acesso exclusivo ao relógio
	safeClock.mutex.Lock()
	defer safeClock.mutex.Unlock()

	// Retorna o valor atual do relógio
	return safeClock.clock
}
