package llm

import "fmt"

type OllamaClient struct {
	delegate OpenAIClient
}

func NewOllamaClient(baseURL, model string) OllamaClient {
	if baseURL == "" {
		baseURL = "http://127.0.0.1:11434/v1"
	}
	return OllamaClient{
		delegate: NewOpenAIClient(baseURL, "ollama", model, "", ""),
	}
}

func (c OllamaClient) Generate(systemPrompt, userPrompt string) (string, error) {
	result, err := c.delegate.Generate(systemPrompt, userPrompt)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	return result, nil
}
