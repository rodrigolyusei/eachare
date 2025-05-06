# EACHare
Exercício programa de Sistemas Distribuídos feito em Go, implementando um sistema de compartilhamento de arquivos peer-to-peer

## Instalação do Go
Para o Ubuntu, é possível instalar pelo arquivo comprimido no site oficial ou usando o gerenciador de pacotes apt:
```cmd
sudo apt update
sudo apt install golang-go
```
É possível verificar a instalação com:
```cmd
go version
```

## Compilação e Execução do Programa
É possível compilar e executar, após estar no diretório /src, com as duas linhas a seguir:
```cmd
go build ./eachare.go
./eachare.go 127.0.0.1:9000 ../neighbors/n1.txt ../shared
```
Caso a versão do go não esteja compatível, crie um novo go.mod:
```cmd
rm go.mod
go mod init EACHare/src
```
Depois tente novamente com build.

## Testes
Para gerar o cover dos unit tests, mostrando a taxa de funções tratadas:

```cmd
go test ./... -coverprofile profile.out
go tool cover -func profile.out
```
