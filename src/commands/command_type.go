package commands

type CommandType uint8

const (
	UNKNOWN CommandType = iota
	GET_PEERS
	PEER_LIST
)

func (ct CommandType) String() string {
	switch ct {
	case GET_PEERS:
		return "GET_PEERS"
	}

	return "UNKNOWN"
}

func GetCommandType(s string) CommandType {
	switch s {
	case "GET_PEERS":
		return GET_PEERS
	}
	return 0
}
