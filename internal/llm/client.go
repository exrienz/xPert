package llm

type Client interface {
	Generate(systemPrompt, userPrompt string) (string, error)
}
