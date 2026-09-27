package store

import (
	"errors"
	"testing"

	"github.com/arturo-and-artura/cyber-kitchen/internal/domain"
)

func TestMemorySnapshotsAreIsolatedAndFailedUpdatesRollBack(t *testing.T) {
	memory := NewMemory(domain.State{Inventory: []domain.InventoryItem{{ID: "rice", Amount: 3}}})
	snapshot := memory.Snapshot()
	snapshot.Inventory[0].Amount = 0
	if got := memory.Snapshot().Inventory[0].Amount; got != 3 {
		t.Fatalf("mutating snapshot changed stored amount to %v", got)
	}

	wantErr := errors.New("stop")
	if _, err := memory.Update(func(state *domain.State) error {
		state.Inventory[0].Amount = 1
		return wantErr
	}); !errors.Is(err, wantErr) {
		t.Fatalf("Update() error = %v, want %v", err, wantErr)
	}
	if got := memory.Snapshot().Inventory[0].Amount; got != 3 {
		t.Errorf("failed update changed stored amount to %v", got)
	}
}
