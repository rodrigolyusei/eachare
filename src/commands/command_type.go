package commands

// Pacotes nativos de go
import (
	"strings"
)

// Tipo int para o comando
type CommandType uint8

// Constantes para os tipos de comando, funcionando como um enum
const (
	UNKNOWN CommandType = iota
	HELLO
	GET_PEERS
	PEERS_LIST
	BYE
)

// Função para converter o tipo de comando em string
func (ct CommandType) String() string {
	switch ct {
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

// Função para obter o tipo de comando a partir de uma string
func GetCommandType(s string) CommandType {
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
