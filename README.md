# EACHare
Exercício programa de Sistemas Distribuídos desenvolvido em Go, implementando um sistema de compartilhamento de arquivos peer-to-peer.

## Instalação do Go
Para o Windows, é possível instalar pelo arquivo `.zip` ou `.msi` no site oficial.\
Para o Ubuntu, é possível instalar pelo arquivo `.tar.gz` no site oficial ou pelo gerenciador de pacotes apt.\
Uma ressalva com apt é que ele instala uma versão mais antiga 1.18 (no caso do nosso programa não tem problema):
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
./eachare 127.0.0.1:9001 ../data/neighbor1.txt ../data/shared1/
```
>**Aviso**: o executável gerado pelo Go varia de SO para SO, no caso do Linux gera um arquivo sem extensão, enquanto no windows gera um `.exe`.

Caso a versão do go não esteja compatível, crie um novo go.mod e tente novamente:
```cmd
rm go.mod
go mod init eachare/src
```
No caso de executar sem compilar, pode ser feito com `go run`.\
Usamos os seguintes comandos para criar peers e avaliar o comportamento:
```cmd
go run ./eachare.go 127.0.0.1:9001 ../data/neighbor1.txt ../data/shared1/
go run ./eachare.go 127.0.0.1:9002 ../data/neighbor2.txt ../data/shared2/
go run ./eachare.go 127.0.0.1:9003 ../data/neighbor3.txt ../data/shared3/
go run ./eachare.go 127.0.0.1:9004 ../data/neighbor4.txt ../data/shared4/
go run ./eachare.go 127.0.0.1:9005 ../data/neighbor5.txt ../data/shared5/
```

## Testes
Para gerar o cover dos unit tests, mostrando a taxa de funções tratadas, basta executar:
```cmd
go test ./... -coverprofile profile.out
go tool cover -func profile.out
```

## Docker
Para trabalhar com o docker, é necessário estar na pasta src e siga as etapas.\
Para subir os conteiners, use
```
docker compose up
```
Para removê-los,
```
docker compose down
```
Para entrar dentro do terminal de um docker para executar o programa,
```
docker exec -it peer1 /bin/sh
```
O número precisa ser o mesmo do docker-compose.
Depois, para executar a aplicação:
```
./peer peer1:9001 /app/neighbor.txt /app/shared
./peer peer2:9002 /app/neighbor.txt /app/shared
./peer peer3:9003 /app/neighbor.txt /app/shared
./peer peer4:9004 /app/neighbor.txt /app/shared
./peer peer5:9005 /app/neighbor.txt /app/shared
```
Atente-se que cada conteiner possui somente aberta a porta 900- de acordo com seu número de peer. Portanto, se tentar executar no conteiner peer2 com o endereço de peer1:9001, não irá funcionar. Para simplificar, siga os comandos em ordem. Recomenda-se abrir um terminal  para cada docker exec.