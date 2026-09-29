package agent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/arturo-and-artura/cyber-kitchen/internal/seed"
)

type modelFunc func(context.Context, string) ([]byte, error)

func (f modelFunc) Complete(ctx context.Context, prompt string) ([]byte, error) {
	return f(ctx, prompt)
}

func TestRunnerAcceptsOnlyValidatedKitchenResponse(t *testing.T) {
	state := seed.InitialState()
	valid, _ := jsonResponse(Response{Recommendations: state.Meals})
	runner := New(modelFunc(func(_ context.Context, prompt string) ([]byte, error) {
		if !strings.Contains(prompt, "Peanut-free") || !strings.Contains(prompt, "exactly three") {
			t.Fatalf("prompt omitted application context or contract")
		}
		return valid, nil
	}), nil)
	meals, source := runner.Recommend(context.Background(), Context{Household: state.Household, Inventory: state.Inventory, History: state.History})
	if source != "model" || len(meals) != 3 {
		t.Fatalf("source=%s meals=%d", source, len(meals))
	}
}

func TestRunnerFallsBackWithoutPublishingInvalidOutput(t *testing.T) {
	state := seed.InitialState()
	for _, payload := range [][]byte{[]byte(`{"recommendations":[]}`), []byte(`{"recommendations":[],"toolCall":"read"}`)} {
		runner := New(modelFunc(func(context.Context, string) ([]byte, error) { return payload, nil }), state.Meals)
		meals, source := runner.Recommend(context.Background(), Context{Household: state.Household, Inventory: state.Inventory})
		if source != "fallback" || len(meals) != 3 {
			t.Fatalf("source=%s meals=%d", source, len(meals))
		}
	}
	runner := New(modelFunc(func(context.Context, string) ([]byte, error) { return nil, errors.New("offline") }), state.Meals)
	_, source := runner.Recommend(context.Background(), Context{Inventory: state.Inventory})
	if source != "fallback" {
		t.Fatalf("source=%s", source)
	}
}

func jsonResponse(response Response) ([]byte, error) {
	return json.Marshal(response)
}
