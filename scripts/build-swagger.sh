#!/usr/bin/env bash

set -euo pipefail

APP_NAME="server"
MAIN_FILE="./cmd/server.go"

echo "========================================"
echo " Verificando versão atual do swaggo/swag"
echo "========================================"
go list -m github.com/swaggo/swag || true

echo
echo "========================================"
echo " Atualizando dependências do Swagger"
echo "========================================"
go get github.com/swaggo/swag@latest
go get github.com/swaggo/gin-swagger@latest
go get github.com/swaggo/files@latest

echo
echo "========================================"
echo " Executando go mod tidy"
echo "========================================"
go mod tidy

echo
echo "========================================"
echo " Gerando documentação Swagger"
echo "========================================"
swag init -g "${MAIN_FILE}" --parseInternal

echo
echo "========================================"
echo " Compilando aplicação"
echo "========================================"
go build -v -o "${APP_NAME}" "${MAIN_FILE}"

echo
echo "========================================"
echo " Processo concluído com sucesso"
echo "========================================"
echo "Binário gerado: ./${APP_NAME}"
echo "Swagger gerado em:"
echo "  docs/docs.go"
echo "  docs/swagger.json"
echo "  docs/swagger.yaml"