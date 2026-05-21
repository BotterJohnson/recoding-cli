package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Provider.Model != "gpt-4o" {
		t.Errorf("expected default model gpt-4o, got %s", cfg.Provider.Model)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	os.Setenv("RECODING_API_KEY", "test-key")
	defer os.Unsetenv("RECODING_API_KEY")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Provider.APIKey != "test-key" {
		t.Errorf("expected api_key test-key, got %s", cfg.Provider.APIKey)
	}
}
