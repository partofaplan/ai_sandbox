package ollama

import "context"

// ChatRequest represents the payload coming from Zachbot's API consumers.
type ChatRequest struct {
	Prompt  string `json:"prompt"`
	Model   string `json:"model,omitempty"`
	Persona string `json:"persona,omitempty"`
}

// ChatResponse mirrors the response emitted by the Ollama generate endpoint.
type ChatResponse struct {
	Response string `json:"response"`
}

// StreamChunk captures a single NDJSON object returned by the Ollama stream API.
type StreamChunk struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// Client abstracts the Ollama API surface that Zachbot needs.
type Client interface {
	Generate(ctx context.Context, req ChatRequest) (ChatResponse, error)
	Stream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk) error) error
	ReloadModel(ctx context.Context) error
	Tags(ctx context.Context) (string, error)
}
