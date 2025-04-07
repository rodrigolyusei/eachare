package commands

// import (
// 	"EACHare/src/commands/message"
// 	"EACHare/src/peers"
// 	"os"
// 	"testing"
// )

// // https://gobyexample.com/writing-files
// func setupTestDir(path string, files []string) {
// 	os.Mkdir(path, 0755)
// 	for _, file := range files {
// 		os.WriteFile(path+"/"+file, []byte("test content"), 0644)
// 	}
// }

// func teardownTestDir(path string) {
// 	os.RemoveAll(path)
// }

// // func TestGetSharedDirectory(t *testing.T) {
// // 	sharedPath := "../shared"
// // 	setupTestDir(sharedPath, []string{"loren.txt", "ipsum.txt"})
// // 	// this scheadules the teardown function to run after the test finishes
// // 	defer teardownTestDir(sharedPath)

// // 	entries := GetSharedDirectory(sharedPath)
// // 	if len(entries) != 2 {
// // 		t.Errorf("Expected two entries, got %d", len(entries))
// // 	}
// // 	if entries[0].Name() != "ipsum.txt" {
// // 		t.Errorf("Expected first entry to be 'loren.txt', got %s", entries[0].Name())
// // 	}
// // }

// // func TestSendMessageWithArguments(t *testing.T) {
// // 	clock.ResetClock()
// // 	conn := &mockConn{}
// // 	message := BaseMessage{
// // 		Clock:     0,
// // 		Type:      UNKNOWN,
// // 		Arguments: []string{"arg1", "arg2"},
// // 	}

// // 	err := sendMessage(conn, message, "")
// // 	if err != nil {
// // 		t.Fatalf("Expected no error, got %v", err)
// // 	}

// // 	expected := "localhost 1 UNKNOWN arg1 arg2"
// // 	if string(conn.data) != expected {
// // 		t.Fatalf("Expected %s, got %s", expected, string(conn.data))
// // 	}
// // }

// func TestPeerListReceive(t *testing.T) {
// 	initialPeers := make(map[string]peers.PeerStatus)
// 	initialPeers["127.0.0.1:9001"] = peers.ONLINE
// 	initialPeers["127.0.0.1:9002"] = peers.OFFLINE

// 	message := message.BaseMessage{
// 		Clock:     0,
// 		Type:      message.PEERS_LIST,
// 		Arguments: []string{"2", "127.0.0.1:9002:ONLINE:3", "127.0.0.1:9003:ONLINE:0"},
// 	}

// 	expectedPeers := []peers.Peer{
// 		{Address: "127.0.0.1", Port: "9001", Status: peers.ONLINE},
// 		{Address: "127.0.0.1", Port: "9002", Status: peers.OFFLINE},
// 		{Address: "127.0.0.1", Port: "9003", Status: peers.ONLINE},
// 	}

// 	PeersListResponse(message, initialPeers)

// 	for _, peer := range expectedPeers {
// 		_, exists := initialPeers[peer.FullAddress()]
// 		if !exists || peer.Status != initialPeers[peer.FullAddress()] {
// 			t.Fatalf("Expected peer %v, got %v", initialPeers[peer.FullAddress()], peer.Status)
// 		}
// 	}
// }

// func TestPeerListResponseArgumentsNil(t *testing.T) {
// 	initialPeers := make(map[string]peers.PeerStatus)

// 	message := message.BaseMessage{
// 		Clock:     0,
// 		Type:      message.PEERS_LIST,
// 		Arguments: []string{"0"},
// 	}

// 	PeersListResponse(message, initialPeers)

// 	if len(initialPeers) != 0 {
// 		t.Fatalf("Expected 0 peers, got %d", len(initialPeers))
// 	}
// }

// // func TestGetPeersResponse(t *testing.T) {
// // 	clock.ResetClock()
// // 	conn := &mockConn{}

// // 	// Mock received message
// // 	receivedMessage := BaseMessage{
// // 		Origin:    "127.0.0.2:9001",
// // 		Clock:     1,
// // 		Type:      GET_PEERS,
// // 		Arguments: []string{},
// // 	}

// // 	// Mock known peers
// // 	knowPeers := make(map[string]peers.PeerStatus)
// // 	knowPeers["127.0.0.1:8080"] = peers.ONLINE
// // 	knowPeers["127.0.0.2:8081"] = peers.OFFLINE

// // 	expected := "localhost 1 PEERS_LIST 2 127.0.0.1:8080:ONLINE:0 127.0.0.2:8081:OFFLINE:0"

// // 	// Call the function
// // 	GetPeersResponse(conn, receivedMessage, knowPeers)

// // 	if clock.GetClock() != 1 {
// // 		t.Fatalf("Expected clock to be 1, got %d", clock.GetClock())
// // 	}

// // 	if string(conn.data) != expected {
// // 		t.Fatalf("Expected %s, got %s", expected, conn.data)
// // 	}

// // }
