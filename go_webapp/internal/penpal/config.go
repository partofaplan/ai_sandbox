package penpal

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	MailhogAPI     string
	SMTPHost       string
	SMTPPort       int
	FromAddress    string
	PollInterval   time.Duration
	DefaultModel   string
	DefaultPersona string
}

const (
	defaultMailhogAPI   = "http://mailhog:8025"
	defaultSMTPHost     = "mailhog"
	defaultSMTPPort     = 1025
	defaultFromAddress  = "penpal@zachbot.local"
	defaultPollInterval = "15s"
	defaultModel        = "prospector:latest"
	defaultPersona      = "Write a concise, friendly email reply."
)

func LoadConfig() Config {
	poll := parseDurationEnv("PENPAL_POLL_INTERVAL", defaultPollInterval)
	port := parseIntEnv("PENPAL_SMTP_PORT", defaultSMTPPort)

	return Config{
		MailhogAPI:     getenv("PENPAL_MAILHOG_API", defaultMailhogAPI),
		SMTPHost:       getenv("PENPAL_SMTP_HOST", defaultSMTPHost),
		SMTPPort:       port,
		FromAddress:    getenv("PENPAL_FROM", defaultFromAddress),
		PollInterval:   poll,
		DefaultModel:   getenv("PENPAL_MODEL", defaultModel),
		DefaultPersona: getenv("PENPAL_PERSONA", defaultPersona),
	}
}

func getenv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func parseDurationEnv(key, fallback string) time.Duration {
	raw := getenv(key, fallback)
	d, err := time.ParseDuration(raw)
	if err != nil {
		log.Printf("invalid duration for %s=%q, using %s: %v", key, raw, fallback, err)
		d, _ = time.ParseDuration(fallback)
	}
	return d
}

func parseIntEnv(key string, fallback int) int {
	raw := getenv(key, "")
	if raw == "" {
		return fallback
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		log.Printf("invalid int for %s=%q, using %d: %v", key, raw, fallback, err)
		return fallback
	}
	return val
}
