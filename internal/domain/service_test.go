package domain_test

import (
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/arturo-and-artura/cyber-kitchen/internal/domain"
	"github.com/arturo-and-artura/cyber-kitchen/internal/store"
)

func float64Pointer(value float64) *float64 { return &value }

func TestUpdateHouseholdReplacesProfileButPreservesServerOwnedName(t *testing.T) {
	initial := domain.State{Household: domain.Household{Name: "Server household", Members: []domain.HouseholdMember{{ID: "old", Name: "Old"}}}}
	memory := store.NewMemory(initial)
	service := domain.NewService(memory, time.Now, func() string { return "unused" })
	members := []domain.HouseholdMember{{ID: "member-1", Name: "Member One", Initials: "MO", HeightCm: float64Pointer(172.5), Notes: []string{"Prefers smaller portions"}}}

	got, err := service.UpdateHousehold(members, []string{"No peanuts"}, []string{"Reduce waste"}, []string{"Quick weeknight meals"})
	if err != nil {
		t.Fatalf("UpdateHousehold() error = %v", err)
	}
	if got.Household.Name != initial.Household.Name || !reflect.DeepEqual(got.Household.Members, members) ||
		!reflect.DeepEqual(got.Household.Preferences, []string{"Quick weeknight meals"}) {
		t.Fatalf("household = %#v", got.Household)
	}

	// Returned and input slices cannot mutate persisted nested profile state.
	got.Household.Members[0].Notes[0] = "changed"
	members[0].Notes[0] = "also changed"
	if stored := memory.Snapshot().Household.Members[0].Notes[0]; stored != "Prefers smaller portions" {
		t.Fatalf("stored nested member notes mutated: %q", stored)
	}
}

func TestUpdateHouseholdRejectsInvalidProfilesWithoutMutation(t *testing.T) {
	initial := domain.State{Household: domain.Household{Name: "Server household", Goals: []string{"Existing"}}}
	memory := store.NewMemory(initial)
	service := domain.NewService(memory, time.Now, func() string { return "unused" })

	invalidMembers := [][]domain.HouseholdMember{
		{{ID: "bad id", Name: "Member"}},
		{{ID: "member", Name: " "}},
		{{ID: "duplicate", Name: "One"}, {ID: "duplicate", Name: "Two"}},
		{{ID: "member", Name: "Member", HeightCm: float64Pointer(math.Inf(1))}},
		{{ID: "member", Name: "Member", HeightCm: float64Pointer(301)}},
		{{ID: "member", Name: "Member", Notes: []string{""}}},
	}
	for _, members := range invalidMembers {
		if _, err := service.UpdateHousehold(members, nil, nil, nil); err != domain.ErrInvalidHousehold {
			t.Errorf("UpdateHousehold(%#v) error = %v", members, err)
		}
	}
	got := memory.Snapshot()
	if got.Household.Name != initial.Household.Name || !reflect.DeepEqual(got.Household.Goals, initial.Household.Goals) || len(got.Household.Members) != 0 {
		t.Fatalf("invalid profile mutated state: %#v", got)
	}
}

func TestPutInventoryValidatesAndStoresPreciseMetadata(t *testing.T) {
	memory := store.NewMemory(domain.State{})
	service := domain.NewService(memory, time.Now, func() string { return "unused" })
	valid := domain.InventoryItem{
		ID: "brown-rice", Name: "Brown rice", Amount: 725.5, Unit: "g", Category: "Pantry", LowAt: 200,
		Count: float64Pointer(1.5), CountUnit: "bags", Storage: "Pantry shelf", RecordedOn: "2026-09-29", Notes: "Opened package",
	}
	got, err := service.PutInventory(valid)
	if err != nil || len(got.Inventory) != 1 || !reflect.DeepEqual(got.Inventory[0], valid) {
		t.Fatalf("PutInventory() state=%#v error=%v", got, err)
	}
	*got.Inventory[0].Count = 99
	if stored := *memory.Snapshot().Inventory[0].Count; stored != 1.5 {
		t.Fatalf("returned count pointer mutated stored value: %v", stored)
	}

	invalid := []domain.InventoryItem{
		{ID: "bad id", Name: "Food", Category: "Pantry"},
		{ID: "food", Name: " Food", Category: "Pantry"},
		{ID: "food", Name: "Food", Amount: math.NaN(), Category: "Pantry"},
		{ID: "food", Name: "Food", LowAt: math.Inf(1), Category: "Pantry"},
		{ID: "food", Name: "Food", Category: "Pantry", Count: float64Pointer(1)},
		{ID: "food", Name: "Food", Category: "Pantry", CountUnit: "bags"},
		{ID: "food", Name: "Food", Category: "Pantry", Count: float64Pointer(-1), CountUnit: "bags"},
		{ID: "food", Name: "Food", Category: "Pantry", RecordedOn: "09/29/2026"},
	}
	for _, item := range invalid {
		if _, err := service.PutInventory(item); err != domain.ErrInvalidInventory {
			t.Errorf("PutInventory(%#v) error = %v", item, err)
		}
	}
	if stored := memory.Snapshot().Inventory; len(stored) != 1 || stored[0].ID != valid.ID {
		t.Fatalf("invalid item mutated inventory: %#v", stored)
	}
}

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
