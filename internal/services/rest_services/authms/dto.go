/*
---------------------------------------------------------------------------------------
File: dto.go
Autor: Aldenor
Data: 13-05-2026
Alteração: 13-05-2026
---------------------------------------------------------------------------------------
Este arquivo declara os DTOs utilizados pelo pacote authms na comunicação com o
microsserviço de autenticação auth-srv.

As estruturas aqui definidas representam os contratos de entrada e saída das operações
de autenticação, tais como login, validação de token e renovação de access token.

Também são declaradas estruturas genéricas para tratamento do envelope padrão de
resposta da API, incluindo dados de sucesso, mensagens, metadados da requisição e
informações de erro retornadas pelo serviço remoto.

Esses DTOs são consumidos pela entidade ClientAuth, que utiliza o ClienteHTTP genérico
para realizar chamadas HTTP ao auth-srv por meio da rede interna do Docker.
*/
package authms

import "strings"

type APIError struct {
	Code        int    `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description"`
}

type APIResponse[T any] struct {
	ID        string    `json:"id"`
	RequestID string    `json:"request_id"`
	OK        bool      `json:"ok"`
	Message   string    `json:"message"`
	Data      T         `json:"data,omitempty"`
	Error     *APIError `json:"error,omitempty"`
	Timestamp string    `json:"timestamp,omitempty"`
}

func (r APIResponse[T]) hasErrorContent() bool {
	if r.Error != nil {
		return true
	}

	return strings.TrimSpace(r.Message) != ""
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type LoginData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`

	UserID   int    `json:"user_id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type TokenRequest struct {
	Token string `json:"token"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type RefreshTokenData struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type TokenClaims struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Exp   int64  `json:"exp"`
}
