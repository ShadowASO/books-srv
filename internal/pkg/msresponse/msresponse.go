/*
---------------------------------------------------------------------------------------
File: msresponse.go
Autor: Aldenor
Data: 04-05-2026
Alteração: 04-05-2026
---------------------------------------------------------------------------------------
*/
package msresponse

import (
	"log"
	"microsrv/internal/pkg/mslogger"
	"time"

	"github.com/gin-gonic/gin"
)

type ErrorCode int

const (
	ErrorFormatoInvalido ErrorCode = 1
	ErrorTokenInvalido   ErrorCode = 2
	ErrorNaoAutorizado   ErrorCode = 3
	ErrorNaoEncontrado   ErrorCode = 4
	ErrorValidacao       ErrorCode = 5
	ErrorInterno         ErrorCode = 500
)

// APIResponse representa o envelope padrão de resposta da API.
type APIResponse struct {
	ID      string     `json:"id,omitempty" example:"7f8c2a9b-5f8a-4b0e-91e5-3b9c2d9a1234"`
	OK      bool       `json:"ok" example:"true"`
	Message string     `json:"message,omitempty" example:"Sucesso"`
	Data    any        `json:"data,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
}

// ErrorBody representa os detalhes de erro da resposta padrão.
type ErrorBody struct {
	Code        ErrorCode `json:"code,omitempty" example:"1"`
	Description string    `json:"description,omitempty" example:"Campo nm_obra é obrigatório"`
}

// LogTime registra um marco temporal de um processo.
func LogTime(msg string) {
	log.Printf("%s: %s", msg, time.Now().Format("2006-01-02 15:04:05"))
}

func getRequestID(c *gin.Context) string {
	if v, ok := c.Get("id"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}

	return ""
}

func SetRequestID(c *gin.Context, status int, resp APIResponse) {
	if resp.ID == "" {
		resp.ID = getRequestID(c)
	}

	c.JSON(status, resp)
}

// OK responde sucesso com data ou sem payload.
func OK(c *gin.Context, status int, message string, data ...any) {
	resp := APIResponse{
		OK:      true,
		Message: message,
	}

	if len(data) > 0 {
		resp.Data = data[0]
	}

	SetRequestID(c, status, resp)
}

// Fail responde erro padronizado.
func Fail(
	c *gin.Context,
	status int,
	message string,
	code ErrorCode,
	description string,
) {
	rid := getRequestID(c)

	mslogger.LoggerGlobal.Errorf(
		"id=%s status=%d code=%d message=%s description=%s",
		rid,
		status,
		code,
		message,
		description,
	)

	SetRequestID(c, status, APIResponse{
		OK:      false,
		ID:      rid,
		Message: message,
		Error: &ErrorBody{
			Code:        code,
			Description: description,
		},
	})
}
