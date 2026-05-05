package npc

import (
	"math/rand"
	"testing"

	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/world"
)

func TestNPCSpawnSystemRunsOnce(t *testing.T) {
	wm := makeBiomeWorld("plains", 64, 64)
	w := ecs.NewWorld()
	RegisterStores(w)

	raceDefs := []RaceDef{
		{ID: "human", SpawnWeight: 1.0, Roles: []RoleWeight{{ID: "farmer", Weight: 1.0}}, NamePool: NamePool{First: []string{"A"}, Last: []string{"B"}}},
	}
	roleDefs := []RoleDef{
		{ID: "farmer", Symbol: "@", Color: "#FFD700", Biomes: []string{"plains"}},
	}

	sys := NewNPCSpawnSystem(wm, SpawnConfig{Count: 5}, 42, raceDefs, roleDefs)
	w.AddSystem(sys)

	if err := w.Update(0); err != nil {
		t.Fatalf("Update error: %v", err)
	}
	count1 := countNPCs(w)
	if count1 != 5 {
		t.Errorf("expected 5 NPCs after first tick, got %d", count1)
	}

	if err := w.Update(0); err != nil {
		t.Fatalf("Update error: %v", err)
	}
	count2 := countNPCs(w)
	if count2 != 5 {
		t.Errorf("expected 5 NPCs after second tick, got %d", count2)
	}
}

func TestLODSystemTransitions(t *testing.T) {
	w := ecs.NewWorld()
	RegisterStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, ecs.Position{X: 10, Y: 10})
	ecs.AddComponent(w, e, LOD{Level: LODLocal})

	px, py := 10, 10
	playerPos := func() (int, int) { return px, py }
	sys := NewLODSystem(playerPos)
	w.AddSystem(sys)

	if err := w.Update(0); err != nil {
		t.Fatalf("Update error: %v", err)
	}
	lod, _ := ecs.GetComponent[LOD](w, e)
	if lod.Level != LODLocal {
		t.Errorf("expected LODLocal at distance 0, got %d", lod.Level)
	}

	px, py = 5, 5 // Chebyshev distance = 5
	if err := w.Update(0); err != nil {
		t.Fatalf("Update error: %v", err)
	}
	lod, _ = ecs.GetComponent[LOD](w, e)
	if lod.Level != LODLocal {
		t.Errorf("expected LODLocal at distance 5, got %d", lod.Level)
	}

	px, py = 0, 0 // Chebyshev distance = 10
	if err := w.Update(0); err != nil {
		t.Fatalf("Update error: %v", err)
	}
	lod, _ = ecs.GetComponent[LOD](w, e)
	if lod.Level != LODNear {
		t.Errorf("expected LODNear at distance 10, got %d", lod.Level)
	}

	px, py = -10, -10 // Chebyshev distance = 20
	if err := w.Update(0); err != nil {
		t.Fatalf("Update error: %v", err)
	}
	lod, _ = ecs.GetComponent[LOD](w, e)
	if lod.Level != LODDistant {
		t.Errorf("expected LODDistant at distance 20, got %d", lod.Level)
	}
}

func TestWanderSystemWithinBounds(t *testing.T) {
	wm := world.NewWorldMap(10, 10)
	for y := 0; y < wm.Height; y++ {
		for x := 0; x < wm.Width; x++ {
			wm.TileAt(x, y).BiomeID = "plains"
		}
	}

	w := ecs.NewWorld()
	RegisterStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, ecs.Position{X: 5, Y: 5})
	ecs.AddComponent(w, e, Job{Role: "farmer"})
	ecs.AddComponent(w, e, AIState{})
	ecs.AddComponent(w, e, LOD{Level: LODLocal})

	roleDefs := []RoleDef{{ID: "farmer", Symbol: "@", Color: "#FFD700", Biomes: []string{"plains"}}}
	sys := NewWanderSystem(wm, roleDefs, rand.New(rand.NewSource(1)))
	w.AddSystem(sys)

	for i := 0; i < 100; i++ {
		if err := w.Update(0); err != nil {
			t.Fatalf("Update error: %v", err)
		}
	}

	pos, _ := ecs.GetComponent[ecs.Position](w, e)
	if int(pos.X) < 0 || int(pos.X) >= 10 || int(pos.Y) < 0 || int(pos.Y) >= 10 {
		t.Errorf("NPC wandered out of bounds: %+v", pos)
	}
}

func TestWanderSystemStaysWhenSurrounded(t *testing.T) {
	wm := world.NewWorldMap(3, 3)
	wm.TileAt(1, 1).BiomeID = "plains"
	// All other tiles default to empty BiomeID (weight 0)

	w := ecs.NewWorld()
	RegisterStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, ecs.Position{X: 1, Y: 1})
	ecs.AddComponent(w, e, Job{Role: "farmer"})
	ecs.AddComponent(w, e, AIState{})
	ecs.AddComponent(w, e, LOD{Level: LODLocal})

	roleDefs := []RoleDef{{ID: "farmer", Symbol: "@", Color: "#FFD700", Biomes: []string{"plains"}}}
	sys := NewWanderSystem(wm, roleDefs, rand.New(rand.NewSource(99)))
	w.AddSystem(sys)

	for i := 0; i < 50; i++ {
		if err := w.Update(0); err != nil {
			t.Fatalf("Update error: %v", err)
		}
	}

	pos, _ := ecs.GetComponent[ecs.Position](w, e)
	if int(pos.X) != 1 || int(pos.Y) != 1 {
		t.Errorf("NPC moved when surrounded by incompatible biomes: %+v", pos)
	}
}

func TestNPCRenderSystemSkipsLOD0(t *testing.T) {
	w := ecs.NewWorld()
	RegisterStores(w)

	e1 := w.NewEntity()
	ecs.AddComponent(w, e1, ecs.Position{X: 0, Y: 0})
	ecs.AddComponent(w, e1, Appearance{Symbol: '@'})
	ecs.AddComponent(w, e1, LOD{Level: LODDistant})

	e2 := w.NewEntity()
	ecs.AddComponent(w, e2, ecs.Position{X: 1, Y: 1})
	ecs.AddComponent(w, e2, Appearance{Symbol: '@'})
	ecs.AddComponent(w, e2, LOD{Level: LODLocal})

	sys := NewNPCRenderSystem()
	w.AddSystem(sys)
	if err := w.Update(0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	infos := sys.RenderInfos()
	if len(infos) != 1 {
		t.Fatalf("expected 1 render info, got %d", len(infos))
	}
	if infos[0].WorldX != 1 || infos[0].WorldY != 1 {
		t.Errorf("wrong NPC rendered: %+v", infos[0])
	}
}

func TestAllSystemsExecutePerTick(t *testing.T) {
	wm := makeBiomeWorld("plains", 10, 10)
	w := ecs.NewWorld()
	RegisterStores(w)

	raceDefs := []RaceDef{
		{ID: "human", SpawnWeight: 1.0, Roles: []RoleWeight{{ID: "farmer", Weight: 1.0}}, NamePool: NamePool{First: []string{"A"}, Last: []string{"B"}}},
	}
	roleDefs := []RoleDef{
		{ID: "farmer", Symbol: "@", Color: "#FFD700", Biomes: []string{"plains"}},
	}

	renderSys := NewNPCRenderSystem()
	w.AddSystem(NewNPCSpawnSystem(wm, SpawnConfig{Count: 3}, 1, raceDefs, roleDefs))
	w.AddSystem(NewWanderSystem(wm, roleDefs, rand.New(rand.NewSource(1))))
	w.AddSystem(NewLODSystem(func() (int, int) { return 5, 5 }))
	w.AddSystem(renderSys)

	if err := w.Update(0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	if len(renderSys.RenderInfos()) != 3 {
		t.Errorf("expected 3 render infos, got %d", len(renderSys.RenderInfos()))
	}
}
