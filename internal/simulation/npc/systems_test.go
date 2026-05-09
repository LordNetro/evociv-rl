package npc

import (
	"math/rand"
	"testing"

	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/simulation/settlement"
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

func TestNeedsDecaySystem(t *testing.T) {
	w := ecs.NewWorld()
	RegisterStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, Needs{Hunger: 0.0, Fatigue: 0.0})
	ecs.AddComponent(w, e, LOD{Level: LODLocal})

	sys := NewNeedsDecaySystem()
	w.AddSystem(sys)

	// Run 10 ticks with dt=1.0
	for i := 0; i < 10; i++ {
		if err := w.Update(1.0); err != nil {
			t.Fatalf("Update error: %v", err)
		}
	}

	needs, ok := ecs.GetComponent[Needs](w, e)
	if !ok {
		t.Fatal("expected Needs component")
	}
	if needs.Hunger < 0.09 || needs.Hunger > 0.11 {
		t.Errorf("expected Hunger ~0.10 after 10 ticks, got %f", needs.Hunger)
	}
	if needs.Fatigue < 0.045 || needs.Fatigue > 0.055 {
		t.Errorf("expected Fatigue ~0.05 after 10 ticks, got %f", needs.Fatigue)
	}
}

func TestNeedsDecaySystemClamps(t *testing.T) {
	w := ecs.NewWorld()
	RegisterStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, Needs{Hunger: 0.99, Fatigue: 0.0})
	ecs.AddComponent(w, e, LOD{Level: LODLocal})

	sys := NewNeedsDecaySystem()
	w.AddSystem(sys)

	// 5 ticks would take Hunger to 1.04 without clamp
	for i := 0; i < 5; i++ {
		if err := w.Update(1.0); err != nil {
			t.Fatalf("Update error: %v", err)
		}
	}

	needs, ok := ecs.GetComponent[Needs](w, e)
	if !ok {
		t.Fatal("expected Needs component")
	}
	if needs.Hunger > 1.0 {
		t.Errorf("expected Hunger clamped at 1.0, got %f", needs.Hunger)
	}
}

func TestNeedsDecaySystemLOD0(t *testing.T) {
	w := ecs.NewWorld()
	RegisterStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, Needs{Hunger: 0.0, Fatigue: 0.0})
	ecs.AddComponent(w, e, LOD{Level: LODDistant})

	sys := NewNeedsDecaySystem()
	w.AddSystem(sys)

	// Run 10 ticks with dt=1.0 at LOD0 (0.5x rate)
	for i := 0; i < 10; i++ {
		if err := w.Update(1.0); err != nil {
			t.Fatalf("Update error: %v", err)
		}
	}

	needs, ok := ecs.GetComponent[Needs](w, e)
	if !ok {
		t.Fatal("expected Needs component")
	}
	if needs.Hunger < 0.045 || needs.Hunger > 0.055 {
		t.Errorf("expected Hunger ~0.05 at LOD0, got %f", needs.Hunger)
	}
	if needs.Fatigue < 0.022 || needs.Fatigue > 0.028 {
		t.Errorf("expected Fatigue ~0.025 at LOD0, got %f", needs.Fatigue)
	}
}

func makeTestWorldMap(biome string, w, h int) *world.WorldMap {
	wm := world.NewWorldMap(w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			wm.TileAt(x, y).BiomeID = biome
		}
	}
	return wm
}

func TestGOAPSystemFullPlanAtLOD2(t *testing.T) {
	wm := makeTestWorldMap("plains", 10, 10)
	w := ecs.NewWorld()
	RegisterStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, ecs.Position{X: 5, Y: 5})
	ecs.AddComponent(w, e, Needs{Hunger: 0.8, Fatigue: 0.1})
	ecs.AddComponent(w, e, AIState{})
	ecs.AddComponent(w, e, LOD{Level: LODLocal})

	actions := []ActionDef{
		{ID: "harvest", Name: "Harvest", Requires: ActionRequires{Biomes: []string{"plains"}, NeedsMin: NeedsValues{Hunger: 0, Fatigue: 0}, NeedsMax: NeedsValues{Hunger: 1, Fatigue: 1}}, Effects: ActionEffects{HungerChange: -0.3}, Reward: ActionReward{Base: 1.0}},
		{ID: "rest", Name: "Rest", Requires: ActionRequires{NeedsMin: NeedsValues{Hunger: 0, Fatigue: 0}, NeedsMax: NeedsValues{Hunger: 1, Fatigue: 1}}, Effects: ActionEffects{FatigueChange: -0.4}, Reward: ActionReward{Base: 1.0}},
	}

	sys := NewGOAPSystem(wm, actions)
	w.AddSystem(sys)

	if err := w.Update(0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	ai, ok := ecs.GetComponent[AIState](w, e)
	if !ok {
		t.Fatal("expected AIState component")
	}
	if ai.CurrentAction == "" {
		t.Error("expected a current action at LOD2")
	}
	if len(ai.Goals) == 0 {
		t.Error("expected goals set for full plan at LOD2")
	}
}

func TestGOAPSystemOneStepAtLOD1(t *testing.T) {
	wm := makeTestWorldMap("plains", 10, 10)
	w := ecs.NewWorld()
	RegisterStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, ecs.Position{X: 5, Y: 5})
	ecs.AddComponent(w, e, Needs{Hunger: 0.8, Fatigue: 0.1})
	ecs.AddComponent(w, e, AIState{})
	ecs.AddComponent(w, e, LOD{Level: LODNear})

	actions := []ActionDef{
		{ID: "harvest", Name: "Harvest", Requires: ActionRequires{Biomes: []string{"plains"}, NeedsMin: NeedsValues{Hunger: 0, Fatigue: 0}, NeedsMax: NeedsValues{Hunger: 1, Fatigue: 1}}, Effects: ActionEffects{HungerChange: -0.3}, Reward: ActionReward{Base: 1.0}},
	}

	sys := NewGOAPSystem(wm, actions)
	w.AddSystem(sys)

	if err := w.Update(0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	ai, ok := ecs.GetComponent[AIState](w, e)
	if !ok {
		t.Fatal("expected AIState component")
	}
	if ai.CurrentAction == "" {
		t.Error("expected a current action at LOD1")
	}
	// LOD1 should not set goals (1-step only)
	if len(ai.Goals) != 0 {
		t.Errorf("expected no goals at LOD1, got %v", ai.Goals)
	}
}

func TestGOAPSystemSkipsLOD0(t *testing.T) {
	wm := makeTestWorldMap("plains", 10, 10)
	w := ecs.NewWorld()
	RegisterStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, ecs.Position{X: 5, Y: 5})
	ecs.AddComponent(w, e, Needs{Hunger: 0.8, Fatigue: 0.1})
	ecs.AddComponent(w, e, AIState{})
	ecs.AddComponent(w, e, LOD{Level: LODDistant})

	actions := []ActionDef{
		{ID: "harvest", Name: "Harvest", Requires: ActionRequires{Biomes: []string{"plains"}, NeedsMin: NeedsValues{Hunger: 0, Fatigue: 0}, NeedsMax: NeedsValues{Hunger: 1, Fatigue: 1}}, Effects: ActionEffects{HungerChange: -0.3}, Reward: ActionReward{Base: 1.0}},
	}

	sys := NewGOAPSystem(wm, actions)
	w.AddSystem(sys)

	if err := w.Update(0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	ai, ok := ecs.GetComponent[AIState](w, e)
	if !ok {
		t.Fatal("expected AIState component")
	}
	if ai.CurrentAction != "" {
		t.Errorf("expected no action at LOD0, got %s", ai.CurrentAction)
	}
}

func TestGOAPSystemSelectsCorrectAction(t *testing.T) {
	wm := makeTestWorldMap("plains", 10, 10)
	w := ecs.NewWorld()
	RegisterStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, ecs.Position{X: 5, Y: 5})
	ecs.AddComponent(w, e, Needs{Hunger: 0.1, Fatigue: 0.8})
	ecs.AddComponent(w, e, AIState{})
	ecs.AddComponent(w, e, LOD{Level: LODLocal})

	actions := []ActionDef{
		{ID: "harvest", Name: "Harvest", Requires: ActionRequires{Biomes: []string{"plains"}, NeedsMin: NeedsValues{Hunger: 0, Fatigue: 0}, NeedsMax: NeedsValues{Hunger: 1, Fatigue: 1}}, Effects: ActionEffects{HungerChange: -0.3}, Reward: ActionReward{Base: 1.0}},
		{ID: "rest", Name: "Rest", Requires: ActionRequires{NeedsMin: NeedsValues{Hunger: 0, Fatigue: 0}, NeedsMax: NeedsValues{Hunger: 1, Fatigue: 1}}, Effects: ActionEffects{FatigueChange: -0.4}, Reward: ActionReward{Base: 1.0}},
	}

	sys := NewGOAPSystem(wm, actions)
	w.AddSystem(sys)

	if err := w.Update(0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	ai, ok := ecs.GetComponent[AIState](w, e)
	if !ok {
		t.Fatal("expected AIState component")
	}
	if ai.CurrentAction != "rest" {
		t.Errorf("expected rest for high fatigue, got %s", ai.CurrentAction)
	}
}

func TestQLearningSystemFallback(t *testing.T) {
	wm := makeTestWorldMap("plains", 10, 10)
	w := ecs.NewWorld()
	RegisterStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, ecs.Position{X: 5, Y: 5})
	ecs.AddComponent(w, e, Needs{Hunger: 0.8, Fatigue: 0.1})
	ecs.AddComponent(w, e, AIState{CurrentAction: "harvest"})
	ecs.AddComponent(w, e, LOD{Level: LODLocal})

	actions := []ActionDef{
		{ID: "harvest", Name: "Harvest", Requires: ActionRequires{Biomes: []string{"plains"}, NeedsMin: NeedsValues{Hunger: 0, Fatigue: 0}, NeedsMax: NeedsValues{Hunger: 1, Fatigue: 1}}, Effects: ActionEffects{HungerChange: -0.3}, Reward: ActionReward{Base: 1.0}},
		{ID: "rest", Name: "Rest", Requires: ActionRequires{NeedsMin: NeedsValues{Hunger: 0, Fatigue: 0}, NeedsMax: NeedsValues{Hunger: 1, Fatigue: 1}}, Effects: ActionEffects{FatigueChange: -0.4}, Reward: ActionReward{Base: 1.0}},
	}

	qt := NewTestQLearningSystem(wm, actions)
	w.AddSystem(qt)

	if err := w.Update(1.0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	// Since Q-table is empty, fallback to GOAP action should happen
	// The action should have been executed (needs changed)
	needs, ok := ecs.GetComponent[Needs](w, e)
	if !ok {
		t.Fatal("expected Needs component")
	}
	// harvest reduces hunger by 0.3: 0.8 - 0.3 = 0.5
	if needs.Hunger < 0.4 || needs.Hunger > 0.6 {
		t.Errorf("expected hunger ~0.5 after harvest, got %f", needs.Hunger)
	}
}

func TestQLearningSystemSkipsLOD0(t *testing.T) {
	wm := makeTestWorldMap("plains", 10, 10)
	w := ecs.NewWorld()
	RegisterStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, ecs.Position{X: 5, Y: 5})
	ecs.AddComponent(w, e, Needs{Hunger: 0.8, Fatigue: 0.1})
	ecs.AddComponent(w, e, AIState{CurrentAction: "harvest"})
	ecs.AddComponent(w, e, LOD{Level: LODDistant})

	actions := []ActionDef{
		{ID: "harvest", Name: "Harvest", Effects: ActionEffects{HungerChange: -0.3}, Reward: ActionReward{Base: 1.0}},
	}

	qt := NewTestQLearningSystem(wm, actions)
	w.AddSystem(qt)

	if err := w.Update(1.0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	needs, ok := ecs.GetComponent[Needs](w, e)
	if !ok {
		t.Fatal("expected Needs component")
	}
	// LOD0 should skip Q-learning; needs unchanged
	if needs.Hunger != 0.8 {
		t.Errorf("expected hunger unchanged at LOD0, got %f", needs.Hunger)
	}
}

func NewTestQLearningSystem(wm *world.WorldMap, actions []ActionDef) *QLearningSystem {
	return NewQLearningSystem(wm, actions, rand.New(rand.NewSource(1)))
}

func TestIntegrationThreeSystemsInOrder(t *testing.T) {
	wm := makeTestWorldMap("plains", 10, 10)
	w := ecs.NewWorld()
	RegisterStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, ecs.Position{X: 5, Y: 5})
	ecs.AddComponent(w, e, Needs{Hunger: 0.0, Fatigue: 0.0})
	ecs.AddComponent(w, e, AIState{})
	ecs.AddComponent(w, e, LOD{Level: LODLocal})

	actions := []ActionDef{
		{ID: "harvest", Name: "Harvest", Requires: ActionRequires{Biomes: []string{"plains"}, NeedsMin: NeedsValues{Hunger: 0, Fatigue: 0}, NeedsMax: NeedsValues{Hunger: 1, Fatigue: 1}}, Effects: ActionEffects{HungerChange: -0.3}, Reward: ActionReward{Base: 1.0}},
		{ID: "rest", Name: "Rest", Requires: ActionRequires{NeedsMin: NeedsValues{Hunger: 0, Fatigue: 0}, NeedsMax: NeedsValues{Hunger: 1, Fatigue: 1}}, Effects: ActionEffects{FatigueChange: -0.4}, Reward: ActionReward{Base: 1.0}},
	}

	w.AddSystem(NewNeedsDecaySystem())
	w.AddSystem(NewGOAPSystem(wm, actions))
	w.AddSystem(NewTestQLearningSystem(wm, actions))

	// Run one tick
	if err := w.Update(1.0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	// Verify order: NeedsDecay ran first, then GOAP planned, then QL executed
	needs, ok := ecs.GetComponent[Needs](w, e)
	if !ok {
		t.Fatal("expected Needs component")
	}
	// After decay: Hunger = 0.01, then harvest reduces by 0.3, clamped to 0
	if needs.Hunger != 0.0 {
		t.Errorf("expected hunger clamped to 0.0 after decay+harvest, got %f", needs.Hunger)
	}

	ai, ok := ecs.GetComponent[AIState](w, e)
	if !ok {
		t.Fatal("expected AIState component")
	}
	if ai.CurrentAction == "" {
		t.Error("expected an action was executed by QLearningSystem")
	}
}

func TestIntegrationLOD0SkipsGOAPAndQL(t *testing.T) {
	wm := makeTestWorldMap("plains", 10, 10)
	w := ecs.NewWorld()
	RegisterStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, ecs.Position{X: 5, Y: 5})
	ecs.AddComponent(w, e, Needs{Hunger: 0.0, Fatigue: 0.0})
	ecs.AddComponent(w, e, AIState{})
	ecs.AddComponent(w, e, LOD{Level: LODDistant})

	actions := []ActionDef{
		{ID: "harvest", Name: "Harvest", Requires: ActionRequires{Biomes: []string{"plains"}, NeedsMin: NeedsValues{Hunger: 0, Fatigue: 0}, NeedsMax: NeedsValues{Hunger: 1, Fatigue: 1}}, Effects: ActionEffects{HungerChange: -0.3}, Reward: ActionReward{Base: 1.0}},
	}

	w.AddSystem(NewNeedsDecaySystem())
	w.AddSystem(NewGOAPSystem(wm, actions))
	w.AddSystem(NewTestQLearningSystem(wm, actions))

	if err := w.Update(1.0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	needs, ok := ecs.GetComponent[Needs](w, e)
	if !ok {
		t.Fatal("expected Needs component")
	}
	// LOD0: only decay at 0.5x rate
	if needs.Hunger < 0.004 || needs.Hunger > 0.006 {
		t.Errorf("expected hunger ~0.005 at LOD0, got %f", needs.Hunger)
	}

	ai, ok := ecs.GetComponent[AIState](w, e)
	if !ok {
		t.Fatal("expected AIState component")
	}
	if ai.CurrentAction != "" {
		t.Error("expected no action at LOD0")
	}
}

func TestQLearningSystemWritesLastReward(t *testing.T) {
	wm := makeTestWorldMap("plains", 10, 10)
	w := ecs.NewWorld()
	RegisterStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, ecs.Position{X: 5, Y: 5})
	ecs.AddComponent(w, e, Needs{Hunger: 0.8, Fatigue: 0.1})
	ecs.AddComponent(w, e, AIState{CurrentAction: "harvest"})
	ecs.AddComponent(w, e, LOD{Level: LODLocal})

	actions := []ActionDef{
		{ID: "harvest", Name: "Harvest", Requires: ActionRequires{Biomes: []string{"plains"}, NeedsMin: NeedsValues{Hunger: 0, Fatigue: 0}, NeedsMax: NeedsValues{Hunger: 1, Fatigue: 1}}, Effects: ActionEffects{HungerChange: -0.3}, Reward: ActionReward{Base: 1.0}},
	}

	qt := NewTestQLearningSystem(wm, actions)
	w.AddSystem(qt)

	if err := w.Update(1.0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	ai, ok := ecs.GetComponent[AIState](w, e)
	if !ok {
		t.Fatal("expected AIState component")
	}
	if ai.LastReward <= 0 {
		t.Errorf("expected LastReward > 0 after Q-learning, got %f", ai.LastReward)
	}
	if ai.RewardTick == 0 {
		t.Error("expected RewardTick to be set")
	}
}

func TestNPCRenderSystemIncludesRewardAndRole(t *testing.T) {
	w := ecs.NewWorld()
	RegisterStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, ecs.Position{X: 1, Y: 2})
	ecs.AddComponent(w, e, Appearance{Symbol: '@'})
	ecs.AddComponent(w, e, LOD{Level: LODLocal})
	ecs.AddComponent(w, e, AIState{LastReward: 0.87, RewardTick: 42})
	ecs.AddComponent(w, e, Job{Role: "farmer"})

	sys := NewNPCRenderSystem()
	w.AddSystem(sys)
	if err := w.Update(0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	infos := sys.RenderInfos()
	if len(infos) != 1 {
		t.Fatalf("expected 1 render info, got %d", len(infos))
	}
	if infos[0].LastReward != 0.87 {
		t.Errorf("LastReward = %f, want 0.87", infos[0].LastReward)
	}
	if infos[0].RewardTick != 42 {
		t.Errorf("RewardTick = %d, want 42", infos[0].RewardTick)
	}
	if infos[0].JobRole != "farmer" {
		t.Errorf("JobRole = %q, want farmer", infos[0].JobRole)
	}
}

func TestNPCRenderSystemZeroRewardWithoutAIState(t *testing.T) {
	w := ecs.NewWorld()
	RegisterStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, ecs.Position{X: 1, Y: 2})
	ecs.AddComponent(w, e, Appearance{Symbol: '@'})
	ecs.AddComponent(w, e, LOD{Level: LODLocal})
	ecs.AddComponent(w, e, Job{Role: "merchant"})
	// No AIState component

	sys := NewNPCRenderSystem()
	w.AddSystem(sys)
	if err := w.Update(0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	infos := sys.RenderInfos()
	if len(infos) != 1 {
		t.Fatalf("expected 1 render info, got %d", len(infos))
	}
	if infos[0].LastReward != 0.0 {
		t.Errorf("LastReward = %f, want 0.0", infos[0].LastReward)
	}
	if infos[0].RewardTick != 0 {
		t.Errorf("RewardTick = %d, want 0", infos[0].RewardTick)
	}
	if infos[0].JobRole != "merchant" {
		t.Errorf("JobRole = %q, want merchant", infos[0].JobRole)
	}
}

func TestNPCRenderSystemRenderInfosForSettlement(t *testing.T) {
	w := ecs.NewWorld()
	RegisterStores(w)
	settlement.RegisterSettlementStores(w)

	settlementEntity := w.NewEntity()
	ecs.AddComponent(w, settlementEntity, settlement.Settlement{Name: "Aldea"})

	e1 := w.NewEntity()
	ecs.AddComponent(w, e1, ecs.Position{X: 1, Y: 2})
	ecs.AddComponent(w, e1, Appearance{Symbol: '@'})
	ecs.AddComponent(w, e1, LOD{Level: LODLocal})
	ecs.AddComponent(w, e1, settlement.HomeReference{SettlementEntity: settlementEntity})

	e2 := w.NewEntity()
	ecs.AddComponent(w, e2, ecs.Position{X: 3, Y: 4})
	ecs.AddComponent(w, e2, Appearance{Symbol: '#'})
	ecs.AddComponent(w, e2, LOD{Level: LODLocal})
	// No HomeReference — should not appear

	sys := NewNPCRenderSystem()
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

// TestBuildingActionsWorkerCount verifies that enter_building increments WorkersInside
// and exit_building decrements WorkersInside.
func TestBuildingActionsWorkerCount(t *testing.T) {
	wm := makeTestWorldMap("plains", 50, 50)
	w := ecs.NewWorld()
	RegisterStores(w)

	// Register settlement stores (required for building actions)
	settlement.RegisterSettlementStores(w)

	// Create a building with BuildingInterior
	buildingEntity := w.NewEntity()
	ecs.AddComponent(w, buildingEntity, ecs.Position{X: 10, Y: 10})
	ecs.AddComponent(w, buildingEntity, settlement.Building{ID: "house", Name: "Casa"})
	ecs.AddComponent(w, buildingEntity, settlement.BuildingInterior{
		Width:         5,
		Height:        5,
		WorkersInside: 0,
		MaxWorkers:    2,
		BuildingSeed:   12345,
	})

	// Create a worker NPC
	worker := w.NewEntity()
	ecs.AddComponent(w, worker, ecs.Position{X: 12, Y: 12})
	ecs.AddComponent(w, worker, Needs{Hunger: 0.5, Fatigue: 0.5})
	ecs.AddComponent(w, worker, AIState{CurrentAction: "enter_building", InsideBuilding: 0})
	ecs.AddComponent(w, worker, LOD{Level: LODLocal})

	// Create worker action (similar to actions.yaml)
	buildingActions := []ActionDef{
		{ID: "enter_building", Name: "Enter Building", Requires: ActionRequires{NeedsMin: NeedsValues{Hunger: 0, Fatigue: 0}, NeedsMax: NeedsValues{Hunger: 1, Fatigue: 1}}, Effects: ActionEffects{HungerChange: 0}, Reward: ActionReward{Base: 0.5}},
		{ID: "work_inside", Name: "Work Inside", Requires: ActionRequires{NeedsMin: NeedsValues{Hunger: 0, Fatigue: 0}, NeedsMax: NeedsValues{Hunger: 0.8, Fatigue: 0.8}}, Effects: ActionEffects{HungerChange: 0.05}, Reward: ActionReward{Base: 1.5}},
		{ID: "exit_building", Name: "Exit Building", Requires: ActionRequires{NeedsMin: NeedsValues{Hunger: 0, Fatigue: 0}, NeedsMax: NeedsValues{Hunger: 1, Fatigue: 1}}, Effects: ActionEffects{HungerChange: 0}, Reward: ActionReward{Base: 0.2}},
	}

	qlSys := NewQLearningSystem(wm, buildingActions, rand.New(rand.NewSource(42)))
	w.AddSystem(qlSys)

	// Verify initial state
	bi, ok := ecs.GetComponent[settlement.BuildingInterior](w, buildingEntity)
	if !ok {
		t.Fatal("expected BuildingInterior component")
	}
	if bi.WorkersInside != 0 {
		t.Errorf("expected WorkersInside=0 initially, got %d", bi.WorkersInside)
	}

	// Execute enter_building action
	if err := w.Update(0.2); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	// Verify WorkersInside incremented
	bi, ok = ecs.GetComponent[settlement.BuildingInterior](w, buildingEntity)
	if !ok {
		t.Fatal("expected BuildingInterior component after update")
	}
	if bi.WorkersInside != 1 {
		t.Errorf("expected WorkersInside=1 after enter, got %d", bi.WorkersInside)
	}

	// Verify AIState updated
	ai, ok := ecs.GetComponent[AIState](w, worker)
	if !ok {
		t.Fatal("expected AIState component")
	}
	if ai.InsideBuilding != buildingEntity {
		t.Errorf("expected InsideBuilding=%d, got %d", buildingEntity, ai.InsideBuilding)
	}

	// Execute work_inside (should keep WorkersInside at 1)
	ai.CurrentAction = "work_inside"
	ecs.AddComponent(w, worker, ai)
	if err := w.Update(0.2); err != nil {
		t.Fatalf("Update error: %v", err)
	}
	bi, ok = ecs.GetComponent[settlement.BuildingInterior](w, buildingEntity)
	if !ok {
		t.Fatal("expected BuildingInterior component after work_inside")
	}
	if bi.WorkersInside != 1 {
		t.Errorf("expected WorkersInside=1 after work_inside, got %d", bi.WorkersInside)
	}

	// Execute exit_building
	ai, _ = ecs.GetComponent[AIState](w, worker)
	ai.CurrentAction = "exit_building"
	ecs.AddComponent(w, worker, ai)
	if err := w.Update(0.2); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	// Verify WorkersInside decremented
	bi, ok = ecs.GetComponent[settlement.BuildingInterior](w, buildingEntity)
	if !ok {
		t.Fatal("expected BuildingInterior component after exit")
	}
	if bi.WorkersInside != 0 {
		t.Errorf("expected WorkersInside=0 after exit, got %d", bi.WorkersInside)
	}

	// Verify AIState cleared
	ai, ok = ecs.GetComponent[AIState](w, worker)
	if !ok {
		t.Fatal("expected AIState component after exit")
	}
	if ai.InsideBuilding != 0 {
		t.Errorf("expected InsideBuilding=0 after exit, got %d", ai.InsideBuilding)
	}
}

// TestBuildingActionsCapacityCap verifies that WorkersInside is capped at MaxWorkers.
func TestBuildingActionsCapacityCap(t *testing.T) {
	wm := makeTestWorldMap("plains", 50, 50)
	w := ecs.NewWorld()
	RegisterStores(w)
	settlement.RegisterSettlementStores(w)

	// Create a building with MaxWorkers=1
	buildingEntity := w.NewEntity()
	ecs.AddComponent(w, buildingEntity, ecs.Position{X: 10, Y: 10})
	ecs.AddComponent(w, buildingEntity, settlement.Building{ID: "house", Name: "Casa"})
	ecs.AddComponent(w, buildingEntity, settlement.BuildingInterior{
		Width:         5,
		Height:        5,
		WorkersInside: 1, // Already at capacity
		MaxWorkers:    1,
		BuildingSeed:   12345,
	})

	// Create a worker NPC
	worker := w.NewEntity()
	ecs.AddComponent(w, worker, ecs.Position{X: 12, Y: 12})
	ecs.AddComponent(w, worker, Needs{Hunger: 0.5, Fatigue: 0.5})
	ecs.AddComponent(w, worker, AIState{CurrentAction: "enter_building", InsideBuilding: 0})
	ecs.AddComponent(w, worker, LOD{Level: LODLocal})

	buildingActions := []ActionDef{
		{ID: "enter_building", Name: "Enter Building", Requires: ActionRequires{NeedsMin: NeedsValues{Hunger: 0, Fatigue: 0}, NeedsMax: NeedsValues{Hunger: 1, Fatigue: 1}}, Effects: ActionEffects{HungerChange: 0}, Reward: ActionReward{Base: 0.5}},
	}

	qlSys := NewQLearningSystem(wm, buildingActions, rand.New(rand.NewSource(42)))
	w.AddSystem(qlSys)

	// Execute enter_building (should be blocked by capacity)
	if err := w.Update(0.2); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	// Verify WorkersInside stayed at 1 (capacity cap)
	bi, ok := ecs.GetComponent[settlement.BuildingInterior](w, buildingEntity)
	if !ok {
		t.Fatal("expected BuildingInterior component")
	}
	if bi.WorkersInside != 1 {
		t.Errorf("expected WorkersInside=1 (capacity capped), got %d", bi.WorkersInside)
	}
}
