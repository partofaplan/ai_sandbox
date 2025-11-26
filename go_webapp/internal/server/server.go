package server

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"zachbot/internal/config"
	"zachbot/internal/ollama"
)

// Server wires HTTP routes to the Ollama client.
type Server struct {
	cfg    config.Config
	client ollama.Client
	logger *logrus.Logger
	router *gin.Engine
}

func New(cfg config.Config, client ollama.Client, logger *logrus.Logger) *Server {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins: cfg.AllowOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	r.Static("/static", cfg.StaticDir)

	templatePattern := filepath.Join(cfg.TemplateDir, "*")
	if matches, _ := filepath.Glob(templatePattern); len(matches) > 0 {
		r.LoadHTMLGlob(templatePattern)
	} else {
		logger.Warnf("No templates found at %s; skipping HTML templates", templatePattern)
	}

	s := &Server{
		cfg:    cfg,
		client: client,
		logger: logger,
		router: r,
	}
	s.registerRoutes()
	return s
}

// Router returns the underlying Gin router for tests and embedding.
func (s *Server) Router() *gin.Engine {
	return s.router
}

func (s *Server) registerRoutes() {
	s.router.GET("/", s.homeHandler)
	s.router.GET("/health/", s.healthCheckHandler)
	s.router.GET("/models/", s.modelsHandler)
	s.router.GET("/models", s.modelsHandler)
	s.router.POST("/chat/", s.chatHandler)
	s.router.POST("/chat/stream/", s.chatStreamHandler)
	s.router.POST("/reload-model/", s.reloadModelHandler)
	s.router.GET("/test-ollama/", s.testOllamaHandler)
}

func (s *Server) homeHandler(c *gin.Context) {
	s.logger.Infof("Serving home page on port %s", s.cfg.ServerPort)
	c.HTML(http.StatusOK, "index.html", nil)
}

func (s *Server) healthCheckHandler(c *gin.Context) {
	if _, err := s.client.Tags(c.Request.Context()); err != nil {
		s.logger.WithError(err).Error("Ollama health check failed")
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "healthy", "ollama_connection": "ok"})
}

func (s *Server) chatHandler(c *gin.Context) {
	var req ollama.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Prompt) == "" {
		s.logger.Error("Invalid or missing prompt in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prompt is required"})
		return
	}

	resp, err := s.client.Generate(c.Request.Context(), req)
	if err != nil {
		s.logger.WithError(err).Error("Error communicating with Ollama API")
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to reach Ollama"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (s *Server) chatStreamHandler(c *gin.Context) {
	var req ollama.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Prompt) == "" {
		s.logger.Error("Invalid or missing prompt in stream request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prompt is required"})
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		s.logger.Error("Streaming not supported by response writer")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	c.Writer.Header().Set("Content-Type", "application/x-ndjson")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Status(http.StatusOK)

	if err := s.client.Stream(c.Request.Context(), req, func(chunk ollama.StreamChunk) error {
		b, err := json.Marshal(chunk)
		if err != nil {
			return err
		}
		if _, err := c.Writer.Write(append(b, '\n')); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}); err != nil {
		s.logger.WithError(err).Error("Error streaming from Ollama")
	}
}

func (s *Server) reloadModelHandler(c *gin.Context) {
	if err := s.client.ReloadModel(c.Request.Context()); err != nil {
		s.logger.WithError(err).Error("Error reloading model")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload model"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Model reloaded successfully"})
}

func (s *Server) testOllamaHandler(c *gin.Context) {
	body, err := s.client.Tags(c.Request.Context())
	if err != nil {
		s.logger.WithError(err).Error("Error communicating with Ollama API")
		c.JSON(http.StatusBadGateway, gin.H{"error": "Request failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "response": body})
}

func (s *Server) modelsHandler(c *gin.Context) {
	body, err := s.client.Tags(c.Request.Context())
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch models from Ollama")
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to fetch models"})
		return
	}

	var parsed struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		s.logger.WithError(err).Warn("Failed to parse model list; returning raw string")
		c.JSON(http.StatusOK, gin.H{"models": []string{body}})
		return
	}

	names := make([]string, 0, len(parsed.Models))
	for _, m := range parsed.Models {
		if trimmed := strings.TrimSpace(m.Name); trimmed != "" {
			names = append(names, trimmed)
		}
	}

	c.JSON(http.StatusOK, gin.H{"models": names})
}
