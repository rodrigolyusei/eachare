package connection

import (
	"EACHare/src/message"
	"EACHare/src/peers"
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
	var knownPeers peers.SafePeers
	knownPeers.Add(peers.Peer{Address: "127.0.0.1:9001", Status: peers.ONLINE, Clock: 0})

	SendMessage(&knownPeers, conn, message, "127.0.0.1:9001")

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
	var knownPeers peers.SafePeers
	knownPeers.Add(peers.Peer{Address: "127.0.0.1:9001", Status: peers.ONLINE, Clock: 0})

	SendMessage(&knownPeers, nil, message, "127.0.0.1:9001")

	neighbor, _ := knownPeers.Get("127.0.0.1:9001")
	if neighbor.Status != peers.OFFLINE {
		t.Fatalf("Expected peer status to be OFFLINE, got %s", neighbor.Status.String())
	}
}
