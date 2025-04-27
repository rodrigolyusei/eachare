package main

import "testing"

func TestGetArgs(t *testing.T) {
	myargs := getArgs([]string{"eachare", "localhost:8080", "vizinhos", "../shared"})

	if myargs.Address != "localhost" {
		t.Errorf("Addrs is casting invalid!")
	}
	if myargs.Port != "8080" {
		t.Errorf("Port is casting invalid!")
	}

	if myargs.Neighbors != "vizinhos" {
		t.Errorf("Neighbors is casting invalid!")
	}

	if myargs.Shared != "../shared" {
		t.Errorf("Shared is casting invalid!")
	}
}
