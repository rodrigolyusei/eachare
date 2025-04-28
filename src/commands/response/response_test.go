package response

import (
	"sync"
	"testing"

	"EACHare/src/commands/message"
	"EACHare/src/peers"
)

func TestPeerListReceive(t *testing.T) {
	var initialPeers sync.Map
	initialPeers.Store("127.0.0.1:9001", peers.ONLINE)
	initialPeers.Store("127.0.0.1:9002", peers.OFFLINE)

	message := message.BaseMessage{
		Clock:     0,
		Type:      message.PEERS_LIST,
		Arguments: []string{"2", "127.0.0.1:9002:ONLINE:3", "127.0.0.1:9003:ONLINE:0"},
	}

	expectedPeers := []peers.Peer{
		{Address: "127.0.0.1", Port: "9001", Status: peers.ONLINE},
		{Address: "127.0.0.1", Port: "9002", Status: peers.OFFLINE},
		{Address: "127.0.0.1", Port: "9003", Status: peers.ONLINE},
	}

	PeersListResponse(message, &initialPeers)

	for _, peer := range expectedPeers {
		status, exists := initialPeers.Load(peer.FullAddress())
		if !exists || peer.Status != status {
			t.Fatalf("Expected peer %v, got %v", status, peer.Status)
		}
	}
}

func TestPeerListResponseArgumentsNil(t *testing.T) {
	var initialPeers sync.Map

	message := message.BaseMessage{
		Clock:     0,
		Type:      message.PEERS_LIST,
		Arguments: []string{"0"},
	}

	PeersListResponse(message, &initialPeers)

	peersCount := 0
	initialPeers.Range(func(_, _ any) bool {
		peersCount++
		return true
	})

	if peersCount != 0 {
		t.Fatalf("Expected 0 peers, got %d", peersCount)
	}
}
