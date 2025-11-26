package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"zachbot/internal/config"
	"zachbot/internal/ollama"
)

type fakeClient struct {
	generateResp ollama.ChatResponse
	generateErr  error

	streamChunks []ollama.StreamChunk
	streamErr    error

	reloadErr error

	tagsResp string
	tagsErr  error
}

func (f *fakeClient) Generate(ctx context.Context, req ollama.ChatRequest) (ollama.ChatResponse, error) {
	return f.generateResp, f.generateErr
}

func (f *fakeClient) Stream(ctx context.Context, req ollama.ChatRequest, onChunk func(ollama.StreamChunk) error) error {
	for _, chunk := range f.streamChunks {
		if err := onChunk(chunk); err != nil {
			return err
		}
	}
	return f.streamErr
}

func (f *fakeClient) ReloadModel(ctx context.Context) error {
	return f.reloadErr
}

func (f *fakeClient) Tags(ctx context.Context) (string, error) {
	return f.tagsResp, f.tagsErr
}

type flushRecorder struct{ *httptest.ResponseRecorder }

func (flushRecorder) Flush() {}

func newTestServer(fc *fakeClient) *Server {
	cfg := config.Config{
		ServerPort:      "6600",
		Model:           "prospector:latest",
		APIBaseURL:      "http://ollama:11434",
		ReloadModelName: "prospector",
		ReloadModelPath: "/root/models/Modelfile",
		AllowOrigins:    []string{"*"},
		StaticDir:       filepath.Join("..", "..", "static"),
		TemplateDir:     filepath.Join("..", "..", "templates"),
	}

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	return New(cfg, fc, logger)
}

func TestHealthCheckHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	srv := newTestServer(&fakeClient{tagsResp: "{}"})

	req := httptest.NewRequest(http.MethodGet, "/health/", nil)
	w := httptest.NewRecorder()

	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["status"] != "healthy" {
		t.Fatalf("expected status healthy, got %q", resp["status"])
	}
}

func TestChatHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	srv := newTestServer(&fakeClient{generateResp: ollama.ChatResponse{Response: "pong"}})

	body := bytes.NewBufferString(`{"prompt":"ping"}`)
	req := httptest.NewRequest(http.MethodPost, "/chat/", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp ollama.ChatResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Response != "pong" {
		t.Fatalf("expected response pong, got %q", resp.Response)
	}
}

func TestChatHandlerValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	srv := newTestServer(&fakeClient{})

	req := httptest.NewRequest(http.MethodPost, "/chat/", bytes.NewBufferString(`{"prompt":""}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestChatStreamHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	chunks := []ollama.StreamChunk{
		{Response: "hel"},
		{Response: "lo", Done: true},
	}
	srv := newTestServer(&fakeClient{streamChunks: chunks})

	req := httptest.NewRequest(http.MethodPost, "/chat/stream/", bytes.NewBufferString(`{"prompt":"hi"}`))
	req.Header.Set("Content-Type", "application/json")
	w := &flushRecorder{httptest.NewRecorder()}

	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	lines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
	if len(lines) != len(chunks) {
		t.Fatalf("expected %d chunks, got %d", len(chunks), len(lines))
	}

	for i, line := range lines {
		var got ollama.StreamChunk
		if err := json.Unmarshal([]byte(line), &got); err != nil {
			t.Fatalf("failed to decode chunk %d: %v", i, err)
		}
		if got != chunks[i] {
			t.Fatalf("unexpected chunk %d: %+v", i, got)
		}
	}
}

func TestReloadModelHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	srv := newTestServer(&fakeClient{})

	req := httptest.NewRequest(http.MethodPost, "/reload-model/", nil)
	w := httptest.NewRecorder()

	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestModelsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tags := `{"models":[{"name":"a"},{"name":"b"}]}`
	srv := newTestServer(&fakeClient{tagsResp: tags})

	req := httptest.NewRequest(http.MethodGet, "/models/", nil)
	w := httptest.NewRecorder()

	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Models []string `json:"models"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Models) != 2 || resp.Models[0] != "a" || resp.Models[1] != "b" {
		t.Fatalf("unexpected models: %+v", resp.Models)
	}
}
