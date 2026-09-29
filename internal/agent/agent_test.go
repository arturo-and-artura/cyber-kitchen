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
	}))
	meals, err := runner.Recommend(context.Background(), Context{Household: state.Household, Inventory: state.Inventory, History: state.History})
	if err != nil || len(meals) != 3 {
		t.Fatalf("meals=%d err=%v", len(meals), err)
	}
}

func TestRunnerReportsModelAndValidationErrors(t *testing.T) {
	state := seed.InitialState()
	valid, _ := jsonResponse(Response{Recommendations: state.Meals})
	for _, payload := range [][]byte{[]byte(`{"recommendations":[]}`), []byte(`{"recommendations":[],"toolCall":"read"}`), append(valid, []byte(` {}`)...)} {
		runner := New(modelFunc(func(context.Context, string) ([]byte, error) { return payload, nil }))
		meals, err := runner.Recommend(context.Background(), Context{Household: state.Household, Inventory: state.Inventory})
		if !errors.Is(err, ErrInvalidResponse) || meals != nil {
			t.Fatalf("meals=%v err=%v", meals, err)
		}
	}
	runner := New(modelFunc(func(context.Context, string) ([]byte, error) { return nil, errors.New("offline") }))
	if _, err := runner.Recommend(context.Background(), Context{Inventory: state.Inventory}); err == nil {
		t.Fatalf("provider error was ignored")
	}
	if _, err := New(nil).Recommend(context.Background(), Context{}); err == nil {
		t.Fatalf("missing model was accepted")
	}
}

func jsonResponse(response Response) ([]byte, error) {
	return json.Marshal(response)
}
