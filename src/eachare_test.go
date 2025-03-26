package main

import "testing"

func TestGetArgs(t *testing.T) {
	all_args := getArgs([]string{"eachare", "localhost:8080", "vizinhos", "../shared"})

	if all_args.Address != "localhost" {
		t.Errorf("Addrs is casting invalid!")
	}
	if all_args.Port != "8080" {
		t.Errorf("Port is casting invalid!")
	}

	if all_args.Neighbors != "vizinhos" {
		t.Errorf("Neighbors is casting invalid!")
	}

	if all_args.Shared != "../shared" {
		t.Errorf("Shared is casting invalid!")
	}
}
