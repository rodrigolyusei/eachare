package commands

import (
	"strings"
)

type CommandType uint8

const (
	UNKNOWN CommandType = iota
	HELLO
	GET_PEERS
	PEER_LIST
	BYE
)

func (ct CommandType) String() string {
	switch ct {
	case HELLO:
		return "HELLO"
	case GET_PEERS:
		return "GET_PEERS"
	case PEER_LIST:
		return "PEER_LIST"
	case BYE:
		return "BYE"
	}

	return "UNKNOWN"
}

func GetCommandType(s string) CommandType {
	s = strings.TrimSpace(s)       // Remove common whitespace
	s = strings.Trim(s, "\x00")    // Remove null bytes
	s = strings.Trim(s, "\r\n\t ") // Remove common control characters

	switch s {
	case "HELLO":
		return HELLO
	case "GET_PEERS":
		return GET_PEERS
	case "PEER_LIST":
		return PEER_LIST
	case "BYE":
		return BYE
	}
	return 0
}
