package logger

// Pacote nativo de go
import (
	"strconv"
)

func MessageForwardingLog(message string, receiver string) {
	Info("\tEncaminhando mensagem \"" + message + "\" para " + receiver)
}

func UpdateClockLog(clock int) {
	Info("\t=> Atualizando relogio para " + strconv.Itoa(clock))
}

func UpdatePeerLog(address string, status string) {
	Info("\tAtualizando peer " + address + " status " + status)
}

func AddPeerLog(address string, peerStatus string) {
	Info("\tAdicionando novo peer " + address + " status " + peerStatus)
}

func ReceiveMessageLog(message string) {
	Info("\tMensagem recebida: \"" + message + "\"")
}

func ReceiveAnswerLog(message string) {
	Info("\tResposta recebida: \"" + message + "\"")
}

func ExitLog() {
	Info("\tSaindo...")
}
