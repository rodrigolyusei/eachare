# EACHare
Exercício Programa de Sistemas Distribuídos

# Tests

Para gerar o cover dos unit tests,

```cmd
cd src
go test ./... -coverprofile profile.out
go tool cover -func profile.out
```