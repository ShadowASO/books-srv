/*
---------------------------------------------------------------------------------------
File: server.go
---------------------------------------------------------------------------------------
Autor: Aldenor
Data: 14-04-2026
---------------------------------------------------------------------------------------
Compilação: go build -v -o server ./cmd/server.go
Execução:   ./server
---------------------------------------------------------------------------------------
*/
package main

import (
	"context"

	"fmt"
	"log"
	"microsrv/internal/config"
	"microsrv/internal/routes"

	"microsrv/internal/middleware"
	"microsrv/internal/repository/mongodb"
	"microsrv/internal/repository/postgres"

	"microsrv/internal/pkg/mslogger"

	"os"
	"path/filepath"

	_ "microsrv/docs"

	"github.com/gin-gonic/gin"
)

// @title           Server API
// @version         1.0
// @description     This is a template server.
// @termsOfService  http://swagger.io

// @contact.name   API Support
// @contact.url    http://swagger.io
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1
func main() {
	fmt.Println("Iniciando o microsserviço")

	/* Carregando as Configurações da aplicação do .env */
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Errorf("erro ao carregar configuração: %v", err))
	}

	/* Set o mode de execução do GIN: release ou debug */
	gin.SetMode(cfg.GinMode)

	/* Inicializa o serviço de logger  do sistema */
	logDir := "logs"
	logFile := filepath.Join(logDir, "app.log")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(fmt.Errorf("erro ao criar diretório de logs: %v", err))
	}

	err = mslogger.InitGlobal(mslogger.Options{
		FilePath: logFile,
		Stdout:   true,
		Rotate:   true,
		Compress: true,
		Level:    mslogger.ParseLevel(os.Getenv("LOG_LEVEL")),
	})
	if err != nil {
		panic(err)
	}
	defer mslogger.LoggerGlobal.Close()

	mslogger.LoggerGlobal.Infof(
		"app iniciou | mode=%s | env=%s",
		cfg.GinMode,
		cfg.ApplicationMode,
	)

	/* Conexão com o Banco de Dados - MongoDB  */
	ctx := context.Background()
	mongoClient := mongodb.NewMongoDB(cfg)
	if err := mongoClient.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := mongoClient.DisconnectMongo(ctx); err != nil {
			log.Println("erro ao encerrar conexão com MongoDB:", err)
		}
	}()
	log.Println("MongoDB conectado com sucesso")

	/* Criação do Cliente Collection */
	booksCollection, err := mongoClient.GetMongoCollection(ctx, cfg.MdDB, "books")
	if err != nil {
		log.Fatal(err)
	}

	/* Conexão com o POSTGRESQL*/

	dbConfig := postgres.PGConfig{
		Host:     cfg.PgHost,
		Port:     cfg.PgPort,
		User:     cfg.PgUser,
		Password: cfg.PgPass,
		DBName:   cfg.PgDB,
		PoolSize: cfg.PGPoolSize,
	}
	db, err := postgres.NewPGConn(dbConfig)
	if err != nil {
		log.Fatalf("erro ao criar pool de conexões com o database: %v", err)
	}
	defer db.Close()

	/*Criação do Router para tratar as chamadas de API */
	router := gin.New()

	/* Aplicação dos middleware */
	router.Use(
		middleware.Logging(),
		middleware.RequestIDMiddleware(),
		/* Configuração do CORS */
		middleware.ConfigureCors(cfg.AllowedOrigins),
		gin.Recovery(),
	)

	// ROTAS - Configuração das rotas
	routes.SetRotasSistema(router, cfg, db, booksCollection)
	/*-----------------------------------------------------------*/

	/* Starting ther Server */
	addr := cfg.ServerPort
	mslogger.LoggerGlobal.Infof("servidor escutando em %s", addr)

	if err := router.Run(addr); err != nil {
		mslogger.LoggerGlobal.Errorf("erro ao iniciar servidor: %v", err)
	}
}
