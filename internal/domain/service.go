package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrMealNotFound      = errors.New("meal not found")
	ErrInventoryNotFound = errors.New("inventory item not found")
	ErrInvalidRating     = errors.New("rating must be one of loved, okay, or not-for-us")
	ErrInvalidHousehold  = errors.New("household constraints and goals must be non-empty values")
	ErrInvalidInventory  = errors.New("inventory item is invalid")
)

type StateStore interface {
	Snapshot() State
	Update(func(*State) error) (State, error)
}

type Service struct {
	store StateStore
	now   func() time.Time
	newID func() string
}

func NewService(store StateStore, now func() time.Time, newID func() string) *Service {
	return &Service{store: store, now: now, newID: newID}
}

func (s *Service) State() State {
	return s.store.Snapshot()
}

func (s *Service) UpdateHousehold(constraints, goals []string) (State, error) {
	if !validStrings(constraints) || !validStrings(goals) {
		return State{}, ErrInvalidHousehold
	}
	return s.store.Update(func(state *State) error {
		state.Household.Constraints = append([]string{}, constraints...)
		state.Household.Goals = append([]string{}, goals...)
		return nil
	})
}

func (s *Service) PutInventory(item InventoryItem) (State, error) {
	if !validInventory(item) {
		return State{}, ErrInvalidInventory
	}
	return s.store.Update(func(state *State) error {
		for i := range state.Inventory {
			if state.Inventory[i].ID == item.ID {
				state.Inventory[i] = item
				return nil
			}
		}
		state.Inventory = append(state.Inventory, item)
		return nil
	})
}

func (s *Service) DeleteInventory(id string) (State, error) {
	return s.store.Update(func(state *State) error {
		for i := range state.Inventory {
			if state.Inventory[i].ID == id {
				state.Inventory = append(state.Inventory[:i], state.Inventory[i+1:]...)
				return nil
			}
		}
		return ErrInventoryNotFound
	})
}

func (s *Service) ReplaceRecommendations(meals []Meal) (State, error) {
	return s.store.Update(func(state *State) error {
		state.Meals = append([]Meal{}, meals...)
		state.SelectedMealID = nil
		return nil
	})
}

func validStrings(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return false
		}
	}
	return true
}

func validInventory(item InventoryItem) bool {
	if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.Name) == "" || item.Amount < 0 || item.LowAt < 0 {
		return false
	}
	switch item.Category {
	case "Produce", "Protein", "Pantry", "Dairy":
		return true
	}
	return false
}

func (s *Service) ConfirmMeal(mealID string, rating Rating, note string) (State, error) {
	if !rating.Valid() {
		return State{}, ErrInvalidRating
	}

	return s.store.Update(func(state *State) error {
		var meal *Meal
		for i := range state.Meals {
			if state.Meals[i].ID == mealID {
				meal = &state.Meals[i]
				break
			}
		}
		if meal == nil {
			return ErrMealNotFound
		}

		used := make(map[string]float64, len(meal.Ingredients))
		for _, ingredient := range meal.Ingredients {
			used[ingredient.InventoryID] += ingredient.Amount
		}
		for i := range state.Inventory {
			remaining := state.Inventory[i].Amount - used[state.Inventory[i].ID]
			if remaining < 0 {
				remaining = 0
			}
			state.Inventory[i].Amount = remaining
		}

		entry := HistoryEntry{
			ID:       s.newID(),
			MealID:   meal.ID,
			MealName: meal.Name,
			Emoji:    meal.Emoji,
			CookedAt: s.now().UTC(),
			Rating:   rating,
			Note:     note,
		}
		state.History = append([]HistoryEntry{entry}, state.History...)
		state.SelectedMealID = nil
		return nil
	})
}
