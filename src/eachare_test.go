package main

import "testing"

func TestGetArgs(t *testing.T) {
	addr, port, neighbors, shared := getArgs([]string{"eachare", "localhost:8080", "vizinhos", "shared"})

	if addr != "localhost" {
		t.Errorf("Addrs is casting invalid!")
	}
	if port != "8080" {
		t.Errorf("Port is casting invalid!")
	}

	if neighbors != "vizinhos" {
		t.Errorf("Neighbors is casting invalid!")
	}

	if shared != "shared" {
		t.Errorf("Shared is casting invalid!")
	}
}
