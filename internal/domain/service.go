package domain

import (
	"errors"
	"time"
)

var (
	ErrMealNotFound  = errors.New("meal not found")
	ErrInvalidRating = errors.New("rating must be one of loved, okay, or not-for-us")
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
