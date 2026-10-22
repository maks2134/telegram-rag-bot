package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds runtime settings from the environment.
type Config struct {
	TelegramToken string
	AdminChatID   int64
	LLMBaseURL    string
	LLMAPIKey     string
	LLMModel      string
	DocsDir       string
	Addr          string
}

// Load reads configuration from environment variables.
func Load() (Config, error) {
	cfg := Config{
		TelegramToken: strings.TrimSpace(os.Getenv("TELEGRAM_TOKEN")),
		LLMBaseURL:    env("LLM_BASE_URL", "https://api.deepseek.com"),
		LLMAPIKey:     strings.TrimSpace(os.Getenv("LLM_API_KEY")),
		LLMModel:      env("LLM_MODEL", "deepseek-chat"),
		DocsDir:       env("DOCS_DIR", "./docs"),
		Addr:          env("ADDR", ":8083"),
	}

	if raw := strings.TrimSpace(os.Getenv("ADMIN_CHAT_ID")); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return Config{}, fmt.Errorf("ADMIN_CHAT_ID: %w", err)
		}
		cfg.AdminChatID = id
	}

	if cfg.LLMAPIKey == "" {
		return Config{}, fmt.Errorf("LLM_API_KEY is required")
	}
	return cfg, nil
}

func env(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}
