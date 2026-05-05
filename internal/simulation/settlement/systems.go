package settlement

import (
	"math"
	"math/rand"

	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/world"
)

// SettlementSpawnSystem spawns settlements once on the first tick.
type SettlementSpawnSystem struct {
	spawned        bool
	wm             *world.WorldMap
	seed           int64
	settlementDefs []SettlementDef
	buildingDefs   []BuildingDef
}

// NewSettlementSpawnSystem creates a spawn system.
func NewSettlementSpawnSystem(wm *world.WorldMap, seed int64, settlementDefs []SettlementDef, buildingDefs []BuildingDef) *SettlementSpawnSystem {
	return &SettlementSpawnSystem{
		wm:             wm,
		seed:           seed,
		settlementDefs: settlementDefs,
		buildingDefs:   buildingDefs,
	}
}

// Name returns the system name.
func (s *SettlementSpawnSystem) Name() string { return "SettlementSpawnSystem" }

// Update runs the spawn logic on the first tick only.
func (s *SettlementSpawnSystem) Update(w *ecs.World, dt float64) error {
	if s.spawned {
		return nil
	}
	s.spawned = true

	if s.wm == nil || len(s.settlementDefs) == 0 {
		return nil
	}

	settlementRNG := rand.New(rand.NewSource(s.seed + 777))
	buildingRNG := rand.New(rand.NewSource(s.seed + 888))

	targetCount := 5 + settlementRNG.Intn(6) // 5-10

	// Track settlement positions for min distance check
	var settlementPositions []ecs.Position

	for i := 0; i < targetCount; i++ {
		setType := pickWeightedSettlement(s.settlementDefs, settlementRNG)
		if setType == nil {
			continue
		}

		found := false
		var sx, sy int
		for attempt := 0; attempt < 200; attempt++ {
			x := settlementRNG.Intn(s.wm.Width)
			y := settlementRNG.Intn(s.wm.Height)
			tile := s.wm.TileAt(x, y)
			if tile == nil {
				continue
			}
			if !biomeCompatible(tile.BiomeID, setType.Biomes) {
				continue
			}
			if !minDistanceOK(x, y, settlementPositions, 20) {
				continue
			}
			sx, sy = x, y
			found = true
			break
		}
		if !found {
			continue
		}

		// Generate procedural name
		name := generateSettlementName(setType.Name, settlementRNG)

		// Create settlement entity
		e := w.NewEntity()
		ecs.AddComponent(w, e, ecs.Position{X: float64(sx), Y: float64(sy)})
		ecs.AddComponent(w, e, Settlement{
			Name:      name,
			Type:      setType.ID,
			Symbol:    setType.Symbol,
			Color:     setType.Color,
			Radius:    setType.Radius,
			Level:     1,
			Buildings: append([]string(nil), setType.Buildings...),
		})
		ecs.AddComponent(w, e, ecs.Name{Name: name})

		settlementPositions = append(settlementPositions, ecs.Position{X: float64(sx), Y: float64(sy)})

		// Spawn buildings
		for _, bID := range setType.Buildings {
			bx := sx - setType.Radius + buildingRNG.Intn(setType.Radius*2+1)
			by := sy - setType.Radius + buildingRNG.Intn(setType.Radius*2+1)
			if !s.wm.InBounds(bx, by) {
				continue
			}
			be := w.NewEntity()
			ecs.AddComponent(w, be, ecs.Position{X: float64(bx), Y: float64(by)})
			ecs.AddComponent(w, be, Building{ID: bID, Name: buildingName(bID), Level: 1})
		}
	}

	return nil
}

func pickWeightedSettlement(defs []SettlementDef, rng *rand.Rand) *SettlementDef {
	var total float64
	for _, d := range defs {
		total += d.SpawnWeight
	}
	if total <= 0 {
		return nil
	}
	target := rng.Float64() * total
	var cumulative float64
	for i := range defs {
		cumulative += defs[i].SpawnWeight
		if target < cumulative {
			return &defs[i]
		}
	}
	return &defs[len(defs)-1]
}

func biomeCompatible(biome string, allowed []string) bool {
	for _, b := range allowed {
		if b == biome {
			return true
		}
	}
	return false
}

func minDistanceOK(x, y int, existing []ecs.Position, minDist int) bool {
	for _, p := range existing {
		dx := math.Abs(float64(x - int(p.X)))
		dy := math.Abs(float64(y - int(p.Y)))
		dist := dx
		if dy > dist {
			dist = dy
		}
		if dist < float64(minDist) {
			return false
		}
	}
	return true
}

func generateSettlementName(typeName string, rng *rand.Rand) string {
	prefix := settlementNamePrefixes[rng.Intn(len(settlementNamePrefixes))]
	suffix := settlementNameSuffixes[rng.Intn(len(settlementNameSuffixes))]
	name := prefix + " " + typeName + " " + suffix
	return truncateName(name, 15)
}

func truncateName(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}
	return name[:maxLen]
}

func buildingName(id string) string {
	switch id {
	case "house":
		return "Casa"
	case "farm":
		return "Granja"
	case "market":
		return "Mercado"
	case "tavern":
		return "Taberna"
	case "temple":
		return "Templo"
	case "blacksmith":
		return "Herreria"
	default:
		return id
	}
}

// PopulationSystem counts NPCs per settlement and updates Population.
type PopulationSystem struct {
	settlementID ecs.ComponentID
}

// NewPopulationSystem creates a population counting system.
func NewPopulationSystem() *PopulationSystem {
	return &PopulationSystem{
		settlementID: SettlementID,
	}
}

// Name returns the system name.
func (s *PopulationSystem) Name() string { return "PopulationSystem" }

// Update counts NPCs with HomeReference matching each settlement.
func (s *PopulationSystem) Update(w *ecs.World, dt float64) error {
	homeRefID := ecs.NewComponentID("home_reference")
	homeStore, ok := w.GetStore(homeRefID).(*ecs.ComponentStore[HomeReference])
	if !ok {
		return nil
	}
	setStore, ok := w.GetStore(SettlementID).(*ecs.ComponentStore[Settlement])
	if !ok {
		return nil
	}

	// Count NPCs per settlement
	counts := make(map[ecs.Entity]int)
	for _, h := range homeStore.All() {
		counts[h.SettlementEntity]++
	}

	// Update Population on each settlement
	for e, set := range setStore.All() {
		set.Population = counts[e]
		setStore.Set(e, set)
	}
	return nil
}

// SettlementRenderSystem collects render information for settlements.
type SettlementRenderSystem struct {
	renderInfos []SettlementRenderInfo
}

// NewSettlementRenderSystem creates a render system.
func NewSettlementRenderSystem() *SettlementRenderSystem {
	return &SettlementRenderSystem{}
}

// Name returns the system name.
func (s *SettlementRenderSystem) Name() string { return "SettlementRenderSystem" }

// Update gathers renderable settlements.
func (s *SettlementRenderSystem) Update(w *ecs.World, dt float64) error {
	s.renderInfos = nil
	posStore := w.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])
	setStore, _ := w.GetStore(SettlementID).(*ecs.ComponentStore[Settlement])
	if setStore == nil {
		return nil
	}

	for e, set := range setStore.All() {
		pos, ok := posStore.Get(e)
		if !ok {
			continue
		}
		s.renderInfos = append(s.renderInfos, SettlementRenderInfo{
			Entity:     int(e),
			Symbol:     []rune(set.Symbol)[0],
			Color:      set.Color,
			Name:       set.Name,
			WorldX:     int(pos.X),
			WorldY:     int(pos.Y),
			Population: set.Population,
		})
	}
	return nil
}

// RenderInfos returns the render information gathered in the last Update.
func (s *SettlementRenderSystem) RenderInfos() []SettlementRenderInfo {
	return s.renderInfos
}
