package config

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Gin
	GinMode string

	// Servidor
	ServerPort           string // aceita "9001" ou ":9001"
	ServerHost           string
	UploadTimeoutSeconds int

	// Postgres
	PgHost     string
	PgPort     string
	PgDB       string
	PgUser     string
	PgPass     string
	PGPoolSize int

	// MongoDB
	MdHost string
	MdPort string
	MdDB   string
	MdUser string
	MdPass string

	// JWT
	JWTSecretKey       string
	AccessTokenExpire  time.Duration
	RefreshTokenExpire time.Duration

	// OpenAI
	OpenApiKey                    string
	OpenOptionMaxCompletionTokens int
	OpenOptionModelTop            string
	OpenOptionModel               string
	OpenOptiondomainecundary      string
	OpenOptionTimeoutSeconds      int

	// OpenSearch
	OpenSearchHost     string // ex: http://192.168.0.30
	OpenSearchPort     string // ex: 9200
	OpenSearchUser     string
	OpenSearchPassword string
	OpenSearchRagName  string

	// CORS
	AllowedOrigins []string

	// App mode
	ApplicationMode string // development|staging|production
}

var (
	GlobalConfig   *Config
	onceLoadConfig sync.Once
	loadErr        error
)

func LoadConfig() (*Config, error) {
	onceLoadConfig.Do(func() {
		loadDotEnvIfPresent()

		cfg := &Config{}
		if err := initEnv(cfg); err != nil {
			loadErr = err
			return
		}
		GlobalConfig = cfg

		if shouldPrintConfig(cfg) {
			showEnv(cfg)
		}
	})

	return GlobalConfig, loadErr
}

func loadDotEnvIfPresent() {
	// Procura .env no cwd
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Printf("Erro ao carregar .env: %v (seguindo com variáveis de ambiente)", err)
		} else {
			log.Println(".env carregado")
		}
		return
	}

	// Procura .env no diretório do executável (deploys)
	exe, err := os.Executable()
	if err != nil {
		return
	}
	dir := filepath.Dir(exe)
	p := filepath.Join(dir, ".env")
	if _, err := os.Stat(p); err == nil {
		if err := godotenv.Load(p); err != nil {
			log.Printf("Erro ao carregar %s: %v", p, err)
		} else {
			log.Printf("✔️  .env carregado de %s", p)
		}
	}
}

// Helpers -----------------------------

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvRequired(key string) (string, error) {
	v, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(v) == "" {
		return "", fmt.Errorf("variável de ambiente obrigatória %s não definida", key)
	}
	return strings.TrimSpace(v), nil
}

func getEnvRequiredIf(cond bool, key string) (string, error) {
	if !cond {
		// não é required nesse modo
		return strings.TrimSpace(getEnv(key, "")), nil
	}
	return getEnvRequired(key)
}

func parseInt(key, val string, def, min, max int) int {
	val = strings.TrimSpace(val)
	if val == "" {
		return def
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("%s inválido (%q), usando default=%d: %v", key, val, def, err)
		return def
	}
	if n < min {
		log.Printf("%s=%d abaixo do mínimo (%d). Usando %d.", key, n, min, min)
		return min
	}
	if n > max {
		log.Printf("%s=%d acima do máximo (%d). Usando %d.", key, n, max, max)
		return max
	}
	return n
}

// Aceita "15" (minutos) ou durações do Go: "15m", "2h"
func parseDurationFlexible(key, val string, def time.Duration) time.Duration {
	val = strings.TrimSpace(val)
	if val == "" {
		return def
	}
	if d, err := time.ParseDuration(val); err == nil {
		if d < 0 {
			log.Printf("%s negativo (%q). Usando default=%s.", key, val, def)
			return def
		}
		return d
	}
	if n, err := strconv.Atoi(val); err == nil {
		if n < 0 {
			log.Printf("%s negativo (%q). Usando default=%s.", key, val, def)
			return def
		}
		return time.Duration(n) * time.Minute
	}
	log.Printf("%s inválido (%q), usando default=%s", key, val, def)
	return def
}

func normalizePort(p string, fallback string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		p = strings.TrimSpace(fallback)
	}
	if p == "" {
		return "", errors.New("porta vazia")
	}
	if p[0] == ':' {
		p = p[1:]
	}
	// valida faixa (1..65535)
	n, err := strconv.Atoi(p)
	if err != nil || n < 1 || n > 65535 {
		return "", fmt.Errorf("porta inválida: %q", p)
	}
	return ":" + strconv.Itoa(n), nil
}

func normalizeURLHost(h string) (string, error) {
	h = strings.TrimSpace(h)
	if h == "" {
		return "", errors.New("host vazio")
	}
	u, err := url.Parse(h)
	if err != nil {
		return "", fmt.Errorf("host inválido: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("host deve incluir esquema http(s)://, recebido: %q", h)
	}
	if u.Host == "" {
		return "", fmt.Errorf("host sem endereço: %q", h)
	}
	return strings.TrimRight(h, "/"), nil
}

func splitAndTrimCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		pp := strings.TrimSpace(p)
		if pp != "" {
			out = append(out, pp)
		}
	}
	return out
}

func mask(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "(vazio)"
	}
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + strings.Repeat("*", len(s)-4) + s[len(s)-2:]
}

func normalizeGinMode(m string) string {
	m = strings.TrimSpace(strings.ToLower(m))
	switch m {
	case "debug", "release", "test":
		return m
	default:
		log.Printf("GIN_MODE inválido (%q). Usando release.", m)
		return "release"
	}
}

func normalizeAppMode(m string) string {
	m = strings.TrimSpace(strings.ToLower(m))
	switch m {
	case "development", "staging", "production":
		return m
	case "":
		return "production"
	default:
		log.Printf("APPLICATION_MODE inválido (%q). Usando production.", m)
		return "production"
	}
}

func isProduction(cfg *Config) bool { return cfg != nil && cfg.ApplicationMode == "production" }

func shouldPrintConfig(cfg *Config) bool {
	// PRINT_CONFIG=1 força imprimir em qualquer modo
	v := strings.TrimSpace(strings.ToLower(getEnv("PRINT_CONFIG", "")))
	if v == "1" || v == "true" || v == "yes" {
		return true
	}
	// default: só imprime em development
	return cfg != nil && cfg.ApplicationMode == "development"
}

func validateHostOrIP(h string) bool {
	h = strings.TrimSpace(h)
	if h == "" {
		return false
	}
	// aceita IP literal ou hostname simples
	if net.ParseIP(h) != nil {
		return true
	}
	// hostname “básico” (sem ser muito rígido)
	return len(h) >= 1 && len(h) <= 253
}

// ------------------------------------

func initEnv(cfg *Config) error {
	// App mode (definir cedo para regras "required em prod")
	cfg.ApplicationMode = normalizeAppMode(getEnv("APPLICATION_MODE", "production"))

	// Gin
	cfg.GinMode = normalizeGinMode(getEnv("GIN_MODE", "release"))

	// CORS
	originsRaw := getEnv("CORS_ORIGINS_ALLOWED", "http://localhost:4012")
	if strings.TrimSpace(originsRaw) == "" {
		log.Println("ℹ️  CORS_ORIGINS_ALLOWED vazio. Usando fallback: http://localhost:4012")
		cfg.AllowedOrigins = []string{"http://localhost:4012"}
	} else {
		list := splitAndTrimCSV(originsRaw)
		if len(list) == 0 {
			log.Println("CORS_ORIGINS_ALLOWED definido, mas sem valores válidos. Usando fallback: http://localhost:4012")
			cfg.AllowedOrigins = []string{"http://localhost:4012"}
		} else {
			if len(list) == 1 && list[0] == "*" {
				log.Println("CORS com '*' detectado. Atenção: com credenciais habilitadas, alguns middlewares rejeitam '*'.")
			}
			cfg.AllowedOrigins = list
		}
	}

	// Servidor
	var err error
	cfg.ServerPort, err = normalizePort(getEnv("SERVER_PORT", ""), "4010")
	//cfg.ServerPort = normalizePort(getEnv("SERVER_PORT", "4003"))
	if err != nil {
		return fmt.Errorf("SERVER_PORT inválido: %w", err)
	}
	cfg.ServerHost = strings.TrimSpace(getEnv("SERVER_HOST", "localhost"))
	if !validateHostOrIP(cfg.ServerHost) {
		log.Printf("SERVER_HOST inválido (%q). Usando localhost.", cfg.ServerHost)
		cfg.ServerHost = "localhost"
	}
	cfg.UploadTimeoutSeconds = parseInt(
		"UPLOAD_TIMEOUT_SECONDS",
		getEnv("UPLOAD_TIMEOUT_SECONDS", "300"),
		300, 60, 600, // eu ampliaria teto p/ 10 min; ajuste como preferir
	)

	// Postgres
	cfg.PgHost = strings.TrimSpace(getEnv("PG_HOST", "192.168.0.30"))
	cfg.PgPort = strings.TrimSpace(getEnv("PG_PORT", "5432"))
	// valida porta PG (opcional mas útil)
	if _, err := normalizePort(cfg.PgPort, "5432"); err != nil {
		return fmt.Errorf("PG_PORT inválido: %w", err)
	}
	//cfg.PgPort = getEnv("PG_PORT", "5432")
	cfg.PgDB = strings.TrimSpace(getEnv("PG_DB", "assjurdb"))
	cfg.PgUser = strings.TrimSpace(getEnv("PG_USER", "assjurpg"))

	// PG_PASS: required em production (recomendado sempre)
	if cfg.PgPass, err = getEnvRequiredIf(isProduction(cfg), "PG_PASS"); err != nil {
		return err
	}

	// MongoDB
	cfg.MdHost = strings.TrimSpace(getEnv("MD_HOST", "localhost"))
	cfg.MdPort = strings.TrimSpace(getEnv("MD_PORT", "27017"))
	// valida porta PG (opcional mas útil)
	if _, err := normalizePort(cfg.MdPort, "27017"); err != nil {
		return fmt.Errorf("MD_PORT inválido: %w", err)
	}

	cfg.MdDB = strings.TrimSpace(getEnv("MD_DB", "books"))
	cfg.MdUser = strings.TrimSpace(getEnv("MD_USER", "root"))

	// PG_PASS: required em production (recomendado sempre)
	if cfg.MdPass, err = getEnvRequiredIf(isProduction(cfg), "MD_PASS"); err != nil {
		return err
	}

	// OpenSearch
	rawOSHost := getEnv("OPENSEARCH_HOST", "http://192.168.0.30")
	if cfg.OpenSearchHost, err = normalizeURLHost(rawOSHost); err != nil {
		return fmt.Errorf("OPENSEARCH_HOST inválido: %v", err)
	}
	cfg.OpenSearchPort = strings.TrimSpace(getEnv("OPENSEARCH_PORT", "9200"))
	if _, err := normalizePort(cfg.OpenSearchPort, "9200"); err != nil {
		return fmt.Errorf("OPENSEARCH_PORT inválido: %v", err)
	}
	//cfg.OpenSearchPort = getEnv("OPENSEARCH_PORT", "9200")
	cfg.OpenSearchUser = strings.TrimSpace(getEnv("OPENSEARCH_USER", "admin"))

	// OPENSEARCH_PASSWORD: NÃO ter fallback; required em production
	if cfg.OpenSearchPassword, err = getEnvRequiredIf(isProduction(cfg), "OPENSEARCH_PASSWORD"); err != nil {
		return err
	}

	cfg.OpenSearchRagName = strings.TrimSpace(getEnv("OPENSEARCH_RAG_NAME", "base_doc_embedding"))

	// OpenAI timeout
	cfg.OpenOptionTimeoutSeconds = parseInt(
		"OPENAI_OPTION_TIMEOUT_SECONDS",
		getEnv("OPENAI_OPTION_TIMEOUT_SECONDS", "120"),
		120, 30, 600,
	)

	// JWT: required em production
	if cfg.JWTSecretKey, err = getEnvRequiredIf(isProduction(cfg), "JWT_SECRET"); err != nil {
		return err
	}

	// OpenAI: required em production
	if cfg.OpenApiKey, err = getEnvRequiredIf(isProduction(cfg), "OPENAI_API_KEY"); err != nil {
		return err
	}
	cfg.OpenOptionMaxCompletionTokens = parseInt(
		"OPENAI_OPTION_MAX_COMPLETION_TOKENS",
		getEnv("OPENAI_OPTION_MAX_COMPLETION_TOKENS", "16384"),
		16384, 256, 128000,
	)

	cfg.OpenOptionModel = strings.TrimSpace(getEnv("OPENAI_OPTION_MODEL", "gpt-5-mini"))
	cfg.OpenOptionModelTop = strings.TrimSpace(getEnv("OPENAI_OPTION_MODEL_TOP", ""))
	if cfg.OpenOptionModelTop == "" {
		cfg.OpenOptionModelTop = cfg.OpenOptionModel
	}
	cfg.OpenOptiondomainecundary = strings.TrimSpace(getEnv("OPENAI_OPTION_MODEL_SECUNDARY", "gpt-5-mini"))

	// Pool do DB
	cfg.PGPoolSize = parseInt("PG_POOLSIZE", getEnv("PG_POOLSIZE", "25"), 25, 5, 200)

	// Expiração de tokens
	cfg.AccessTokenExpire = parseDurationFlexible("ACCESSTOKEN_EXPIRE", getEnv("ACCESSTOKEN_EXPIRE", "10m"), 10*time.Minute)
	cfg.RefreshTokenExpire = parseDurationFlexible("REFRESHTOKEN_EXPIRE", getEnv("REFRESHTOKEN_EXPIRE", "60m"), 60*time.Minute)

	return nil
}

func showEnv(cfg *Config) {
	fmt.Println("--------- CONFIG ---------")
	fmt.Println("APPLICATION_MODE:", cfg.ApplicationMode)

	fmt.Println("GIN_MODE:", cfg.GinMode)
	fmt.Println("SERVER_HOST:", cfg.ServerHost)
	fmt.Println("SERVER_PORT:", cfg.ServerPort)
	fmt.Println("UPLOAD_TIMEOUT_SECONDS:", cfg.UploadTimeoutSeconds)

	fmt.Println("JWT_SECRET:", cfg.JWTSecretKey)

	fmt.Println("POSTGRES_HOST:", cfg.PgHost)
	fmt.Println("POSTGRES_PORT:", cfg.PgPort)
	fmt.Println("POSTGRES_DB:", cfg.PgDB)
	fmt.Println("POSTGRES_USER:", cfg.PgUser)
	fmt.Println("POSTGRES_PASSWORD:", mask(cfg.PgPass))

	fmt.Println("OPENSEARCH_HOST:", cfg.OpenSearchHost)
	fmt.Println("OPENSEARCH_PORT:", cfg.OpenSearchPort)
	fmt.Println("OPENSEARCH_USER:", cfg.OpenSearchUser)
	fmt.Println("OPENSEARCH_PASSWORD:", mask(cfg.OpenSearchPassword))
	fmt.Println("OPENSEARCH_RAG_NAME:", cfg.OpenSearchRagName)

	fmt.Println("OPENAI_OPTION_TIMEOUT_SECONDS:", cfg.OpenOptionTimeoutSeconds)
	fmt.Println("OPENAI_API_KEY:", mask(cfg.OpenApiKey))
	fmt.Println("OPENAI_OPTION_MODEL_TOP:", cfg.OpenOptionModelTop)
	fmt.Println("OPENAI_OPTION_MODEL:", cfg.OpenOptionModel)
	fmt.Println("OPENAI_OPTION_MODEL_SECUNDARY:", cfg.OpenOptiondomainecundary)
	fmt.Println("OPENAI_OPTION_MAX_COMPLETION_TOKENS:", cfg.OpenOptionMaxCompletionTokens)

	fmt.Println("PG_POOLSIZE:", cfg.PGPoolSize)
	fmt.Println("ACCESS_TOKEN_EXPIRE:", cfg.AccessTokenExpire)
	fmt.Println("REFRESH_TOKEN_EXPIRE:", cfg.RefreshTokenExpire)

	fmt.Println("CORS_ORIGINS_ALLOWED:", strings.Join(cfg.AllowedOrigins, ","))
	fmt.Println("--------------------------")
}
