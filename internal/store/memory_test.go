package store

import (
	"errors"
	"testing"

	"github.com/arturo-and-artura/cyber-kitchen/internal/domain"
)

func TestMemorySnapshotsAreIsolatedAndFailedUpdatesRollBack(t *testing.T) {
	height := 170.0
	count := 2.0
	memory := NewMemory(domain.State{
		Household: domain.Household{
			Members:     []domain.HouseholdMember{{ID: "member", Name: "Member", HeightCm: &height, Notes: []string{"Original note"}}},
			Preferences: []string{"Original preference"},
		},
		Inventory: []domain.InventoryItem{{ID: "rice", Amount: 3, Count: &count}},
	})
	snapshot := memory.Snapshot()
	snapshot.Inventory[0].Amount = 0
	*snapshot.Inventory[0].Count = 0
	*snapshot.Household.Members[0].HeightCm = 0
	snapshot.Household.Members[0].Notes[0] = "Changed note"
	snapshot.Household.Preferences[0] = "Changed preference"
	if got := memory.Snapshot().Inventory[0].Amount; got != 3 {
		t.Fatalf("mutating snapshot changed stored amount to %v", got)
	}
	stored := memory.Snapshot()
	if *stored.Inventory[0].Count != 2 || *stored.Household.Members[0].HeightCm != 170 || stored.Household.Members[0].Notes[0] != "Original note" || stored.Household.Preferences[0] != "Original preference" {
		t.Fatalf("mutating snapshot changed nested metadata: %#v", stored)
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
