package domain_test

import (
	"testing"
	"time"

	"github.com/arturo-and-artura/cyber-kitchen/internal/domain"
	"github.com/arturo-and-artura/cyber-kitchen/internal/store"
)

func TestConfirmMealUpdatesStateAtomically(t *testing.T) {
	selected := "meal-1"
	initial := domain.State{
		SelectedMealID: &selected,
		Inventory:      []domain.InventoryItem{{ID: "food", Amount: 1}},
		Meals:          []domain.Meal{{ID: "meal-1", Name: "Dinner", Emoji: "🍽️", Ingredients: []domain.MealIngredient{{InventoryID: "food", Amount: 3}}}},
		History:        []domain.HistoryEntry{{ID: "old"}},
	}
	wantTime := time.Date(2026, 9, 26, 18, 30, 0, 0, time.FixedZone("PDT", -7*60*60))
	service := domain.NewService(store.NewMemory(initial), func() time.Time { return wantTime }, func() string { return "new" })

	got, err := service.ConfirmMeal("meal-1", domain.RatingLoved, "Excellent")
	if err != nil {
		t.Fatalf("ConfirmMeal() error = %v", err)
	}
	if got.Inventory[0].Amount != 0 {
		t.Errorf("inventory amount = %v, want 0", got.Inventory[0].Amount)
	}
	if got.SelectedMealID != nil {
		t.Errorf("selected meal = %v, want nil", *got.SelectedMealID)
	}
	if len(got.History) != 2 || got.History[0].ID != "new" || got.History[0].Note != "Excellent" {
		t.Fatalf("history = %#v, want new entry prepended", got.History)
	}
	if !got.History[0].CookedAt.Equal(wantTime) || got.History[0].CookedAt.Location() != time.UTC {
		t.Errorf("cookedAt = %v, want UTC %v", got.History[0].CookedAt, wantTime.UTC())
	}
}

func TestConfirmMealRejectsInvalidInputWithoutMutation(t *testing.T) {
	initial := domain.State{Inventory: []domain.InventoryItem{{ID: "food", Amount: 1}}}
	memory := store.NewMemory(initial)
	service := domain.NewService(memory, time.Now, func() string { return "unused" })

	if _, err := service.ConfirmMeal("missing", domain.RatingLoved, ""); err != domain.ErrMealNotFound {
		t.Fatalf("missing meal error = %v, want %v", err, domain.ErrMealNotFound)
	}
	if _, err := service.ConfirmMeal("missing", domain.Rating("great"), ""); err != domain.ErrInvalidRating {
		t.Fatalf("invalid rating error = %v, want %v", err, domain.ErrInvalidRating)
	}
	if got := memory.Snapshot().Inventory[0].Amount; got != 1 {
		t.Errorf("inventory changed after rejected confirmation: %v", got)
	}
}
