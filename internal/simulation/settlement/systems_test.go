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

func TestBuildingRenderSystemCollectsBuildings(t *testing.T) {
	w := ecs.NewWorld()
	ecs.RegisterComponentStore[ecs.Position](w, ecs.NewComponentID("position"), ecs.NewComponentStore[ecs.Position]())
	RegisterSettlementStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, ecs.Position{X: 10, Y: 20})
	ecs.AddComponent(w, e, Building{ID: "farm", Name: "Granja", InteriorSymbol: "╬", Color: "#DEB887", SettlementEntity: ecs.Entity(1)})

	defs := []BuildingDef{
		{ID: "farm", Name: "Granja", Role: "farmer", MaxWorkers: 3, Produces: map[string]float64{"food": 2.0}},
	}
	sys := NewBuildingRenderSystem(defs)
	w.AddSystem(sys)
	if err := w.Update(0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	infos := sys.RenderInfos()
	if len(infos) != 1 {
		t.Fatalf("expected 1 render info, got %d", len(infos))
	}
	if infos[0].ID != "farm" {
		t.Errorf("ID = %q, want farm", infos[0].ID)
	}
	if infos[0].Symbol != '╬' {
		t.Errorf("Symbol = %q, want ╬", infos[0].Symbol)
	}
	if infos[0].Color != "#DEB887" {
		t.Errorf("Color = %q, want #DEB887", infos[0].Color)
	}
	if infos[0].WorldX != 10 || infos[0].WorldY != 20 {
		t.Errorf("position = (%d,%d), want (10,20)", infos[0].WorldX, infos[0].WorldY)
	}
	if infos[0].Role != "farmer" {
		t.Errorf("Role = %q, want farmer", infos[0].Role)
	}
	if infos[0].MaxWorkers != 3 {
		t.Errorf("MaxWorkers = %d, want 3", infos[0].MaxWorkers)
	}
	if len(infos[0].Produces) != 1 || infos[0].Produces["food"] != 2.0 {
		t.Errorf("Produces = %v, want map[food:2.0]", infos[0].Produces)
	}
	if len(infos[0].Consumes) != 0 {
		t.Errorf("Consumes = %v, want empty", infos[0].Consumes)
	}
}

func TestBuildingRenderSystemCollectsProducesAndConsumes(t *testing.T) {
	w := ecs.NewWorld()
	ecs.RegisterComponentStore[ecs.Position](w, ecs.NewComponentID("position"), ecs.NewComponentStore[ecs.Position]())
	RegisterSettlementStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, ecs.Position{X: 10, Y: 20})
	ecs.AddComponent(w, e, Building{ID: "market", Name: "Mercado", InteriorSymbol: "§", Color: "#FFD700", SettlementEntity: ecs.Entity(1)})

	defs := []BuildingDef{
		{ID: "market", Name: "Mercado", Role: "merchant", MaxWorkers: 2, Produces: map[string]float64{"gold": 1.0}, Consumes: map[string]float64{"food": 0.5}},
	}
	sys := NewBuildingRenderSystem(defs)
	w.AddSystem(sys)
	if err := w.Update(0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	infos := sys.RenderInfos()
	if len(infos) != 1 {
		t.Fatalf("expected 1 render info, got %d", len(infos))
	}
	if len(infos[0].Produces) != 1 || infos[0].Produces["gold"] != 1.0 {
		t.Errorf("Produces = %v, want map[gold:1.0]", infos[0].Produces)
	}
	if len(infos[0].Consumes) != 1 || infos[0].Consumes["food"] != 0.5 {
		t.Errorf("Consumes = %v, want map[food:0.5]", infos[0].Consumes)
	}
}

func TestBuildingRenderSystemSkipsZeroSymbol(t *testing.T) {
	w := ecs.NewWorld()
	ecs.RegisterComponentStore[ecs.Position](w, ecs.NewComponentID("position"), ecs.NewComponentStore[ecs.Position]())
	RegisterSettlementStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, ecs.Position{X: 10, Y: 20})
	ecs.AddComponent(w, e, Building{ID: "farm", Name: "Granja", InteriorSymbol: ""})

	sys := NewBuildingRenderSystem(nil)
	w.AddSystem(sys)
	if err := w.Update(0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	infos := sys.RenderInfos()
	if len(infos) != 0 {
		t.Errorf("expected 0 render info for zero-symbol building, got %d", len(infos))
	}
}

func TestBuildingRenderSystemRenderInfosForSettlement(t *testing.T) {
	w := ecs.NewWorld()
	ecs.RegisterComponentStore[ecs.Position](w, ecs.NewComponentID("position"), ecs.NewComponentStore[ecs.Position]())
	RegisterSettlementStores(w)

	settlementEntity := ecs.Entity(1)

	e1 := w.NewEntity()
	ecs.AddComponent(w, e1, ecs.Position{X: 10, Y: 20})
	ecs.AddComponent(w, e1, Building{ID: "farm", Name: "Granja", InteriorSymbol: "╬", Color: "#DEB887", SettlementEntity: settlementEntity})

	e2 := w.NewEntity()
	ecs.AddComponent(w, e2, ecs.Position{X: 30, Y: 40})
	ecs.AddComponent(w, e2, Building{ID: "house", Name: "Casa", InteriorSymbol: "⌂", Color: "#8B4513", SettlementEntity: ecs.Entity(2)})

	sys := NewBuildingRenderSystem(nil)
	w.AddSystem(sys)
	if err := w.Update(0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	filtered := sys.RenderInfosForSettlement(w, settlementEntity)
	if len(filtered) != 1 {
		t.Fatalf("expected 1 filtered render info, got %d", len(filtered))
	}
	if filtered[0].Entity != e1 {
		t.Errorf("expected entity %d, got %d", e1, filtered[0].Entity)
	}
}

func TestSettlementSpawnSystemCopiesInteriorFields(t *testing.T) {
	wm := makeBiomeWorld("plains", 256, 256)
	w := ecs.NewWorld()
	ecs.RegisterComponentStore[ecs.Position](w, ecs.NewComponentID("position"), ecs.NewComponentStore[ecs.Position]())
	ecs.RegisterComponentStore[ecs.Name](w, ecs.NewComponentID("name"), ecs.NewComponentStore[ecs.Name]())
	RegisterSettlementStores(w)

	settlementDefs := []SettlementDef{
		{ID: "village", Name: "Aldea", Symbol: "♦", Color: "#8B7355", Radius: 3, Biomes: []string{"plains"}, Buildings: []string{"house", "farm"}, SpawnWeight: 1.0},
	}
	buildingDefs := []BuildingDef{
		{ID: "house", Name: "Casa", InteriorSymbol: "⌂", Color: "#8B4513"},
		{ID: "farm", Name: "Granja", InteriorSymbol: "╬", Color: "#DEB887", Role: "farmer", MaxWorkers: 3},
	}

	sys := NewSettlementSpawnSystem(wm, 42, settlementDefs, buildingDefs)
	w.AddSystem(sys)
	_ = w.Update(0)

	buildStore, _ := w.GetStore(BuildingID).(*ecs.ComponentStore[Building])
	if buildStore == nil {
		t.Fatal("expected Building store")
	}

	foundHouse := false
	foundFarm := false
	for _, b := range buildStore.All() {
		if b.ID == "house" {
			foundHouse = true
			if b.InteriorSymbol != "⌂" {
				t.Errorf("house InteriorSymbol = %q, want ⌂", b.InteriorSymbol)
			}
			if b.Color != "#8B4513" {
				t.Errorf("house Color = %q, want #8B4513", b.Color)
			}
			if b.SettlementEntity == 0 {
				t.Error("expected house SettlementEntity to be set")
			}
		}
		if b.ID == "farm" {
			foundFarm = true
			if b.InteriorSymbol != "╬" {
				t.Errorf("farm InteriorSymbol = %q, want ╬", b.InteriorSymbol)
			}
			if b.Color != "#DEB887" {
				t.Errorf("farm Color = %q, want #DEB887", b.Color)
			}
		}
	}
	if !foundHouse {
		t.Error("expected at least one house building")
	}
	if !foundFarm {
		t.Error("expected at least one farm building")
	}
}
