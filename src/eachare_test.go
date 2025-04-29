package main

import "testing"

func TestGetArgs(t *testing.T) {
	myargs := getArgs([]string{"eachare", "localhost:8080", "../neighbors/n1.txt", "../shared"})

	if myargs.Address != "localhost:8080" {
		t.Errorf("Addrs is casting invalid!")
	}

	if myargs.Neighbors != "../neighbors/n1.txt" {
		t.Errorf("Neighbors is casting invalid!")
	}

	if myargs.Shared != "../shared" {
		t.Errorf("Shared is casting invalid!")
	}
}
