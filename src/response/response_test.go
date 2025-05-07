package response

import (
	"bytes"
	"testing"

	"eachare/src/logger"
	"eachare/src/peers"
)

func TestGetPeersResponse(t *testing.T) {
	var initialPeers peers.SafePeers
	initialPeers.Add(peers.Peer{Address: "127.0.0.1:9001", Status: peers.ONLINE, Clock: 0})
	initialPeers.Add(peers.Peer{Address: "127.0.0.1:9002", Status: peers.ONLINE, Clock: 3})
	initialPeers.Add(peers.Peer{Address: "127.0.0.1:9003", Status: peers.OFFLINE, Clock: 3})

	var buffer bytes.Buffer
	logger.SetOutput(&buffer)

	GetPeersResponse(&initialPeers, "127.0.0.1:9001", "127.0.0.1:9002", nil)

	out := buffer.String()
	expected := `Saindo...
    => Atualizando relogio para 1
    Encaminhando mensagem "127.0.0.1:9002 1 BYE" para 127.0.0.1:9001
    Atualizando peer 127.0.0.1:9001 status OFFLINE`

	if expected != out {
		t.Errorf("\nExpected %d:\n%s\nGot %d:\n%s", len(expected), expected, len(out), out)
	}
}
