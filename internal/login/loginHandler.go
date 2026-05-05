/*
---------------------------------------------------------------------------------------
File: loginHandler.go
Autor: Aldenor
Data: 04-05-2026
Alteração: 04-05-2026
---------------------------------------------------------------------------------------
*/
package login

import (
	auth "microsrv/internal/pkg/msauth"
	"microsrv/internal/pkg/mslogger"
	"microsrv/internal/pkg/msresponse"
	"microsrv/internal/services"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type LoginHandlerType struct {
	service       *services.UserService
	jwt           *auth.JWTService
	refreshExpire time.Duration
	accessExpire  time.Duration
}

func NewLoginHandlers(service *services.UserService, jwt *auth.JWTService, refresh_expire time.Duration, access_expire time.Duration) *LoginHandlerType {
	return &LoginHandlerType{
		service:       service,
		jwt:           jwt,
		refreshExpire: refresh_expire,
		accessExpire:  access_expire,
	}
}

// VerifyToken verifica se o access token ainda é válido.
// @Summary Verificar token
// @Description Verifica se um access token JWT é válido e retorna os dados principais das claims.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body TokenRequest true "Token JWT"
// @Success 200 {object} msresponse.APIResponse "Login realizado com sucesso"
// @Failure 400 {object} msresponse.APIResponse "Formato inválido ou campos obrigatórios ausentes"
// @Failure 401 {object} msresponse.APIResponse "Usuário ou senha inválidos"
// @Router /auth/token/verify [post]
func (obj *LoginHandlerType) VerifyToken(c *gin.Context) {

	var body TokenRequest

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

	//claims, err := obj.jwt.ValidateString(body.Token)
	claims, err := obj.jwt.ParseTokenByType(body.Token, auth.TokenTypeAccess)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Access token inválido: %v", err)

		msresponse.Fail(
			c,
			http.StatusUnauthorized,
			"Token inválido ou expirado",
			msresponse.ErrorTokenInvalido,
			"O token informado é inválido, expirou ou não é um access token.",
		)
		return
	}

	rsp := VerifyTokenResponseData{
		ID:    claims.ID,
		Name:  claims.Name,
		Email: claims.Email,
		Role:  claims.Role,
		Exp:   claims.ExpiresAt.Time.Unix(),
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
// @Success 200 {object} msresponse.APIResponse "Login realizado com sucesso"
// @Failure 400 {object} msresponse.APIResponse "Formato inválido ou campos obrigatórios ausentes"
// @Failure 401 {object} msresponse.APIResponse "Usuário ou senha inválidos"
// @Failure 500 {object} msresponse.APIResponse "Erro interno ao gerar token"
// @Router /auth/token/refresh [post]
func (obj *LoginHandlerType) RefreshToken(c *gin.Context) {
	var body TokenRequest

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
			"Refresh token não enviado",
			msresponse.ErrorTokenInvalido,
			"O campo token é obrigatório.",
		)
		return
	}

	//claims, err := obj.jwt.ValidateString(body.Token)
	claims, err := obj.jwt.ParseTokenByType(body.Token, auth.TokenTypeRefresh)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Refresh token inválido: %v", err)

		msresponse.Fail(
			c,
			http.StatusUnauthorized,
			"Refresh token inválido ou expirado",
			msresponse.ErrorTokenInvalido,
			"O token informado é inválido, expirou ou não é um refresh token.",
		)
		return
	}

	accessToken, err := obj.jwt.GenerateToken(
		claims.ID,
		claims.Name,
		claims.Email,
		claims.Role,
		auth.TokenTypeAccess,
		obj.accessExpire,
	)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao gerar access token: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao gerar token",
			msresponse.ErrorInterno,
			"Não foi possível gerar o access token.",
		)
		return
	}

	rsp := RefreshTokenResponseData{
		AccessToken: accessToken,
	}

	msresponse.OK(c, http.StatusOK, "Token renovado com sucesso", rsp)
}

// Login autentica o usuário e retorna access token e refresh token.
// @Summary Login do usuário
// @Description Valida usuário e senha e retorna um access token e um refresh token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Dados de login"
// @Success 200 {object} msresponse.APIResponse "Login realizado com sucesso"
// @Failure 400 {object} msresponse.APIResponse "Formato inválido ou campos obrigatórios ausentes"
// @Failure 401 {object} msresponse.APIResponse "Usuário ou senha inválidos"
// @Failure 500 {object} msresponse.APIResponse "Erro interno ao gerar token"
// @Router /auth/login [post]
func (obj *LoginHandlerType) Login(c *gin.Context) {
	var body LoginRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		msresponse.Fail(
			c,
			http.StatusBadRequest,
			"Formato inválido",
			msresponse.ErrorFormatoInvalido,
			"Informe username e password em formato JSON válido.",
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

	usr, err := obj.service.SelectUserByName(c.Request.Context(), body.Username)
	if err != nil || usr == nil {
		mslogger.LoggerGlobal.Errorf("Falha na autenticação do usuário %q: %v", body.Username, err)

		msresponse.Fail(
			c,
			http.StatusUnauthorized,
			"Usuário ou senha inválidos",
			msresponse.ErrorTokenInvalido,
			"Credenciais inválidas.",
		)
		return
	}

	/* Valida a senha do usuário */
	if !auth.CheckPassword(body.Password, usr.Password) {
		mslogger.LoggerGlobal.Warnf("Falha de autenticação para o usuário %q", body.Username)

		msresponse.Fail(
			c,
			http.StatusUnauthorized,
			"Usuário ou senha inválidos",
			msresponse.ErrorTokenInvalido,
			"Credenciais inválidas.",
		)
		return
	}

	accessToken, err := obj.jwt.GenerateToken(
		uint(usr.UserID),
		usr.Username,
		usr.Email,
		usr.Userrole,
		auth.TokenTypeAccess,
		obj.accessExpire,
	)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao gerar access token: %v", err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao gerar token",
			msresponse.ErrorInterno,
			"Não foi possível gerar o access token.",
		)
		return
	}

	refreshToken, err := obj.jwt.GenerateToken(
		uint(usr.UserID),
		usr.Username,
		usr.Email,
		usr.Userrole,
		auth.TokenTypeRefresh,
		obj.refreshExpire,
	)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao gerar refresh token para o usuário %q: %v", usr.Username, err)

		msresponse.Fail(
			c,
			http.StatusInternalServerError,
			"Erro ao gerar refresh token",
			msresponse.ErrorInterno,
			"Não foi possível gerar o refresh token.",
		)
		return
	}

	rsp := LoginResponseData{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	msresponse.OK(c, http.StatusOK, "Login realizado com sucesso", rsp)
}

// Logout encerra a sessão de forma orientativa.
// @Summary Logout do usuário
// @Description Realiza logout em modelo stateless. Como o JWT é stateless, a invalidação efetiva depende da política do cliente ou de blacklist no backend.
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} msresponse.APIResponse "Login realizado com sucesso"
// @Router /auth/logout [post]
func (obj *LoginHandlerType) Logout(c *gin.Context) {
	msresponse.OK(c, http.StatusOK, "Logout realizado com sucesso", nil)
}
