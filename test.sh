#!/bin/bash

if [ -z "$1" ]; then
    set -- "127.0.0.1:8080" "neighbors" "shared"
fi
# Compila o arquivo eachare.go
go run src/eachare.go "$1" "$2" "$3" "$4"
# Define argumentos padrão para testes
