package commands

type CommandType uint8

const (
	GET_PEERS CommandType = iota
)

func (ct CommandType) String() string {
	switch ct {
	case GET_PEERS:
		return "GET_PEERS"
	}

	return "UNKNOWN"
}
