package config

import (
	"testing"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "feature enabled with empty pool",
			cfg: Config{
				EnableRandomModelSelection: true,
				AIModelPool:                nil,
			},
			wantErr: true,
		},
		{
			name: "feature disabled with empty pool",
			cfg: Config{
				EnableRandomModelSelection: false,
				AIModelPool:                nil,
			},
			wantErr: false,
		},
		{
			name: "valid pool",
			cfg: Config{
				EnableRandomModelSelection: true,
				AIModelPool:                []string{"gpt-4o", "claude-3"},
			},
			wantErr: false,
		},
		{
			name: "invalid model identifier with spaces",
			cfg: Config{
				EnableRandomModelSelection: true,
				AIModelPool:                []string{"valid-model", "invalid model"},
			},
			wantErr: true,
		},
		{
			name: "invalid model identifier with special chars",
			cfg: Config{
				EnableRandomModelSelection: true,
				AIModelPool:                []string{"valid-model", "invalid!model"},
			},
			wantErr: true,
		},
		{
			name: "valid model with slashes and colons",
			cfg: Config{
				EnableRandomModelSelection: true,
				AIModelPool:                []string{"anthropic/claude-3:latest", "openai/gpt-4o"},
			},
			wantErr: false,
		},
		{
			name: "valid model with underscores",
			cfg: Config{
				EnableRandomModelSelection: true,
				AIModelPool:                []string{"gpt_4o_mini", "claude_3_opus"},
			},
			wantErr: false,
		},
		{
			name: "empty model in pool",
			cfg: Config{
				EnableRandomModelSelection: true,
				AIModelPool:                []string{"valid-model", ""},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidModelIdentifier(t *testing.T) {
	tests := []struct {
		model string
		valid bool
	}{
		{"gpt-4o", true},
		{"gpt-4o-mini", true},
		{"claude-3.5-sonnet", true},
		{"anthropic/claude-3:latest", true},
		{"model_with_underscores", true},
		{"GPT4", true},
		{"", false},
		{"model with spaces", false},
		{"model!invalid", false},
		{"model@invalid", false},
		{"model#invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			if got := isValidModelIdentifier(tt.model); got != tt.valid {
				t.Errorf("isValidModelIdentifier(%q) = %v, want %v", tt.model, got, tt.valid)
			}
		})
	}
}
