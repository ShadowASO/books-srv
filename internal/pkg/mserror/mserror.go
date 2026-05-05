/*
---------------------------------------------------------------------------------------
File: mserror.go
Autor: Aldenor
Data: 04-05-2026
Alteração: 04-05-2026
---------------------------------------------------------------------------------------

func handle(c *gin.Context) {
	err := doSomething()
	if err != nil {
		// extrai code do AppError
		if code, ok := appError.CodeOf(err); ok {
			mslogger.Global.ErrorErr("handle/doSomething", err)
			response.Fail(c, code, "Erro na operação", "APP_ERROR", err.Error())
			return
		}

		mslogger.Global.ErrorErr("handle/doSomething", err)
		response.Fail(c, 500, "Erro interno", "INTERNAL", err.Error())
		return
	}

	response.OK(c, 200, "OK")
}

*/

package mserror

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// AppError representa um erro de aplicação com código e causa (opcional).
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Cause   error  `cause:"details,omitempty"`
}

func (e *AppError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Cause != nil {
		// Ex.: "invalid input (code=400): strconv.Atoi: parsing..."
		return fmt.Sprintf("%s (code=%d): %v", e.Message, e.Code, e.Cause)
	}
	return fmt.Sprintf("%s (code=%d)", e.Message, e.Code)
}

func (e *AppError) Unwrap() error { return e.Cause }

// CodeOf extrai o código de um AppError (se houver).
func CodeOf(err error) (int, bool) {
	var ae *AppError
	if errors.As(err, &ae) && ae != nil {
		return ae.Code, true
	}
	return 0, false
}

// MessageOf retorna a mensagem “de topo” do AppError quando houver.
func MessageOf(err error) (string, bool) {
	var ae *AppError
	if errors.As(err, &ae) && ae != nil {
		return ae.Message, true
	}
	return "", false
}

func New(code int, msg string) *AppError {
	return &AppError{Code: code, Message: msg}
}

func Newf(code int, format string, args ...any) *AppError {
	return &AppError{Code: code, Message: fmt.Sprintf(format, args...)}
}

// Wrap cria AppError mantendo a causa (para errors.Is/As funcionar).
func Wrap(code int, msg string, cause error) *AppError {
	return &AppError{Code: code, Message: msg, Cause: cause}
}

func Wrapf(code int, cause error, format string, args ...any) *AppError {
	return &AppError{Code: code, Message: fmt.Sprintf(format, args...), Cause: cause}
}

// NewError cria erro simples com detalhes concatenados de forma legível.
func NewError(message string, details ...string) error {
	if len(details) == 0 {
		return errors.New(message)
	}
	return fmt.Errorf("%s: %s", message, strings.Join(details, "; "))
}

func NewErrorf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

// ---------------- Backoff ----------------

var (
	rngMu sync.Mutex
	rng   = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// RetryBackoff calcula backoff exponencial com cap e jitter.
// attempt: 1..N (se vier <=0, trata como 1)
func RetryBackoff(attempt int) time.Duration {
	if attempt <= 0 {
		attempt = 1
	}

	base := 200 * time.Millisecond
	max := 2 * time.Second

	// Exponencial: base * 2^(attempt-1), com cap em max.
	d := base
	// Evita overflow/shift esquisito: multiplica iterativamente
	for i := 1; i < attempt; i++ {
		if d >= max/2 {
			d = max
			break
		}
		d *= 2
	}
	if d > max {
		d = max
	}

	// jitter: até 100ms
	jitterMax := 100 * time.Millisecond
	rngMu.Lock()
	j := time.Duration(rng.Int63n(int64(jitterMax + 1)))
	rngMu.Unlock()

	return d + j
}
