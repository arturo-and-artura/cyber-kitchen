package main

import (
	"testing"

	"github.com/arturo-and-artura/cyber-kitchen/internal/agent"
)

func TestConfiguredModelIsOptional(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "")
	model, err := configuredModel()
	if err != nil {
		t.Fatalf("configure without deepseek: %v", err)
	}
	if model != nil {
		t.Fatalf("model = %T, want nil", model)
	}

	t.Setenv("DEEPSEEK_API_KEY", "test-key")
	model, err = configuredModel()
	if err != nil {
		t.Fatalf("configure deepseek: %v", err)
	}
	if _, ok := model.(*agent.DeepSeekModel); !ok {
		t.Fatalf("model = %T", model)
	}
}
