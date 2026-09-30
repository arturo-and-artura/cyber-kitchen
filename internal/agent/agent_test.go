package agent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/arturo-and-artura/cyber-kitchen/internal/domain"
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
	meals, err := runner.Recommend(context.Background(), Context{Locale: "en", Household: state.Household, Inventory: state.Inventory, History: state.History})
	if err != nil || len(meals) != 3 {
		t.Fatalf("meals=%d err=%v", len(meals), err)
	}
}

func TestRunnerReportsModelAndValidationErrors(t *testing.T) {
	state := seed.InitialState()
	valid, _ := jsonResponse(Response{Recommendations: state.Meals})
	for _, payload := range [][]byte{[]byte(`{"recommendations":[]}`), []byte(`{"recommendations":[],"toolCall":"read"}`), append(valid, []byte(` {}`)...)} {
		runner := New(modelFunc(func(context.Context, string) ([]byte, error) { return payload, nil }))
		meals, err := runner.Recommend(context.Background(), Context{Locale: "en", Household: state.Household, Inventory: state.Inventory})
		if !errors.Is(err, ErrInvalidResponse) || meals != nil {
			t.Fatalf("meals=%v err=%v", meals, err)
		}
	}
	runner := New(modelFunc(func(context.Context, string) ([]byte, error) { return nil, errors.New("offline") }))
	if _, err := runner.Recommend(context.Background(), Context{Locale: "en", Inventory: state.Inventory}); err == nil {
		t.Fatalf("provider error was ignored")
	}
	if _, err := New(nil).Recommend(context.Background(), Context{Locale: "en"}); err == nil {
		t.Fatalf("missing model was accepted")
	}
}

func TestBuildPromptDefinesCompleteMealSchemaAndInventoryRules(t *testing.T) {
	state := seed.InitialState()
	height := 170.5
	count := 2.0
	state.Household.Members = []domain.HouseholdMember{{ID: "member-1", Name: "Member One", HeightCm: &height, Notes: []string{"Smaller portions"}}}
	state.Household.Preferences = []string{"Quick dinners"}
	state.Inventory = []domain.InventoryItem{{
		ID: "rice", Name: "Rice", Amount: 500, Unit: "g", Category: "Pantry", LowAt: 100,
		Count: &count, CountUnit: "bags", Storage: "Pantry", RecordedOn: "2026-09-29", Notes: "Opened",
	}}
	prompt := BuildPrompt(Context{Locale: "en", Household: state.Household, Inventory: state.Inventory})
	for _, required := range []string{
		`"locale":"en or zh-CN matching the requested locale"`,
		`"inventoryId":"exact inventory item id"`,
		`"difficulty":"Easy"`,
		`"ingredients"`,
		`"steps"`,
		"exactly three complete meals",
		"no greater than the available amount",
		"must not introduce unlisted ingredients",
		`"id":"rice"`,
		`"heightCm":170.5`,
		`"preferences":["Quick dinners"]`,
		`"countUnit":"bags"`,
		`"recordedOn":"2026-09-29"`,
	} {
		if !strings.Contains(prompt, required) {
			t.Fatalf("prompt missing %q: %s", required, prompt)
		}
	}
}

func TestRunnerRequiresRequestedLocale(t *testing.T) {
	state := seed.InitialState()
	for i := range state.Meals {
		state.Meals[i].Locale = "zh-CN"
	}
	payload, _ := jsonResponse(Response{Recommendations: state.Meals})
	runner := New(modelFunc(func(_ context.Context, prompt string) ([]byte, error) {
		if !strings.Contains(prompt, `"locale":"zh-CN"`) || !strings.Contains(prompt, "Simplified Chinese") {
			t.Fatalf("localized prompt omitted locale contract: %s", prompt)
		}
		return payload, nil
	}))
	meals, err := runner.Recommend(context.Background(), Context{Locale: "zh-CN", Household: state.Household, Inventory: state.Inventory})
	if err != nil || len(meals) != 3 || meals[0].Locale != "zh-CN" {
		t.Fatalf("meals=%v err=%v", meals, err)
	}
	state.Meals[0].Locale = "en"
	payload, _ = jsonResponse(Response{Recommendations: state.Meals})
	if _, err := runner.Recommend(context.Background(), Context{Locale: "zh-CN", Inventory: state.Inventory}); !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("mismatched locale err=%v", err)
	}
}

func jsonResponse(response Response) ([]byte, error) {
	return json.Marshal(response)
}
