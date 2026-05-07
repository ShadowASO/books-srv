/*
---------------------------------------------------------------------------------------
File: rotas.go
Autor: Aldenor
Data: 04-05-2026
Alteração: 04-05-2026
---------------------------------------------------------------------------------------
*/
package routes

import (
	"microsrv/internal/config"

	"microsrv/internal/handlers"
	"microsrv/internal/login"
	"microsrv/internal/middleware"
	"microsrv/internal/repository/mongodb"
	"microsrv/internal/repository/postgres"

	"microsrv/internal/pkg/msauth"

	"microsrv/internal/services"

	_ "microsrv/docs"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// SetRotasSistema registra todas as rotas e injeta dependências
func SetRotasSistema(router *gin.Engine, cfg *config.Config, dbPG *postgres.PGPool, mdCollect *mongo.Collection) {

	/* Criação do Serviço de Autenticação por JWT */
	jwt := auth.NewJWTService(cfg.JWTSecretKey)

	/* Criação do Repositório, Service e  Handler. */
	//Books
	booksRepo := mongodb.NewBooksMongoRepository(mdCollect)
	booksService := services.NewBooksService(booksRepo)
	booksHandler := handlers.NewBooksHandler(*booksService)

	/* Users */
	userRepo := postgres.NewUserPGRepository(dbPG.Pool)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(*userService)

	//Login

	loginHandler := login.NewLoginHandlers(userService, jwt, cfg.RefreshTokenExpire, cfg.AccessTokenExpire)

	/* Configuração das Rotas da API */
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	verApi := router.Group("/api/v1")
	{

		authApi := verApi.Group("/auth")
		{
			authApi.POST("/login", loginHandler.Login)
			authApi.POST("/token/verify", loginHandler.VerifyToken)
			authApi.POST("/token/refresh", loginHandler.RefreshToken)
		}
		sysApi := verApi.Group("/sys")
		{
			sysApi.GET("/version", jwt.AuthMiddleware(), handlers.VersionHandler)
			sysApi.GET("/health", handlers.HealthCheck, middleware.LogMemStats())

		}

		tabelasApi := verApi.Group("/tabelas", jwt.AuthMiddleware())
		{
			tabelasApi.POST("/books", booksHandler.Insert)
			tabelasApi.PUT("/books/:id", booksHandler.Update)
			tabelasApi.GET("/books/:id", booksHandler.Select)
			tabelasApi.DELETE("/books/:id", booksHandler.Delete)
			tabelasApi.POST("/books/search", booksHandler.SearchByNmObra)

		}

		usersApi := verApi.Group("/auth/users")
		{
			usersApi.POST("", userHandler.Insert)
			usersApi.GET("", userHandler.SelectRows)

			usersApi.POST("/search", userHandler.Search)
			usersApi.GET("/username/:username", userHandler.SelectUserByName)
			usersApi.GET("/email/:email", userHandler.SelectByEmail)

			usersApi.GET("/:id", userHandler.Select)
			usersApi.PUT("/:id", userHandler.Update)
			usersApi.DELETE("/:id", userHandler.Delete)
		}

	}
}
