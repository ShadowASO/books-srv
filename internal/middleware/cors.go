/*
---------------------------------------------------------------------------------------
File: cors.go
Autor: Aldenor
Data: 04-05-2026
Alteração: 04-05-2026
---------------------------------------------------------------------------------------
*/
package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

/* Faz a configuração de CORS e devolve um middleware */
//func ConfigureCors(cfg *config.Config) gin.HandlerFunc {
func ConfigureCors(allowed_origins []string) gin.HandlerFunc {
	/* Configuração do CORS */
	corsCfg := cors.Config{
		//AllowOrigins:     cfg.AllowedOrigins,
		AllowOrigins:     allowed_origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	return cors.New(corsCfg)
}
