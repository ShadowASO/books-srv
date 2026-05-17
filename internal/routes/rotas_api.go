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
	"microsrv/internal/middleware/grpc_middleware"
	"microsrv/internal/services/grpc_services/authgrpc"

	"microsrv/internal/handlers"

	"microsrv/internal/middleware"
	"microsrv/internal/repository/mongodb"
	"microsrv/internal/repository/postgres"

	"microsrv/internal/services"

	_ "microsrv/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// SetRotasSistema registra todas as rotas e injeta dependências
func SetRotasSistema(router *gin.Engine, cfg *config.Config, dbPG *postgres.PGPool, mdCollect *mongo.Collection, authClient *authgrpc.ClientAuth) {

	// authClient, err := authms.New(msclientehttp.ConfigClienteHTTP{
	// 	Name:               cfg.AuthName,
	// 	BaseURL:            cfg.AuthServiceURL,
	// 	Timeout:            5 * time.Second,
	// 	Debug:              cfg.AuthClientDebug,
	// 	InsecureSkipVerify: cfg.AuthInsecureSkipVerify,
	// })
	// if err != nil {
	// 	panic(err)
	// }

	/* Criação do Serviço de Autenticação por JWT */
	//jwt := auth.NewJWTService(cfg.JWTSecretKey)

	/* Criação do Repositório, Service e  Handler. */
	//Books
	booksRepo := mongodb.NewBooksMongoRepository(mdCollect)
	booksService := services.NewBooksService(booksRepo)
	booksHandler := handlers.NewBooksHandler(*booksService)

	/* Users */
	// userRepo := postgres.NewUserPGRepository(dbPG.Pool)
	// userService := services.NewUserService(userRepo)
	// userHandler := handlers.NewUserHandler(*userService)

	//Login

	//authHandler := authms.NewAuthHandler(authClient)
	authHandler := handlers.NewAuthHandler(authClient)

	/* Configuração das Rotas da API */
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	verApi := router.Group("/api/v1")
	{

		authApi := verApi.Group("/auth")
		{
			authApi.POST("/login", authHandler.Login)
			authApi.POST("/token/verify", authHandler.VerifyToken)
			authApi.POST("/token/refresh", authHandler.RefreshToken)

		}
		sysApi := verApi.Group("/sys")
		{
			sysApi.GET("/version", handlers.VersionHandler)
			sysApi.GET("/health", handlers.HealthCheck, middleware.LogMemStats())

		}

		tabelasApi := verApi.Group("/tabelas", grpc_middleware.GRPC_authMiddleware(authClient))
		{
			tabelasApi.POST("/books", booksHandler.Insert)
			tabelasApi.PUT("/books/:id", booksHandler.Update)
			tabelasApi.GET("/books/:id", booksHandler.Select)
			tabelasApi.DELETE("/books/:id", booksHandler.Delete)
			tabelasApi.POST("/books/search", booksHandler.SearchByNmObra)

		}

		// usersApi := verApi.Group("/auth/users")
		// {
		// 	usersApi.POST("", userHandler.Insert)
		// 	usersApi.GET("", userHandler.SelectRows)

		// 	usersApi.POST("/search", userHandler.Search)
		// 	usersApi.GET("/username/:username", userHandler.SelectUserByName)
		// 	usersApi.GET("/email/:email", userHandler.SelectByEmail)

		// 	usersApi.GET("/:id", userHandler.Select)
		// 	usersApi.PUT("/:id", userHandler.Update)
		// 	usersApi.DELETE("/:id", userHandler.Delete)
		// }

	}
}
