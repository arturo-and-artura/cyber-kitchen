package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/arturo-and-artura/cyber-kitchen/internal/domain"
	"github.com/arturo-and-artura/cyber-kitchen/internal/seed"
)

func TestSQLiteSeedsAndReloadsRelationalState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "kitchen.db")
	store, err := NewSQLite(path, seed.InitialState())
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	_, err = store.Update(func(state *domain.State) error {
		state.Inventory[0].Amount = 7
		state.History = []domain.HistoryEntry{}
		return nil
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	reloaded, err := NewSQLite(path, domain.State{})
	if err != nil {
		t.Fatalf("reload NewSQLite() error = %v", err)
	}
	defer reloaded.Close()
	state := reloaded.Snapshot()
	if state.Inventory[0].Amount != 7 {
		t.Fatalf("reloaded inventory amount = %v, want 7", state.Inventory[0].Amount)
	}
	if state.History == nil || len(state.History) != 0 {
		t.Fatalf("reloaded history = %#v, want non-nil empty slice", state.History)
	}
	if len(state.Household.Members) != 3 || len(state.Meals) != 3 || len(state.Meals[0].Ingredients) == 0 {
		t.Fatalf("relational seed was not reloaded completely: %#v", state)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("database permissions = %o, want 600", got)
	}
}

func TestSQLiteCommitsWholeStateTransaction(t *testing.T) {
	path := filepath.Join(t.TempDir(), "kitchen.db")
	store, err := NewSQLite(path, seed.InitialState())
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	_, err = store.Update(func(state *domain.State) error {
		state.Inventory[0].Amount = 1
		state.History = append([]domain.HistoryEntry{{ID: "new", MealID: "miso-salmon", MealName: "Miso salmon", Rating: domain.RatingLoved}}, state.History...)
		state.SelectedMealID = nil
		return nil
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	reloaded, err := NewSQLite(path, domain.State{})
	if err != nil {
		t.Fatalf("reload NewSQLite() error = %v", err)
	}
	defer reloaded.Close()
	state := reloaded.Snapshot()
	if state.Inventory[0].Amount != 1 || len(state.History) != 3 || state.History[0].ID != "new" || state.SelectedMealID != nil {
		t.Fatalf("transaction did not persist complete state: inventory=%v history=%#v selected=%v", state.Inventory[0].Amount, state.History, state.SelectedMealID)
	}
}

func TestSQLiteDoesNotPersistRejectedUpdate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "kitchen.db")
	store, err := NewSQLite(path, seed.InitialState())
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	defer store.Close()

	wantErr := errors.New("reject change")
	if _, err := store.Update(func(state *domain.State) error {
		state.Inventory[0].Amount = 99
		return wantErr
	}); !errors.Is(err, wantErr) {
		t.Fatalf("Update() error = %v, want %v", err, wantErr)
	}
	if got := store.Snapshot().Inventory[0].Amount; got == 99 {
		t.Fatal("rejected update changed cached state")
	}
	var amount float64
	if err := store.db.QueryRow("SELECT amount FROM inventory ORDER BY position LIMIT 1").Scan(&amount); err != nil {
		t.Fatalf("query persisted inventory: %v", err)
	}
	if amount == 99 {
		t.Fatal("rejected update changed persisted state")
	}
}

func TestSQLiteDoesNotPublishFailedTransaction(t *testing.T) {
	path := filepath.Join(t.TempDir(), "kitchen.db")
	store, err := NewSQLite(path, seed.InitialState())
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	before := store.Snapshot().Inventory[0].Amount
	if err := store.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if _, err := store.Update(func(state *domain.State) error {
		state.Inventory[0].Amount = before + 1
		return nil
	}); err == nil {
		t.Fatal("Update() error = nil, want closed database error")
	}
	if got := store.Snapshot().Inventory[0].Amount; got != before {
		t.Fatalf("cached amount = %v, want %v", got, before)
	}
}
