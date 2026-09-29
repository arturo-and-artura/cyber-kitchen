package main

import (
	"os"
	"testing"

	"github.com/arturo-and-artura/cyber-kitchen/internal/agent"
)

func TestConfiguredModelReadsOptionalKeyFile(t *testing.T) {
	missingPath := t.TempDir() + "/missing"
	model, err := configuredModel(missingPath)
	if err != nil {
		t.Fatalf("configure without deepseek: %v", err)
	}
	if model != nil {
		t.Fatalf("model = %T, want nil", model)
	}

	keyPath := t.TempDir() + "/deepseek-api-key"
	if err := os.WriteFile(keyPath, []byte("test-key\n"), 0o600); err != nil {
		t.Fatalf("write key file: %v", err)
	}
	model, err = configuredModel(keyPath)
	if err != nil {
		t.Fatalf("configure deepseek: %v", err)
	}
	if _, ok := model.(*agent.DeepSeekModel); !ok {
		t.Fatalf("model = %T", model)
	}
}
