package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/simulation/npc"
	"github.com/marco/evociv-rl/internal/simulation/settlement"
	"github.com/marco/evociv-rl/internal/world"
)

func TestDebugStatusBar(t *testing.T) {
	wm := world.NewWorldMap(10, 10)
	for y := 0; y < wm.Height; y++ {
		for x := 0; x < wm.Width; x++ {
			wm.TileAt(x, y).BiomeID = "plains"
		}
	}

	w := ecs.NewWorld()
	npc.RegisterStores(w)
	settlement.RegisterSettlementStores(w)

	// Create an NPC with a name
	ne := w.NewEntity()
	ecs.AddComponent(w, ne, ecs.Name{Name: "Gorim"})
	ecs.AddComponent(w, ne, npc.Health{Current: 80, Max: 100})
	ecs.AddComponent(w, ne, npc.Job{Role: "farmer"})
	ecs.AddComponent(w, ne, npc.AIState{})
	ecs.AddComponent(w, ne, npc.Appearance{Symbol: '@', Color: lipgloss.Color("#FF0000")})
	ecs.AddComponent(w, ne, ecs.Position{X: 5, Y: 5})
	ecs.AddComponent(w, ne, settlement.HomeReference{SettlementEntity: ecs.Entity(1)})
	ecs.AddComponent(w, ne, npc.LOD{Level: npc.LODLocal})

	m := NewModel()
	m.SetWorldMap(wm)
	m.SetECSWorld(w)
	m.screen = "settlement"
	m.width = 80
	m.height = 24
	m.ready = true
	m.settlementViewState = &SettlementViewState{
		SettlementEntity:  ecs.Entity(1),
		SettlementCenterX: 5,
		SettlementCenterY: 5,
		CursorX:           3,
		CursorY:           3,
		ViewportRadius:    3,
	}
	m.settlementNPCs = []npc.NPCRenderInfo{
		{Entity: ne, WorldX: 5, WorldY: 5, Symbol: '@', Color: lipgloss.Color("#FF0000")},
	}

	// Test with NPC under cursor
	v := m.View()
	fmt.Printf("View output with NPC:\n%s\n", v)
	fmt.Printf("Contains 'Gorim': %v\n", strings.Contains(v, "Gorim"))
	fmt.Printf("Contains '(3,3)': %v\n", strings.Contains(v, "(3,3)"))

	// Test with building under cursor (no NPC)
	m.settlementNPCs = nil
	m.settlementBuildings = []settlement.BuildingRenderInfo{
		{Entity: ecs.Entity(10), WorldX: 5, WorldY: 5, Symbol: '╬', Color: "#DEB887", Name: "Granja", Role: "farmer", MaxWorkers: 3},
	}
	v2 := m.View()
	fmt.Printf("\nView output with building (role):\n%s\n", v2)
	fmt.Printf("Contains 'Granja': %v\n", strings.Contains(v2, "Granja"))
	fmt.Printf("Contains 'workers': %v\n", strings.Contains(v2, "workers"))

	// Test with building under cursor (no role)
	m.settlementBuildings = []settlement.BuildingRenderInfo{
		{Entity: ecs.Entity(10), WorldX: 5, WorldY: 5, Symbol: '╬', Color: "#DEB887", Name: "Granja"},
	}
	v3 := m.View()
	fmt.Printf("\nView output with building (no role):\n%s\n", v3)
	fmt.Printf("Contains 'Granja': %v\n", strings.Contains(v3, "Granja"))

	// Test with small terminal height
	m.height = 5
	m.settlementBuildings = nil
	m.settlementNPCs = []npc.NPCRenderInfo{
		{Entity: ne, WorldX: 5, WorldY: 5, Symbol: '@', Color: lipgloss.Color("#FF0000")},
	}
	v4 := m.View()
	fmt.Printf("\nView output with height=5:\n%s\n", v4)
	fmt.Printf("Number of lines: %d\n", strings.Count(v4, "\n")+1)
	fmt.Printf("Contains 'Gorim': %v\n", strings.Contains(v4, "Gorim"))
}
