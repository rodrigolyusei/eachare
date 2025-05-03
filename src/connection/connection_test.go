package connection

import (
	"EACHare/src/commands/message"
	"EACHare/src/peers"
	"sync"
	"testing"
)

func TestSendMessageArgumentsNilOK(t *testing.T) {
	conn := &mockConn{}
	message := message.BaseMessage{
		Origin:    "localhost",
		Clock:     0,
		Type:      message.UNKNOWN,
		Arguments: nil,
	}
	var knownPeers sync.Map
	knownPeers.Store("127.0.0.1:9001", peers.Peer{Status: peers.ONLINE, Clock: 0})

	SendMessage(conn, message, "127.0.0.1:9001", &knownPeers)

	if string(conn.data) != "localhost 1 UNKNOWN\n" {
		t.Fatalf("Expected %s, got %s", "localhost 1 UNKNOWN\n", string(conn.data))
	}
}

func TestSendMessageConnNil(t *testing.T) {
	message := message.BaseMessage{
		Origin:    "localhost",
		Clock:     0,
		Type:      message.UNKNOWN,
		Arguments: nil,
	}
	var knownPeers sync.Map
	knownPeers.Store("127.0.0.1:9001", peers.Peer{Status: peers.ONLINE, Clock: 0})

	SendMessage(nil, message, "127.0.0.1:9001", &knownPeers)

	neighbor, _ := knownPeers.Load("127.0.0.1:9001")
	neighborStatus := neighbor.(peers.Peer).Status
	if neighborStatus != peers.OFFLINE {
		t.Fatalf("Expected peer status to be OFFLINE, got %s", neighborStatus.String())
	}
}
