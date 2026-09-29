package domain

import (
	"errors"
	"math"
	"regexp"
	"strings"
	"time"
)

var (
	ErrMealNotFound      = errors.New("meal not found")
	ErrInventoryNotFound = errors.New("inventory item not found")
	ErrInvalidRating     = errors.New("rating must be one of loved, okay, or not-for-us")
	ErrInvalidHousehold  = errors.New("household profile is invalid")
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

func (s *Service) UpdateHousehold(members []HouseholdMember, constraints, goals, preferences []string) (State, error) {
	if !validHousehold(members, constraints, goals, preferences) {
		return State{}, ErrInvalidHousehold
	}
	return s.store.Update(func(state *State) error {
		state.Household.Members = cloneMembers(members)
		state.Household.Constraints = append([]string{}, constraints...)
		state.Household.Goals = append([]string{}, goals...)
		state.Household.Preferences = append([]string{}, preferences...)
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

const (
	maxMembers       = 20
	maxProfileValues = 50
	maxMemberNotes   = 20
	maxNameLength    = 120
	maxShortLength   = 80
	maxTextLength    = 500
)

var validID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)

func validHousehold(members []HouseholdMember, constraints, goals, preferences []string) bool {
	if len(members) > maxMembers || !validStrings(constraints, maxProfileValues, maxTextLength) ||
		!validStrings(goals, maxProfileValues, maxTextLength) || !validStrings(preferences, maxProfileValues, maxTextLength) {
		return false
	}
	seen := make(map[string]struct{}, len(members))
	for _, member := range members {
		if !validID.MatchString(member.ID) || !validRequiredString(member.Name, maxNameLength) ||
			!validOptionalString(member.Initials, maxShortLength) || !validStrings(member.Notes, maxMemberNotes, maxTextLength) {
			return false
		}
		if member.HeightCm != nil && (!finiteNonnegative(*member.HeightCm) || *member.HeightCm == 0 || *member.HeightCm > 300) {
			return false
		}
		if _, exists := seen[member.ID]; exists {
			return false
		}
		seen[member.ID] = struct{}{}
	}
	return true
}

func validStrings(values []string, maxItems, maxLength int) bool {
	if len(values) > maxItems {
		return false
	}
	for _, value := range values {
		if !validRequiredString(value, maxLength) {
			return false
		}
	}
	return true
}

func validInventory(item InventoryItem) bool {
	if !validID.MatchString(item.ID) || !validRequiredString(item.Name, maxNameLength) ||
		!finiteNonnegative(item.Amount) || !finiteNonnegative(item.LowAt) ||
		!validOptionalString(item.Unit, maxShortLength) || !validOptionalString(item.CountUnit, maxShortLength) ||
		!validOptionalString(item.Storage, maxNameLength) || !validOptionalString(item.Notes, maxTextLength) {
		return false
	}
	if (item.Count == nil) != (strings.TrimSpace(item.CountUnit) == "") {
		return false
	}
	if item.Count != nil && !finiteNonnegative(*item.Count) {
		return false
	}
	if item.RecordedOn != "" {
		parsed, err := time.Parse("2006-01-02", item.RecordedOn)
		if err != nil || parsed.Format("2006-01-02") != item.RecordedOn {
			return false
		}
	}
	if strings.TrimSpace(item.RecordedOn) != item.RecordedOn {
		return false
	}
	switch item.Category {
	case "Produce", "Protein", "Pantry", "Dairy":
		return true
	}
	return false
}

func validRequiredString(value string, maxLength int) bool {
	return strings.TrimSpace(value) != "" && strings.TrimSpace(value) == value && len([]rune(value)) <= maxLength
}

func validOptionalString(value string, maxLength int) bool {
	return value == "" || validRequiredString(value, maxLength)
}

func finiteNonnegative(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0
}

func cloneMembers(members []HouseholdMember) []HouseholdMember {
	result := make([]HouseholdMember, len(members))
	for i, member := range members {
		result[i] = member
		result[i].Notes = append([]string{}, member.Notes...)
		if member.HeightCm != nil {
			height := *member.HeightCm
			result[i].HeightCm = &height
		}
	}
	return result
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
