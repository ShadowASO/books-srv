/*
---------------------------------------------------------------------------------------
File: authms.go
Autor: Aldenor
Data: 13-05-2026
Alteração: 13-05-2026
---------------------------------------------------------------------------------------
A entidade "ClientAuth" implementar a biblioteca ClienteHTTP para realiza a comunicação
com o microsserviço "auth-srv",  realizando chamadas aos métodos por meio da rede interna
do docker. Esta biblioteca é espécifica para o microsserviço "auth-srv".

type ClientAuth struct { // size=8

	    http *msclientehttp.ClienteHTTP
	}

func (c *ClientAuth) Login(ctx context.Context, username string, password string) (*LoginData, error)
func (c *ClientAuth) RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenData, error)
func (c *ClientAuth) VerifyToken(ctx context.Context, token string) (*VerifyTokenData, error)
func (c *ClientAuth) validate() error

USO:

	authClient, err := authms.New(msclientehttp.ConfigClienteHTTP{
			Name:               cfg.AuthName,
			BaseURL:            cfg.AuthServiceURL,
			Timeout:            5 * time.Second,
			Debug:              cfg.AuthClientDebug,
			InsecureSkipVerify: cfg.AuthInsecureSkipVerify,
		})
		if err != nil {
			panic(err)
		}
*/
package authms

import (
	"context"
	"errors"
	"fmt"
	"microsrv/internal/pkg/msclientehttp"
	"strings"
	"time"
)

const (
	defaultServiceName = "auth-srv"
	defaultTimeout     = 5 * time.Second

	routeLogin        = "/auth/login"
	routeVerifyToken  = "/auth/token/verify"
	routeRefreshToken = "/auth/token/refresh"
)

type ClientAuth struct {
	http *msclientehttp.ClienteHTTP
}

func (c *ClientAuth) validate() error {
	if c == nil || c.http == nil {
		return errors.New("authclient não inicializado")
	}

	return nil
}

func New(cfg msclientehttp.ConfigClienteHTTP) (*ClientAuth, error) {
	httpClient, err := msclientehttp.New(msclientehttp.ConfigClienteHTTP{
		Name:               "auth-srv",
		BaseURL:            cfg.BaseURL,
		Timeout:            cfg.Timeout,
		Debug:              cfg.Debug,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
	})
	if err != nil {
		return nil, err
	}

	return &ClientAuth{
		http: httpClient,
	}, nil
}

func MustNew(cfg msclientehttp.ConfigClienteHTTP) *ClientAuth {
	client, err := New(cfg)
	if err != nil {
		panic(err)
	}

	return client
}

func NewFromURL(baseURL string) (*ClientAuth, error) {
	return New(msclientehttp.ConfigClienteHTTP{
		Name:               defaultServiceName,
		BaseURL:            baseURL,
		Timeout:            defaultTimeout,
		Debug:              false,
		InsecureSkipVerify: false,
	})
}

func (c *ClientAuth) VerifyToken(ctx context.Context, token string) (*TokenClaims, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("token não informado")
	}

	payload := TokenRequest{
		Token: token,
	}

	var resp APIResponse[*TokenClaims]

	if err := c.http.PostJSON(ctx, routeVerifyToken, nil, payload, &resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, buildAPIError(resp.Error, resp.Message, "token inválido")
	}

	if resp.Data == nil {
		return nil, errors.New("auth-srv retornou resposta sem data")
	}

	return resp.Data, nil
}

func (c *ClientAuth) Login(ctx context.Context, username string, password string) (*LoginData, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	username = strings.TrimSpace(username)
	if username == "" {
		return nil, errors.New("username não informado")
	}

	if password == "" {
		return nil, errors.New("password não informado")
	}

	payload := LoginRequest{
		Username: username,
		Password: password,
	}

	var resp APIResponse[*LoginData]

	if err := c.http.PostJSON(ctx, routeLogin, nil, payload, &resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, buildAPIError(resp.Error, resp.Message, "login não autorizado")
	}

	if resp.Data == nil {
		return nil, errors.New("auth-srv retornou resposta sem data")
	}

	return resp.Data, nil
}

func (c *ClientAuth) RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenData, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, errors.New("refresh token não informado")
	}

	payload := TokenRequest{
		Token: refreshToken,
	}

	var resp APIResponse[*RefreshTokenData]

	if err := c.http.PostJSON(ctx, routeRefreshToken, nil, payload, &resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, buildAPIError(resp.Error, resp.Message, "refresh token inválido")
	}

	if resp.Data == nil {
		return nil, errors.New("auth-srv retornou resposta sem data")
	}

	return resp.Data, nil
}

func buildAPIError(apiErr *APIError, message string, fallback string) error {
	if apiErr != nil {
		if strings.TrimSpace(apiErr.Description) != "" {
			return errors.New(apiErr.Description)
		}

		if strings.TrimSpace(apiErr.Message) != "" {
			return errors.New(apiErr.Message)
		}
	}

	if strings.TrimSpace(message) != "" {
		return errors.New(message)
	}

	return fmt.Errorf("%s", fallback)
}
