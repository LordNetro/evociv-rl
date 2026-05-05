package npc

import (
	"testing"

	"github.com/marco/evociv-rl/internal/ecs"
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
	if err := Spawn(w, wm, config, 42, raceDefs, roleDefs); err != nil {
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
	if err := Spawn(w1, wm, config, 123, raceDefs, roleDefs); err != nil {
		t.Fatalf("Spawn error: %v", err)
	}

	w2 := ecs.NewWorld()
	RegisterStores(w2)
	if err := Spawn(w2, wm, config, 123, raceDefs, roleDefs); err != nil {
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
	if err := Spawn(w, wm, config, 42, raceDefs, roleDefs); err != nil {
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
	if err := Spawn(w, wm, config, 42, raceDefs, roleDefs); err != nil {
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
	if err := Spawn(w1, wm, config, 111, raceDefs, roleDefs); err != nil {
		t.Fatalf("Spawn error: %v", err)
	}

	w2 := ecs.NewWorld()
	RegisterStores(w2)
	if err := Spawn(w2, wm, config, 222, raceDefs, roleDefs); err != nil {
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
	if err := Spawn(w, wm, config, 42, raceDefs, roleDefs); err != nil {
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
	if err := Spawn(w, wm, config, 42, raceDefs, roleDefs); err != nil {
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
	if err := Spawn(w, wm, config, 42, raceDefs, roleDefs); err != nil {
		t.Fatalf("Spawn error: %v", err)
	}

	count := countNPCs(w)
	if count != 0 {
		t.Errorf("expected 0 NPCs when race roles are unknown, got %d", count)
	}
}
