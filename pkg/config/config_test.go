package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	content := `
endpoints:
  - name: openai
    endpoint_url: "https://api.openai.com/v1"
    api_key_env: "OPENAI_API_KEY"
    model: "gpt-4"
    context_window_size: 8192
    chunk_size: 4096
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test loading the config
	cfg, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(cfg.Endpoints) != 1 {
		t.Fatalf("Expected 1 endpoint, got %d", len(cfg.Endpoints))
	}

	endpoint := cfg.Endpoints[0]
	if endpoint.Name != "openai" {
		t.Errorf("Expected endpoint name 'openai', got '%s'", endpoint.Name)
	}
	if endpoint.EndpointURL != "https://api.openai.com/v1" {
		t.Errorf("Expected endpoint URL 'https://api.openai.com/v1', got '%s'", endpoint.EndpointURL)
	}
	if endpoint.APIKeyEnv != "OPENAI_API_KEY" {
		t.Errorf("Expected api key env 'OPENAI_API_KEY', got '%s'", endpoint.APIKeyEnv)
	}
	if endpoint.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%s'", endpoint.Model)
	}
	if endpoint.ContextWindowSize != 8192 {
		t.Errorf("Expected context window size 8192, got %d", endpoint.ContextWindowSize)
	}
	if endpoint.ChunkSize != 4096 {
		t.Errorf("Expected chunk size 4096, got %d", endpoint.ChunkSize)
	}
}
