/*
---------------------------------------------------------------------------------------
File: requestid.go
Autor: Aldenor
Data: 04-05-2026
Alteração: 04-05-2026
---------------------------------------------------------------------------------------
*/
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const CtxRequestID = "id"

func NewRequestID() string {
	if v7, err := uuid.NewV7(); err == nil {
		return v7.String()
	}

	return uuid.NewString()
}

/*
Essa versão permite que um NGINX, API Gateway ou frontend envie um ID já existente,
mantendo o rastreamento ponta a ponta.
*/
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetString(CtxRequestID)

		if rid == "" {
			rid = c.GetHeader("X-Request-Id")
		}

		if rid == "" {
			rid = NewRequestID()
		}

		c.Set(CtxRequestID, rid)
		c.Writer.Header().Set("X-Request-Id", rid)

		c.Next()
	}
}

func GetRequestID(c *gin.Context) string {
	rid := c.GetString(CtxRequestID)

	if rid == "" {
		rid = c.GetHeader("X-Request-Id")
	}
	return rid

}
