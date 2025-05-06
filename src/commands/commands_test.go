package commands

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"EACHare/src/logger"
	"EACHare/src/peers"
)

var senderAddress = "localhost"

func TestGetPeersRequest(t *testing.T) {
	var initialPeers peers.SafePeers
	initialPeers.Add(peers.Peer{Address: "127.0.0.1:9001", Status: peers.ONLINE, Clock: 0})
	initialPeers.Add(peers.Peer{Address: "127.0.0.2:9002", Status: peers.OFFLINE, Clock: 0})

	GetPeersRequest(&initialPeers, senderAddress)

	for _, peer := range initialPeers.GetAll() {
		if peer.Status {
			t.Errorf("Expected peer status to be false, got true")
		}
	}
}

func TestByeRequest(t *testing.T) {
	var initialPeers peers.SafePeers
	initialPeers.Add(peers.Peer{Address: "127.0.0.1:9001", Status: peers.ONLINE, Clock: 0})
	initialPeers.Add(peers.Peer{Address: "127.0.0.2:9002", Status: peers.OFFLINE, Clock: 0})

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
