package llm

import (
	"errors"
	"log"
	"math/rand"
	"sync"
)

// ModelSelector manages model pool and selection strategy.
type ModelSelector struct {
	models       []string
	enabled      bool
	defaultModel string
	mu           sync.Mutex
	usageStats   map[string]int
}

// NewModelSelector creates a selector from config.
func NewModelSelector(defaultModel string, pool []string, enabled bool) *ModelSelector {
	return &ModelSelector{
		models:       pool,
		enabled:      enabled,
		defaultModel: defaultModel,
		usageStats:   make(map[string]int),
	}
}

// Select returns a model to use. If pool is enabled, returns random selection.
// Otherwise returns the default model.
func (s *ModelSelector) Select() string {
	if !s.enabled || len(s.models) == 0 {
		return s.defaultModel
	}

	model := s.models[rand.Intn(len(s.models))]

	s.mu.Lock()
	s.usageStats[model]++
	s.mu.Unlock()

	log.Printf("[llm] selected model: %s", model)
	return model
}

// SelectExcluding returns a model excluding the specified ones (for fallback).
func (s *ModelSelector) SelectExcluding(exclude map[string]bool) (string, error) {
	if !s.enabled || len(s.models) == 0 {
		if exclude[s.defaultModel] {
			return "", errors.New("no models available for fallback")
		}
		return s.defaultModel, nil
	}

	available := make([]string, 0, len(s.models))
	for _, m := range s.models {
		if !exclude[m] {
			available = append(available, m)
		}
	}

	if len(available) == 0 {
		return "", errors.New("all models exhausted")
	}

	model := available[rand.Intn(len(available))]

	s.mu.Lock()
	s.usageStats[model]++
	s.mu.Unlock()

	log.Printf("[llm] fallback model selected: %s", model)
	return model, nil
}

// GetModels returns the list of available models.
func (s *ModelSelector) GetModels() []string {
	if !s.enabled || len(s.models) == 0 {
		return []string{s.defaultModel}
	}
	return s.models
}

// Stats returns usage statistics.
func (s *ModelSelector) Stats() map[string]int {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make(map[string]int, len(s.usageStats))
	for k, v := range s.usageStats {
		result[k] = v
	}
	return result
}

// ModelCount returns the number of available models.
func (s *ModelSelector) ModelCount() int {
	if !s.enabled || len(s.models) == 0 {
		return 1
	}
	return len(s.models)
}
