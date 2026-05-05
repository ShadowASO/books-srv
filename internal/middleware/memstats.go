/*
---------------------------------------------------------------------------------------
File: memstats.go
Autor: Aldenor
Data: 04-05-2026
Alteração: 04-05-2026
---------------------------------------------------------------------------------------
*/
package middleware

import (
	"log"
	"runtime"

	"github.com/gin-gonic/gin"
)

var memStats runtime.MemStats

/* Atribui um UUID para todas as requisições recebidas. Ela será usada para identificar a resposta */
func LogMemStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		runtime.ReadMemStats(&memStats)
		log.Println("Memory Status:")
		log.Printf("Alloc = %v MiB", memStats.Alloc/1024/1024)
		log.Printf("TotalAlloc = %v MiB", memStats.TotalAlloc/1024/1024)
		log.Printf("Sys = %v MiB", memStats.Sys/1024/1024)
		log.Printf("NumGC = %v", memStats.NumGC)

		c.Next()
	}
}
