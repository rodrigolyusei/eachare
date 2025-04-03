package message

// Pacotes nativos de go
import (
	"strconv"
	"strings"
)

// Tipo int para o comando
type MessageType uint8

// Constantes para os tipos de comando, funcionando como um enum
const (
	UNKNOWN MessageType = iota
	HELLO
	GET_PEERS
	PEERS_LIST
	BYE
)

// Estrutura para armazenar as informações da mensagem
type BaseMessage struct {
	Origin    string
	Clock     int
	Type      MessageType
	Arguments []string
}

// Função para retornar a string do tipo de comando
func (messageType MessageType) String() string {
	switch messageType {
	case HELLO:
		return "HELLO"
	case GET_PEERS:
		return "GET_PEERS"
	case PEERS_LIST:
		return "PEERS_LIST"
	case BYE:
		return "BYE"
	default:
		return "UNKNOWN"
	}
}

// Função para obter o tipo de mensagem a partir de uma string
func GetMessageType(s string) MessageType {
	s = strings.TrimSpace(s)       // Remove espaços em branco
	s = strings.Trim(s, "\x00")    // Remove caracteres nulos
	s = strings.Trim(s, "\r\n\t ") // Remove caracteres de nova linha, tabulação e espaços

	switch s {
	case "HELLO":
		return HELLO
	case "GET_PEERS":
		return GET_PEERS
	case "PEERS_LIST":
		return PEERS_LIST
	case "BYE":
		return BYE
	default:
		return UNKNOWN
	}
}

// Função para retornar a string da mensagem
func (message BaseMessage) String() string {
	// Cria a string apenas do argumento da mensagem
	arguments := ""
	if message.Arguments != nil {
		arguments = " " + strings.Join(message.Arguments, " ")
	}

	// Cria a string da mensagem inteira e retorna
	messageStr := message.Origin + " " + strconv.Itoa(message.Clock) + " " + message.Type.String() + arguments
	return messageStr
}
