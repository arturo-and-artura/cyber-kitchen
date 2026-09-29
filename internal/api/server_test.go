package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/arturo-and-artura/cyber-kitchen/internal/api"
	"github.com/arturo-and-artura/cyber-kitchen/internal/domain"
	"github.com/arturo-and-artura/cyber-kitchen/internal/seed"
	"github.com/arturo-and-artura/cyber-kitchen/internal/store"
)

func TestConfirmMealAPI(t *testing.T) {
	clock := func() time.Time { return time.Date(2026, 9, 27, 1, 30, 0, 0, time.UTC) }
	service := domain.NewService(store.NewMemory(seed.InitialState()), clock, func() string { return "history-new" })
	handler := api.New(service, api.Config{AllowedOrigins: []string{"http://localhost:5173"}})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/meals/miso-salmon/confirm", strings.NewReader(`{"rating":"loved","note":"Weeknight winner"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "http://localhost:5173")
	confirmResponse := httptest.NewRecorder()
	handler.ServeHTTP(confirmResponse, request)
	if confirmResponse.Code != http.StatusOK {
		t.Fatalf("POST confirm status = %d, want 200: %s", confirmResponse.Code, confirmResponse.Body.String())
	}
	if got := confirmResponse.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("CORS origin = %q", got)
	}
	assertJSONKeys(t, confirmResponse.Body.Bytes(), "inventory", "history", "selectedMealId")
	var after struct {
		Inventory []struct {
			ID     string  `json:"id"`
			Amount float64 `json:"amount"`
		} `json:"inventory"`
		History []struct {
			ID       string `json:"id"`
			MealID   string `json:"mealId"`
			CookedAt string `json:"cookedAt"`
		} `json:"history"`
		SelectedMealID *string `json:"selectedMealId"`
	}
	if err := json.Unmarshal(confirmResponse.Body.Bytes(), &after); err != nil {
		t.Fatalf("decode POST confirm: %v", err)
	}
	if len(after.History) != 3 || after.History[0].ID != "history-new" || after.History[0].MealID != "miso-salmon" || after.History[0].CookedAt != "2026-09-27T01:30:00.000Z" {
		t.Fatalf("unexpected history after confirm: %#v", after.History)
	}
	if after.SelectedMealID != nil {
		t.Errorf("selectedMealId = %v, want null", after.SelectedMealID)
	}
	wantAmounts := map[string]float64{"salmon": 0, "rice": 2, "cucumber": 0, "spinach": 3, "miso": 6}
	for _, item := range after.Inventory {
		if want, ok := wantAmounts[item.ID]; ok && item.Amount != want {
			t.Errorf("%s amount = %v, want %v", item.ID, item.Amount, want)
		}
	}
}

func TestAggregateStateEndpointIsNotAvailable(t *testing.T) {
	service := domain.NewService(store.NewMemory(seed.InitialState()), time.Now, func() string { return "unused" })
	handler := api.New(service, api.Config{})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/state", nil))

	if response.Code != http.StatusNotFound {
		t.Fatalf("GET state status = %d, want 404: %s", response.Code, response.Body.String())
	}
}

func TestConfirmMealErrorsDoNotMutateState(t *testing.T) {
	memory := store.NewMemory(seed.InitialState())
	service := domain.NewService(memory, time.Now, func() string { return "unused" })
	handler := api.New(service, api.Config{})

	for _, test := range []struct {
		name string
		path string
		body string
		want int
	}{
		{name: "unknown meal", path: "/api/v1/meals/unknown/confirm", body: `{"rating":"okay","note":""}`, want: http.StatusNotFound},
		{name: "invalid rating", path: "/api/v1/meals/miso-salmon/confirm", body: `{"rating":"excellent","note":""}`, want: http.StatusBadRequest},
		{name: "unknown field", path: "/api/v1/meals/miso-salmon/confirm", body: `{"rating":"okay","extra":true}`, want: http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.want {
				t.Errorf("status = %d, want %d: %s", response.Code, test.want, response.Body.String())
			}
		})
	}
	state := memory.Snapshot()
	if len(state.History) != 2 || state.Inventory[0].Amount != 2 {
		t.Errorf("rejected requests mutated state: history=%d salmon=%v", len(state.History), state.Inventory[0].Amount)
	}
}

func TestResourceReadAPIs(t *testing.T) {
	service := domain.NewService(store.NewMemory(seed.InitialState()), time.Now, func() string { return "unused" })
	handler := api.New(service, api.Config{})

	tests := []struct {
		name     string
		path     string
		wantKeys []string
		assert   func(*testing.T, []byte)
	}{
		{
			name:     "household",
			path:     "/api/v1/household",
			wantKeys: []string{"name", "members", "constraints", "goals", "preferences"},
			assert: func(t *testing.T, body []byte) {
				var response struct {
					Name    string            `json:"name"`
					Members []json.RawMessage `json:"members"`
				}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("decode household: %v", err)
				}
				if response.Name == "" || len(response.Members) != 3 {
					t.Errorf("unexpected household: %#v", response)
				}
			},
		},
		{
			name:     "inventory",
			path:     "/api/v1/inventory",
			wantKeys: []string{"inventory"},
			assert: func(t *testing.T, body []byte) {
				var response struct {
					Inventory []json.RawMessage `json:"inventory"`
				}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("decode inventory: %v", err)
				}
				if len(response.Inventory) != 13 {
					t.Errorf("inventory length = %d, want 13", len(response.Inventory))
				}
			},
		},
		{
			name:     "meals",
			path:     "/api/v1/meals",
			wantKeys: []string{"meals", "selectedMealId"},
			assert: func(t *testing.T, body []byte) {
				var response struct {
					Meals          []json.RawMessage `json:"meals"`
					SelectedMealID *string           `json:"selectedMealId"`
				}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("decode meals: %v", err)
				}
				if len(response.Meals) != 3 || response.SelectedMealID != nil {
					t.Errorf("unexpected meals response: %#v", response)
				}
			},
		},
		{
			name:     "history",
			path:     "/api/v1/history",
			wantKeys: []string{"history"},
			assert: func(t *testing.T, body []byte) {
				var response struct {
					History []struct {
						CookedAt string `json:"cookedAt"`
					} `json:"history"`
				}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("decode history: %v", err)
				}
				if len(response.History) != 2 || response.History[0].CookedAt != "2026-09-24T18:30:00.000Z" {
					t.Errorf("unexpected history: %#v", response.History)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want 200: %s", test.path, response.Code, response.Body.String())
			}
			if got := response.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}
			assertJSONKeys(t, response.Body.Bytes(), test.wantKeys...)
			test.assert(t, response.Body.Bytes())
		})
	}
}

func TestHouseholdProfileAPIReplacesEditableFieldsAndKeepsNameServerOwned(t *testing.T) {
	initial := seed.InitialState()
	originalName := initial.Household.Name
	memory := store.NewMemory(initial)
	service := domain.NewService(memory, time.Now, func() string { return "unused" })
	handler := api.New(service, api.Config{})
	body := `{"members":[{"id":"member-1","name":"Member One","initials":"MO","heightCm":171.5,"notes":["Prefers mild food"]}],"constraints":["No peanuts"],"goals":["Reduce waste"],"preferences":["Fast dinners"]}`

	request := httptest.NewRequest(http.MethodPut, "/api/v1/household", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("PUT household status = %d: %s", response.Code, response.Body.String())
	}
	assertJSONKeys(t, response.Body.Bytes(), "name", "members", "constraints", "goals", "preferences")
	var got struct {
		Name        string   `json:"name"`
		Preferences []string `json:"preferences"`
		Members     []struct {
			HeightCm *float64 `json:"heightCm"`
			Notes    []string `json:"notes"`
		} `json:"members"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode household response: %v", err)
	}
	if got.Name != originalName || len(got.Preferences) != 1 || len(got.Members) != 1 || got.Members[0].HeightCm == nil || *got.Members[0].HeightCm != 171.5 || len(got.Members[0].Notes) != 1 {
		t.Fatalf("household response = %#v", got)
	}

	// The household name is intentionally absent from the write contract.
	rejected := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPut, "/api/v1/household", strings.NewReader(`{"name":"Client override","members":[],"constraints":[],"goals":[],"preferences":[]}`))
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(rejected, request)
	if rejected.Code != http.StatusBadRequest || memory.Snapshot().Household.Name != originalName || len(memory.Snapshot().Household.Members) != 1 {
		t.Fatalf("server-owned name request status=%d state=%#v", rejected.Code, memory.Snapshot().Household)
	}
}

func TestInventoryMetadataAPIAndStrictValidation(t *testing.T) {
	memory := store.NewMemory(seed.InitialState())
	service := domain.NewService(memory, time.Now, func() string { return "unused" })
	handler := api.New(service, api.Config{})
	valid := `{"name":"Brown rice","amount":725.5,"unit":"g","category":"Pantry","lowAt":200,"count":1.5,"countUnit":"bags","storage":"Pantry shelf","recordedOn":"2026-09-29","notes":"Opened package"}`

	request := httptest.NewRequest(http.MethodPut, "/api/v1/inventory/brown-rice", strings.NewReader(valid))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("PUT inventory status = %d: %s", response.Code, response.Body.String())
	}
	var body struct {
		Inventory []map[string]json.RawMessage `json:"inventory"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode inventory response: %v", err)
	}
	item := body.Inventory[len(body.Inventory)-1]
	for _, key := range []string{"id", "name", "amount", "unit", "category", "lowAt", "count", "countUnit", "storage", "recordedOn", "notes"} {
		if _, ok := item[key]; !ok {
			t.Errorf("inventory response missing %q: %#v", key, item)
		}
	}

	for _, invalid := range []string{
		`{"name":"Food","amount":1,"unit":"g","category":"Pantry","lowAt":0,"count":2}`,
		`{"name":"Food","amount":1,"unit":"g","category":"Pantry","lowAt":0,"recordedOn":"2026/09/29"}`,
		`{"name":"Food","amount":1,"unit":"g","category":"Pantry","lowAt":0,"private":true}`,
	} {
		request = httptest.NewRequest(http.MethodPut, "/api/v1/inventory/rejected", strings.NewReader(invalid))
		request.Header.Set("Content-Type", "application/json")
		rejected := httptest.NewRecorder()
		handler.ServeHTTP(rejected, request)
		if rejected.Code != http.StatusBadRequest {
			t.Errorf("invalid inventory status = %d: %s", rejected.Code, rejected.Body.String())
		}
	}
	for _, stored := range memory.Snapshot().Inventory {
		if stored.ID == "rejected" {
			t.Fatalf("rejected inventory mutation was stored")
		}
	}
}

func assertJSONKeys(t *testing.T, body []byte, want ...string) {
	t.Helper()
	var object map[string]json.RawMessage
	if err := json.Unmarshal(body, &object); err != nil {
		t.Fatalf("decode JSON object: %v", err)
	}
	if len(object) != len(want) {
		t.Fatalf("JSON has %d keys, want %d", len(object), len(want))
	}
	for _, key := range want {
		if _, ok := object[key]; !ok {
			t.Errorf("JSON is missing key %q", key)
		}
	}
}
