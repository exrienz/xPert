package llm

import (
	"testing"
)

func TestModelSelectorDisabled(t *testing.T) {
	selector := NewModelSelector("default-model", nil, false)

	for i := 0; i < 10; i++ {
		if got := selector.Select(); got != "default-model" {
			t.Errorf("Select() = %v, want default-model", got)
		}
	}
}

func TestModelSelectorEnabled(t *testing.T) {
	pool := []string{"model-a", "model-b", "model-c"}
	selector := NewModelSelector("default", pool, true)

	selections := make(map[string]int)
	for i := 0; i < 100; i++ {
		model := selector.Select()
		selections[model]++
	}

	// Verify all pool models were selected (statistically very likely with 100 tries)
	for _, m := range pool {
		if selections[m] == 0 {
			t.Errorf("Model %s was never selected in 100 tries", m)
		}
	}

	// Default should not be selected when pool is enabled
	if selections["default"] > 0 {
		t.Error("Default model should not be selected when pool is enabled")
	}
}

func TestModelSelectorSelectExcluding(t *testing.T) {
	pool := []string{"model-a", "model-b", "model-c"}
	selector := NewModelSelector("default", pool, true)

	exclude := map[string]bool{"model-a": true, "model-b": true}

	for i := 0; i < 10; i++ {
		model, err := selector.SelectExcluding(exclude)
		if err != nil {
			t.Fatalf("SelectExcluding failed: %v", err)
		}
		if model != "model-c" {
			t.Errorf("Expected model-c, got %s", model)
		}
	}
}

func TestModelSelectorSelectExcludingDisabled(t *testing.T) {
	selector := NewModelSelector("default-model", nil, false)

	// When disabled and default not excluded, should return default
	model, err := selector.SelectExcluding(map[string]bool{})
	if err != nil {
		t.Fatalf("SelectExcluding failed: %v", err)
	}
	if model != "default-model" {
		t.Errorf("Expected default-model, got %s", model)
	}

	// When disabled and default is excluded, should return error
	_, err = selector.SelectExcluding(map[string]bool{"default-model": true})
	if err == nil {
		t.Error("Expected error when default model is excluded and pool is disabled")
	}
}

func TestModelSelectorAllExhausted(t *testing.T) {
	pool := []string{"model-a", "model-b"}
	selector := NewModelSelector("default", pool, true)

	exclude := map[string]bool{"model-a": true, "model-b": true}

	_, err := selector.SelectExcluding(exclude)
	if err == nil {
		t.Error("Expected error when all models exhausted")
	}
}

func TestModelSelectorStats(t *testing.T) {
	pool := []string{"model-a", "model-b"}
	selector := NewModelSelector("default", pool, true)

	// Select some models
	for i := 0; i < 10; i++ {
		selector.Select()
	}

	stats := selector.Stats()

	// Should have some selections recorded
	total := 0
	for _, count := range stats {
		total += count
	}
	if total != 10 {
		t.Errorf("Expected 10 total selections, got %d", total)
	}
}

func TestModelSelectorGetModels(t *testing.T) {
	tests := []struct {
		name         string
		defaultModel string
		pool         []string
		enabled      bool
		wantCount    int
	}{
		{
			name:         "disabled returns default",
			defaultModel: "default",
			pool:         []string{"a", "b"},
			enabled:      false,
			wantCount:    1,
		},
		{
			name:         "enabled returns pool",
			defaultModel: "default",
			pool:         []string{"a", "b", "c"},
			enabled:      true,
			wantCount:    3,
		},
		{
			name:         "enabled with empty pool returns default",
			defaultModel: "default",
			pool:         nil,
			enabled:      true,
			wantCount:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := NewModelSelector(tt.defaultModel, tt.pool, tt.enabled)
			models := selector.GetModels()
			if len(models) != tt.wantCount {
				t.Errorf("GetModels() returned %d models, want %d", len(models), tt.wantCount)
			}
		})
	}
}

func TestModelSelectorModelCount(t *testing.T) {
	tests := []struct {
		name      string
		pool      []string
		enabled   bool
		wantCount int
	}{
		{
			name:      "disabled",
			pool:      []string{"a", "b"},
			enabled:   false,
			wantCount: 1,
		},
		{
			name:      "enabled with pool",
			pool:      []string{"a", "b", "c"},
			enabled:   true,
			wantCount: 3,
		},
		{
			name:      "enabled empty pool",
			pool:      nil,
			enabled:   true,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := NewModelSelector("default", tt.pool, tt.enabled)
			if got := selector.ModelCount(); got != tt.wantCount {
				t.Errorf("ModelCount() = %d, want %d", got, tt.wantCount)
			}
		})
	}
}
