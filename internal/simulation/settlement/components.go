package settlement

import (
	"github.com/marco/evociv-rl/internal/ecs"
)

// Settlement represents a settlement entity in the world.
type Settlement struct {
	Name       string
	Type       string // "village" | "town" | "city"
	Symbol     string
	Color      string
	Radius     int      // tiles que abarca (3, 5, 8)
	Population int      // NPCs asignados (counter)
	Level      int      // nivel de desarrollo (1-5, MVP siempre 1)
	Buildings  []string // building IDs present in this settlement
}

// Building represents a building within a settlement.
type Building struct {
	ID    string // "house" | "farm" | "market" | "tavern" | "temple" | "blacksmith"
	Name  string
	Level int // 1-3, MVP siempre 1
}

// HomeReference links an NPC to their home settlement.
type HomeReference struct {
	SettlementEntity ecs.Entity // entity ID del settlement hogar
}

// ResourceStore holds resources for a settlement or building.
type ResourceStore struct {
	Resources map[string]float64 // {"food": 100, "gold": 50} — para futuro
}

// Component IDs for the settlement component types.
var (
	SettlementID = ecs.NewComponentID("settlement")
	BuildingID   = ecs.NewComponentID("building")
	HomeRefID    = ecs.NewComponentID("home_reference")
	ResourceID   = ecs.NewComponentID("resource_store")
)

// RegisterSettlementStores registers the four settlement component stores
// on the given world.
func RegisterSettlementStores(w *ecs.World) {
	ecs.RegisterComponentStore[Settlement](w, SettlementID, ecs.NewComponentStore[Settlement]())
	ecs.RegisterComponentStore[Building](w, BuildingID, ecs.NewComponentStore[Building]())
	ecs.RegisterComponentStore[HomeReference](w, HomeRefID, ecs.NewComponentStore[HomeReference]())
	ecs.RegisterComponentStore[ResourceStore](w, ResourceID, ecs.NewComponentStore[ResourceStore]())
}
