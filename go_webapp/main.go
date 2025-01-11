package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
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
	HOST   = getEnv("OLLAMA_HOST", "localhost")
	PORT   = getEnv("OLLAMA_PORT", "11434")
	MODEL  = getEnv("OLLAMA_MODEL", "prospector:latest")
	TIMEOUT, _ = time.ParseDuration(getEnv("API_TIMEOUT", "300s"))

	CREATE_API_URL   = "http://" + HOST + ":" + PORT + "/api/create"
	GENERATE_API_URL = "http://" + HOST + ":" + PORT + "/api/generate"

	logger = logrus.New()
)

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	logger.SetFormatter(&logrus.JSONFormatter{})

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")

	r.GET("/", homeHandler)
	r.GET("/health", healthCheckHandler)
	r.POST("/chat", chatHandler)
	r.POST("/reload-model", reloadModelHandler)
	r.GET("/test-ollama", testOllamaHandler)

	r.Run(":6500")
}

func homeHandler(c *gin.Context) {
	logger.Info("Serving home page")
	c.HTML(http.StatusOK, "index.html", nil)
}

func healthCheckHandler(c *gin.Context) {
	resp, err := http.Get("http://" + HOST + ":" + PORT + "/api/tags")
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
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
	payloadBytes, _ := json.Marshal(payload)

	client := &http.Client{Timeout: TIMEOUT}
	resp, err := client.Post(GENERATE_API_URL, "application/json", ioutil.NopCloser(bytes.NewReader(payloadBytes)))
	if err != nil {
		logger.Error("Error communicating with Ollama API: ", err)
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Request timed out"})
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var chatResp ChatResponse
	json.Unmarshal(body, &chatResp)

	if resp.StatusCode == http.StatusOK {
		c.JSON(http.StatusOK, chatResp)
	} else {
		c.JSON(resp.StatusCode, gin.H{"error": "Failed to get valid response"})
	}
}

func reloadModelHandler(c *gin.Context) {
	payload := map[string]string{
		"model": "prospector",
		"path":  "/root/models/Modelfile",
	}
	payloadBytes, _ := json.Marshal(payload)
	client := &http.Client{Timeout: TIMEOUT}
	resp, err := client.Post(CREATE_API_URL, "application/json", ioutil.NopCloser(bytes.NewReader(payloadBytes)))
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload model"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Model reloaded successfully"})
}

func testOllamaHandler(c *gin.Context) {
	resp, err := http.Get("http://" + HOST + ":" + PORT + "/api/tags")
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	c.JSON(http.StatusOK, gin.H{"status": "success", "response": string(body)})
}