package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/arturo-and-artura/cyber-kitchen/internal/domain"
)

var ErrInvalidResponse = errors.New("invalid kitchen agent response")

// Context is the complete application-owned context for one focused kitchen
// assistance turn. The model has no filesystem, mutation, or session access.
type Context struct {
	Household domain.Household       `json:"household"`
	Inventory []domain.InventoryItem `json:"inventory"`
	History   []domain.HistoryEntry  `json:"history"`
}

// Response is the only accepted kitchen-agent response format.
type Response struct {
	Recommendations []domain.Meal `json:"recommendations"`
}

// Model executes one stateless kitchen-assistance turn. Implementations receive
// an application-built prompt and must return only a Response JSON document.
type Model interface {
	Complete(context.Context, string) ([]byte, error)
}

type Runner struct {
	model Model
}

func New(model Model) *Runner {
	return &Runner{model: model}
}

func (r *Runner) Recommend(ctx context.Context, input Context) ([]domain.Meal, error) {
	if r.model == nil {
		return nil, errors.New("kitchen model is not configured")
	}
	payload, err := r.model.Complete(ctx, BuildPrompt(input))
	if err != nil {
		return nil, fmt.Errorf("complete kitchen recommendation: %w", err)
	}
	var response Response
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("%w: decode JSON: %v", ErrInvalidResponse, err)
	}
	if !decodeEOF(decoder) {
		return nil, fmt.Errorf("%w: response must contain one JSON object", ErrInvalidResponse)
	}
	if err := Validate(response, input.Inventory); err != nil {
		return nil, err
	}
	return cloneMeals(response.Recommendations), nil
}

func decodeEOF(decoder *json.Decoder) bool {
	var extra any
	return errors.Is(decoder.Decode(&extra), io.EOF)
}

func BuildPrompt(input Context) string {
	contextJSON, _ := json.Marshal(input)
	return "You are Cyber Kitchen, a focused household kitchen-assistance agent. " +
		"Recommend exactly three safe, practical meals using the supplied household, inventory, and history state. " +
		"Respect every constraint. Return only one JSON object matching this exact shape: " +
		`{"recommendations":[{"id":"unique-kebab-case-id","name":"meal name","description":"short description","reason":"why it fits this household","emoji":"one emoji","accent":"non-empty color name","minutes":30,"difficulty":"Easy","tags":["tag"],"ingredients":[{"inventoryId":"exact inventory item id","name":"ingredient name","amount":1,"unit":"inventory unit"}],"steps":["complete cooking step"]}]}. ` +
		"The recommendations array must contain exactly three complete meals. Difficulty must be Easy or Medium. " +
		"Every ingredient must reference an exact supplied inventory id, use a positive amount no greater than the available amount, and must not introduce unlisted ingredients. " +
		"Do not request tools, use markdown, add unknown fields, or describe mutations. " +
		"Application context: " + string(contextJSON)
}

func Validate(response Response, inventory []domain.InventoryItem) error {
	if len(response.Recommendations) != 3 {
		return fmt.Errorf("%w: exactly three recommendations are required", ErrInvalidResponse)
	}
	available := make(map[string]domain.InventoryItem, len(inventory))
	for _, item := range inventory {
		available[item.ID] = item
	}
	ids := map[string]bool{}
	for _, meal := range response.Recommendations {
		if strings.TrimSpace(meal.ID) == "" || strings.TrimSpace(meal.Name) == "" || strings.TrimSpace(meal.Description) == "" || strings.TrimSpace(meal.Reason) == "" || strings.TrimSpace(meal.Emoji) == "" || strings.TrimSpace(meal.Accent) == "" || meal.Minutes <= 0 || (meal.Difficulty != "Easy" && meal.Difficulty != "Medium") || len(meal.Ingredients) == 0 || len(meal.Steps) == 0 {
			return fmt.Errorf("%w: incomplete meal", ErrInvalidResponse)
		}
		if ids[meal.ID] {
			return fmt.Errorf("%w: duplicate meal id", ErrInvalidResponse)
		}
		ids[meal.ID] = true
		for _, ingredient := range meal.Ingredients {
			item, ok := available[ingredient.InventoryID]
			if !ok || ingredient.Amount <= 0 || ingredient.Amount > item.Amount || strings.TrimSpace(ingredient.Name) == "" {
				return fmt.Errorf("%w: unsafe ingredient reference", ErrInvalidResponse)
			}
		}
		for _, step := range meal.Steps {
			if strings.TrimSpace(step) == "" {
				return fmt.Errorf("%w: empty step", ErrInvalidResponse)
			}
		}
	}
	return nil
}

func cloneMeals(meals []domain.Meal) []domain.Meal {
	result := make([]domain.Meal, len(meals))
	for i, meal := range meals {
		result[i] = meal
		result[i].Tags = append([]string{}, meal.Tags...)
		result[i].Ingredients = append([]domain.MealIngredient{}, meal.Ingredients...)
		result[i].Steps = append([]string{}, meal.Steps...)
	}
	return result
}
