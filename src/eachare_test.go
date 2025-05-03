package main

import "testing"

func TestGetArgs(t *testing.T) {
	getArgs([]string{"eachare", "localhost:8080", "../neighbors/n1.txt", "../shared"})

	if myAddress != "localhost:8080" {
		t.Errorf("Addrs is casting invalid!")
		t.Errorf("Expected: %s, got: %s", "localhost:8080", myAddress)
	}

	if myNeighbors != "../neighbors/n1.txt" {
		t.Errorf("Neighbors is casting invalid!")
		t.Errorf("Expected: %s, got: %s", "../neighbors/n1.txt", myNeighbors)
	}

	if myShared != "../shared" {
		t.Errorf("Shared is casting invalid!")
		t.Errorf("Expected: %s, got: %s", "../shared", myShared)
	}
}
