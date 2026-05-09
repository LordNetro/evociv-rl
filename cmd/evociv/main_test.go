package main

import (
	"testing"
	"testing/fstest"

	"github.com/marco/evociv-rl/internal/data"
	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/simulation/economy"
	"github.com/marco/evociv-rl/internal/simulation/npc"
	"github.com/marco/evociv-rl/internal/simulation/settlement"
	"github.com/marco/evociv-rl/internal/world"
)

func TestSmokeEconomyIntegration(t *testing.T) {
	fsys := fstest.MapFS{
		"data/settlements.yaml": &fstest.MapFile{
			Data: []byte(`kind: settlement-types
data:
  - id: village
    name: Aldea
    symbol: "♦"
    color: "#8B7355"
    radius: 3
    biomes: [plains]
    buildings: [house, farm]
    spawn_weight: 1.0
`),
		},
		"data/buildings.yaml": &fstest.MapFile{
			Data: []byte(`kind: building-types
data:
  - id: house
    name: Casa
  - id: farm
    name: Granja
    role: farmer
    produces:
      food: 2.0
    max_workers: 3
`),
		},
		"data/growth.yaml": &fstest.MapFile{
			Data: []byte(`kind: growth-thresholds
data:
  - level: 2
    food: 10
    tools: 1
    gold: 1
    new_radius: 4
`),
		},
		"data/npcs.yaml": &fstest.MapFile{
			Data: []byte(`kind: npc-races
data:
  - id: human
    name: Humano
    spawn_weight: 1.0
    roles:
      - id: farmer
        weight: 1.0
    name_pool:
      first: ["A"]
      last: ["B"]
`),
		},
		"data/npc-roles.yaml": &fstest.MapFile{
			Data: []byte(`kind: npc-roles
data:
  - id: farmer
    symbol: "@"
    color: "#FFD700"
    biomes: [plains]
`),
		},
	}

	loader := data.NewLoader(fsys)
	registry := data.NewRegistry()
	if err := loader.LoadAll("data", registry); err != nil {
		t.Fatalf("load data: %v", err)
	}

	settlementDefs, _ := settlement.LoadSettlementTypes(registry)
	buildingDefs, _ := settlement.LoadBuildingTypes(registry)
	thresholds, _ := settlement.LoadGrowthThresholds(registry)
	raceDefs, _ := npc.LoadNpcRaces(registry)
	roleDefs, _ := npc.LoadNpcRoles(registry)

	wm := world.NewWorldMap(64, 64)
	for y := 0; y < wm.Height; y++ {
		for x := 0; x < wm.Width; x++ {
			wm.TileAt(x, y).BiomeID = "plains"
		}
	}

	w := ecs.NewWorld()
	npc.RegisterStores(w)
	settlement.RegisterSettlementStores(w)

	// Spawn settlement and NPCs
	w.AddSystem(settlement.NewSettlementSpawnSystem(wm, 42, settlementDefs, buildingDefs))
	w.AddSystem(npc.NewNPCSpawnSystem(wm, npc.SpawnConfig{Count: 5}, 42, raceDefs, roleDefs))
	w.AddSystem(settlement.NewPopulationSystem())
	w.AddSystem(economy.NewSettlementEconomySystem(buildingDefs))
	w.AddSystem(economy.NewSettlementGrowthSystem(thresholds))
	w.AddSystem(economy.NewFamineSystem())
	_ = w.Update(0)

	setStore := w.GetStore(settlement.SettlementID).(*ecs.ComponentStore[settlement.Settlement])
	resStore := w.GetStore(settlement.ResourceID).(*ecs.ComponentStore[settlement.ResourceStore])
	homeStore := w.GetStore(settlement.HomeRefID).(*ecs.ComponentStore[settlement.HomeReference])

	if setStore.Len() == 0 {
		t.Fatal("expected settlements")
	}

	// Run multiple ticks to see economy in action
	for i := 0; i < 10; i++ {
		_ = w.Update(1.0)
	}

	// Verify resources were produced or consumed
	for e, rs := range resStore.All() {
		_ = e
		if _, ok := rs.Resources["food"]; !ok {
			t.Error("expected food resource to exist")
		}
	}

	// Find a settlement that has NPCs with HomeReference
	var targetSettle ecs.Entity
	for e := range setStore.All() {
		for _, h := range homeStore.All() {
			if h.SettlementEntity == e {
				targetSettle = e
				break
			}
		}
		if targetSettle != 0 {
			break
		}
	}
	if targetSettle == 0 {
		t.Fatal("expected at least one settlement with NPCs")
	}

	// Force famine on that settlement
	rs, _ := resStore.Get(targetSettle)
	rs.Resources["food"] = -5.0
	resStore.Set(targetSettle, rs)

	before := 0
	for _, h := range homeStore.All() {
		if h.SettlementEntity == targetSettle {
			before++
		}
	}
	_ = w.Update(1.0)
	after := 0
	for _, h := range homeStore.All() {
		if h.SettlementEntity == targetSettle {
			after++
		}
	}
	if after >= before {
		t.Errorf("expected famine to remove at least one HomeReference, before=%d after=%d", before, after)
	}
}

func TestSmokeSettlementIntegration(t *testing.T) {
	fsys := fstest.MapFS{
		"data/settlements.yaml": &fstest.MapFile{
			Data: []byte(`kind: settlement-types
data:
  - id: village
    name: Aldea
    symbol: "♦"
    color: "#8B7355"
    radius: 3
    biomes: [plains]
    buildings: [house, farm]
    spawn_weight: 1.0
`),
		},
		"data/buildings.yaml": &fstest.MapFile{
			Data: []byte(`kind: building-types
data:
  - id: house
    name: Casa
  - id: farm
    name: Granja
`),
		},
		"data/npcs.yaml": &fstest.MapFile{
			Data: []byte(`kind: npc-races
data:
  - id: human
    name: Humano
    spawn_weight: 1.0
    roles:
      - id: farmer
        weight: 1.0
    name_pool:
      first: ["A"]
      last: ["B"]
`),
		},
		"data/npc-roles.yaml": &fstest.MapFile{
			Data: []byte(`kind: npc-roles
data:
  - id: farmer
    symbol: "@"
    color: "#FFD700"
    biomes: [plains]
`),
		},
	}

	loader := data.NewLoader(fsys)
	registry := data.NewRegistry()
	if err := loader.LoadAll("data", registry); err != nil {
		t.Fatalf("load data: %v", err)
	}

	settlementDefs, err := settlement.LoadSettlementTypes(registry)
	if err != nil {
		t.Fatalf("load settlements: %v", err)
	}
	buildingDefs, err := settlement.LoadBuildingTypes(registry)
	if err != nil {
		t.Fatalf("load buildings: %v", err)
	}
	raceDefs, err := npc.LoadNpcRaces(registry)
	if err != nil {
		t.Fatalf("load races: %v", err)
	}
	roleDefs, err := npc.LoadNpcRoles(registry)
	if err != nil {
		t.Fatalf("load roles: %v", err)
	}

	wm := world.NewWorldMap(64, 64)
	for y := 0; y < wm.Height; y++ {
		for x := 0; x < wm.Width; x++ {
			wm.TileAt(x, y).BiomeID = "plains"
		}
	}

	w := ecs.NewWorld()
	npc.RegisterStores(w)
	settlement.RegisterSettlementStores(w)

	// Settlement spawn first
	setSys := settlement.NewSettlementSpawnSystem(wm, 42, settlementDefs, buildingDefs)
	w.AddSystem(setSys)
	if err := w.Update(0); err != nil {
		t.Fatalf("settlement update: %v", err)
	}

	setStore := w.GetStore(settlement.SettlementID).(*ecs.ComponentStore[settlement.Settlement])
	if setStore.Len() == 0 {
		t.Fatal("expected settlements after spawn")
	}

	// NPC spawn second (should find settlements)
	npcSys := npc.NewNPCSpawnSystem(wm, npc.SpawnConfig{Count: 5}, 42, raceDefs, roleDefs)
	w.AddSystem(npcSys)
	if err := w.Update(0); err != nil {
		t.Fatalf("npc update: %v", err)
	}

	homeStore := w.GetStore(settlement.HomeRefID).(*ecs.ComponentStore[settlement.HomeReference])
	if homeStore.Len() == 0 {
		t.Fatal("expected some NPCs to have HomeReference after settlement-aware spawn")
	}

	// Settlement render system produces overlay data
	renderSys := settlement.NewSettlementRenderSystem()
	w.AddSystem(renderSys)
	if err := w.Update(0); err != nil {
		t.Fatalf("render update: %v", err)
	}
	infos := renderSys.RenderInfos()
	if len(infos) == 0 {
		t.Fatal("expected settlement render infos")
	}
}
