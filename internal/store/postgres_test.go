package store

import (
	"os"
	"testing"

	"github.com/arturo-and-artura/cyber-kitchen/internal/domain"
	"github.com/arturo-and-artura/cyber-kitchen/internal/seed"
)

func TestPostgresPersistsAtomicStateAcrossRestart(t *testing.T) {
	url := os.Getenv("CYBER_KITCHEN_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set CYBER_KITCHEN_TEST_DATABASE_URL to run PostgreSQL integration")
	}
	first, err := NewPostgres(url, seed.InitialState())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := first.Update(func(state *domain.State) error {
		count := 1.5
		state.Inventory[0].Amount = 0
		state.Inventory[0].Count = &count
		state.Inventory[0].CountUnit = "packages"
		state.Inventory[0].Storage = "Freezer"
		state.Inventory[0].RecordedOn = "2026-09-29"
		state.Inventory[0].Notes = "Opened"
		state.Household.Preferences = []string{"Quick dinners"}
		state.History = state.History[:1]
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := NewPostgres(url, seed.InitialState())
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	state := second.Snapshot()
	if state.Inventory[0].Amount != 0 || state.Inventory[0].Count == nil || *state.Inventory[0].Count != 1.5 ||
		state.Inventory[0].CountUnit != "packages" || state.Inventory[0].Storage != "Freezer" ||
		state.Inventory[0].RecordedOn != "2026-09-29" || state.Inventory[0].Notes != "Opened" ||
		len(state.Household.Preferences) != 1 || state.Household.Preferences[0] != "Quick dinners" || len(state.History) != 1 {
		t.Fatalf("state did not persist: %#v", state)
	}
}
