package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/arturo-and-artura/cyber-kitchen/internal/agent"
	"github.com/arturo-and-artura/cyber-kitchen/internal/api"
	"github.com/arturo-and-artura/cyber-kitchen/internal/domain"
	"github.com/arturo-and-artura/cyber-kitchen/internal/seed"
	"github.com/arturo-and-artura/cyber-kitchen/internal/store"
)

type modelFunc func(context.Context, string) ([]byte, error)

func (f modelFunc) Complete(ctx context.Context, prompt string) ([]byte, error) {
	return f(ctx, prompt)
}

func TestRecommendationCapabilityReturnsFriendlyErrorsWithoutChangingMeals(t *testing.T) {
	initial := seed.InitialState()
	tests := []struct {
		name       string
		runner     *agent.Runner
		wantStatus int
		wantCode   string
		wantText   string
	}{
		{name: "not configured", wantStatus: http.StatusServiceUnavailable, wantCode: "ai_recommendations_unavailable", wantText: "You can still explore and manage your kitchen"},
		{name: "provider failure", runner: agent.New(modelFunc(func(context.Context, string) ([]byte, error) { return nil, errors.New("provider unavailable") })), wantStatus: http.StatusBadGateway, wantCode: "ai_recommendation_failed", wantText: "current meals are unchanged"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := domain.NewService(store.NewMemory(initial), time.Now, func() string { return "unused" })
			handler := api.New(service, api.Config{}, test.runner)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/recommendations/generate", strings.NewReader(`{"locale":"en"}`)))

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d: %s", response.Code, response.Body.String())
			}
			var body struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body.Code != test.wantCode || !strings.Contains(body.Message, test.wantText) {
				t.Fatalf("response = %#v", body)
			}
			if meals := service.State().Meals; len(meals) != len(initial.Meals) || meals[0].ID != initial.Meals[0].ID {
				t.Fatalf("recommendations changed after capability failure")
			}
		})
	}
}

func TestRecommendationRejectsUnsupportedLocale(t *testing.T) {
	initial := seed.InitialState()
	service := domain.NewService(store.NewMemory(initial), time.Now, func() string { return "unused" })
	handler := api.New(service, api.Config{}, agent.New(modelFunc(func(context.Context, string) ([]byte, error) { return nil, nil })))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/recommendations/generate", strings.NewReader(`{"locale":"fr"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "unsupported_locale") {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}

func TestPilotEditingRecommendationAndConfirmationFlow(t *testing.T) {
	initial := seed.InitialState()
	service := domain.NewService(store.NewMemory(initial), func() time.Time { return time.Date(2026, 9, 28, 20, 0, 0, 0, time.UTC) }, func() string { return "history-pilot" })
	model := modelFunc(func(_ context.Context, prompt string) ([]byte, error) {
		if !strings.Contains(prompt, `"locale":"en"`) {
			t.Fatalf("prompt missing locale: %s", prompt)
		}
		return json.Marshal(agent.Response{Recommendations: initial.Meals})
	})
	handler := api.New(service, api.Config{}, agent.New(model))

	requests := []struct {
		method, path, body string
		want               int
	}{
		{http.MethodPut, "/api/v1/household", `{"constraints":["Peanut-free"],"goals":["Use produce first"]}`, 200},
		{http.MethodPut, "/api/v1/inventory/tomato", `{"name":"Tomato","amount":4,"unit":"","category":"Produce","lowAt":1}`, 200},
		{http.MethodPost, "/api/v1/recommendations/generate", `{"locale":"en"}`, 200},
		{http.MethodPost, "/api/v1/meals/miso-salmon/confirm", `{"rating":"loved","note":"Pilot flow"}`, 200},
		{http.MethodDelete, "/api/v1/inventory/tomato", ``, 200},
	}
	for _, step := range requests {
		request := httptest.NewRequest(step.method, step.path, strings.NewReader(step.body))
		if step.body != "" {
			request.Header.Set("Content-Type", "application/json")
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != step.want {
			t.Fatalf("%s %s = %d: %s", step.method, step.path, response.Code, response.Body.String())
		}
	}
	state := service.State()
	if len(state.Household.Goals) != 1 || state.Household.Goals[0] != "Use produce first" {
		t.Fatalf("household edit not committed")
	}
	if len(state.History) != 3 || state.History[0].ID != "history-pilot" {
		t.Fatalf("confirmation not committed")
	}
	if state.Inventory[0].Amount != 0 {
		t.Fatalf("inventory mutation was not deterministic")
	}
	for _, item := range state.Inventory {
		if item.ID == "tomato" {
			t.Fatalf("delete not committed")
		}
	}
}
