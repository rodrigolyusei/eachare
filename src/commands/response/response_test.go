package response

import (
	"bytes"
	"sync"
	"testing"

	"EACHare/src/commands/message"
	"EACHare/src/logger"
	"EACHare/src/peers"
)

func TestGetPeersResponse(t *testing.T) {
	var initialPeers sync.Map
	initialPeers.Store("127.0.0.1:9001", peers.Peer{Status: peers.ONLINE, Clock: 0})
	initialPeers.Store("127.0.0.1:9002", peers.Peer{Status: peers.ONLINE, Clock: 3})
	initialPeers.Store("127.0.0.1:9003", peers.Peer{Status: peers.OFFLINE, Clock: 3})

	message := message.BaseMessage{
		Origin:    "127.0.0.1:9001",
		Clock:     1,
		Type:      message.GET_PEERS,
		Arguments: nil,
	}

	var buffer bytes.Buffer
	logger.SetOutput(&buffer)

	GetPeersResponse(&initialPeers, message, nil, "127.0.0.1:9002")

	out := buffer.String()
	expected := `Saindo...
    => Atualizando relogio para 1
    Encaminhando mensagem "localhost 1 BYE" para 127.0.0.1:9001
    Atualizando peer 127.0.0.1:9001 status OFFLINE`

	if expected != out {
		t.Errorf("\nExpected %d:\n%s\nGot %d:\n%s", len(expected), expected, len(out), out)
	}
}
