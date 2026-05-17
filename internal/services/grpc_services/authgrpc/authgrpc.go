/*
---------------------------------------------------------------------------------------
File: authms.go
Autor: Aldenor
Data: 15-05-2026
---------------------------------------------------------------------------------------
Cliente gRPC específico para comunicação com o microsserviço auth-srv.

Substitui o ClientAuth baseado em REST, preservando os mesmos métodos públicos:

	Login(ctx, username, password)
	VerifyToken(ctx, token)
	RefreshToken(ctx, refreshToken)
*/
package authgrpc

import (
	"context"
	"errors"
	"fmt"

	"microsrv/internal/grpc/pb/authpb"
	"microsrv/internal/pkg/msclientegrpc"
	"strings"
	"time"

	"google.golang.org/grpc/status"
)

const (
	defaultServiceName = "auth-srv"
	defaultTimeout     = 5 * time.Second
)

type ClientAuth struct {
	grpc   *msclientegrpc.ClienteGRPC
	client authpb.AuthServiceClient
}

func New(cfg msclientegrpc.ConfigClienteGRPC) (*ClientAuth, error) {
	if strings.TrimSpace(cfg.Name) == "" {
		cfg.Name = defaultServiceName
	}

	if strings.TrimSpace(cfg.Host) == "" {
		cfg.Host = defaultServiceName
	}

	if cfg.Port <= 0 {
		cfg.Port = 50051
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}

	grpcClient, err := msclientegrpc.New(cfg)
	if err != nil {
		return nil, err
	}

	authClient := authpb.NewAuthServiceClient(grpcClient.Conn())

	return &ClientAuth{
		grpc:   grpcClient,
		client: authClient,
	}, nil
}

func MustNew(cfg msclientegrpc.ConfigClienteGRPC) *ClientAuth {
	client, err := New(cfg)
	if err != nil {
		panic(err)
	}

	return client
}

func (c *ClientAuth) Close() error {
	if c == nil || c.grpc == nil {
		return nil
	}

	return c.grpc.Close()
}

func (c *ClientAuth) validate() error {
	if c == nil || c.grpc == nil || c.client == nil {
		return errors.New("authclient gRPC não inicializado")
	}

	if err := c.grpc.Validate(); err != nil {
		return err
	}

	return nil
}

func (c *ClientAuth) Login(
	ctx context.Context,
	username string,
	password string,
) (*LoginData, error) {
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

	rpcCtx, cancel := c.grpc.Context(ctx)
	defer cancel()

	resp, err := c.client.Login(rpcCtx, &authpb.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, normalizeGRPCError(err, "login não autorizado")
	}

	return &LoginData{
		AccessToken:  resp.GetAccessToken(),
		RefreshToken: resp.GetRefreshToken(),
	}, nil
}

func (c *ClientAuth) VerifyToken(
	ctx context.Context,
	token string,
) (*TokenClaims, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("token não informado")
	}

	rpcCtx, cancel := c.grpc.Context(ctx)
	defer cancel()

	resp, err := c.client.VerifyToken(rpcCtx, &authpb.TokenRequest{
		Token: token,
	})
	if err != nil {
		return nil, normalizeGRPCError(err, "token inválido")
	}

	return &TokenClaims{
		ID:    int(resp.GetId()),
		Name:  resp.GetName(),
		Email: resp.GetEmail(),
		Role:  resp.GetRole(),
		Exp:   resp.GetExp(),
	}, nil
}

func (c *ClientAuth) RefreshToken(
	ctx context.Context,
	refreshToken string,
) (*RefreshTokenData, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, errors.New("refresh token não informado")
	}

	rpcCtx, cancel := c.grpc.Context(ctx)
	defer cancel()

	resp, err := c.client.RefreshToken(rpcCtx, &authpb.TokenRequest{
		Token: refreshToken,
	})
	if err != nil {
		return nil, normalizeGRPCError(err, "refresh token inválido")
	}

	return &RefreshTokenData{
		AccessToken: resp.GetAccessToken(),
		TokenType:   resp.GetTokenType(),
		ExpiresIn:   resp.GetExpiresIn(),
	}, nil
}

func normalizeGRPCError(err error, fallback string) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if ok {
		msg := strings.TrimSpace(st.Message())
		if msg != "" {
			return errors.New(msg)
		}
	}

	if strings.TrimSpace(err.Error()) != "" {
		return err
	}

	return fmt.Errorf("%s", fallback)
}
