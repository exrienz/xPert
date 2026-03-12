package config

import (
	"os"
	"strconv"
)

type Config struct {
	Host                string
	Port                string
	DataPath            string
	StorageBackend      string
	StorageDSN          string
	LLMProvider         string
	OpenAIBaseURL       string
	OpenAIAPIKey        string
	OpenAIModel         string
	MaxParallelSections int
	MaxJobAttempts      int
	DefaultWordCount    int
}

func Load() Config {
	return Config{
		Host:                envOrDefault("DOCGEN_HOST", "0.0.0.0"),
		Port:                envOrDefault("DOCGEN_PORT", "8080"),
		DataPath:            envOrDefault("DOCGEN_DATA_PATH", "./data/docgen.sqlite"),
		StorageBackend:      envOrDefault("DOCGEN_STORAGE_BACKEND", "sqlite"),
		StorageDSN:          envOrDefault("DOCGEN_STORAGE_DSN", ""),
		LLMProvider:         envOrDefault("DOCGEN_LLM_PROVIDER", "mock"),
		OpenAIBaseURL:       envOrDefault("DOCGEN_OPENAI_BASE_URL", "https://api.openai.com/v1"),
		OpenAIAPIKey:        envOrDefault("DOCGEN_OPENAI_API_KEY", ""),
		OpenAIModel:         envOrDefault("DOCGEN_OPENAI_MODEL", "gpt-4o-mini"),
		MaxParallelSections: envInt("DOCGEN_MAX_PARALLEL_SECTIONS", 4),
		MaxJobAttempts:      envInt("DOCGEN_MAX_JOB_ATTEMPTS", 2),
		DefaultWordCount:    envInt("DOCGEN_DEFAULT_WORD_COUNT", 6000),
	}
}

func (c Config) Address() string {
	return c.Host + ":" + c.Port
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
