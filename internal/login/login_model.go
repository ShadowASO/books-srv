// internal/handlers/login/swagger_models.go
/*
---------------------------------------------------------------------------------------
File: login_model.go
Autor: Aldenor
Data: 04-05-2026
Alteração: 04-05-2026
---------------------------------------------------------------------------------------
*/
package login

type LoginRequest struct {
	Username string `json:"username" example:"aldenor"`
	Password string `json:"password" example:"123456"`
}

type TokenRequest struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type LoginResponseData struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type RefreshTokenResponseData struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type VerifyTokenResponseData struct {
	ID    uint   `json:"id" example:"1"`
	Name  string `json:"name" example:"aldenor"`
	Email string `json:"email" example:"aldenor.oliveira@uol.com.br"`
	Role  string `json:"role" example:"admin"`
	Exp   int64  `json:"exp" example:"1767225599"`
}
