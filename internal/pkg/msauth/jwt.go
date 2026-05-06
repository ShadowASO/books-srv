/*
---------------------------------------------------------------------------------------
File: jwt.go
Autor: Aldenor
Inspiração: Enterprise Applications with Gin
Data: 03-05-2025
Alteração: 04-05-2026
---------------------------------------------------------------------------------------
*/
package auth

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"microsrv/internal/pkg/mslogger"
	"microsrv/internal/pkg/msresponse"
)

/*
=========================

	Context keys (padronização)

=========================
*/

const (
	CtxUserID    = "user_id"
	CtxUserName  = "user_name"
	CtxUserEmail = "user_email"
	CtxUserRole  = "user_role"
)

/*
=========================

	Claims

=========================
*/
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type Claims struct {
	ID        uint      `json:"user_id"`
	Email     string    `json:"user_email"`
	Role      string    `json:"user_role"`
	Name      string    `json:"user_name"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

/*
=========================

	Serviço JWT

=========================
*/

type JWTService struct {
	secretKey []byte
	issuer    string
	leeway    time.Duration
}

func NewJWTService(secret_key string) *JWTService {
	return &JWTService{
		secretKey: []byte(secret_key),
		issuer:    "assjur",
		leeway:    30 * time.Second,
	}
}

func NewID() string {
	if v7, err := uuid.NewV7(); err == nil {
		return v7.String()
	}

	return uuid.NewString()
}

func (j *JWTService) GenerateToken(id uint, name, email, role string, tokenType TokenType, ttl time.Duration) (string, error) {
	now := time.Now().UTC()

	claims := &Claims{
		ID:        id,
		Email:     strings.TrimSpace(email),
		Role:      normalizeRole(role),
		Name:      strings.TrimSpace(name),
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   strconv.FormatUint(uint64(id), 10),
			ID:        NewID(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-j.leeway)),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(j.secretKey)
}

// ParseToken valida o token JWT.
func (j *JWTService) ParseToken(tokenString string) (*Claims, error) {
	if strings.TrimSpace(tokenString) == "" {
		return nil, errors.New("token vazio")
	}

	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		jwt.WithLeeway(j.leeway),
	)

	token, err := parser.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		return j.secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)

	if !ok || !token.Valid {
		return nil, errors.New("token inválido")
	}

	if claims.Issuer != j.issuer {
		return nil, errors.New("issuer inválido")
	}

	if claims.ExpiresAt == nil {
		return nil, errors.New("token sem expiração")
	}
	if claims.TokenType != TokenTypeAccess && claims.TokenType != TokenTypeRefresh {
		return nil, errors.New("tipo de token ausente ou inválido")
	}

	claims.Role = normalizeRole(claims.Role)

	return claims, nil
}

func (j *JWTService) ParseTokenByType(tokenString string, expectedType TokenType) (*Claims, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != expectedType {
		return nil, errors.New("tipo de token inválido")
	}

	return claims, nil
}

/*
=========================

	Helpers HTTP

=========================
*/

func ExtractBearerToken(authHeader string) (string, error) {
	parts := strings.Fields(authHeader)

	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		if strings.TrimSpace(parts[1]) == "" {
			return "", errors.New("bearer token vazio")
		}

		return parts[1], nil
	}

	return "", errors.New("authorization header inválido: esperado Bearer <token>")
}

func normalizeRole(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

/*
=========================

	Middlewares

=========================
*/

// AuthMiddleware valida JWT, injeta claims no contexto e segue.
func (j *JWTService) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")

		if h == "" {
			//mslogger.LoggerGlobal.Warn("Cabeçalho Authorization ausente")

			msresponse.Fail(
				c,
				http.StatusUnauthorized,
				"Cabeçalho Authorization ausente",
				msresponse.ErrorTokenInvalido,
				"A requisição deve informar o cabeçalho Authorization no formato Bearer <token>.",
			)

			c.Abort()
			return
		}

		tokenStr, err := ExtractBearerToken(h)
		if err != nil {
			//mslogger.LoggerGlobal.Warnf("Token mal formatado: %v", err)

			msresponse.Fail(
				c,
				http.StatusUnauthorized,
				"Token mal formatado",
				msresponse.ErrorTokenInvalido,
				"A requisição deve informar o cabeçalho Authorization no formato Bearer <token>.",
			)

			c.Abort()
			return
		}

		claims, err := j.ParseTokenByType(tokenStr, TokenTypeAccess)
		if err != nil {
			//mslogger.LoggerGlobal.Warnf("Token inválido ou expirado: %v", err)

			msresponse.Fail(
				c,
				http.StatusUnauthorized,
				"Token inválido ou expirado",
				msresponse.ErrorTokenInvalido,
				"O token informado é inválido, expirou ou não é um access token.",
			)

			c.Abort()
			return
		}

		c.Set(CtxUserID, claims.ID)
		c.Set(CtxUserName, claims.Name)
		c.Set(CtxUserEmail, claims.Email)
		c.Set(CtxUserRole, normalizeRole(claims.Role))

		c.Next()
	}
}

// AuthorizeMiddleware verifica se o usuário autenticado possui role autorizada.
// - Se allowedRoles estiver vazio: qualquer usuário autenticado passa.
// - Admin sempre passa.
func (j *JWTService) AuthorizeMiddleware(allowedRoles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedRoles))

	for _, r := range allowedRoles {
		rr := normalizeRole(r)
		if rr != "" {
			allowed[rr] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		roleVal, ok := c.Get(CtxUserRole)
		if !ok {
			mslogger.LoggerGlobal.Error("Usuário não autenticado: role ausente no contexto")

			msresponse.Fail(
				c,
				http.StatusUnauthorized,
				"Usuário não autenticado",
				msresponse.ErrorTokenInvalido,
				"As credenciais do usuário não foram encontradas no contexto da requisição.",
			)

			c.Abort()
			return
		}

		role, _ := roleVal.(string)
		role = normalizeRole(role)

		if role == "" {
			mslogger.LoggerGlobal.Error("Usuário não autenticado: role vazia no contexto")

			msresponse.Fail(
				c,
				http.StatusUnauthorized,
				"Usuário não autenticado",
				msresponse.ErrorTokenInvalido,
				"O perfil do usuário autenticado não foi encontrado.",
			)

			c.Abort()
			return
		}

		if len(allowed) == 0 {
			c.Next()
			return
		}

		if role == "admin" {
			c.Next()
			return
		}

		if _, ok := allowed[role]; ok {
			c.Next()
			return
		}

		mslogger.LoggerGlobal.Infof(
			"Acesso negado: role=%q precisa de uma das roles %v",
			role,
			allowedRoles,
		)

		msresponse.Fail(
			c,
			http.StatusForbidden,
			"Usuário sem permissão suficiente para esta ação",
			msresponse.ErrorNaoAutorizado,
			"O usuário autenticado não possui permissão para acessar este recurso.",
		)

		c.Abort()
	}
}
