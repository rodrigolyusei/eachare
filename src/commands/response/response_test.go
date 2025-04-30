package response

import (
	"sync"
	"testing"

	"EACHare/src/commands/message"
	"EACHare/src/peers"
)

func TestPeerListReceive(t *testing.T) {
	var initialPeers sync.Map
	initialPeers.Store("127.0.0.1:9001", peers.Peer{Status: peers.ONLINE, Clock: 0})
	initialPeers.Store("127.0.0.1:9002", peers.Peer{Status: peers.OFFLINE, Clock: 0})

	message := message.BaseMessage{
		Clock:     1,
		Type:      message.PEERS_LIST,
		Arguments: []string{"2", "127.0.0.1:9002:ONLINE:3", "127.0.0.1:9003:ONLINE:0"},
	}

	var expectedPeers sync.Map
	expectedPeers.Store("127.0.0.1:9001", peers.Peer{Status: peers.ONLINE, Clock: 0})
	expectedPeers.Store("127.0.0.1:9002", peers.Peer{Status: peers.OFFLINE, Clock: 0})
	expectedPeers.Store("127.0.0.1:9003", peers.Peer{Status: peers.ONLINE, Clock: 0})

	PeersListResponse(message, &initialPeers)

	expectedPeers.Range(func(key, value any) bool {
		peerAddress := key.(string)
		peerStatus := value.(peers.Peer).Status
		peerClock := value.(peers.Peer).Clock

		peer, exists := initialPeers.Load(peerAddress)
		if !exists {
			t.Fatalf("Expected peer %s not found", peerAddress)
		}
		if peerStatus != peer.(peers.Peer).Status {
			t.Fatalf("Expected peer %s status %v, got %v", peerAddress, peerStatus, peer.(peers.Peer).Status)
		}
		if peerClock != peer.(peers.Peer).Clock {
			t.Fatalf("Expected peer %s clock %d, got %d", peerAddress, peerClock, peer.(peers.Peer).Clock)
		}
		return true
	})
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
