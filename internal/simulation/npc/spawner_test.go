package npc

import (
	"testing"

	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/simulation/settlement"
	"github.com/marco/evociv-rl/internal/world"
)

func makeBiomeWorld(biome string, w, h int) *world.WorldMap {
	wm := world.NewWorldMap(w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			wm.TileAt(x, y).BiomeID = biome
		}
	}
	return wm
}

func countNPCs(w *ecs.World) int {
	store := w.GetStore(HealthID)
	if store == nil {
		return 0
	}
	cs := store.(*ecs.ComponentStore[Health])
	return cs.Len()
}

func TestSpawnCount(t *testing.T) {
	wm := makeBiomeWorld("plains", 64, 64)
	w := ecs.NewWorld()
	RegisterStores(w)

	raceDefs := []RaceDef{
		{ID: "human", SpawnWeight: 1.0, Roles: []RoleWeight{{ID: "farmer", Weight: 1.0}}, NamePool: NamePool{First: []string{"A"}, Last: []string{"B"}}},
	}
	roleDefs := []RoleDef{
		{ID: "farmer", Symbol: "@", Color: "#FFD700", Biomes: []string{"plains"}},
	}

	config := SpawnConfig{Density: 0.0015}
	if err := Spawn(w, wm, config, 42, raceDefs, roleDefs, nil); err != nil {
		t.Fatalf("Spawn error: %v", err)
	}

	count := countNPCs(w)
	if count < 50 || count > 100 {
		t.Errorf("spawned %d NPCs, want [50,100]", count)
	}
}

func TestSpawnDeterminism(t *testing.T) {
	wm := makeBiomeWorld("plains", 64, 64)
	config := SpawnConfig{Density: 0.0015}

	raceDefs := []RaceDef{
		{ID: "human", SpawnWeight: 1.0, Roles: []RoleWeight{{ID: "farmer", Weight: 1.0}}, NamePool: NamePool{First: []string{"A"}, Last: []string{"B"}}},
	}
	roleDefs := []RoleDef{
		{ID: "farmer", Symbol: "@", Color: "#FFD700", Biomes: []string{"plains"}},
	}

	w1 := ecs.NewWorld()
	RegisterStores(w1)
	if err := Spawn(w1, wm, config, 123, raceDefs, roleDefs, nil); err != nil {
		t.Fatalf("Spawn error: %v", err)
	}

	w2 := ecs.NewWorld()
	RegisterStores(w2)
	if err := Spawn(w2, wm, config, 123, raceDefs, roleDefs, nil); err != nil {
		t.Fatalf("Spawn error: %v", err)
	}

	// Compare positions of all spawned NPCs
	posStore1 := w1.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])
	posStore2 := w2.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])

	if len(posStore1.All()) != len(posStore2.All()) {
		t.Fatalf("different NPC counts: %d vs %d", len(posStore1.All()), len(posStore2.All()))
	}

	for e, p1 := range posStore1.All() {
		p2, ok := posStore2.Get(e)
		if !ok {
			t.Fatalf("entity %d missing in second world", e)
		}
		if p1 != p2 {
			t.Errorf("position mismatch for entity %d: %+v vs %+v", e, p1, p2)
		}
	}
}

func TestSpawnZeroInOcean(t *testing.T) {
	wm := makeBiomeWorld("ocean", 64, 64)
	w := ecs.NewWorld()
	RegisterStores(w)

	raceDefs := []RaceDef{
		{ID: "human", SpawnWeight: 1.0, Roles: []RoleWeight{{ID: "farmer", Weight: 1.0}}, NamePool: NamePool{First: []string{"A"}, Last: []string{"B"}}},
	}
	roleDefs := []RoleDef{
		{ID: "farmer", Symbol: "@", Color: "#FFD700", Biomes: []string{"plains"}},
	}

	config := SpawnConfig{Density: 0.0015}
	if err := Spawn(w, wm, config, 42, raceDefs, roleDefs, nil); err != nil {
		t.Fatalf("Spawn error: %v", err)
	}

	count := countNPCs(w)
	if count != 0 {
		t.Errorf("expected 0 NPCs in ocean-only world, got %d", count)
	}
}

func TestSpawnPlainsGreaterThanTundra(t *testing.T) {
	wm := world.NewWorldMap(64, 64)
	for y := 0; y < wm.Height; y++ {
		for x := 0; x < wm.Width; x++ {
			if x < 32 {
				wm.TileAt(x, y).BiomeID = "plains"
			} else {
				wm.TileAt(x, y).BiomeID = "tundra"
			}
		}
	}

	w := ecs.NewWorld()
	RegisterStores(w)

	raceDefs := []RaceDef{
		{ID: "human", SpawnWeight: 1.0, Roles: []RoleWeight{{ID: "farmer", Weight: 1.0}}, NamePool: NamePool{First: []string{"A"}, Last: []string{"B"}}},
	}
	roleDefs := []RoleDef{
		{ID: "farmer", Symbol: "@", Color: "#FFD700", Biomes: []string{"plains", "tundra"}},
	}

	config := SpawnConfig{Count: 200} // force many spawns for statistical significance
	if err := Spawn(w, wm, config, 42, raceDefs, roleDefs, nil); err != nil {
		t.Fatalf("Spawn error: %v", err)
	}

	posStore := w.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])
	plainsCount, tundraCount := 0, 0
	for _, p := range posStore.All() {
		x, y := int(p.X), int(p.Y)
		tile := wm.TileAt(x, y)
		if tile == nil {
			continue
		}
		if tile.BiomeID == "plains" {
			plainsCount++
		} else if tile.BiomeID == "tundra" {
			tundraCount++
		}
	}

	if plainsCount <= tundraCount {
		t.Errorf("expected more NPCs in plains (%d) than tundra (%d)", plainsCount, tundraCount)
	}
}

func TestSpawnDifferentSeedsDifferentPositions(t *testing.T) {
	wm := makeBiomeWorld("plains", 64, 64)
	config := SpawnConfig{Density: 0.0015}

	raceDefs := []RaceDef{
		{ID: "human", SpawnWeight: 1.0, Roles: []RoleWeight{{ID: "farmer", Weight: 1.0}}, NamePool: NamePool{First: []string{"A"}, Last: []string{"B"}}},
	}
	roleDefs := []RoleDef{
		{ID: "farmer", Symbol: "@", Color: "#FFD700", Biomes: []string{"plains"}},
	}

	w1 := ecs.NewWorld()
	RegisterStores(w1)
	if err := Spawn(w1, wm, config, 111, raceDefs, roleDefs, nil); err != nil {
		t.Fatalf("Spawn error: %v", err)
	}

	w2 := ecs.NewWorld()
	RegisterStores(w2)
	if err := Spawn(w2, wm, config, 222, raceDefs, roleDefs, nil); err != nil {
		t.Fatalf("Spawn error: %v", err)
	}

	posStore1 := w1.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])
	posStore2 := w2.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])

	// Same count but different positions
	if len(posStore1.All()) == 0 || len(posStore2.All()) == 0 {
		t.Fatal("expected NPCs in both worlds")
	}

	allSame := true
	for e, p1 := range posStore1.All() {
		p2, ok := posStore2.Get(e)
		if !ok || p1 != p2 {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("different seeds produced identical positions")
	}
}

func TestAppearanceVariesByRole(t *testing.T) {
	wm := makeBiomeWorld("plains", 64, 64)
	w := ecs.NewWorld()
	RegisterStores(w)

	raceDefs := []RaceDef{
		{ID: "human", SpawnWeight: 1.0, Roles: []RoleWeight{{ID: "farmer", Weight: 1.0}, {ID: "hunter", Weight: 1.0}}, NamePool: NamePool{First: []string{"A"}, Last: []string{"B"}}},
	}
	roleDefs := []RoleDef{
		{ID: "farmer", Symbol: "@", Color: "#FFD700", Biomes: []string{"plains"}},
		{ID: "hunter", Symbol: "h", Color: "#228B22", Biomes: []string{"plains"}},
	}

	config := SpawnConfig{Count: 10}
	if err := Spawn(w, wm, config, 42, raceDefs, roleDefs, nil); err != nil {
		t.Fatalf("Spawn error: %v", err)
	}

	appStore := w.GetStore(AppearanceID).(*ecs.ComponentStore[Appearance])
	farmerFound, hunterFound := false, false
	for _, app := range appStore.All() {
		if app.Symbol == '@' {
			farmerFound = true
		}
		if app.Symbol == 'h' {
			hunterFound = true
		}
	}
	if !farmerFound || !hunterFound {
		t.Error("expected different appearances for different roles")
	}
}

func TestAppearanceSameRole(t *testing.T) {
	wm := makeBiomeWorld("plains", 64, 64)
	w := ecs.NewWorld()
	RegisterStores(w)

	raceDefs := []RaceDef{
		{ID: "human", SpawnWeight: 1.0, Roles: []RoleWeight{{ID: "farmer", Weight: 1.0}}, NamePool: NamePool{First: []string{"A"}, Last: []string{"B"}}},
	}
	roleDefs := []RoleDef{
		{ID: "farmer", Symbol: "@", Color: "#FFD700", Biomes: []string{"plains"}},
	}

	config := SpawnConfig{Count: 5}
	if err := Spawn(w, wm, config, 42, raceDefs, roleDefs, nil); err != nil {
		t.Fatalf("Spawn error: %v", err)
	}

	appStore := w.GetStore(AppearanceID).(*ecs.ComponentStore[Appearance])
	var firstApp Appearance
	first := true
	for _, app := range appStore.All() {
		if first {
			firstApp = app
			first = false
			continue
		}
		if app.Symbol != firstApp.Symbol || app.Color != firstApp.Color {
			t.Error("same role should produce same appearance")
		}
	}
}

func TestSpawnRaceRoleRejection(t *testing.T) {
	wm := makeBiomeWorld("plains", 64, 64)
	w := ecs.NewWorld()
	RegisterStores(w)

	// Race references a role that does not exist in roleDefs
	raceDefs := []RaceDef{
		{ID: "human", SpawnWeight: 1.0, Roles: []RoleWeight{{ID: "wizard", Weight: 1.0}}, NamePool: NamePool{First: []string{"A"}, Last: []string{"B"}}},
	}
	roleDefs := []RoleDef{
		{ID: "farmer", Symbol: "@", Color: "#FFD700", Biomes: []string{"plains"}},
	}

	config := SpawnConfig{Count: 10}
	if err := Spawn(w, wm, config, 42, raceDefs, roleDefs, nil); err != nil {
		t.Fatalf("Spawn error: %v", err)
	}

	count := countNPCs(w)
	if count != 0 {
		t.Errorf("expected 0 NPCs when race roles are unknown, got %d", count)
	}
}

func TestSpawnFarmerInSettlement(t *testing.T) {
	wm := makeBiomeWorld("plains", 64, 64)
	w := ecs.NewWorld()
	RegisterStores(w)
	ecs.RegisterComponentStore[ecs.Position](w, ecs.NewComponentID("position"), ecs.NewComponentStore[ecs.Position]())
	ecs.RegisterComponentStore[ecs.Name](w, ecs.NewComponentID("name"), ecs.NewComponentStore[ecs.Name]())
	settlement.RegisterSettlementStores(w)

	// Create a village with a farm at (10,10)
	se := w.NewEntity()
	ecs.AddComponent(w, se, ecs.Position{X: 10, Y: 10})
	ecs.AddComponent(w, se, settlement.Settlement{
		Name: "Test Village", Type: "village", Radius: 3,
		Buildings: []string{"house", "farm"},
	})

	raceDefs := []RaceDef{
		{ID: "human", SpawnWeight: 1.0, Roles: []RoleWeight{{ID: "farmer", Weight: 1.0}}, NamePool: NamePool{First: []string{"A"}, Last: []string{"B"}}},
	}
	roleDefs := []RoleDef{
		{ID: "farmer", Symbol: "@", Color: "#FFD700", Biomes: []string{"plains"}},
	}

	config := SpawnConfig{Count: 1}
	if err := Spawn(w, wm, config, 42, raceDefs, roleDefs, []ecs.Entity{se}); err != nil {
		t.Fatalf("Spawn error: %v", err)
	}

	// Verify the NPC has a HomeReference to the settlement
	homeStore := w.GetStore(settlement.HomeRefID).(*ecs.ComponentStore[settlement.HomeReference])
	found := false
	for _, h := range homeStore.All() {
		if h.SettlementEntity == se {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected farmer to have HomeReference to settlement")
	}
}

func TestSpawnNomadNoHomeReference(t *testing.T) {
	wm := makeBiomeWorld("plains", 64, 64)
	w := ecs.NewWorld()
	RegisterStores(w)
	settlement.RegisterSettlementStores(w)

	raceDefs := []RaceDef{
		{ID: "human", SpawnWeight: 1.0, Roles: []RoleWeight{{ID: "farmer", Weight: 1.0}}, NamePool: NamePool{First: []string{"A"}, Last: []string{"B"}}},
	}
	roleDefs := []RoleDef{
		{ID: "farmer", Symbol: "@", Color: "#FFD700", Biomes: []string{"plains"}},
	}

	config := SpawnConfig{Count: 1}
	if err := Spawn(w, wm, config, 42, raceDefs, roleDefs, nil); err != nil {
		t.Fatalf("Spawn error: %v", err)
	}

	homeStore := w.GetStore(settlement.HomeRefID).(*ecs.ComponentStore[settlement.HomeReference])
	if homeStore.Len() != 0 {
		t.Errorf("expected 0 HomeReferences for nomad, got %d", homeStore.Len())
	}
}

func TestSpawnCapacityOverflow(t *testing.T) {
	wm := makeBiomeWorld("plains", 64, 64)
	w := ecs.NewWorld()
	RegisterStores(w)
	ecs.RegisterComponentStore[ecs.Position](w, ecs.NewComponentID("position"), ecs.NewComponentStore[ecs.Position]())
	ecs.RegisterComponentStore[ecs.Name](w, ecs.NewComponentID("name"), ecs.NewComponentStore[ecs.Name]())
	settlement.RegisterSettlementStores(w)

	// Create a small village with radius 1 (capacity 2)
	se := w.NewEntity()
	ecs.AddComponent(w, se, ecs.Position{X: 10, Y: 10})
	ecs.AddComponent(w, se, settlement.Settlement{
		Name: "Tiny Village", Type: "village", Radius: 1,
		Buildings: []string{"house", "farm"},
	})

	raceDefs := []RaceDef{
		{ID: "human", SpawnWeight: 1.0, Roles: []RoleWeight{{ID: "farmer", Weight: 1.0}}, NamePool: NamePool{First: []string{"A"}, Last: []string{"B"}}},
	}
	roleDefs := []RoleDef{
		{ID: "farmer", Symbol: "@", Color: "#FFD700", Biomes: []string{"plains"}},
	}

	config := SpawnConfig{Count: 5}
	if err := Spawn(w, wm, config, 42, raceDefs, roleDefs, []ecs.Entity{se}); err != nil {
		t.Fatalf("Spawn error: %v", err)
	}

	// Count how many NPCs have HomeReference to this settlement
	homeStore := w.GetStore(settlement.HomeRefID).(*ecs.ComponentStore[settlement.HomeReference])
	assigned := 0
	for _, h := range homeStore.All() {
		if h.SettlementEntity == se {
			assigned++
		}
	}
	if assigned > 2 {
		t.Errorf("expected at most 2 NPCs assigned to settlement (capacity 2), got %d", assigned)
	}
}

func TestSpawnMerchantInTown(t *testing.T) {
	wm := makeBiomeWorld("plains", 64, 64)
	w := ecs.NewWorld()
	RegisterStores(w)
	ecs.RegisterComponentStore[ecs.Position](w, ecs.NewComponentID("position"), ecs.NewComponentStore[ecs.Position]())
	ecs.RegisterComponentStore[ecs.Name](w, ecs.NewComponentID("name"), ecs.NewComponentStore[ecs.Name]())
	settlement.RegisterSettlementStores(w)

	// Create a town with market
	se := w.NewEntity()
	ecs.AddComponent(w, se, ecs.Position{X: 20, Y: 20})
	ecs.AddComponent(w, se, settlement.Settlement{
		Name: "Market Town", Type: "town", Radius: 5,
		Buildings: []string{"market", "house", "tavern"},
	})

	raceDefs := []RaceDef{
		{ID: "human", SpawnWeight: 1.0, Roles: []RoleWeight{{ID: "merchant", Weight: 1.0}}, NamePool: NamePool{First: []string{"A"}, Last: []string{"B"}}},
	}
	roleDefs := []RoleDef{
		{ID: "merchant", Symbol: "$", Color: "#FFD700", Biomes: []string{"plains"}},
	}

	config := SpawnConfig{Count: 3}
	if err := Spawn(w, wm, config, 42, raceDefs, roleDefs, []ecs.Entity{se}); err != nil {
		t.Fatalf("Spawn error: %v", err)
	}

	homeStore := w.GetStore(settlement.HomeRefID).(*ecs.ComponentStore[settlement.HomeReference])
	found := false
	for _, h := range homeStore.All() {
		if h.SettlementEntity == se {
			found = true
			break
		}
	}
	if !found {
		t.Error("merchant should spawn in town with market")
	}
}

func TestSpawnDeterminismWithSettlements(t *testing.T) {
	wm := makeBiomeWorld("plains", 64, 64)
	config := SpawnConfig{Density: 0.0015}

	raceDefs := []RaceDef{
		{ID: "human", SpawnWeight: 1.0, Roles: []RoleWeight{{ID: "farmer", Weight: 1.0}}, NamePool: NamePool{First: []string{"A"}, Last: []string{"B"}}},
	}
	roleDefs := []RoleDef{
		{ID: "farmer", Symbol: "@", Color: "#FFD700", Biomes: []string{"plains"}},
	}

	// Create a settlement for NPCs to spawn in
	createWorld := func(w *ecs.World, seed int64) {
		ecs.RegisterComponentStore[ecs.Position](w, ecs.NewComponentID("position"), ecs.NewComponentStore[ecs.Position]())
		ecs.RegisterComponentStore[ecs.Name](w, ecs.NewComponentID("name"), ecs.NewComponentStore[ecs.Name]())
		settlement.RegisterSettlementStores(w)
		se := w.NewEntity()
		ecs.AddComponent(w, se, ecs.Position{X: 20, Y: 20})
		ecs.AddComponent(w, se, settlement.Settlement{
			Name: "Test Village", Type: "village", Radius: 5,
			Buildings: []string{"farm", "house"},
		})
		if err := Spawn(w, wm, config, seed, raceDefs, roleDefs, []ecs.Entity{se}); err != nil {
			t.Fatalf("Spawn error with seed %d: %v", seed, err)
		}
	}

	w1 := ecs.NewWorld()
	RegisterStores(w1)
	createWorld(w1, 123)

	w2 := ecs.NewWorld()
	RegisterStores(w2)
	createWorld(w2, 123)

	// Check same number of NPCs
	posStore1 := w1.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])
	posStore2 := w2.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])
	if len(posStore1.All()) != len(posStore2.All()) {
		t.Errorf("different NPC counts: %d vs %d", len(posStore1.All()), len(posStore2.All()))
	}
}

func TestSpawnInsideSettlementIgnoresBiomeWeight(t *testing.T) {
	wm := makeBiomeWorld("ocean", 64, 64)
	// Override center tiles to be plains so settlement can spawn
	for y := 28; y < 36; y++ {
		for x := 28; x < 36; x++ {
			wm.TileAt(x, y).BiomeID = "plains"
		}
	}

	w := ecs.NewWorld()
	RegisterStores(w)
	ecs.RegisterComponentStore[ecs.Position](w, ecs.NewComponentID("position"), ecs.NewComponentStore[ecs.Position]())
	ecs.RegisterComponentStore[ecs.Name](w, ecs.NewComponentID("name"), ecs.NewComponentStore[ecs.Name]())
	settlement.RegisterSettlementStores(w)

	se := w.NewEntity()
	ecs.AddComponent(w, se, ecs.Position{X: 32, Y: 32})
	ecs.AddComponent(w, se, settlement.Settlement{
		Name: "Coastal Village", Type: "village", Radius: 3,
		Buildings: []string{"farm", "house"},
	})

	raceDefs := []RaceDef{
		{ID: "human", SpawnWeight: 1.0, Roles: []RoleWeight{{ID: "farmer", Weight: 1.0}}, NamePool: NamePool{First: []string{"A"}, Last: []string{"B"}}},
	}
	roleDefs := []RoleDef{
		{ID: "farmer", Symbol: "@", Color: "#FFD700", Biomes: []string{"plains"}},
	}

	config := SpawnConfig{Count: 5}
	if err := Spawn(w, wm, config, 42, raceDefs, roleDefs, []ecs.Entity{se}); err != nil {
		t.Fatalf("Spawn error: %v", err)
	}

	// NPCs should be inside settlement radius despite ocean biome
	posStore := w.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])
	homeStore := w.GetStore(settlement.HomeRefID).(*ecs.ComponentStore[settlement.HomeReference])
	if homeStore.Len() == 0 {
		t.Error("expected NPCs to spawn in settlement despite ocean biome")
	}
	for e := range homeStore.All() {
		pos, ok := posStore.Get(e)
		if !ok {
			continue
		}
		dist := max(abs(int(pos.X)-32), abs(int(pos.Y)-32))
		if dist > 3 {
			t.Errorf("NPC assigned to settlement but at distance %d from center (radius 3)", dist)
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
