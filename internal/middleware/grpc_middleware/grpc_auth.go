// package middlewaregrpc
package grpc_middleware

import (
	"microsrv/internal/services/grpc_services/authgrpc"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	CtxUserID   = "user_id"
	CtxUsername = "username"
	CtxEmail    = "email"
	CtxRole     = "role"
	CtxTokenExp = "token_exp"
)

func GRPC_authMiddleware(authClient *authgrpc.ClientAuth) gin.HandlerFunc {
	return func(c *gin.Context) {
		if authClient == nil {
			abortAuth(
				c,
				http.StatusInternalServerError,
				5001,
				"Serviço de autenticação não configurado",
			)
			return
		}

		token, ok := extractBearerToken(c.GetHeader("Authorization"))
		if !ok {
			abortAuth(
				c,
				http.StatusUnauthorized,
				1,
				"Informe o token no formato Authorization: Bearer <token>.",
			)
			return
		}

		data, err := authClient.VerifyToken(c.Request.Context(), token)
		if err != nil {
			abortAuth(
				c,
				http.StatusUnauthorized,
				2,
				err.Error(),
			)
			return
		}

		c.Set(CtxUserID, data.ID)
		c.Set(CtxUsername, data.Name)
		c.Set(CtxEmail, data.Email)
		c.Set(CtxRole, data.Role)
		c.Set(CtxTokenExp, data.Exp)

		c.Next()
	}
}

func extractBearerToken(authHeader string) (string, bool) {
	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" {
		return "", false
	}

	parts := strings.Fields(authHeader)
	if len(parts) != 2 {
		return "", false
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}

	return token, true
}

func abortAuth(c *gin.Context, status int, code int, description string) {
	c.AbortWithStatusJSON(status, gin.H{
		"ok":      false,
		"message": "Não autorizado",
		"error": gin.H{
			"code":        code,
			"description": description,
		},
	})
}
