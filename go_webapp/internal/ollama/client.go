package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"zachbot/internal/config"
)

// HTTPClient is a thin wrapper around Ollama's HTTP API.
type HTTPClient struct {
	baseURL         string
	model           string
	reloadModelName string
	reloadModelPath string
	logger          *logrus.Logger
	httpClient      *http.Client
	streamClient    *http.Client
}

// NewHTTPClient builds a configured Ollama client.
func NewHTTPClient(cfg config.Config, logger *logrus.Logger) *HTTPClient {
	return &HTTPClient{
		baseURL:         cfg.APIBaseURL,
		model:           cfg.Model,
		reloadModelName: cfg.ReloadModelName,
		reloadModelPath: cfg.ReloadModelPath,
		logger:          logger,
		httpClient: &http.Client{
			Timeout: cfg.APITimeout,
		},
		streamClient: &http.Client{
			Timeout: 0,
		},
	}
}

func (c *HTTPClient) generateURL() string { return c.baseURL + "/api/generate" }
func (c *HTTPClient) createURL() string   { return c.baseURL + "/api/create" }
func (c *HTTPClient) tagsURL() string     { return c.baseURL + "/api/tags" }

func (c *HTTPClient) Generate(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	payload, err := c.marshalGeneratePayload(req, false)
	if err != nil {
		return ChatResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.generateURL(), bytes.NewReader(payload))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("create generate request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("call generate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body := readBody(resp.Body)
		return ChatResponse{}, fmt.Errorf("generate failed: status=%d body=%s", resp.StatusCode, body)
	}

	var out ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return ChatResponse{}, fmt.Errorf("decode generate response: %w", err)
	}
	return out, nil
}

func (c *HTTPClient) Stream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk) error) error {
	payload, err := c.marshalGeneratePayload(req, true)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.generateURL(), bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create stream request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.streamClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("call stream: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body := readBody(resp.Body)
		return fmt.Errorf("stream failed: status=%d body=%s", resp.StatusCode, body)
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var chunk StreamChunk
		if err := json.Unmarshal(line, &chunk); err != nil {
			c.logger.WithError(err).Warn("failed to decode stream chunk")
			continue
		}

		if err := onChunk(chunk); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read stream: %w", err)
	}
	return nil
}

func (c *HTTPClient) ReloadModel(ctx context.Context) error {
	payload := map[string]string{
		"model": c.reloadModelName,
		"path":  c.reloadModelPath,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal reload payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.createURL(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create reload request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call reload: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("reload failed: status=%d body=%s", resp.StatusCode, readBody(resp.Body))
	}

	return nil
}

func (c *HTTPClient) Tags(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.tagsURL(), nil)
	if err != nil {
		return "", fmt.Errorf("create tags request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call tags: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read tags response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("tags failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return string(body), nil
}

func (c *HTTPClient) marshalGeneratePayload(req ChatRequest, stream bool) ([]byte, error) {
	prompt := formatPrompt(req.Prompt, req.Persona)
	modelName := c.model
	if strings.TrimSpace(req.Model) != "" {
		modelName = strings.TrimSpace(req.Model)
	}

	payload := map[string]interface{}{
		"model":  modelName,
		"prompt": prompt,
		"stream": stream,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal generate payload: %w", err)
	}
	return body, nil
}

func formatPrompt(prompt, persona string) string {
	prompt = strings.TrimSpace(prompt)
	persona = strings.TrimSpace(persona)
	if persona == "" {
		return prompt
	}

	return fmt.Sprintf("Persona: %s\n\nUser: %s", persona, prompt)
}

func readBody(r io.Reader) string {
	body, err := io.ReadAll(r)
	if err != nil {
		return fmt.Sprintf("<failed to read body: %v>", err)
	}
	return strings.TrimSpace(string(body))
}
