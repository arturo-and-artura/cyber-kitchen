package store

import (
	"sync"

	"github.com/arturo-and-artura/cyber-kitchen/internal/domain"
)

type Memory struct {
	mu    sync.RWMutex
	state domain.State
}

func NewMemory(initial domain.State) *Memory {
	return &Memory{state: cloneState(initial)}
}

func (m *Memory) Snapshot() domain.State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return cloneState(m.state)
}

func (m *Memory) Update(change func(*domain.State) error) (domain.State, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	next := cloneState(m.state)
	if err := change(&next); err != nil {
		return domain.State{}, err
	}
	m.state = next
	return cloneState(m.state), nil
}

func cloneState(state domain.State) domain.State {
	clone := state
	clone.Household.Members = make([]domain.HouseholdMember, len(state.Household.Members))
	for i, member := range state.Household.Members {
		clone.Household.Members[i] = member
		clone.Household.Members[i].Notes = append([]string{}, member.Notes...)
		if member.HeightCm != nil {
			height := *member.HeightCm
			clone.Household.Members[i].HeightCm = &height
		}
	}
	clone.Household.Constraints = append([]string{}, state.Household.Constraints...)
	clone.Household.Goals = append([]string{}, state.Household.Goals...)
	clone.Household.Preferences = append([]string{}, state.Household.Preferences...)
	clone.Inventory = append([]domain.InventoryItem{}, state.Inventory...)
	for i, item := range state.Inventory {
		if item.Count != nil {
			count := *item.Count
			clone.Inventory[i].Count = &count
		}
	}
	clone.History = append([]domain.HistoryEntry{}, state.History...)
	clone.Meals = make([]domain.Meal, len(state.Meals))
	for i, meal := range state.Meals {
		clone.Meals[i] = meal
		clone.Meals[i].Tags = append([]string{}, meal.Tags...)
		clone.Meals[i].Ingredients = append([]domain.MealIngredient{}, meal.Ingredients...)
		clone.Meals[i].Steps = append([]string{}, meal.Steps...)
	}
	if state.SelectedMealID != nil {
		selected := *state.SelectedMealID
		clone.SelectedMealID = &selected
	}
	return clone
}
