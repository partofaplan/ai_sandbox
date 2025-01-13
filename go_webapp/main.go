package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ChatRequest struct {
	Prompt string `json:"prompt"`
}

type ChatResponse struct {
	Response string `json:"response"`
}

var (
	HOST   = getEnv("OLLAMA_HOST", "ollama")
	PORT   = getEnv("OLLAMA_PORT", "11434")
	MODEL  = getEnv("OLLAMA_MODEL", "prospector:latest")
	TIMEOUT, _ = time.ParseDuration(getEnv("API_TIMEOUT", "300s"))

	API_BASE_URL     = "http://" + HOST + ":" + PORT
	CREATE_API_URL   = API_BASE_URL + "/api/create"
	GENERATE_API_URL = API_BASE_URL + "/api/generate"

	logger = logrus.New()
	client = &http.Client{Timeout: TIMEOUT}
)

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func main() {
	r := gin.Default()

	// CORS settings
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*", "http://localhost:6600"},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	logger.SetFormatter(&logrus.JSONFormatter{})

	// Serve static files (CSS, JS, images)
	r.Static("/static", "./static")

	// Load templates
	r.LoadHTMLGlob("templates/*")

	// Routes
	r.GET("/", homeHandler)
	r.GET("/health/", healthCheckHandler)
	r.POST("/chat/", chatHandler)
	r.POST("/reload-model/", reloadModelHandler)
	r.GET("/test-ollama/", testOllamaHandler)

	r.Run(":6600")
}

func homeHandler(c *gin.Context) {
	logger.Info("Serving home page on port 6600")
	c.HTML(http.StatusOK, "index.html", nil)
}

func healthCheckHandler(c *gin.Context) {
	logger.Infof("Sending health check request to: %s/api/tags", API_BASE_URL)
	req, err := http.NewRequest("GET", API_BASE_URL+"/api/tags", nil)
	if err != nil {
		logger.Error("Error creating health check request: ", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error performing health check request: ", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "ollama_connection": "ok"})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "ollama_connection": "failed"})
	}
}

func chatHandler(c *gin.Context) {
	var request ChatRequest
	if err := c.ShouldBindJSON(&request); err != nil || request.Prompt == "" {
		logger.Error("Invalid or missing prompt in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prompt is required"})
		return
	}

	payload := map[string]interface{}{
		"model":  MODEL,
		"prompt": request.Prompt,
		"stream": false,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Error("Error marshaling payload: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request payload"})
		return
	}

	req, err := http.NewRequest("POST", GENERATE_API_URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		logger.Error("Error creating request: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error communicating with Ollama API: ", err)
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Request timed out"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading response body: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		logger.Error("Error parsing response: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response format"})
		return
	}

	c.JSON(http.StatusOK, chatResp)
}

func reloadModelHandler(c *gin.Context) {
	payload := map[string]string{
		"model": "prospector",
		"path":  "/root/models/Modelfile",
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Error("Error marshaling payload: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request payload"})
		return
	}

	req, err := http.NewRequest("POST", CREATE_API_URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		logger.Error("Error creating request: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		logger.Error("Error reloading model: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload model"})
		return
	}
	defer resp.Body.Close()

	c.JSON(http.StatusOK, gin.H{"message": "Model reloaded successfully"})
}

func testOllamaHandler(c *gin.Context) {
	req, err := http.NewRequest("GET", API_BASE_URL+"/api/tags", nil)
	if err != nil {
		logger.Error("Error creating test request: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create test request"})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error communicating with Ollama API: ", err)
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Request timed out"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading response body: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "response": string(body)})
}
