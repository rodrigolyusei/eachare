package main

import "testing"

func TestGetArgs(t *testing.T) {
	getArgs([]string{"eachare", "localhost:8080", "../neighbors/n1.txt", "../shared"})

	if myArgs.address != "localhost:8080" {
		t.Errorf("Addrs is casting invalid!")
		t.Errorf("Expected: %s, got: %s", "localhost:8080", myArgs.address)
	}

	if myArgs.neighbors != "../neighbors/n1.txt" {
		t.Errorf("Neighbors is casting invalid!")
		t.Errorf("Expected: %s, got: %s", "../neighbors/n1.txt", myArgs.neighbors)
	}

	if myArgs.shared != "../shared" {
		t.Errorf("Shared is casting invalid!")
		t.Errorf("Expected: %s, got: %s", "../shared", myArgs.shared)
	}
}
