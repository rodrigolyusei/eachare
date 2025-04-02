package message

import (
	"strconv"
	"strings"
)

// Estrutura para armazenar as informações da mensagem
type BaseMessage struct {
	Origin    string
	Clock     int
	Type      MessageType
	Arguments []string
}

// Função para retornar a string da mensagem
func (message BaseMessage) String() string {
	// Cria a string do argumento da mensagem enviada
	arguments := ""
	if message.Arguments != nil {
		arguments = " " + strings.Join(message.Arguments, " ")
	}

	// Cria a string do tipo de mensagem
	messageStr := message.Origin + " " + strconv.Itoa(message.Clock) + " " + message.Type.String() + arguments
	return messageStr
}
