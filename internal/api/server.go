package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/arturo-and-artura/cyber-kitchen/internal/domain"
)

type Config struct {
	AllowedOrigins []string
}

type Server struct {
	service *domain.Service
	handler http.Handler
}

func New(service *domain.Service, config Config) *Server {
	server := &Server{service: service}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", server.health)
	mux.HandleFunc("GET /api/v1/state", server.getState)
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

func (s *Server) getState(response http.ResponseWriter, _ *http.Request) {
	writeJSON(response, http.StatusOK, stateResponseFrom(s.service.State()))
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
	writeJSON(response, http.StatusOK, stateResponseFrom(state))
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

type stateResponse struct {
	Household      householdResponse   `json:"household"`
	Inventory      []inventoryResponse `json:"inventory"`
	Meals          []mealResponse      `json:"meals"`
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

func stateResponseFrom(state domain.State) stateResponse {
	result := stateResponse{
		Household: householdResponse{
			Name:        state.Household.Name,
			Members:     make([]householdMemberResponse, len(state.Household.Members)),
			Constraints: state.Household.Constraints,
			Goals:       state.Household.Goals,
		},
		SelectedMealID: state.SelectedMealID,
		Inventory:      make([]inventoryResponse, len(state.Inventory)),
		Meals:          make([]mealResponse, len(state.Meals)),
		History:        make([]historyResponse, len(state.History)),
	}
	for i, member := range state.Household.Members {
		result.Household.Members[i] = householdMemberResponse{ID: member.ID, Name: member.Name, Initials: member.Initials}
	}
	for i, item := range state.Inventory {
		result.Inventory[i] = inventoryResponse{ID: item.ID, Name: item.Name, Amount: item.Amount, Unit: item.Unit, Category: item.Category, LowAt: item.LowAt}
	}
	for i, meal := range state.Meals {
		result.Meals[i] = mealResponse{
			ID: meal.ID, Name: meal.Name, Description: meal.Description, Reason: meal.Reason, Emoji: meal.Emoji,
			Accent: meal.Accent, Minutes: meal.Minutes, Difficulty: meal.Difficulty, Tags: meal.Tags, Steps: meal.Steps,
			Ingredients: make([]ingredientResponse, len(meal.Ingredients)),
		}
		for j, ingredient := range meal.Ingredients {
			result.Meals[i].Ingredients[j] = ingredientResponse{
				InventoryID: ingredient.InventoryID, Name: ingredient.Name, Amount: ingredient.Amount, Unit: ingredient.Unit, Optional: ingredient.Optional,
			}
		}
	}
	for i, entry := range state.History {
		result.History[i] = historyResponse{
			ID: entry.ID, MealID: entry.MealID, MealName: entry.MealName, Emoji: entry.Emoji,
			CookedAt: entry.CookedAt.UTC().Format("2006-01-02T15:04:05.000Z"), Rating: entry.Rating, Note: entry.Note,
		}
	}
	return result
}
