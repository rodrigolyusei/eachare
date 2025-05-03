package request

import (
	"EACHare/src/logger"
	"EACHare/src/peers"
	"bytes"
	"strings"
	"sync"
	"testing"
)

var senderAddress = "localhost"

func TestHelloRequestOffline(t *testing.T) {
	var initialPeers sync.Map
	initialPeers.Store("127.0.0.2:9002", peers.Peer{Status: peers.OFFLINE, Clock: 0})
	receiverAddress := "invalid-address:9999"

	HelloRequest(&initialPeers, senderAddress, receiverAddress)
}

func TestGetPeersRequest(t *testing.T) {
	var initialPeers sync.Map
	initialPeers.Store("127.0.0.1:9001", peers.Peer{Status: peers.ONLINE, Clock: 0})
	initialPeers.Store("127.0.0.2:9002", peers.Peer{Status: peers.OFFLINE, Clock: 0})

	GetPeersRequest(&initialPeers, senderAddress)

	initialPeers.Range(func(key, value any) bool {
		peerStatus := value.(peers.PeerStatus)
		if peerStatus {
			t.Errorf("Expected peer status to be false, got true")
		}

		return true
	})
}

func TestByeRequest(t *testing.T) {
	var initialPeers sync.Map
	initialPeers.Store("127.0.0.1:9001", peers.Peer{Status: peers.ONLINE, Clock: 0})
	initialPeers.Store("127.0.0.2:9002", peers.Peer{Status: peers.OFFLINE, Clock: 0})

	var buffer bytes.Buffer
	logger.SetOutput(&buffer)

	ByeRequest(&initialPeers, senderAddress)

	out := buffer.String()
	expected := `Saindo...
    => Atualizando relogio para 1
    Encaminhando mensagem "localhost 1 BYE" para 127.0.0.1:9001
    Atualizando peer 127.0.0.1:9001 status OFFLINE`

	if strings.TrimSpace(expected) != strings.TrimSpace(out) {
		t.Errorf("\nExpected %d:\n%s\nGot %d:\n%s", len(expected), expected, len(out), out)
	}
}
