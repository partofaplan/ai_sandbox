package main

import (
	"github.com/sirupsen/logrus"

	"zachbot/internal/config"
	"zachbot/internal/ollama"
	"zachbot/internal/server"
)

func main() {
	cfg := config.Load()

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	client := ollama.NewHTTPClient(cfg, logger)
	srv := server.New(cfg, client, logger)

	logger.Infof("Starting Zachbot on %s", cfg.ListenAddr())
	if err := srv.Router().Run(cfg.ListenAddr()); err != nil {
		logger.WithError(err).Fatal("Server exited with error")
	}
}
