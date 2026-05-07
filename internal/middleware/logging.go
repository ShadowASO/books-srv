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

		request_id := c.Writer.Header().Get("X-Request-Id")
		if request_id == "" {
			request_id = c.GetString("request_id")
		}

		var errorCode string
		if v, ok := c.Get("error_code"); ok && v != nil {
			errorCode = fmt.Sprint(v)
		}

		var errorDetail string
		if v, ok := c.Get("error_detail"); ok && v != nil {
			errorDetail = fmt.Sprint(v)
		}

		mslogger.LoggerGlobal.HTTP(mslogger.HTTPLogData{
			RequestID:   request_id,
			Status:      c.Writer.Status(),
			Method:      c.Request.Method,
			Path:        c.Request.URL.Path,
			Route:       c.FullPath(),
			Handler:     c.HandlerName(),
			ClientIP:    c.ClientIP(),
			Duration:    duration,
			ErrorCode:   errorCode,
			ErrorDetail: errorDetail,
		})
	}
}
