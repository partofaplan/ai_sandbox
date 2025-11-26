package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"zachbot/internal/config"
	"zachbot/internal/ollama"
	"zachbot/internal/penpal"
)

func main() {
	cfg := penpal.LoadConfig()
	llmCfg := config.Load()

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	client := ollama.NewHTTPClient(llmCfg, logger)
	service := penpal.NewService(cfg, client, logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := service.Run(ctx); err != nil {
		logger.WithError(err).Fatal("penpal service exited")
		os.Exit(1)
	}
}
