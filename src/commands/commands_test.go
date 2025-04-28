package commands

import (
	"bytes"
	"os"
	"strings"
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
