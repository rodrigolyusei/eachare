package commands

import (
	"strings"
)

type CommandType uint8

const (
	UNKNOWN CommandType = iota
	HELLO_PEER
	GET_PEERS
	PEER_LIST
)

func (ct CommandType) String() string {
	switch ct {
	case HELLO_PEER:
		return "HELLO_PEER"
	case GET_PEERS:
		return "GET_PEERS"
	case PEER_LIST:
		return "PEER_LIST"
	}

	return "UNKNOWN"
}

func GetCommandType(s string) CommandType {
	s = strings.TrimSpace(s)       // Remove common whitespace
	s = strings.Trim(s, "\x00")    // Remove null bytes
	s = strings.Trim(s, "\r\n\t ") // Remove common control characters

	switch s {
	case "HELLO_PEER":
		return HELLO_PEER
	case "GET_PEERS":
		return GET_PEERS
	case "PEER_LIST":
		return PEER_LIST
	}
	return 0
}
