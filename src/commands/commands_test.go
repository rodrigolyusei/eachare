package commands

import (
	"EACHare/src/clock"
	"EACHare/src/peers"
	"os"
	"testing"
)

// https://gobyexample.com/writing-files
func setupTestDir(path string, files []string) {
	os.Mkdir(path, 0755)
	for _, file := range files {
		os.WriteFile(path+"/"+file, []byte("test content"), 0644)
	}
}

func teardownTestDir(path string) {
	os.RemoveAll(path)
}

func TestGetSharedDirectory(t *testing.T) {
	sharedPath := "../shared"
	setupTestDir(sharedPath, []string{"loren.txt", "ipsum.txt"})
	// this scheadules the teardown function to run after the test finishes
	defer teardownTestDir(sharedPath)

	entries := GetSharedDirectory(sharedPath)
	if len(entries) != 2 {
		t.Errorf("Expected two entries, got %d", len(entries))
	}
	if entries[0].Name() != "ipsum.txt" {
		t.Errorf("Expected first entry to be 'loren.txt', got %s", entries[0].Name())
	}
}

func TestSendMessageWithArguments(t *testing.T) {
	clock.ResetClock()
	conn := &mockConn{}
	message := BaseMessage{
		Clock:     0,
		Type:      UNKNOWN,
		Arguments: []string{"arg1", "arg2"},
	}

	err := sendMessage(conn, message, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := "localhost 1 UNKNOWN arg1 arg2"
	if string(conn.data) != expected {
		t.Fatalf("Expected %s, got %s", expected, string(conn.data))
	}
}

func TestSendMessageArgumentsNil(t *testing.T) {
	clock.ResetClock()
	conn := &mockConn{}
	message := BaseMessage{
		Clock:     0,
		Type:      UNKNOWN,
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

func TestPeerListReceive(t *testing.T) {
	message := BaseMessage{
		Clock:     0,
		Type:      PEER_LIST,
		Arguments: []string{"2", "127.0.0.1:9002:ONLINE:3", "127.0.0.1:9004:ONLINE:0"},
	}

	expected := []string{"127.0.0.1:9002", "127.0.0.1:9004"}

	receivePeers := PeerListReceive(message)

	for i, peer := range receivePeers {
		if peer.FullAddress() != expected[i] {
			t.Fatalf("Expected %s, got %s", message.Arguments[i], peer.FullAddress())
		}
		if peer.Status != peers.ONLINE {
			t.Fatalf("Expected ONLINE, got %s", peer.Status)
		}
	}
}

func TestPeerListReceiveOffline(t *testing.T) {
	message := BaseMessage{
		Clock:     0,
		Type:      PEER_LIST,
		Arguments: []string{"1", "127.0.0.1:9002:OFFLINE:3"},
	}

	expected := "127.0.0.1:9002"

	receivePeers := PeerListReceive(message)

	for i, peer := range receivePeers {
		if peer.FullAddress() != expected {
			t.Fatalf("Expected %s, got %s", message.Arguments[i], peer.FullAddress())
		}
		if peer.Status != peers.OFFLINE {
			t.Fatalf("Expected OFFLINE, got %s", peer.Status)
		}
	}
}

func TestPeerListReceiveArgumentsNil(t *testing.T) {
	message := BaseMessage{
		Clock:     0,
		Type:      PEER_LIST,
		Arguments: []string{"0"},
	}

	peers := PeerListReceive(message)

	if len(peers) != 0 {
		t.Fatalf("Expected 0 peers, got %d", len(peers))
	}
}

func TestUpdatePeersList(t *testing.T) {
	initialPeers := []peers.Peer{
		{Address: "127.0.0.1:9001", Status: peers.ONLINE},
		{Address: "127.0.0.1:9002", Status: peers.OFFLINE},
	}

	newPeers := []peers.Peer{
		{Address: "127.0.0.1:9001", Status: peers.OFFLINE},
		{Address: "127.0.0.1:9003", Status: peers.ONLINE},
	}

	expectedPeers := []peers.Peer{
		{Address: "127.0.0.1:9001", Status: peers.OFFLINE},
		{Address: "127.0.0.1:9002", Status: peers.OFFLINE},
		{Address: "127.0.0.1:9003", Status: peers.ONLINE},
	}

	updatedPeers := UpdatePeersList(initialPeers, newPeers)

	if len(updatedPeers) != len(expectedPeers) {
		t.Fatalf("Expected %d peers, got %d", len(expectedPeers), len(updatedPeers))
	}

	for i, peer := range updatedPeers {
		if peer.Address != expectedPeers[i].Address || peer.Status != expectedPeers[i].Status {
			t.Fatalf("Expected peer %v, got %v", expectedPeers[i], peer)
		}
	}
}

func TestGetPeersResponse(t *testing.T) {
	clock.ResetClock()
	conn := &mockConn{}

	// Mock received message
	receivedMessage := BaseMessage{
		Origin:    "127.0.0.2:9001",
		Clock:     1,
		Type:      GET_PEERS,
		Arguments: []string{},
	}

	// Mock known peers
	knowPeers := []peers.Peer{
		{Address: "192.168.1.1", Port: "8080", Status: peers.ONLINE},
		{Address: "192.168.1.2", Port: "8081", Status: peers.OFFLINE},
	}

	expected := "localhost 1 PEER_LIST 2 192.168.1.1:8080:ONLINE:0 192.168.1.2:8081:OFFLINE:0"

	// Call the function
	GetPeersResponse(conn, receivedMessage, knowPeers)

	if clock.GetClock() != 1 {
		t.Fatalf("Expected clock to be 1, got %d", clock.GetClock())
	}

	if string(conn.data) != expected {
		t.Fatalf("Expected %s, got %s", expected, conn.data)
	}

}

func TestGetPeersRequest(t *testing.T) {
	// Mock peers
	knowPeers := []peers.Peer{
		{Address: "127.0.0.1", Port: "8080", Status: true},
		{Address: "127.0.0.2", Port: "8081", Status: true},
	}

	GetPeersRequest(knowPeers)

	for _, peer := range knowPeers {
		if peer.Status {
			t.Errorf("Expected peer status to be false, got true")
		}
	}
}
