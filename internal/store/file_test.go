package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/arturo-and-artura/cyber-kitchen/internal/domain"
	"github.com/arturo-and-artura/cyber-kitchen/internal/seed"
)

func TestFileCreatesAndReloadsState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "kitchen.json")
	store, err := NewFile(path, seed.InitialState())
	if err != nil {
		t.Fatalf("NewFile() error = %v", err)
	}

	_, err = store.Update(func(state *domain.State) error {
		state.Inventory[0].Amount = 7
		state.History = nil
		return nil
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	reloaded, err := NewFile(path, domain.State{})
	if err != nil {
		t.Fatalf("reload NewFile() error = %v", err)
	}
	state := reloaded.Snapshot()
	if state.Inventory[0].Amount != 7 {
		t.Fatalf("reloaded inventory amount = %v, want 7", state.Inventory[0].Amount)
	}
	if state.History == nil || len(state.History) != 0 {
		t.Fatalf("reloaded history = %#v, want non-nil empty slice", state.History)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("data file permissions = %o, want 600", got)
	}
}

func TestFileRejectsInvalidExistingData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "kitchen.json")
	if err := os.WriteFile(path, []byte("not json"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if _, err := NewFile(path, seed.InitialState()); err == nil {
		t.Fatal("NewFile() error = nil, want invalid data error")
	}
}

func TestFileDoesNotPersistRejectedUpdate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "kitchen.json")
	store, err := NewFile(path, seed.InitialState())
	if err != nil {
		t.Fatalf("NewFile() error = %v", err)
	}

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	wantErr := errors.New("reject change")
	if _, err := store.Update(func(state *domain.State) error {
		state.Inventory[0].Amount = 99
		return wantErr
	}); !errors.Is(err, wantErr) {
		t.Fatalf("Update() error = %v, want %v", err, wantErr)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(after) != string(before) {
		t.Fatal("rejected update changed persisted data")
	}
	if got := store.Snapshot().Inventory[0].Amount; got == 99 {
		t.Fatal("rejected update changed in-memory state")
	}
}

func TestFileReturnsPersistenceFailureWithoutPublishingUpdate(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "kitchen.json")
	store, err := NewFile(path, seed.InitialState())
	if err != nil {
		t.Fatalf("NewFile() error = %v", err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if err := os.Remove(directory); err != nil {
		t.Fatalf("RemoveAll() error = %v", err)
	}
	if err := os.WriteFile(directory, []byte("blocks directory recreation"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	before := store.Snapshot().Inventory[0].Amount
	if _, err := store.Update(func(state *domain.State) error {
		state.Inventory[0].Amount = before + 1
		return nil
	}); err == nil {
		t.Fatal("Update() error = nil, want persistence error")
	}
	if got := store.Snapshot().Inventory[0].Amount; got != before {
		t.Fatalf("in-memory amount = %v, want %v", got, before)
	}
}
