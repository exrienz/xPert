package llm

import (
	"strings"

	"xpert/internal/config"
)

// ClientFactory creates LLM clients with specific models.
type ClientFactory struct {
	cfg      config.Config
	provider string
}

// NewClientFactory creates a new client factory.
func NewClientFactory(cfg config.Config) *ClientFactory {
	return &ClientFactory{
		cfg:      cfg,
		provider: strings.ToLower(cfg.LLMProvider),
	}
}

// CreateClient creates a client for the specified model.
func (f *ClientFactory) CreateClient(model string) Client {
	switch f.provider {
	case "", "mock":
		return MockClient{Provider: "mock", Model: model}
	case "openai":
		return NewOpenAIClient(f.cfg.OpenAIBaseURL, f.cfg.OpenAIAPIKey, model, "", "")
	case "openrouter":
		return NewOpenAIClient(f.cfg.OpenAIBaseURL, f.cfg.OpenAIAPIKey, model, "https://openrouter.ai", "xpert")
	case "deepseek", "custom":
		return NewOpenAIClient(f.cfg.OpenAIBaseURL, f.cfg.OpenAIAPIKey, model, "", "")
	case "ollama":
		return NewOllamaClient(f.cfg.OpenAIBaseURL, model)
	default:
		return nil
	}
}

// Provider returns the configured provider name.
func (f *ClientFactory) Provider() string {
	return f.provider
}
