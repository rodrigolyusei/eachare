# EACHare
Exercício programa de Sistemas Distribuídos desenvolvido em Go, implementando um sistema de compartilhamento de arquivos peer-to-peer.

## Instalação do Go
Para o Windows, é possível instalar pelo arquivo `.zip` ou `.msi` no site oficial.\
Para o Ubuntu, é possível instalar pelo arquivo `.tar.gz` no site oficial ou usando o gerenciador de pacotes apt:
```cmd
sudo apt update
sudo apt install golang-go
```
É possível verificar a instalação com:
```cmd
go version
```

## Compilação e Execução do Programa
Todos os comandos abaixos, inclusive a de teste, deve ser feito estando no diretório `/src`.\
É possível compilar e executar com as duas linhas a seguir:
```cmd
go build ./eachare.go
./eachare.exe 127.0.0.1:9001 ../data/neighbor1.txt ../data/shared1/
```
Caso a versão do go não esteja compatível, crie um novo go.mod e tente novamente:
```cmd
rm go.mod
go mod init EACHare/src
```
Se quiser executar sem compilar, pode ser feito com `go run`:
```cmd
go run ./eachare.go 127.0.0.1:9001 ../data/neighbor1.txt ../data/shared1/
```

## Testes
Para gerar o cover dos unit tests, mostrando a taxa de funções tratadas, basta executar:
```cmd
go test ./... -coverprofile profile.out
go tool cover -func profile.out
```
