package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/arturo-and-artura/cyber-kitchen/internal/agent"
	"github.com/arturo-and-artura/cyber-kitchen/internal/domain"
)

type Config struct {
	AllowedOrigins []string
}

type Server struct {
	service *domain.Service
	agent   *agent.Runner
	handler http.Handler
}

func New(service *domain.Service, config Config, runners ...*agent.Runner) *Server {
	server := &Server{service: service}
	if len(runners) > 0 {
		server.agent = runners[0]
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", server.health)
	mux.HandleFunc("GET /api/v1/household", server.getHousehold)
	mux.HandleFunc("PUT /api/v1/household", server.updateHousehold)
	mux.HandleFunc("GET /api/v1/inventory", server.getInventory)
	mux.HandleFunc("PUT /api/v1/inventory/{id}", server.putInventory)
	mux.HandleFunc("DELETE /api/v1/inventory/{id}", server.deleteInventory)
	mux.HandleFunc("GET /api/v1/meals", server.getMeals)
	mux.HandleFunc("POST /api/v1/recommendations/generate", server.generateRecommendations)
	mux.HandleFunc("GET /api/v1/history", server.getHistory)
	mux.HandleFunc("POST /api/v1/meals/{id}/confirm", server.confirmMeal)
	server.handler = withCORS(config.AllowedOrigins, mux)
	return server
}

func (s *Server) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	s.handler.ServeHTTP(response, request)
}

func (s *Server) health(response http.ResponseWriter, _ *http.Request) {
	writeJSON(response, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) getHousehold(response http.ResponseWriter, _ *http.Request) {
	state := s.service.State()
	writeJSON(response, http.StatusOK, householdResponseFrom(state.Household))
}

func (s *Server) getInventory(response http.ResponseWriter, _ *http.Request) {
	state := s.service.State()
	writeJSON(response, http.StatusOK, inventoryReadResponse{Inventory: inventoryResponsesFrom(state.Inventory)})
}

func (s *Server) getMeals(response http.ResponseWriter, _ *http.Request) {
	state := s.service.State()
	writeJSON(response, http.StatusOK, mealsReadResponse{
		Meals:          mealResponsesFrom(state.Meals),
		SelectedMealID: state.SelectedMealID,
	})
}

func (s *Server) getHistory(response http.ResponseWriter, _ *http.Request) {
	state := s.service.State()
	writeJSON(response, http.StatusOK, historyReadResponse{History: historyResponsesFrom(state.History)})
}

func (s *Server) updateHousehold(response http.ResponseWriter, request *http.Request) {
	var input struct {
		Constraints []string `json:"constraints"`
		Goals       []string `json:"goals"`
	}
	if !decodeBody(response, request, &input) {
		return
	}
	state, err := s.service.UpdateHousehold(input.Constraints, input.Goals)
	if errors.Is(err, domain.ErrInvalidHousehold) {
		writeError(response, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeError(response, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(response, http.StatusOK, householdResponseFrom(state.Household))
}

func (s *Server) putInventory(response http.ResponseWriter, request *http.Request) {
	var input domain.InventoryItem
	if !decodeBody(response, request, &input) {
		return
	}
	input.ID = request.PathValue("id")
	state, err := s.service.PutInventory(input)
	if errors.Is(err, domain.ErrInvalidInventory) {
		writeError(response, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeError(response, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(response, http.StatusOK, inventoryReadResponse{Inventory: inventoryResponsesFrom(state.Inventory)})
}

func (s *Server) deleteInventory(response http.ResponseWriter, request *http.Request) {
	state, err := s.service.DeleteInventory(request.PathValue("id"))
	if errors.Is(err, domain.ErrInventoryNotFound) {
		writeError(response, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(response, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(response, http.StatusOK, inventoryReadResponse{Inventory: inventoryResponsesFrom(state.Inventory)})
}

func (s *Server) generateRecommendations(response http.ResponseWriter, request *http.Request) {
	if s.agent == nil {
		writeError(response, http.StatusServiceUnavailable, "kitchen agent is unavailable")
		return
	}
	state := s.service.State()
	meals, source := s.agent.Recommend(request.Context(), agent.Context{Household: state.Household, Inventory: state.Inventory, History: state.History})
	state, err := s.service.ReplaceRecommendations(meals)
	if err != nil {
		writeError(response, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(response, http.StatusOK, struct {
		Meals  []mealResponse `json:"meals"`
		Source string         `json:"source"`
	}{Meals: mealResponsesFrom(state.Meals), Source: source})
}

func decodeBody(response http.ResponseWriter, request *http.Request, destination any) bool {
	if contentType := request.Header.Get("Content-Type"); contentType != "" && !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		writeError(response, http.StatusUnsupportedMediaType, "content type must be application/json")
		return false
	}
	decoder := json.NewDecoder(http.MaxBytesReader(response, request.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		writeError(response, http.StatusBadRequest, "invalid JSON request body")
		return false
	}
	if err := ensureEOF(decoder); err != nil {
		writeError(response, http.StatusBadRequest, "request body must contain one JSON object")
		return false
	}
	return true
}

func (s *Server) confirmMeal(response http.ResponseWriter, request *http.Request) {
	if contentType := request.Header.Get("Content-Type"); contentType != "" && !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		writeError(response, http.StatusUnsupportedMediaType, "content type must be application/json")
		return
	}

	var input struct {
		Rating domain.Rating `json:"rating"`
		Note   string        `json:"note"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(response, request.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeError(response, http.StatusBadRequest, "invalid JSON request body")
		return
	}
	if err := ensureEOF(decoder); err != nil {
		writeError(response, http.StatusBadRequest, "request body must contain one JSON object")
		return
	}

	state, err := s.service.ConfirmMeal(request.PathValue("id"), input.Rating, input.Note)
	if errors.Is(err, domain.ErrInvalidRating) {
		writeError(response, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, domain.ErrMealNotFound) {
		writeError(response, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(response, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(response, http.StatusOK, confirmationResponse{
		Inventory:      inventoryResponsesFrom(state.Inventory),
		History:        historyResponsesFrom(state.History),
		SelectedMealID: state.SelectedMealID,
	})
}

func ensureEOF(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return errors.New("extra JSON value")
	}
	return err
}

func withCORS(origins []string, next http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowed[origin] = struct{}{}
		}
	}

	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		origin := request.Header.Get("Origin")
		if origin != "" {
			_, exact := allowed[origin]
			_, wildcard := allowed["*"]
			if exact || wildcard {
				if wildcard {
					response.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					response.Header().Set("Access-Control-Allow-Origin", origin)
					response.Header().Add("Vary", "Origin")
				}
				response.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				response.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			}
		}
		if request.Method == http.MethodOptions {
			response.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(response, request)
	})
}

func writeError(response http.ResponseWriter, status int, message string) {
	writeJSONStatus(response, status, map[string]string{"error": message})
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	writeJSONStatus(response, status, value)
}

func writeJSONStatus(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(value)
}

type inventoryReadResponse struct {
	Inventory []inventoryResponse `json:"inventory"`
}

type mealsReadResponse struct {
	Meals          []mealResponse `json:"meals"`
	SelectedMealID *string        `json:"selectedMealId"`
}

type historyReadResponse struct {
	History []historyResponse `json:"history"`
}

type confirmationResponse struct {
	Inventory      []inventoryResponse `json:"inventory"`
	History        []historyResponse   `json:"history"`
	SelectedMealID *string             `json:"selectedMealId"`
}

type householdMemberResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Initials string `json:"initials"`
}

type householdResponse struct {
	Name        string                    `json:"name"`
	Members     []householdMemberResponse `json:"members"`
	Constraints []string                  `json:"constraints"`
	Goals       []string                  `json:"goals"`
}

type inventoryResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Amount   float64 `json:"amount"`
	Unit     string  `json:"unit"`
	Category string  `json:"category"`
	LowAt    float64 `json:"lowAt"`
}

type ingredientResponse struct {
	InventoryID string  `json:"inventoryId"`
	Name        string  `json:"name"`
	Amount      float64 `json:"amount"`
	Unit        string  `json:"unit"`
	Optional    bool    `json:"optional,omitempty"`
}

type mealResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Reason      string               `json:"reason"`
	Emoji       string               `json:"emoji"`
	Accent      string               `json:"accent"`
	Minutes     int                  `json:"minutes"`
	Difficulty  string               `json:"difficulty"`
	Tags        []string             `json:"tags"`
	Ingredients []ingredientResponse `json:"ingredients"`
	Steps       []string             `json:"steps"`
}

type historyResponse struct {
	ID       string        `json:"id"`
	MealID   string        `json:"mealId"`
	MealName string        `json:"mealName"`
	Emoji    string        `json:"emoji"`
	CookedAt string        `json:"cookedAt"`
	Rating   domain.Rating `json:"rating"`
	Note     string        `json:"note"`
}

func householdResponseFrom(household domain.Household) householdResponse {
	result := householdResponse{
		Name:        household.Name,
		Members:     make([]householdMemberResponse, len(household.Members)),
		Constraints: append([]string{}, household.Constraints...),
		Goals:       append([]string{}, household.Goals...),
	}
	for i, member := range household.Members {
		result.Members[i] = householdMemberResponse{ID: member.ID, Name: member.Name, Initials: member.Initials}
	}
	return result
}

func inventoryResponsesFrom(inventory []domain.InventoryItem) []inventoryResponse {
	result := make([]inventoryResponse, len(inventory))
	for i, item := range inventory {
		result[i] = inventoryResponse{ID: item.ID, Name: item.Name, Amount: item.Amount, Unit: item.Unit, Category: item.Category, LowAt: item.LowAt}
	}
	return result
}

func mealResponsesFrom(meals []domain.Meal) []mealResponse {
	result := make([]mealResponse, len(meals))
	for i, meal := range meals {
		result[i] = mealResponse{
			ID: meal.ID, Name: meal.Name, Description: meal.Description, Reason: meal.Reason, Emoji: meal.Emoji,
			Accent: meal.Accent, Minutes: meal.Minutes, Difficulty: meal.Difficulty,
			Tags: append([]string{}, meal.Tags...), Steps: append([]string{}, meal.Steps...),
			Ingredients: make([]ingredientResponse, len(meal.Ingredients)),
		}
		for j, ingredient := range meal.Ingredients {
			result[i].Ingredients[j] = ingredientResponse{
				InventoryID: ingredient.InventoryID, Name: ingredient.Name, Amount: ingredient.Amount, Unit: ingredient.Unit, Optional: ingredient.Optional,
			}
		}
	}
	return result
}

func historyResponsesFrom(history []domain.HistoryEntry) []historyResponse {
	result := make([]historyResponse, len(history))
	for i, entry := range history {
		result[i] = historyResponse{
			ID: entry.ID, MealID: entry.MealID, MealName: entry.MealName, Emoji: entry.Emoji,
			CookedAt: entry.CookedAt.UTC().Format("2006-01-02T15:04:05.000Z"), Rating: entry.Rating, Note: entry.Note,
		}
	}
	return result
}
