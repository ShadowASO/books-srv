/*
File: msclientegrpc.go
Criado: 16-05-2026
Interface geneŕica para o Cliente do gRPC. Ela deve ser utilizada como base para a criação
dos clientes específicos pelo usuário.
Exemplo:

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
*/
package msclientegrpc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultServiceName = "remote-grpc-service"
	defaultTimeout     = 5 * time.Second
	defaultPort        = 50051
)

type ClienteGRPC struct {
	name    string
	host    string
	port    int
	address string

	conn    *grpc.ClientConn
	timeout time.Duration
	debug   bool
}

type ConfigClienteGRPC struct {
	Name string

	// Host deve conter apenas o nome/IP do serviço.
	// Exemplos:
	//   "auth-srv"
	//   "localhost"
	//   "127.0.0.1"
	Host string

	// Port deve conter apenas a porta numérica.
	// Exemplo:
	//   50051
	Port int

	Timeout time.Duration
	Debug   bool
}

func New(cfg ConfigClienteGRPC) (*ClienteGRPC, error) {
	name := strings.TrimSpace(cfg.Name)
	if name == "" {
		name = defaultServiceName
	}

	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		return nil, errors.New("host gRPC não informado")
	}

	port := cfg.Port
	if port <= 0 {
		port = defaultPort
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	address := net.JoinHostPort(host, strconv.Itoa(port))

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("[%s] erro ao criar cliente gRPC para %s: %w", name, address, err)
	}

	client := &ClienteGRPC{
		name:    name,
		host:    host,
		port:    port,
		address: address,
		conn:    conn,
		timeout: timeout,
		debug:   cfg.Debug,
	}

	if client.debug {
		log.Printf(
			"[%s] cliente gRPC criado para host=%s port=%d address=%s state=%s",
			client.name,
			client.host,
			client.port,
			client.address,
			client.State(),
		)
	}

	return client, nil
}

func MustNew(cfg ConfigClienteGRPC) *ClienteGRPC {
	client, err := New(cfg)
	if err != nil {
		panic(err)
	}

	return client
}

func (c *ClienteGRPC) Conn() *grpc.ClientConn {
	if c == nil {
		return nil
	}

	return c.conn
}

func (c *ClienteGRPC) Timeout() time.Duration {
	if c == nil || c.timeout <= 0 {
		return defaultTimeout
	}

	return c.timeout
}

func (c *ClienteGRPC) Name() string {
	if c == nil || strings.TrimSpace(c.name) == "" {
		return defaultServiceName
	}

	return c.name
}

func (c *ClienteGRPC) Host() string {
	if c == nil {
		return ""
	}

	return c.host
}

func (c *ClienteGRPC) Port() int {
	if c == nil {
		return 0
	}

	return c.port
}

func (c *ClienteGRPC) Address() string {
	if c == nil {
		return ""
	}

	return c.address
}

func (c *ClienteGRPC) Debug() bool {
	if c == nil {
		return false
	}

	return c.debug
}

func (c *ClienteGRPC) State() string {
	if c == nil || c.conn == nil {
		return "nil"
	}

	return c.conn.GetState().String()
}

func (c *ClienteGRPC) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}

	if c.debug {
		log.Printf(
			"[%s] fechando conexão gRPC com host=%s port=%d address=%s state=%s",
			c.name,
			c.host,
			c.port,
			c.address,
			c.State(),
		)
	}

	return c.conn.Close()
}

func (c *ClienteGRPC) Validate() error {
	if c == nil {
		return errors.New("cliente gRPC não inicializado")
	}

	if c.conn == nil {
		return errors.New("conexão gRPC não inicializada")
	}

	if strings.TrimSpace(c.host) == "" {
		return errors.New("host gRPC não configurado")
	}

	if c.port <= 0 {
		return errors.New("porta gRPC inválida")
	}

	if strings.TrimSpace(c.address) == "" {
		return errors.New("endereço gRPC não configurado")
	}

	return nil
}

func (c *ClienteGRPC) Context(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}

	return context.WithTimeout(parent, c.Timeout())
}

func (c *ClienteGRPC) WaitReady(ctx context.Context) error {
	if err := c.Validate(); err != nil {
		return err
	}

	if ctx == nil {
		ctx = context.Background()
	}

	c.conn.Connect()

	for {
		state := c.conn.GetState()

		if state.String() == "READY" {
			return nil
		}

		if !c.conn.WaitForStateChange(ctx, state) {
			return fmt.Errorf(
				"[%s] serviço gRPC indisponível em host=%s port=%d address=%s; último estado: %s",
				c.name,
				c.host,
				c.port,
				c.address,
				state.String(),
			)
		}
	}
}
