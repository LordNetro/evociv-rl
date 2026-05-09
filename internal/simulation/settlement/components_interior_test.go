package settlement

import (
	"testing"

	"github.com/marco/evociv-rl/internal/ecs"
)

// TestRegisterBuildingInteriorStore verifies BuildingInterior is registered
func TestRegisterBuildingInteriorStore(t *testing.T) {
	t.Run("RegisterSettlementStores includes BuildingInteriorID", func(t *testing.T) {
		w := ecs.NewWorld()
		RegisterSettlementStores(w)

		store := w.GetStore(BuildingInteriorID)
		if store == nil {
			t.Error("BuildingInteriorID store should be registered")
		}

		// Verify it's the correct type
		biStore, ok := store.(*ecs.ComponentStore[BuildingInterior])
		if !ok {
			t.Error("Store should be of type ComponentStore[BuildingInterior]")
		}
		_ = biStore // use the variable to avoid unused warning
	})
	t.Run("BuildingInteriorID is unique", func(t *testing.T) {
		// BuildingInteriorID should be different from existing IDs
		if BuildingInteriorID == SettlementID {
			t.Error("BuildingInteriorID should differ from SettlementID")
		}
		if BuildingInteriorID == BuildingID {
			t.Error("BuildingInteriorID should differ from BuildingID")
		}
		if BuildingInteriorID == HomeRefID {
			t.Error("BuildingInteriorID should differ from HomeRefID")
		}
		if BuildingInteriorID == ResourceID {
			t.Error("BuildingInteriorID should differ from ResourceID")
		}
	})
}