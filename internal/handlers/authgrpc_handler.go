/*
---------------------------------------------------------------------------------------
File: loginHandler.go
Autor: Aldenor
Data: 04-05-2026
Alteração: 13-05-2026
---------------------------------------------------------------------------------------
*/
package handlers

import (
	"microsrv/internal/pkg/mslogger"
	"microsrv/internal/services/grpc_services/authgrpc"

	"microsrv/internal/pkg/msresponse"

	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthHandlerType struct {
	authClient *authgrpc.ClientAuth
}

func NewAuthHandler(authClient *authgrpc.ClientAuth) *AuthHandlerType {
	if authClient == nil {
		panic("authClient não informado")
	}

	return &AuthHandlerType{
		authClient: authClient,
	}
}

// VerifyToken verifica se o access token ainda é válido.
// @Summary Verificar token
// @Description Verifica se um access token JWT é válido e retorna os dados principais das claims.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body TokenRequest true "Token JWT"
// @Success 200 {object} msresponse.APIResponse "Token válido"
// @Failure 400 {object} msresponse.APIResponse "Formato inválido ou campos obrigatórios ausentes"
// @Failure 401 {object} msresponse.APIResponse "Token inválido ou expirado"
// @Failure 500 {object} msresponse.APIResponse "Erro interno"
// @Router /auth/token/verify [post]
func (obj *AuthHandlerType) VerifyToken(c *gin.Context) {

	var body authgrpc.TokenRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Formato inválido",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	body.Token = strings.TrimSpace(body.Token)

	if body.Token == "" {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Token não enviado",
			msresponse.ErrorTokenInvalido,
			"O campo token é obrigatório.",
		)
		return
	}

	data, err := obj.authClient.VerifyToken(c.Request.Context(), body.Token)
	if err != nil {
		msresponse.Fail(
			c,
			http.StatusUnauthorized,
			"token inválido ou expirado",
			msresponse.ErrorValidacao,
			"token inválido ou expirado",
		)
		return
	}

	rsp := authgrpc.TokenClaims{
		ID:    (data.ID),
		Name:  data.Name,
		Email: data.Email,
		Role:  data.Role,
		Exp:   data.Exp,
	}

	msresponse.OK(c, http.StatusOK, "Token válido", rsp)
}

// RefreshToken gera novo access token a partir de um refresh token válido.
// @Summary Renovar access token
// @Description Valida um refresh token e retorna um novo access token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body TokenRequest true "Refresh token"
// @Success 200 {object} msresponse.APIResponse "Token renovado com sucesso"
// @Failure 400 {object} msresponse.APIResponse "Formato inválido ou campos obrigatórios ausentes"
// @Failure 401 {object} msresponse.APIResponse "Refresh token inválido ou expirado"
// @Failure 500 {object} msresponse.APIResponse "Erro interno ao renovar token"
// @Router /auth/token/refresh [post]
func (obj *AuthHandlerType) RefreshToken(c *gin.Context) {
	var body authgrpc.TokenRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		mslogger.LoggerGlobal.Errorf("Refresh token inválido: %v", err)
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Formato inválido",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	body.Token = strings.TrimSpace(body.Token)

	if body.Token == "" {

		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Refresh token não enviado",
			msresponse.ErrorTokenInvalido,
			"O campo token é obrigatório.",
		)
		return
	}

	data, err := obj.authClient.RefreshToken(c.Request.Context(), body.Token)
	if err != nil {
		msresponse.Fail(
			c,
			http.StatusUnauthorized,
			"Refresh token inválido ou expirado",
			msresponse.ErrorValidacao,
			err.Error(),
		)
		return
	}

	rsp := authgrpc.TokenResponse{
		AccessToken: data.AccessToken,
	}

	msresponse.OK(c, http.StatusOK, "Token renovado com sucesso", rsp)
}

// Login autentica o usuário e retorna access token e refresh token.
// @Summary Login do usuário
// @Description Valida usuário e senha e retorna access token e refresh token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Dados de login"
// @Success 200 {object} msresponse.APIResponse "Login realizado com sucesso"
// @Failure 400 {object} msresponse.APIResponse "Formato inválido ou campos obrigatórios ausentes"
// @Failure 401 {object} msresponse.APIResponse "Usuário ou senha inválidos"
// @Failure 500 {object} msresponse.APIResponse "Erro interno ao autenticar usuário"
// @Router /auth/login [post]
func (obj *AuthHandlerType) Login(c *gin.Context) {
	var body authgrpc.LoginRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Formato inválido",
			msresponse.ErrorFormatoInvalido,
			err.Error(),
		)
		return
	}

	body.Username = strings.TrimSpace(body.Username)
	body.Password = strings.TrimSpace(body.Password)

	if body.Username == "" || body.Password == "" {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Usuário e senha são obrigatórios",
			msresponse.ErrorValidacao,
			"Os campos username e password são obrigatórios.",
		)
		return
	}

	data, err := obj.authClient.Login(c.Request.Context(), body.Username, body.Password)
	if err != nil {
		msresponse.Fail(
			c,
			http.StatusUnauthorized,
			"Senha inválido ou usuário incorreto",
			msresponse.ErrorValidacao,
			err.Error(),
		)
		return
	}

	rsp := authgrpc.LoginResponse{
		AccessToken:  data.AccessToken,
		RefreshToken: data.RefreshToken,
	}

	msresponse.OK(c, http.StatusOK, "Login realizado com sucesso", rsp)
}

// Logout encerra a sessão de forma orientativa.
// @Summary Logout do usuário
// @Description Realiza logout em modelo stateless.
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} msresponse.APIResponse "Logout realizado com sucesso"
// @Router /auth/logout [post]
func (obj *AuthHandlerType) Logout(c *gin.Context) {
	msresponse.OK(c, http.StatusOK, "Logout realizado com sucesso", nil)
}
