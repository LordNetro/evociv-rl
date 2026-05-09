package economy

import (
	"testing"

	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/simulation/npc"
	"github.com/marco/evociv-rl/internal/simulation/settlement"
)

func setupWorld(t *testing.T) *ecs.World {
	w := ecs.NewWorld()
	npc.RegisterStores(w)
	settlement.RegisterSettlementStores(w)
	return w
}

func buildingMap(defs []settlement.BuildingDef) map[string]settlement.BuildingDef {
	m := make(map[string]settlement.BuildingDef, len(defs))
	for _, d := range defs {
		m[d.ID] = d
	}
	return m
}

func TestEconomySystemFarmProducesFood(t *testing.T) {
	w := setupWorld(t)

	// Create settlement with farm
	settle := w.NewEntity()
	ecs.AddComponent(w, settle, ecs.Position{X: 10, Y: 10})
	ecs.AddComponent(w, settle, settlement.Settlement{
		Name: "Testville", Type: "village", Radius: 3,
		Buildings: []string{"farm"},
	})

	// Create building entity with BuildingInterior (2 workers inside)
	build := w.NewEntity()
	ecs.AddComponent(w, build, ecs.Position{X: 12, Y: 12})
	ecs.AddComponent(w, build, settlement.Building{
		ID:               "farm",
		Name:             "Granja",
		SettlementEntity: settle,
	})
	ecs.AddComponent(w, build, settlement.BuildingInterior{
		WorkersInside: 2,
		MaxWorkers:    3,
		Grid:          createTestGrid(5, 5),
		Width:         5,
		Height:        5,
	})

	// Create 2 farmer NPCs with HomeReference to this settlement (for fallback)
	for i := 0; i < 2; i++ {
		n := w.NewEntity()
		ecs.AddComponent(w, n, npc.Job{Role: "farmer"})
		ecs.AddComponent(w, n, settlement.HomeReference{SettlementEntity: settle})
	}

	defs := []settlement.BuildingDef{
		{ID: "farm", Name: "Granja", Role: "farmer", Produces: map[string]float64{"food": 2.0}, MaxWorkers: 3},
	}
	sys := NewSettlementEconomySystem(defs)
	if err := sys.Update(w, 1.0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	rs, ok := ecs.GetComponent[settlement.ResourceStore](w, settle)
	if !ok {
		t.Fatal("expected ResourceStore to be created")
	}
	// 2 workers inside * 2.0 food = 4.0 production, 2 NPCs * 0.01 = 0.02 consumption
	if rs.Resources["food"] != 3.98 {
		t.Errorf("food = %f, want 3.98", rs.Resources["food"])
	}
}

func TestEconomySystemBlacksmithProducesTools(t *testing.T) {
	w := setupWorld(t)

	settle := w.NewEntity()
	ecs.AddComponent(w, settle, ecs.Position{X: 10, Y: 10})
	ecs.AddComponent(w, settle, settlement.Settlement{
		Name: "Testville", Type: "village", Radius: 3,
		Buildings: []string{"blacksmith"},
	})

	// Create building entity with BuildingInterior (1 worker inside)
	build := w.NewEntity()
	ecs.AddComponent(w, build, ecs.Position{X: 12, Y: 12})
	ecs.AddComponent(w, build, settlement.Building{
		ID:               "blacksmith",
		Name:             "Herreria",
		SettlementEntity: settle,
	})
	ecs.AddComponent(w, build, settlement.BuildingInterior{
		WorkersInside: 1,
		MaxWorkers:    2,
		Grid:          createTestGrid(5, 5),
		Width:         5,
		Height:        5,
	})

	n := w.NewEntity()
	ecs.AddComponent(w, n, npc.Job{Role: "smith"})
	ecs.AddComponent(w, n, settlement.HomeReference{SettlementEntity: settle})

	defs := []settlement.BuildingDef{
		{ID: "blacksmith", Name: "Herreria", Role: "smith", Produces: map[string]float64{"tools": 1.0}, MaxWorkers: 2},
	}
	sys := NewSettlementEconomySystem(defs)
	if err := sys.Update(w, 1.0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	rs, ok := ecs.GetComponent[settlement.ResourceStore](w, settle)
	if !ok {
		t.Fatal("expected ResourceStore to be created")
	}
	// 1 worker inside * 1.0 tools = 1.0 production
	if rs.Resources["tools"] != 1.0 {
		t.Errorf("tools = %f, want 1.0", rs.Resources["tools"])
	}
}

func TestEconomySystemMarketProducesGoldAndConsumesFood(t *testing.T) {
	w := setupWorld(t)

	settle := w.NewEntity()
	ecs.AddComponent(w, settle, ecs.Position{X: 10, Y: 10})
	ecs.AddComponent(w, settle, settlement.Settlement{
		Name: "Testville", Type: "village", Radius: 3,
		Buildings: []string{"market"},
	})
	ecs.AddComponent(w, settle, settlement.ResourceStore{Resources: map[string]float64{"food": 10.0}})

	// Create building entity with BuildingInterior (1 worker inside)
	build := w.NewEntity()
	ecs.AddComponent(w, build, ecs.Position{X: 12, Y: 12})
	ecs.AddComponent(w, build, settlement.Building{
		ID:               "market",
		Name:             "Mercado",
		SettlementEntity: settle,
	})
	ecs.AddComponent(w, build, settlement.BuildingInterior{
		WorkersInside: 1,
		MaxWorkers:    2,
		Grid:          createTestGrid(5, 5),
		Width:         5,
		Height:        5,
	})

	n := w.NewEntity()
	ecs.AddComponent(w, n, npc.Job{Role: "merchant"})
	ecs.AddComponent(w, n, settlement.HomeReference{SettlementEntity: settle})

	defs := []settlement.BuildingDef{
		{ID: "market", Name: "Mercado", Role: "merchant", Produces: map[string]float64{"gold": 1.0}, Consumes: map[string]float64{"food": 0.5}, MaxWorkers: 2},
	}
	sys := NewSettlementEconomySystem(defs)
	if err := sys.Update(w, 1.0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	rs, ok := ecs.GetComponent[settlement.ResourceStore](w, settle)
	if !ok {
		t.Fatal("expected ResourceStore")
	}
	if rs.Resources["gold"] != 1.0 {
		t.Errorf("gold = %f, want 1.0", rs.Resources["gold"])
	}
	// 10 - 0.5 building consumption - 0.01 NPC consumption = 9.49
	if rs.Resources["food"] != 9.49 {
		t.Errorf("food = %f, want 9.49", rs.Resources["food"])
	}
}

func TestEconomySystemNPCFoodConsumption(t *testing.T) {
	w := setupWorld(t)

	settle := w.NewEntity()
	ecs.AddComponent(w, settle, ecs.Position{X: 10, Y: 10})
	ecs.AddComponent(w, settle, settlement.Settlement{
		Name: "Testville", Type: "village", Radius: 3,
		Buildings: []string{},
	})
	ecs.AddComponent(w, settle, settlement.ResourceStore{Resources: map[string]float64{"food": 100.0}})

	for i := 0; i < 5; i++ {
		n := w.NewEntity()
		ecs.AddComponent(w, n, settlement.HomeReference{SettlementEntity: settle})
	}

	sys := NewSettlementEconomySystem(nil)
	if err := sys.Update(w, 1.0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	rs, ok := ecs.GetComponent[settlement.ResourceStore](w, settle)
	if !ok {
		t.Fatal("expected ResourceStore")
	}
	// 5 NPCs * 0.01 = 0.05
	if rs.Resources["food"] != 99.95 {
		t.Errorf("food = %f, want 99.95", rs.Resources["food"])
	}
}

func TestEconomySystemNoNPCsNoConsumption(t *testing.T) {
	w := setupWorld(t)

	settle := w.NewEntity()
	ecs.AddComponent(w, settle, ecs.Position{X: 10, Y: 10})
	ecs.AddComponent(w, settle, settlement.Settlement{
		Name: "Testville", Type: "village", Radius: 3,
		Buildings: []string{},
	})
	ecs.AddComponent(w, settle, settlement.ResourceStore{Resources: map[string]float64{"food": 50.0}})

	sys := NewSettlementEconomySystem(nil)
	if err := sys.Update(w, 1.0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	rs, ok := ecs.GetComponent[settlement.ResourceStore](w, settle)
	if !ok {
		t.Fatal("expected ResourceStore")
	}
	if rs.Resources["food"] != 50.0 {
		t.Errorf("food = %f, want 50.0", rs.Resources["food"])
	}
}

func TestEconomySystemMaxWorkersCap(t *testing.T) {
	w := setupWorld(t)

	settle := w.NewEntity()
	ecs.AddComponent(w, settle, ecs.Position{X: 10, Y: 10})
	ecs.AddComponent(w, settle, settlement.Settlement{
		Name: "Testville", Type: "village", Radius: 3,
		Buildings: []string{"farm"},
	})

	// Create building entity with BuildingInterior (5 workers inside, but max 2)
	// This tests that production is capped at MaxWorkers
	build := w.NewEntity()
	ecs.AddComponent(w, build, ecs.Position{X: 12, Y: 12})
	ecs.AddComponent(w, build, settlement.Building{
		ID:               "farm",
		Name:             "Granja",
		SettlementEntity: settle,
	})
	ecs.AddComponent(w, build, settlement.BuildingInterior{
		WorkersInside: 5, // More than max
		MaxWorkers:    2, // Should be capped
		Grid:          createTestGrid(5, 5),
		Width:         5,
		Height:        5,
	})

	// 5 farmers (for fallback and NPC consumption)
	for i := 0; i < 5; i++ {
		n := w.NewEntity()
		ecs.AddComponent(w, n, npc.Job{Role: "farmer"})
		ecs.AddComponent(w, n, settlement.HomeReference{SettlementEntity: settle})
	}

	defs := []settlement.BuildingDef{
		{ID: "farm", Name: "Granja", Role: "farmer", Produces: map[string]float64{"food": 2.0}, MaxWorkers: 2},
	}
	sys := NewSettlementEconomySystem(defs)
	if err := sys.Update(w, 1.0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	rs, ok := ecs.GetComponent[settlement.ResourceStore](w, settle)
	if !ok {
		t.Fatal("expected ResourceStore")
	}
	// 2 workers (capped) * 2.0 = 4.0 production, 5 NPCs * 0.01 = 0.05 consumption
	if rs.Resources["food"] != 3.95 {
		t.Errorf("food = %f, want 3.95", rs.Resources["food"])
	}
}

func TestEconomySystemNoWorkersNoProduction(t *testing.T) {
	w := setupWorld(t)

	settle := w.NewEntity()
	ecs.AddComponent(w, settle, ecs.Position{X: 10, Y: 10})
	ecs.AddComponent(w, settle, settlement.Settlement{
		Name: "Testville", Type: "village", Radius: 3,
		Buildings: []string{"farm"},
	})

	defs := []settlement.BuildingDef{
		{ID: "farm", Name: "Granja", Role: "farmer", Produces: map[string]float64{"food": 2.0}, MaxWorkers: 3},
	}
	sys := NewSettlementEconomySystem(defs)
	if err := sys.Update(w, 1.0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	rs, ok := ecs.GetComponent[settlement.ResourceStore](w, settle)
	if !ok {
		t.Fatal("expected ResourceStore")
	}
	if rs.Resources["food"] != 0.0 {
		t.Errorf("food = %f, want 0.0", rs.Resources["food"])
	}
}

func TestEconomySystemHouseIgnored(t *testing.T) {
	w := setupWorld(t)

	settle := w.NewEntity()
	ecs.AddComponent(w, settle, ecs.Position{X: 10, Y: 10})
	ecs.AddComponent(w, settle, settlement.Settlement{
		Name: "Testville", Type: "village", Radius: 3,
		Buildings: []string{"house"},
	})

	defs := []settlement.BuildingDef{
		{ID: "house", Name: "Casa"},
	}
	sys := NewSettlementEconomySystem(defs)
	if err := sys.Update(w, 1.0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	rs, ok := ecs.GetComponent[settlement.ResourceStore](w, settle)
	if !ok {
		t.Fatal("expected ResourceStore")
	}
	if len(rs.Resources) != 3 {
		t.Errorf("expected only default resources, got %v", rs.Resources)
	}
}

func TestGrowthSystemLevelUp(t *testing.T) {
	w := setupWorld(t)

	settle := w.NewEntity()
	ecs.AddComponent(w, settle, settlement.Settlement{
		Name: "Testville", Type: "village", Radius: 3, Level: 1,
	})
	ecs.AddComponent(w, settle, settlement.ResourceStore{Resources: map[string]float64{
		"food": 100, "tools": 10, "gold": 5,
	}})

	thresholds := []settlement.GrowthThreshold{
		{Level: 2, Food: 100, Tools: 10, Gold: 5, NewRadius: 4},
	}
	sys := NewSettlementGrowthSystem(thresholds)
	if err := sys.Update(w, 1.0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	set, ok := ecs.GetComponent[settlement.Settlement](w, settle)
	if !ok {
		t.Fatal("expected settlement")
	}
	if set.Level != 2 {
		t.Errorf("level = %d, want 2", set.Level)
	}
	if set.Radius != 4 {
		t.Errorf("radius = %d, want 4", set.Radius)
	}

	rs, ok := ecs.GetComponent[settlement.ResourceStore](w, settle)
	if !ok {
		t.Fatal("expected ResourceStore")
	}
	if rs.Resources["food"] != 0 || rs.Resources["tools"] != 0 || rs.Resources["gold"] != 0 {
		t.Errorf("expected resources deducted, got %v", rs.Resources)
	}
}

func TestGrowthSystemPartialResources(t *testing.T) {
	w := setupWorld(t)

	settle := w.NewEntity()
	ecs.AddComponent(w, settle, settlement.Settlement{
		Name: "Testville", Type: "village", Radius: 3, Level: 1,
	})
	ecs.AddComponent(w, settle, settlement.ResourceStore{Resources: map[string]float64{
		"food": 100, "tools": 10, "gold": 2,
	}})

	thresholds := []settlement.GrowthThreshold{
		{Level: 2, Food: 100, Tools: 10, Gold: 5, NewRadius: 4},
	}
	sys := NewSettlementGrowthSystem(thresholds)
	if err := sys.Update(w, 1.0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	set, ok := ecs.GetComponent[settlement.Settlement](w, settle)
	if !ok {
		t.Fatal("expected settlement")
	}
	if set.Level != 1 {
		t.Errorf("level = %d, want 1", set.Level)
	}
}

func TestGrowthSystemMaxLevel(t *testing.T) {
	w := setupWorld(t)

	settle := w.NewEntity()
	ecs.AddComponent(w, settle, settlement.Settlement{
		Name: "Testville", Type: "village", Radius: 3, Level: 3,
	})
	ecs.AddComponent(w, settle, settlement.ResourceStore{Resources: map[string]float64{
		"food": 999, "tools": 999, "gold": 999,
	}})

	thresholds := []settlement.GrowthThreshold{
		{Level: 2, Food: 100, Tools: 10, Gold: 5, NewRadius: 4},
	}
	sys := NewSettlementGrowthSystem(thresholds)
	if err := sys.Update(w, 1.0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	set, ok := ecs.GetComponent[settlement.Settlement](w, settle)
	if !ok {
		t.Fatal("expected settlement")
	}
	if set.Level != 3 {
		t.Errorf("level = %d, want 3", set.Level)
	}
}

func TestGrowthSystemMissingThreshold(t *testing.T) {
	w := setupWorld(t)

	settle := w.NewEntity()
	ecs.AddComponent(w, settle, settlement.Settlement{
		Name: "Testville", Type: "village", Radius: 3, Level: 1,
	})
	ecs.AddComponent(w, settle, settlement.ResourceStore{Resources: map[string]float64{
		"food": 100, "tools": 10, "gold": 5,
	}})

	// No threshold for level 2
	thresholds := []settlement.GrowthThreshold{}
	sys := NewSettlementGrowthSystem(thresholds)
	if err := sys.Update(w, 1.0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	set, ok := ecs.GetComponent[settlement.Settlement](w, settle)
	if !ok {
		t.Fatal("expected settlement")
	}
	if set.Level != 1 {
		t.Errorf("level = %d, want 1", set.Level)
	}
}

func TestGrowthSystemLevel2To3(t *testing.T) {
	w := setupWorld(t)

	settle := w.NewEntity()
	ecs.AddComponent(w, settle, settlement.Settlement{
		Name: "Testville", Type: "village", Radius: 4, Level: 2,
	})
	ecs.AddComponent(w, settle, settlement.ResourceStore{Resources: map[string]float64{
		"food": 500, "tools": 50, "gold": 25,
	}})

	thresholds := []settlement.GrowthThreshold{
		{Level: 2, Food: 100, Tools: 10, Gold: 5, NewRadius: 4},
		{Level: 3, Food: 500, Tools: 50, Gold: 25, NewRadius: 6},
	}
	sys := NewSettlementGrowthSystem(thresholds)
	if err := sys.Update(w, 1.0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	set, ok := ecs.GetComponent[settlement.Settlement](w, settle)
	if !ok {
		t.Fatal("expected settlement")
	}
	if set.Level != 3 {
		t.Errorf("level = %d, want 3", set.Level)
	}
	if set.Radius != 6 {
		t.Errorf("radius = %d, want 6", set.Radius)
	}
}

func TestFamineSystemRemovesOneNPC(t *testing.T) {
	w := setupWorld(t)

	settle := w.NewEntity()
	ecs.AddComponent(w, settle, settlement.Settlement{
		Name: "Testville", Type: "village", Radius: 3,
	})
	ecs.AddComponent(w, settle, settlement.ResourceStore{Resources: map[string]float64{
		"food": -2.0,
	}})

	npcs := make([]ecs.Entity, 3)
	for i := 0; i < 3; i++ {
		n := w.NewEntity()
		ecs.AddComponent(w, n, settlement.HomeReference{SettlementEntity: settle})
		npcs[i] = n
	}

	sys := NewFamineSystem()
	if err := sys.Update(w, 1.0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	remaining := 0
	for _, n := range npcs {
		if _, ok := ecs.GetComponent[settlement.HomeReference](w, n); ok {
			remaining++
		}
	}
	if remaining != 2 {
		t.Errorf("remaining NPCs with home = %d, want 2", remaining)
	}
}

func TestFamineSystemMultipleTicks(t *testing.T) {
	w := setupWorld(t)

	settle := w.NewEntity()
	ecs.AddComponent(w, settle, settlement.Settlement{
		Name: "Testville", Type: "village", Radius: 3,
	})
	ecs.AddComponent(w, settle, settlement.ResourceStore{Resources: map[string]float64{
		"food": -10.0,
	}})

	for i := 0; i < 5; i++ {
		n := w.NewEntity()
		ecs.AddComponent(w, n, settlement.HomeReference{SettlementEntity: settle})
	}

	sys := NewFamineSystem()
	for i := 0; i < 3; i++ {
		if err := sys.Update(w, 1.0); err != nil {
			t.Fatalf("Update error: %v", err)
		}
	}

	// Count remaining HomeReferences for this settlement
	homeStore := w.GetStore(settlement.HomeRefID).(*ecs.ComponentStore[settlement.HomeReference])
	count := 0
	for _, h := range homeStore.All() {
		if h.SettlementEntity == settle {
			count++
		}
	}
	if count != 2 {
		t.Errorf("remaining NPCs = %d, want 2", count)
	}
}

func TestFamineSystemAllMigrate(t *testing.T) {
	w := setupWorld(t)

	settle := w.NewEntity()
	ecs.AddComponent(w, settle, settlement.Settlement{
		Name: "Testville", Type: "village", Radius: 3,
	})
	ecs.AddComponent(w, settle, settlement.ResourceStore{Resources: map[string]float64{
		"food": -5.0,
	}})

	for i := 0; i < 2; i++ {
		n := w.NewEntity()
		ecs.AddComponent(w, n, settlement.HomeReference{SettlementEntity: settle})
	}

	sys := NewFamineSystem()
	for i := 0; i < 3; i++ {
		if err := sys.Update(w, 1.0); err != nil {
			t.Fatalf("Update error: %v", err)
		}
	}

	homeStore := w.GetStore(settlement.HomeRefID).(*ecs.ComponentStore[settlement.HomeReference])
	count := 0
	for _, h := range homeStore.All() {
		if h.SettlementEntity == settle {
			count++
		}
	}
	if count != 0 {
		t.Errorf("remaining NPCs = %d, want 0", count)
	}
}

func TestFamineSystemPositiveFoodNoAction(t *testing.T) {
	w := setupWorld(t)

	settle := w.NewEntity()
	ecs.AddComponent(w, settle, settlement.Settlement{
		Name: "Testville", Type: "village", Radius: 3,
	})
	ecs.AddComponent(w, settle, settlement.ResourceStore{Resources: map[string]float64{
		"food": 50.0,
	}})

	npcs := make([]ecs.Entity, 5)
	for i := 0; i < 5; i++ {
		n := w.NewEntity()
		ecs.AddComponent(w, n, settlement.HomeReference{SettlementEntity: settle})
		npcs[i] = n
	}

	sys := NewFamineSystem()
	if err := sys.Update(w, 1.0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	for _, n := range npcs {
		if _, ok := ecs.GetComponent[settlement.HomeReference](w, n); !ok {
			t.Errorf("NPC %d lost HomeReference unexpectedly", n)
		}
	}
}

func TestEconomySystemLazyInitResourceStore(t *testing.T) {
	w := setupWorld(t)

	settle := w.NewEntity()
	ecs.AddComponent(w, settle, ecs.Position{X: 10, Y: 10})
	ecs.AddComponent(w, settle, settlement.Settlement{
		Name: "Testville", Type: "village", Radius: 3,
		Buildings: []string{},
	})

	sys := NewSettlementEconomySystem(nil)
	if err := sys.Update(w, 1.0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	rs, ok := ecs.GetComponent[settlement.ResourceStore](w, settle)
	if !ok {
		t.Fatal("expected ResourceStore to be lazy-initialized")
	}
	if rs.Resources["food"] != 0.0 || rs.Resources["gold"] != 0.0 || rs.Resources["tools"] != 0.0 {
		t.Errorf("expected zeroed default resources, got %v", rs.Resources)
	}
}

// TestProductionFormulaWorkersInside tests the production formula using WorkersInside from BuildingInterior.
// Production = base_output * workers_inside * dt
func TestProductionFormulaWorkersInside(t *testing.T) {
	tests := []struct {
		name           string
		workersInside  int
		maxWorkers     int
		baseOutput     float64
		dt             float64
		expectedOutput float64
	}{
		{
			name:           "0 workers -> 0 output",
			workersInside:  0,
			maxWorkers:     2,
			baseOutput:     2.0,
			dt:             1.0,
			expectedOutput: 0.0,
		},
		{
			name:           "1 worker -> base output",
			workersInside:  1,
			maxWorkers:     2,
			baseOutput:     2.0,
			dt:             1.0,
			expectedOutput: 2.0,
		},
		{
			name:           "2 workers -> 2x base output",
			workersInside:  2,
			maxWorkers:     2,
			baseOutput:     2.0,
			dt:             1.0,
			expectedOutput: 4.0,
		},
		{
			name:           "capped at max workers",
			workersInside:  5,
			maxWorkers:     2,
			baseOutput:     2.0,
			dt:             1.0,
			expectedOutput: 4.0, // capped at maxWorkers
		},
		{
			name:           "with fractional dt",
			workersInside:  1,
			maxWorkers:     2,
			baseOutput:     2.0,
			dt:             0.5,
			expectedOutput: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := setupWorld(t)

			// Create settlement
			settle := w.NewEntity()
			ecs.AddComponent(w, settle, ecs.Position{X: 10, Y: 10})
			ecs.AddComponent(w, settle, settlement.Settlement{
				Name: "Testville", Type: "village", Radius: 3,
			})

			// Create building entity with BuildingInterior
			build := w.NewEntity()
			ecs.AddComponent(w, build, ecs.Position{X: 12, Y: 12})
			ecs.AddComponent(w, build, settlement.Building{
				ID:               "farm",
				Name:             "Granja",
				SettlementEntity: settle,
			})
			ecs.AddComponent(w, build, settlement.BuildingInterior{
				WorkersInside: tt.workersInside,
				MaxWorkers:    tt.maxWorkers,
				Grid:          createTestGrid(5, 5),
				Width:         5,
				Height:        5,
			})

			defs := []settlement.BuildingDef{
				{ID: "farm", Name: "Granja", Role: "farmer", Produces: map[string]float64{"food": tt.baseOutput}, MaxWorkers: tt.maxWorkers},
			}
			sys := NewSettlementEconomySystem(defs)
			if err := sys.Update(w, tt.dt); err != nil {
				t.Fatalf("Update error: %v", err)
			}

			rs, ok := ecs.GetComponent[settlement.ResourceStore](w, settle)
			if !ok {
				t.Fatal("expected ResourceStore to be created")
			}

			actual := rs.Resources["food"]
			// Allow small floating point tolerance
			diff := actual - tt.expectedOutput
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.001 {
				t.Errorf("food = %f, want %f", actual, tt.expectedOutput)
			}
		})
	}
}

// createTestGrid creates a simple test grid with floor tiles.
func createTestGrid(width, height int) [][]settlement.CellType {
	grid := make([][]settlement.CellType, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]settlement.CellType, width)
		for x := 0; x < width; x++ {
			if x == 0 || x == width-1 || y == 0 || y == height-1 {
				grid[y][x] = settlement.CellWall
			} else {
				grid[y][x] = settlement.CellFloor
			}
		}
	}
	return grid
}

// TestProductionFormulaWorkersInsideAndRoleFallback tests that role-based workers
// are used as fallback when BuildingInterior is not present (backward compatibility).
func TestProductionFormulaWorkersInsideAndRoleFallback(t *testing.T) {
	w := setupWorld(t)

	// Create settlement
	settle := w.NewEntity()
	ecs.AddComponent(w, settle, ecs.Position{X: 10, Y: 10})
	ecs.AddComponent(w, settle, settlement.Settlement{
		Name: "Testville", Type: "village", Radius: 3,
	})

	// Create building WITHOUT BuildingInterior (backward compatibility scenario)
	build := w.NewEntity()
	ecs.AddComponent(w, build, ecs.Position{X: 12, Y: 12})
	ecs.AddComponent(w, build, settlement.Building{
		ID:               "farm",
		Name:             "Granja",
		SettlementEntity: settle,
	})

	// Create farmer NPC with job (role-based worker)
	farmer := w.NewEntity()
	ecs.AddComponent(w, farmer, npc.Job{Role: "farmer"})
	ecs.AddComponent(w, farmer, settlement.HomeReference{SettlementEntity: settle})

	defs := []settlement.BuildingDef{
		{ID: "farm", Name: "Granja", Role: "farmer", Produces: map[string]float64{"food": 2.0}, MaxWorkers: 3},
	}
	sys := NewSettlementEconomySystem(defs)
	if err := sys.Update(w, 1.0); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	rs, ok := ecs.GetComponent[settlement.ResourceStore](w, settle)
	if !ok {
		t.Fatal("expected ResourceStore to be created")
	}

	// 1 role-based worker * 2.0 food = 2.0 production
	// minus 0.01 food consumption for 1 NPC = 1.99 net
	if rs.Resources["food"] < 1.98 || rs.Resources["food"] > 2.0 {
		t.Errorf("food = %f, want ~1.99 (2.0 production - 0.01 consumption)", rs.Resources["food"])
	}
}
