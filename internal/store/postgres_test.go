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
		state.Inventory[0].Amount = 0
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
	if state.Inventory[0].Amount != 0 || len(state.History) != 1 {
		t.Fatalf("state did not persist: %#v", state)
	}
}
