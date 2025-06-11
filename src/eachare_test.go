package main

import "testing"

func TestGetArgs(t *testing.T) {
	client := getArgs([]string{"eachare", "localhost:8080", "../neighbors/n1.txt", "../shared"})

	if client.Address != "localhost:8080" {
		t.Errorf("Addrs is casting invalid!")
		t.Errorf("Expected: %s, got: %s", "localhost:8080", client.Address)
	}

	if client.neighbors != "../neighbors/n1.txt" {
		t.Errorf("Neighbors is casting invalid!")
		t.Errorf("Expected: %s, got: %s", "../neighbors/n1.txt", client.neighbors)
	}

	if client.shared != "../shared" {
		t.Errorf("Shared is casting invalid!")
		t.Errorf("Expected: %s, got: %s", "../shared", client.shared)
	}
}
