/*
---------------------------------------------------------------------------------------
File: msclientehttp.go
Autor: Aldenor
Data: 13-05-2026
Alteração: 13-05-2026
---------------------------------------------------------------------------------------
A entidade  ClienteHTTP realiza a comunicação com os microsserviços disponíveis,  reali-
zando chamadas HTTP por meio da rede interna do docker. Esta biblioteca é a base para a
construção das entidades relacionadas a cada microsserviço utilizado.

type ClienteHTTP struct { // size=48 (0x30)

	    name       string
	    baseURL    string
	    httpClient *http.Client
	    debug      bool
	}

Métodos disponíveis:

func (c *ClienteHTTP) DeleteJSON(ctx context.Context, path string, headers map[string]string, out any) error
func (c *ClienteHTTP) DoJSON(ctx context.Context, method string, path string, headers map[string]string, payload any, out any) error
func (c *ClienteHTTP) GetJSON(ctx context.Context, path string, headers map[string]string, out any) error
func (c *ClienteHTTP) PostJSON(ctx context.Context, path string, headers map[string]string, payload any, out any) error
func (c *ClienteHTTP) PutJSON(ctx context.Context, path string, headers map[string]string, payload any, out any) error
func (c *ClienteHTTP) buildURL(path string) string
*/
package msclientehttp

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type ClienteHTTP struct {
	name       string
	baseURL    string
	httpClient *http.Client
	debug      bool
}

type ConfigClienteHTTP struct {
	Name               string
	BaseURL            string
	Timeout            time.Duration
	Debug              bool
	InsecureSkipVerify bool
}

type ErrorResponse struct {
	StatusCode int
	Body       string
	Message    string
}

func (e *ErrorResponse) Error() string {
	if e == nil {
		return ""
	}

	if e.Message != "" {
		return e.Message
	}

	return fmt.Sprintf("serviço retornou HTTP %d: %s", e.StatusCode, e.Body)
}

func (c *ClienteHTTP) validate() error {
	if c == nil || c.httpClient == nil {
		return errors.New("authclient não inicializado")
	}

	return nil
}

func New(cfg ConfigClienteHTTP) (*ClienteHTTP, error) {
	name := strings.TrimSpace(cfg.Name)
	if name == "" {
		name = "remote-service"
	}

	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		return nil, errors.New("baseURL não informada")
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()

	if cfg.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec
		}
	}

	return &ClienteHTTP{
		name:    name,
		baseURL: baseURL,
		debug:   cfg.Debug,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}, nil
}

func MustNew(cfg ConfigClienteHTTP) *ClienteHTTP {
	client, err := New(cfg)
	if err != nil {
		panic(err)
	}

	return client
}

func (c *ClienteHTTP) GetJSON(
	ctx context.Context,
	path string,
	headers map[string]string,
	out any,
) error {
	return c.DoJSON(ctx, http.MethodGet, path, headers, nil, out)
}

func (c *ClienteHTTP) PostJSON(
	ctx context.Context,
	path string,
	headers map[string]string,
	payload any,
	out any,
) error {
	return c.DoJSON(ctx, http.MethodPost, path, headers, payload, out)
}

func (c *ClienteHTTP) PutJSON(
	ctx context.Context,
	path string,
	headers map[string]string,
	payload any,
	out any,
) error {
	return c.DoJSON(ctx, http.MethodPut, path, headers, payload, out)
}

func (c *ClienteHTTP) DeleteJSON(
	ctx context.Context,
	path string,
	headers map[string]string,
	out any,
) error {
	return c.DoJSON(ctx, http.MethodDelete, path, headers, nil, out)
}

func (c *ClienteHTTP) DoJSON(
	ctx context.Context,
	method string,
	path string,
	headers map[string]string,
	payload any,
	out any,
) error {

	if err := c.validate(); err != nil {
		return err
	}

	url := c.buildURL(path)

	var body io.Reader

	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("[%s] erro ao serializar payload: %w", c.name, err)
		}

		body = bytes.NewReader(raw)

		if c.debug {
			log.Printf("[%s] request.body=%s", c.name, string(raw))
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("[%s] erro ao criar request para %s: %w", c.name, url, err)
	}

	req.Header.Set("Accept", "application/json")

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for key, value := range headers {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if key != "" && value != "" {
			req.Header.Set(key, value)
		}
	}

	if c.debug {
		log.Printf("[%s] request.method=%s", c.name, method)
		log.Printf("[%s] request.url=%s", c.name, url)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("[%s] erro ao chamar %s: %w", c.name, url, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("[%s] erro ao ler resposta: %w", c.name, err)
	}

	if c.debug {
		log.Printf("[%s] response.status=%d", c.name, resp.StatusCode)
		log.Printf("[%s] response.body=%s", c.name, string(respBody))
	}

	if len(respBody) > 0 && out != nil {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("[%s] erro ao decodificar resposta: %w. body=%s", c.name, err, string(respBody))
		}
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return &ErrorResponse{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
			Message:    fmt.Sprintf("[%s] serviço retornou HTTP %d", c.name, resp.StatusCode),
		}
	}

	return nil
}

func (c *ClienteHTTP) buildURL(path string) string {
	return strings.TrimRight(c.baseURL, "/") + "/" + strings.TrimLeft(path, "/")
}

func BearerHeader(token string) map[string]string {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil
	}

	return map[string]string{
		"Authorization": "Bearer " + token,
	}
}
