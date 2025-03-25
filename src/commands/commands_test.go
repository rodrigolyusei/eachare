package commands

import (
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
	conn := &mockConn{}
	message := BaseMessage{
		Clock:     1,
		Type:      "TEST",
		Arguments: []string{"arg1", "arg2"},
	}

	err := sendMessage(conn, message)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := "localhost 1 TEST arg1 arg2"
	if string(conn.data) != expected {
		t.Fatalf("Expected %s, got %s", expected, string(conn.data))
	}
}

func TestSendMessageArgumentsNil(t *testing.T) {
	conn := &mockConn{}
	message := BaseMessage{
		Clock:     1,
		Type:      "TEST",
		Arguments: nil,
	}

	err := sendMessage(conn, message)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := "localhost 1 TEST"
	if string(conn.data) != expected {
		t.Fatalf("Expected %s, got %s", expected, string(conn.data))
	}
}
