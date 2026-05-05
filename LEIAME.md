### Create o Projeto e criando o go.mod file:

mkdir nmprojeto
cd nmprojeto
go mod init 

### Atualizando os módulos do projeto

go mod tidy

### Compilação do Projeto 

go build -v -o server ./cmd/server.go

### Execução:

./server

### Instalar Swagger

# Install the Swag CLI to generate docs
go install github.com/swaggo/swag/cmd/swag@latest

# Install the gin-swagger middleware and files handler
go get github.com/swaggo/gin-swagger
go get github.com/swaggo/files

### SWAGGER - Documentação da API com Swagger

./scripts/build-swagger.sh

### JWT - Geração da Chave Forte

### Para HS256 - 32 bytes
openssl rand -hex 32


### Para HS512 - 64 bytes
openssl rand -base64 64


### COMANDOS ÚTEIS:

# Initialize a new module
go mod init example.com/myproject

# Download dependencies
go mod download

# Update dependencies
go get -u ./...
go get -u=patch ./...  # Only patch updates

# Clean up dependencies
go mod tidy

# Verify dependencies
go mod verify

# List all dependencies
go list -m all




