package api_test

import (
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

func TestPilotEditingRecommendationAndConfirmationFlow(t *testing.T) {
	initial := seed.InitialState()
	service := domain.NewService(store.NewMemory(initial), func() time.Time { return time.Date(2026, 9, 28, 20, 0, 0, 0, time.UTC) }, func() string { return "history-pilot" })
	handler := api.New(service, api.Config{}, agent.New(nil, initial.Meals))

	requests := []struct {
		method, path, body string
		want               int
	}{
		{http.MethodPut, "/api/v1/household", `{"constraints":["Peanut-free"],"goals":["Use produce first"]}`, 200},
		{http.MethodPut, "/api/v1/inventory/tomato", `{"name":"Tomato","amount":4,"unit":"","category":"Produce","lowAt":1}`, 200},
		{http.MethodPost, "/api/v1/recommendations/generate", ``, 200},
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
