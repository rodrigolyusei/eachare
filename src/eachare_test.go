package main

import "testing"

func TestGetArgs(t *testing.T) {
	getArgs([]string{"eachare", "localhost:8080", "../neighbors/n1.txt", "../shared"})

	if myArgs.Address != "localhost:8080" {
		t.Errorf("Addrs is casting invalid!")
		t.Errorf("Expected: %s, got: %s", "localhost:8080", myArgs.Address)
	}

	if myArgs.Neighbors != "../neighbors/n1.txt" {
		t.Errorf("Neighbors is casting invalid!")
		t.Errorf("Expected: %s, got: %s", "../neighbors/n1.txt", myArgs.Neighbors)
	}

	if myArgs.Shared != "../shared" {
		t.Errorf("Shared is casting invalid!")
		t.Errorf("Expected: %s, got: %s", "../shared", myArgs.Shared)
	}
}
