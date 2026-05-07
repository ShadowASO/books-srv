/*
---------------------------------------------------------------------------------------
File: msresponse.go
Autor: Aldenor
Data: 04-05-2026
Alteração: 06-05-2026
---------------------------------------------------------------------------------------
*/
package msresponse

import (
	"fmt"
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
	if mslogger.LoggerGlobal != nil {
		mslogger.LoggerGlobal.InfoData("time_marker", mslogger.AppLogData{
			Context: fmt.Sprintf("%s: %s", msg, time.Now().Format("2006-01-02 15:04:05")),
		})
		return
	}

	fmt.Printf("%s: %s\n", msg, time.Now().Format("2006-01-02 15:04:05"))
}

func getRequestID(c *gin.Context) string {
	if c == nil {
		return ""
	}

	if v, ok := c.Get("id"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}

	if v, ok := c.Get("request_id"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}

	if rid := c.Writer.Header().Get("X-Request-Id"); rid != "" {
		return rid
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

	// Não logar 400/401/403/404 como erro de aplicação.
	// Esses casos já serão registrados pelo middleware HTTP.
	if status >= 500 && mslogger.LoggerGlobal != nil {
		mslogger.LoggerGlobal.ErrorData("response_fail", mslogger.AppLogData{
			RequestID:   rid,
			Status:      status,
			Code:        int(code),
			Context:     message,
			Description: description,
		})
	}

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
