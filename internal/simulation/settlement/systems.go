package settlement

import (
	"math"
	"math/rand"

	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/world"
)

// DefaultBuildingSeed is the base seed for building interior generation.
// Each building gets a unique seed based on its world position.
var DefaultBuildingSeed int64 = 12345

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

		// Build a lookup for building definitions
		buildDefMap := make(map[string]BuildingDef)
		for _, bd := range s.buildingDefs {
			buildDefMap[bd.ID] = bd
		}

		// Spawn buildings
		for _, bID := range setType.Buildings {
			bx := sx - setType.Radius + buildingRNG.Intn(setType.Radius*2+1)
			by := sy - setType.Radius + buildingRNG.Intn(setType.Radius*2+1)
			if !s.wm.InBounds(bx, by) {
				continue
			}
			be := w.NewEntity()
			ecs.AddComponent(w, be, ecs.Position{X: float64(bx), Y: float64(by)})
			b := Building{ID: bID, Name: buildingName(bID), Level: 1, SettlementEntity: e}
			maxWorkers := 2 // default
			if bd, ok := buildDefMap[bID]; ok {
				b.InteriorSymbol = bd.InteriorSymbol
				b.Color = bd.Color
				maxWorkers = bd.MaxWorkers
				if maxWorkers <= 0 {
					maxWorkers = 2
				}
			}
			ecs.AddComponent(w, be, b)

			// Generate interior using InteriorGenerator
			interiorSeed := DefaultBuildingSeed + int64(bx)*1000 + int64(by)
			interior := DefaultInteriorGenerator.Generate(interiorSeed, bID, 5, 5)
			interior.WorkersInside = 0
			interior.MaxWorkers = maxWorkers
			// Set door world positions based on building position
			for i := range interior.Doors {
				interior.Doors[i].WorldX = bx + interior.Doors[i].GridX - interior.Width/2
				interior.Doors[i].WorldY = by + interior.Doors[i].GridY - interior.Height/2
			}
			ecs.AddComponent(w, be, interior)
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
	resStore, _ := w.GetStore(ResourceID).(*ecs.ComponentStore[ResourceStore])
	if setStore == nil {
		return nil
	}

	for e, set := range setStore.All() {
		pos, ok := posStore.Get(e)
		if !ok {
			continue
		}
		info := SettlementRenderInfo{
			Entity:     int(e),
			Symbol:     []rune(set.Symbol)[0],
			Color:      set.Color,
			Name:       set.Name,
			WorldX:     int(pos.X),
			WorldY:     int(pos.Y),
			Population: set.Population,
			Level:      set.Level,
		}
		if resStore != nil {
			if rs, ok := resStore.Get(e); ok {
				info.HasResources = true
				info.Food = rs.Resources["food"]
				info.Gold = rs.Resources["gold"]
				info.Tools = rs.Resources["tools"]
			}
		}
		s.renderInfos = append(s.renderInfos, info)
	}
	return nil
}

// RenderInfos returns the render information gathered in the last Update.
func (s *SettlementRenderSystem) RenderInfos() []SettlementRenderInfo {
	return s.renderInfos
}

// BuildingRenderSystem collects render information for buildings.
type BuildingRenderSystem struct {
	renderInfos  []BuildingRenderInfo
	buildingDefs map[string]BuildingDef
}

// NewBuildingRenderSystem creates a render system.
func NewBuildingRenderSystem(defs []BuildingDef) *BuildingRenderSystem {
	m := make(map[string]BuildingDef)
	for _, d := range defs {
		m[d.ID] = d
	}
	return &BuildingRenderSystem{buildingDefs: m}
}

// Name returns the system name.
func (s *BuildingRenderSystem) Name() string { return "BuildingRenderSystem" }

// Update gathers renderable buildings.
func (s *BuildingRenderSystem) Update(w *ecs.World, dt float64) error {
	s.renderInfos = nil
	posStore := w.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])
	buildStore, _ := w.GetStore(BuildingID).(*ecs.ComponentStore[Building])
	interiorStore, _ := w.GetStore(BuildingInteriorID).(*ecs.ComponentStore[BuildingInterior])
	if buildStore == nil {
		return nil
	}

	for e, b := range buildStore.All() {
		if b.InteriorSymbol == "" {
			continue
		}
		pos, ok := posStore.Get(e)
		if !ok {
			continue
		}
		info := BuildingRenderInfo{
			Entity:           e,
			Symbol:           []rune(b.InteriorSymbol)[0],
			Color:            b.Color,
			Name:             b.Name,
			ID:               b.ID,
			Level:            b.Level,
			WorldX:           int(pos.X),
			WorldY:           int(pos.Y),
			SettlementEntity: b.SettlementEntity,
			WorkersInside:    0, // default
		}
		// Get WorkersInside from BuildingInterior if available
		if interiorStore != nil {
			if interior, ok := interiorStore.Get(e); ok {
				info.WorkersInside = interior.WorkersInside
			}
		}
		if def, ok := s.buildingDefs[b.ID]; ok {
			info.Role = def.Role
			info.MaxWorkers = def.MaxWorkers
			if len(def.Produces) > 0 {
				info.Produces = make(map[string]float64, len(def.Produces))
				for k, v := range def.Produces {
					info.Produces[k] = v
				}
			}
			if len(def.Consumes) > 0 {
				info.Consumes = make(map[string]float64, len(def.Consumes))
				for k, v := range def.Consumes {
					info.Consumes[k] = v
				}
			}
		}
		s.renderInfos = append(s.renderInfos, info)
	}
	return nil
}

// RenderInfos returns the render information gathered in the last Update.
func (s *BuildingRenderSystem) RenderInfos() []BuildingRenderInfo {
	return s.renderInfos
}

// RenderInfosForSettlement returns buildings whose SettlementEntity matches the given entity.
func (s *BuildingRenderSystem) RenderInfosForSettlement(w *ecs.World, entity ecs.Entity) []BuildingRenderInfo {
	var result []BuildingRenderInfo
	for _, info := range s.renderInfos {
		if info.SettlementEntity == entity {
			result = append(result, info)
		}
	}
	return result
}
