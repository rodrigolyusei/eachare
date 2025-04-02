package request

import (
	"EACHare/src/clock"
	"EACHare/src/commands/message"
	"EACHare/src/peers"
	"testing"
)

func TestSendMessageArgumentsNil(t *testing.T) {
	clock.ResetClock()
	conn := &mockConn{}
	message := message.BaseMessage{
		Clock:     0,
		Type:      message.UNKNOWN,
		Arguments: nil,
	}

	err := sendMessage(conn, message, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := "localhost 1 UNKNOWN"
	if string(conn.data) != expected {
		t.Fatalf("Expected %s, got %s", expected, string(conn.data))
	}
}

func TestGetPeersRequest(t *testing.T) {
	// Mock peers
	knowPeers := make(map[string]peers.PeerStatus)
	knowPeers["127.0.0.1:8080"] = peers.ONLINE
	knowPeers["127.0.0.2:8081"] = peers.OFFLINE

	GetPeersRequest(knowPeers)

	for _, peerStatus := range knowPeers {
		if peerStatus {
			t.Errorf("Expected peer status to be false, got true")
		}
	}
}
