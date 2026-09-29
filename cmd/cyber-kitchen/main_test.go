package main

import (
	"testing"

	"github.com/arturo-and-artura/cyber-kitchen/internal/agent"
)

func TestConfiguredModelRequiresAPIKey(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "")
	if _, err := configuredModel(); err == nil {
		t.Fatalf("deepseek without API key was accepted")
	}

	t.Setenv("DEEPSEEK_API_KEY", "test-key")
	model, err := configuredModel()
	if err != nil {
		t.Fatalf("configure deepseek: %v", err)
	}
	if _, ok := model.(*agent.DeepSeekModel); !ok {
		t.Fatalf("model = %T", model)
	}
}
