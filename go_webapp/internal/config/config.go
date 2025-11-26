package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

const (
	defaultServerPort   = "6600"
	defaultOllamaHost   = "ollama"
	defaultOllamaPort   = "11434"
	defaultModel        = "prospector:latest"
	defaultReloadName   = "prospector"
	defaultReloadPath   = "/root/models/Modelfile"
	defaultAPITimeout   = "300s"
	defaultAllowOrigins = "*;http://localhost:6600"
	defaultStaticDir    = "static"
	defaultTemplateDir  = "templates"
)

// Config holds runtime configuration for the Zachbot service.
type Config struct {
	ServerPort      string
	Model           string
	APIBaseURL      string
	APITimeout      time.Duration
	ReloadModelName string
	ReloadModelPath string
	AllowOrigins    []string
	StaticDir       string
	TemplateDir     string
}

// Load reads configuration from environment variables, applying defaults when
// values are not present or are invalid.
func Load() Config {
	timeout := parseDuration(env("API_TIMEOUT", defaultAPITimeout))
	host := env("OLLAMA_HOST", defaultOllamaHost)
	port := env("OLLAMA_PORT", defaultOllamaPort)

	return Config{
		ServerPort:      env("SERVER_PORT", defaultServerPort),
		Model:           env("OLLAMA_MODEL", defaultModel),
		APIBaseURL:      fmt.Sprintf("http://%s:%s", host, port),
		APITimeout:      timeout,
		ReloadModelName: env("OLLAMA_RELOAD_NAME", defaultReloadName),
		ReloadModelPath: env("OLLAMA_RELOAD_PATH", defaultReloadPath),
		AllowOrigins:    parseAllowOrigins(env("CORS_ALLOW_ORIGINS", defaultAllowOrigins)),
		StaticDir:       env("STATIC_DIR", defaultStaticDir),
		TemplateDir:     env("TEMPLATE_DIR", defaultTemplateDir),
	}
}

// ListenAddr returns the HTTP listen address for the Gin server.
func (c Config) ListenAddr() string {
	return ":" + c.ServerPort
}

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func parseDuration(raw string) time.Duration {
	dur, err := time.ParseDuration(raw)
	if err != nil {
		log.Printf("Invalid API_TIMEOUT %q, defaulting to %s: %v", raw, defaultAPITimeout, err)
		dur, _ = time.ParseDuration(defaultAPITimeout)
	}
	return dur
}

func parseAllowOrigins(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';'
	})
	if len(parts) == 0 {
		return []string{"*"}
	}
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return []string{"*"}
	}
	return out
}
