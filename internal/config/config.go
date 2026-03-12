package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Host                       string
	Port                       string
	DataPath                   string
	StorageBackend             string
	StorageDSN                 string
	LLMProvider                string
	OpenAIBaseURL              string
	OpenAIAPIKey               string
	OpenAIModel                string
	AIModelPool                []string
	EnableRandomModelSelection bool
	MaxParallelSections        int
	MaxJobAttempts             int
	DefaultWordCount           int
}

func Load() Config {
	return Config{
		Host:                       envOrDefault("XPERT_HOST", "0.0.0.0"),
		Port:                       envOrDefault("XPERT_PORT", "8080"),
		DataPath:                   envOrDefault("XPERT_DATA_PATH", "./data/xpert.sqlite"),
		StorageBackend:             envOrDefault("XPERT_STORAGE_BACKEND", "sqlite"),
		StorageDSN:                 envOrDefault("XPERT_STORAGE_DSN", ""),
		LLMProvider:                envOrDefault("XPERT_LLM_PROVIDER", "mock"),
		OpenAIBaseURL:              envOrDefault("XPERT_OPENAI_BASE_URL", "https://api.openai.com/v1"),
		OpenAIAPIKey:               envOrDefault("XPERT_OPENAI_API_KEY", ""),
		OpenAIModel:                envOrDefault("XPERT_OPENAI_MODEL", "gpt-4o-mini"),
		AIModelPool:                envSlice("XPERT_AI_MODEL_POOL", nil),
		EnableRandomModelSelection: envBool("XPERT_ENABLE_RANDOM_MODEL_SELECTION", false),
		MaxParallelSections:        envInt("XPERT_MAX_PARALLEL_SECTIONS", 4),
		MaxJobAttempts:             envInt("XPERT_MAX_JOB_ATTEMPTS", 2),
		DefaultWordCount:           envInt("XPERT_DEFAULT_WORD_COUNT", 6000),
	}
}

func (c Config) Address() string {
	return c.Host + ":" + c.Port
}

func (c Config) Validate() error {
	if c.EnableRandomModelSelection && len(c.AIModelPool) == 0 {
		return errors.New("XPERT_AI_MODEL_POOL must be set when XPERT_ENABLE_RANDOM_MODEL_SELECTION is enabled")
	}
	for _, model := range c.AIModelPool {
		if !isValidModelIdentifier(model) {
			return fmt.Errorf("invalid model identifier in pool: %q", model)
		}
	}
	return nil
}

func isValidModelIdentifier(model string) bool {
	if len(model) == 0 || len(model) > 128 {
		return false
	}
	for _, r := range model {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '.' || r == ':' || r == '/' || r == '_') {
			return false
		}
	}
	return true
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func envSlice(key string, fallback []string) []string {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}

func envBool(key string, fallback bool) bool {
	raw := strings.ToLower(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	return raw == "true" || raw == "1" || raw == "yes"
}
