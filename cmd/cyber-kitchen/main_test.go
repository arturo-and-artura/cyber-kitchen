package main

import (
	"testing"

	"github.com/arturo-and-artura/cyber-kitchen/internal/agent"
)

func TestConfiguredModelIsExplicitAndValidated(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "")
	model, err := configuredModel("")
	if err != nil || model != nil {
		t.Fatalf("disabled model = %T, %v", model, err)
	}
	if _, err := configuredModel("unknown"); err == nil {
		t.Fatalf("unknown provider was accepted")
	}
	if _, err := configuredModel("deepseek"); err == nil {
		t.Fatalf("deepseek without API key was accepted")
	}

	t.Setenv("DEEPSEEK_API_KEY", "test-key")
	model, err = configuredModel("deepseek")
	if err != nil {
		t.Fatalf("configure deepseek: %v", err)
	}
	if _, ok := model.(*agent.DeepSeekModel); !ok {
		t.Fatalf("model = %T", model)
	}
}
