package main

import (
	"testing"
	"testing/fstest"

	"github.com/marco/evociv-rl/internal/data"
	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/simulation/npc"
	"github.com/marco/evociv-rl/internal/simulation/settlement"
	"github.com/marco/evociv-rl/internal/world"
)

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
