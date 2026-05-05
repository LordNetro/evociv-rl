package settlement

import (
	"math"
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

func countSettlements(w *ecs.World) int {
	store, ok := w.GetStore(SettlementID).(*ecs.ComponentStore[Settlement])
	if !ok {
		return 0
	}
	return store.Len()
}

func countBuildings(w *ecs.World) int {
	store, ok := w.GetStore(BuildingID).(*ecs.ComponentStore[Building])
	if !ok {
		return 0
	}
	return store.Len()
}

func TestSettlementSpawnSystemPlains(t *testing.T) {
	wm := makeBiomeWorld("plains", 256, 256)
	w := ecs.NewWorld()
	ecs.RegisterComponentStore[ecs.Position](w, ecs.NewComponentID("position"), ecs.NewComponentStore[ecs.Position]())
	ecs.RegisterComponentStore[ecs.Name](w, ecs.NewComponentID("name"), ecs.NewComponentStore[ecs.Name]())
	RegisterSettlementStores(w)

	settlementDefs := []SettlementDef{
		{ID: "village", Name: "Aldea", Symbol: "♦", Color: "#8B7355", Radius: 3, Biomes: []string{"plains", "forest"}, Buildings: []string{"house", "farm"}, SpawnWeight: 0.6},
		{ID: "town", Name: "Pueblo", Symbol: "▲", Color: "#B8860B", Radius: 5, Biomes: []string{"plains"}, Buildings: []string{"house", "market", "tavern", "blacksmith", "farm"}, SpawnWeight: 0.3},
		{ID: "city", Name: "Ciudad", Symbol: "●", Color: "#DAA520", Radius: 8, Biomes: []string{"plains"}, Buildings: []string{"house", "market", "temple", "tavern", "blacksmith", "farm"}, SpawnWeight: 0.1},
	}
	buildingDefs := []BuildingDef{
		{ID: "house", Name: "Casa"},
		{ID: "farm", Name: "Granja"},
		{ID: "market", Name: "Mercado"},
		{ID: "tavern", Name: "Taberna"},
		{ID: "temple", Name: "Templo"},
		{ID: "blacksmith", Name: "Herreria"},
	}

	sys := NewSettlementSpawnSystem(wm, 42, settlementDefs, buildingDefs)
	w.AddSystem(sys)

	if err := w.Update(0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	count := countSettlements(w)
	if count < 5 || count > 10 {
		t.Errorf("expected 5-10 settlements on plains, got %d", count)
	}

	// Verify buildings were spawned
	buildings := countBuildings(w)
	if buildings == 0 {
		t.Error("expected buildings to be spawned")
	}
}

func TestSettlementSpawnSystemOcean(t *testing.T) {
	wm := makeBiomeWorld("ocean", 256, 256)
	w := ecs.NewWorld()
	ecs.RegisterComponentStore[ecs.Position](w, ecs.NewComponentID("position"), ecs.NewComponentStore[ecs.Position]())
	ecs.RegisterComponentStore[ecs.Name](w, ecs.NewComponentID("name"), ecs.NewComponentStore[ecs.Name]())
	RegisterSettlementStores(w)

	settlementDefs := []SettlementDef{
		{ID: "village", Name: "Aldea", Symbol: "♦", Color: "#8B7355", Radius: 3, Biomes: []string{"plains", "forest"}, Buildings: []string{"house", "farm"}, SpawnWeight: 0.6},
		{ID: "town", Name: "Pueblo", Symbol: "▲", Color: "#B8860B", Radius: 5, Biomes: []string{"plains"}, Buildings: []string{"house", "market", "tavern", "blacksmith", "farm"}, SpawnWeight: 0.3},
		{ID: "city", Name: "Ciudad", Symbol: "●", Color: "#DAA520", Radius: 8, Biomes: []string{"plains"}, Buildings: []string{"house", "market", "temple", "tavern", "blacksmith", "farm"}, SpawnWeight: 0.1},
	}
	buildingDefs := []BuildingDef{
		{ID: "house", Name: "Casa"},
		{ID: "farm", Name: "Granja"},
		{ID: "market", Name: "Mercado"},
		{ID: "tavern", Name: "Taberna"},
		{ID: "temple", Name: "Templo"},
		{ID: "blacksmith", Name: "Herreria"},
	}

	sys := NewSettlementSpawnSystem(wm, 42, settlementDefs, buildingDefs)
	w.AddSystem(sys)

	if err := w.Update(0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	count := countSettlements(w)
	if count != 0 {
		t.Errorf("expected 0 settlements on ocean, got %d", count)
	}
}

func TestSettlementSpawnSystemDeterminism(t *testing.T) {
	wm := makeBiomeWorld("plains", 256, 256)
	settlementDefs := []SettlementDef{
		{ID: "village", Name: "Aldea", Symbol: "♦", Color: "#8B7355", Radius: 3, Biomes: []string{"plains"}, Buildings: []string{"house"}, SpawnWeight: 1.0},
	}
	buildingDefs := []BuildingDef{
		{ID: "house", Name: "Casa"},
	}

	// Run 1
	w1 := ecs.NewWorld()
	ecs.RegisterComponentStore[ecs.Position](w1, ecs.NewComponentID("position"), ecs.NewComponentStore[ecs.Position]())
	ecs.RegisterComponentStore[ecs.Name](w1, ecs.NewComponentID("name"), ecs.NewComponentStore[ecs.Name]())
	RegisterSettlementStores(w1)
	sys1 := NewSettlementSpawnSystem(wm, 99, settlementDefs, buildingDefs)
	w1.AddSystem(sys1)
	_ = w1.Update(0)

	// Run 2
	w2 := ecs.NewWorld()
	ecs.RegisterComponentStore[ecs.Position](w2, ecs.NewComponentID("position"), ecs.NewComponentStore[ecs.Position]())
	ecs.RegisterComponentStore[ecs.Name](w2, ecs.NewComponentID("name"), ecs.NewComponentStore[ecs.Name]())
	RegisterSettlementStores(w2)
	sys2 := NewSettlementSpawnSystem(wm, 99, settlementDefs, buildingDefs)
	w2.AddSystem(sys2)
	_ = w2.Update(0)

	posStore1, _ := w1.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])
	posStore2, _ := w2.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])

	if posStore1.Len() != posStore2.Len() {
		t.Fatalf("determinism failed: different entity counts %d vs %d", posStore1.Len(), posStore2.Len())
	}

	// Compare settlement positions (first entities are settlements)
	settlements1 := countSettlements(w1)
	settlements2 := countSettlements(w2)
	if settlements1 != settlements2 {
		t.Fatalf("different settlement counts: %d vs %d", settlements1, settlements2)
	}

	// All positions should match
	for e, p1 := range posStore1.All() {
		p2, ok := posStore2.Get(e)
		if !ok {
			t.Errorf("entity %d missing in run 2", e)
			continue
		}
		if p1.X != p2.X || p1.Y != p2.Y {
			t.Errorf("position mismatch for entity %d: (%.0f,%.0f) vs (%.0f,%.0f)", e, p1.X, p1.Y, p2.X, p2.Y)
		}
	}
}

func TestSettlementSpawnSystemMinDistance(t *testing.T) {
	wm := makeBiomeWorld("plains", 256, 256)
	w := ecs.NewWorld()
	ecs.RegisterComponentStore[ecs.Position](w, ecs.NewComponentID("position"), ecs.NewComponentStore[ecs.Position]())
	ecs.RegisterComponentStore[ecs.Name](w, ecs.NewComponentID("name"), ecs.NewComponentStore[ecs.Name]())
	RegisterSettlementStores(w)

	settlementDefs := []SettlementDef{
		{ID: "village", Name: "Aldea", Symbol: "♦", Color: "#8B7355", Radius: 3, Biomes: []string{"plains"}, Buildings: []string{"house"}, SpawnWeight: 1.0},
	}
	buildingDefs := []BuildingDef{
		{ID: "house", Name: "Casa"},
	}

	sys := NewSettlementSpawnSystem(wm, 42, settlementDefs, buildingDefs)
	w.AddSystem(sys)
	_ = w.Update(0)

	posStore, _ := w.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])
	setStore, _ := w.GetStore(SettlementID).(*ecs.ComponentStore[Settlement])

	var positions []ecs.Position
	for e, set := range setStore.All() {
		_ = set
		pos, ok := posStore.Get(e)
		if ok {
			positions = append(positions, pos)
		}
	}

	for i := 0; i < len(positions); i++ {
		for j := i + 1; j < len(positions); j++ {
			dx := math.Abs(float64(positions[i].X - positions[j].X))
			dy := math.Abs(float64(positions[i].Y - positions[j].Y))
			dist := dx
			if dy > dist {
				dist = dy
			}
			if dist < 20 {
				t.Errorf("settlements too close: %.0f tiles (min 20)", dist)
			}
		}
	}
}

func TestSettlementSpawnSystemBuildingsInsideRadius(t *testing.T) {
	wm := makeBiomeWorld("plains", 256, 256)
	w := ecs.NewWorld()
	ecs.RegisterComponentStore[ecs.Position](w, ecs.NewComponentID("position"), ecs.NewComponentStore[ecs.Position]())
	ecs.RegisterComponentStore[ecs.Name](w, ecs.NewComponentID("name"), ecs.NewComponentStore[ecs.Name]())
	RegisterSettlementStores(w)

	settlementDefs := []SettlementDef{
		{ID: "village", Name: "Aldea", Symbol: "♦", Color: "#8B7355", Radius: 3, Biomes: []string{"plains"}, Buildings: []string{"house", "farm"}, SpawnWeight: 1.0},
	}
	buildingDefs := []BuildingDef{
		{ID: "house", Name: "Casa"},
		{ID: "farm", Name: "Granja"},
	}

	sys := NewSettlementSpawnSystem(wm, 42, settlementDefs, buildingDefs)
	w.AddSystem(sys)
	_ = w.Update(0)

	posStore, _ := w.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])
	setStore, _ := w.GetStore(SettlementID).(*ecs.ComponentStore[Settlement])
	buildStore, _ := w.GetStore(BuildingID).(*ecs.ComponentStore[Building])

	// For each building, verify it is within the radius of at least one settlement
	for be, b := range buildStore.All() {
		_ = b
		bPos, ok := posStore.Get(be)
		if !ok {
			continue
		}
		insideAny := false
		for e, set := range setStore.All() {
			setPos, ok := posStore.Get(e)
			if !ok {
				continue
			}
			dx := math.Abs(float64(setPos.X - bPos.X))
			dy := math.Abs(float64(setPos.Y - bPos.Y))
			if dx <= float64(set.Radius) && dy <= float64(set.Radius) {
				insideAny = true
				break
			}
		}
		if !insideAny {
			t.Errorf("building at (%.0f,%.0f) not inside any settlement radius", bPos.X, bPos.Y)
		}
	}
}

func TestSettlementSpawnSystemRunsOnce(t *testing.T) {
	wm := makeBiomeWorld("plains", 256, 256)
	w := ecs.NewWorld()
	ecs.RegisterComponentStore[ecs.Position](w, ecs.NewComponentID("position"), ecs.NewComponentStore[ecs.Position]())
	ecs.RegisterComponentStore[ecs.Name](w, ecs.NewComponentID("name"), ecs.NewComponentStore[ecs.Name]())
	RegisterSettlementStores(w)

	settlementDefs := []SettlementDef{
		{ID: "village", Name: "Aldea", Symbol: "♦", Color: "#8B7355", Radius: 3, Biomes: []string{"plains"}, Buildings: []string{"house"}, SpawnWeight: 1.0},
	}
	buildingDefs := []BuildingDef{
		{ID: "house", Name: "Casa"},
	}

	sys := NewSettlementSpawnSystem(wm, 42, settlementDefs, buildingDefs)
	w.AddSystem(sys)

	_ = w.Update(0)
	count1 := countSettlements(w)

	_ = w.Update(0)
	count2 := countSettlements(w)

	if count1 != count2 {
		t.Errorf("expected same count after second tick, got %d vs %d", count1, count2)
	}
}
