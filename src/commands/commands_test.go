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

	err := sendMessage(conn, message, peers.Peer{})
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

	err := sendMessage(conn, message, peers.Peer{})
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
