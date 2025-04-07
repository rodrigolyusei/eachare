package commands

import (
	"bytes"
	"os"
	"strings"
	"sync"
	"testing"

	"EACHare/src/commands/message"
	"EACHare/src/peers"
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
	defer teardownTestDir(sharedPath)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	ListLocalFiles(sharedPath)

	w.Close()
	os.Stdout = oldStdout
	var buffer bytes.Buffer
	buffer.ReadFrom(r)
	out := buffer.String()

	expected := `	ipsum.txt
	loren.txt
	`
	if strings.TrimSpace(expected) != strings.TrimSpace(out) {
		t.Errorf("Expected %d \n%s, got %d \n%s", len(expected), expected, len(out), out)

	}
}

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
	initialPeers.Range(func(_, _ interface{}) bool {
		peersCount++
		return true
	})

	if peersCount != 0 {
		t.Fatalf("Expected 0 peers, got %d", peersCount)
	}
}
