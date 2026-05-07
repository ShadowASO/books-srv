/*
---------------------------------------------------------------------------------------
File: server.go
---------------------------------------------------------------------------------------
Autor: Aldenor
Data: 14-04-2026
Alteração: 06-05-2026
---------------------------------------------------------------------------------------
Compilação: go build -v -o server ./cmd/server.go
Execução:   ./server
---------------------------------------------------------------------------------------
*/
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"microsrv/internal/config"
	"microsrv/internal/middleware"
	"microsrv/internal/pkg/mslogger"
	"microsrv/internal/repository/mongodb"
	"microsrv/internal/repository/postgres"
	"microsrv/internal/routes"

	_ "microsrv/docs"

	"github.com/gin-gonic/gin"
)

func main() {

	/* Carrega as configurações a partir do .env */
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Errorf("erro ao carregar configuração: %v", err))
	}
	/* Fixa o modo de funcionamento do GIN */
	gin.SetMode(cfg.GinMode)

	/* Faz a inicialização do Logger global */
	err = mslogger.InitGlobal(mslogger.Options{
		FilePath:   "./logs/app.log",
		Stdout:     true,
		Rotate:     true,
		MaxSizeMB:  20,
		MaxBackups: 10,
		MaxAgeDays: 30,
		Compress:   true,
		Level:      mslogger.DebugLevel,
		JSON:       true,
		Service:    "books-srv",
		AddSource:  true,
	})
	if err != nil {
		panic(err)
	}
	// Encerramento do Logger global deferido
	defer func() {
		if mslogger.LoggerGlobal != nil {
			mslogger.LoggerGlobal.InfoData("app encerrado", mslogger.AppLogData{
				Context: "shutdown",
			})

			_ = mslogger.LoggerGlobal.Close()
		}
	}()
	/* Insere mensagem no logger de inicialização*/
	// mslogger.LoggerGlobal.Infof(
	// 	"app iniciou | mode=%s | env=%s",
	// 	cfg.GinMode,
	// 	cfg.ApplicationMode,
	// )
	mslogger.LoggerGlobal.InfoData("app iniciou", mslogger.AppLogData{
		Context: "startup",
		Mode:    gin.Mode(),
		Env:     config.GlobalConfig.ApplicationMode,
	})
	/* Cria um contexto para uso pelo mongo*/
	appCtx := context.Background()

	/* Cria um cliente mongo para uso */
	mongoClient := mongodb.NewMongoDB(cfg)
	if err := mongoClient.Connect(appCtx); err != nil {
		mslogger.LoggerGlobal.ErrorErr("erro ao conectar MongoDB", err)
		return
	}

	defer func() {
		if err := mongoClient.DisconnectMongo(appCtx); err != nil {
			mslogger.LoggerGlobal.ErrorErr("erro ao encerrar conexão com MongoDB", err)
			return
		}

		mslogger.LoggerGlobal.InfoData("MongoDB desconectado com sucesso", mslogger.AppLogData{
			Context: "shutdown",
		})
	}()

	mslogger.LoggerGlobal.InfoData("MongoDB conectado com sucesso", mslogger.AppLogData{
		Context: "startup",
	})

	/* Conexta à coleção que será utilizada na aplicação. */

	booksCollection, err := mongoClient.GetMongoCollection(appCtx, cfg.MdDB, "books")
	if err != nil {
		mslogger.LoggerGlobal.ErrorErr("erro ao obter collection books", err)
		return
	}

	/* Configura e cria uma conexta ao PostgreSQL*/
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
		mslogger.LoggerGlobal.ErrorErr("erro ao criar pool de conexões com PostgreSQL", err)
		return
	}

	defer func() {
		db.Close()

		mslogger.LoggerGlobal.InfoData("PostgreSQL desconectado com sucesso", mslogger.AppLogData{
			Context: "shutdown",
		})
	}()

	/* Cria um router do GIN*/
	router := gin.New()

	/* Faz a atribuição do Middleware ao router. */
	router.Use(
		middleware.Logging(),
		middleware.RequestIDMiddleware(),
		middleware.ConfigureCors(cfg.AllowedOrigins),
		gin.Recovery(),
	)

	/* Chama a função que rotina que cria as rotas do sistema. */
	routes.SetRotasSistema(router, cfg, db, booksCollection)
	/*  -----------------------------------------------------  */

	/* Faz a configuração do Servidor HTTP. */
	addr := cfg.ServerPort

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}
	/* Executa o Servidor HTTP. */
	go func() {
		mslogger.LoggerGlobal.Infof("servidor escutando em %s", addr)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			mslogger.LoggerGlobal.ErrorErr("erro ao iniciar servidor", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit

	mslogger.LoggerGlobal.InfoData("sinal de encerramento recebido", mslogger.AppLogData{
		Context: sig.String(),
	})

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		mslogger.LoggerGlobal.ErrorErr("erro ao encerrar servidor", err)
		return
	}

	mslogger.LoggerGlobal.InfoData("servidor encerrado com sucesso", mslogger.AppLogData{
		Context: "shutdown",
	})
}
