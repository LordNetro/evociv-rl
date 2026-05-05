package settlement

import (
	"testing"

	"github.com/marco/evociv-rl/internal/ecs"
)

func TestRegisterSettlementStores(t *testing.T) {
	w := ecs.NewWorld()
	RegisterSettlementStores(w)

	// Verify all 4 stores are registered by adding and retrieving components
	e := w.NewEntity()

	// Settlement
	ecs.AddComponent(w, e, Settlement{Name: "Test", Type: "village", Radius: 3})
	if s, ok := ecs.GetComponent[Settlement](w, e); !ok || s.Name != "Test" {
		t.Error("Settlement store not registered correctly")
	}

	// Building
	ecs.AddComponent(w, e, Building{ID: "house", Name: "Casa"})
	if b, ok := ecs.GetComponent[Building](w, e); !ok || b.ID != "house" {
		t.Error("Building store not registered correctly")
	}

	// HomeReference
	ecs.AddComponent(w, e, HomeReference{SettlementEntity: ecs.Entity(42)})
	if h, ok := ecs.GetComponent[HomeReference](w, e); !ok || h.SettlementEntity != ecs.Entity(42) {
		t.Error("HomeReference store not registered correctly")
	}

	// ResourceStore
	ecs.AddComponent(w, e, ResourceStore{Resources: map[string]float64{"food": 100}})
	if r, ok := ecs.GetComponent[ResourceStore](w, e); !ok || r.Resources["food"] != 100 {
		t.Error("ResourceStore not registered correctly")
	}
}

func TestComponentIDsUnique(t *testing.T) {
	w := ecs.NewWorld()
	RegisterSettlementStores(w)

	// IDs should be non-zero
	if SettlementID == 0 {
		t.Error("SettlementID is zero")
	}
	if BuildingID == 0 {
		t.Error("BuildingID is zero")
	}
	if HomeRefID == 0 {
		t.Error("HomeRefID is zero")
	}
	if ResourceID == 0 {
		t.Error("ResourceID is zero")
	}
}
