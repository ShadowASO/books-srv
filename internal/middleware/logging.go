/*
---------------------------------------------------------------------------------------
File: logging.go
Autor: Aldenor
Data: 04-05-2026
Alteração: 04-05-2026
---------------------------------------------------------------------------------------
*/
package middleware

import (
	"fmt"

	"time"

	"microsrv/internal/pkg/mslogger"

	"github.com/gin-gonic/gin"
)

func Logging() gin.HandlerFunc {

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		msg := fmt.Sprintf("| %d |  %v | %s  | %s : %s", c.Writer.Status(), duration, c.Request.Method, c.RemoteIP(), c.Request.URL.Path)

		mslogger.LoggerGlobal.Info(msg)

	}
}
