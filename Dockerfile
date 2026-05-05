FROM golang:1.26.2-alpine AS builder

LABEL maintainer="aldenor"


# Diretório de trabalho dentro do container
WORKDIR /app

# Copiar arquivos de dependências do Go
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copiar o código no diretório atual para o diretório de trabalho dentro do container
COPY . .

# Criar diretório de logs com permissões
#RUN useradd -m appuser
#RUN mkdir -p /app/logs && chown -R appuser:appuser /app/logs
#RUN mkdir -p /app/logs

# Compilar o binário da aplicação
RUN go build -v -o server ./cmd/server.go

#------------------------------------------------------------
#    CONCLUÍDA A COMPILAÇÃO - SEGUE A CÓPIA PARA O ALPINE
#------------------------------------------------------------    

FROM alpine:latest

WORKDIR /app

RUN mkdir -p /app/logs

COPY --from=builder /app/server .
COPY --from=builder /app/.env .

# Expor a porta que a aplicação usa
EXPOSE 4010

# Comando para iniciar a aplicação
CMD ["./server"]