package npc

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"

	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/world"
)

// NPCSpawnSystem spawns NPCs once on the first tick.
type NPCSpawnSystem struct {
	spawned   bool
	wm        *world.WorldMap
	config    SpawnConfig
	seed      int64
	raceDefs  []RaceDef
	roleDefs  []RoleDef
}

// NewNPCSpawnSystem creates a spawn system.
func NewNPCSpawnSystem(wm *world.WorldMap, config SpawnConfig, seed int64, raceDefs []RaceDef, roleDefs []RoleDef) *NPCSpawnSystem {
	return &NPCSpawnSystem{
		wm:       wm,
		config:   config,
		seed:     seed,
		raceDefs: raceDefs,
		roleDefs: roleDefs,
	}
}

// Name returns the system name.
func (s *NPCSpawnSystem) Name() string { return "NPCSpawnSystem" }

// Update runs the spawn logic on the first tick only.
func (s *NPCSpawnSystem) Update(w *ecs.World, dt float64) error {
	if s.spawned {
		return nil
	}
	s.spawned = true
	return Spawn(w, s.wm, s.config, s.seed, s.raceDefs, s.roleDefs)
}

// WanderSystem moves NPCs with LOD≥1 to random adjacent compatible tiles.
type WanderSystem struct {
	wm      *world.WorldMap
	roleMap map[string]RoleDef
	rng     *rand.Rand
}

// NewWanderSystem creates a wander system.
func NewWanderSystem(wm *world.WorldMap, roleDefs []RoleDef, rng *rand.Rand) *WanderSystem {
	m := make(map[string]RoleDef)
	for _, rd := range roleDefs {
		m[rd.ID] = rd
	}
	return &WanderSystem{wm: wm, roleMap: m, rng: rng}
}

// Name returns the system name.
func (s *WanderSystem) Name() string { return "WanderSystem" }

// Update processes wandering for eligible NPCs.
func (s *WanderSystem) Update(w *ecs.World, dt float64) error {
	posStore := w.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])
	lodStore, _ := w.GetStore(LODID).(*ecs.ComponentStore[LOD])
	jobStore, _ := w.GetStore(JobID).(*ecs.ComponentStore[Job])
	aiStore, _ := w.GetStore(AIStateID).(*ecs.ComponentStore[AIState])

	if lodStore == nil || jobStore == nil || aiStore == nil {
		return nil
	}

	for e, lod := range lodStore.All() {
		if lod.Level < LODNear {
			continue
		}
		job, ok := jobStore.Get(e)
		if !ok {
			continue
		}
		ai, ok := aiStore.Get(e)
		if !ok {
			continue
		}
		pos, ok := posStore.Get(e)
		if !ok {
			continue
		}

		roleDef, ok := s.roleMap[job.Role]
		if !ok {
			continue
		}

		x, y := int(pos.X), int(pos.Y)

		if len(ai.Plan) == 0 {
			// Generate a new wander goal
			dx := s.rng.Intn(3) - 1 // -1, 0, 1
			dy := s.rng.Intn(3) - 1
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := x+dx, y+dy
			if !s.wm.InBounds(nx, ny) {
				continue
			}
			tile := s.wm.TileAt(nx, ny)
			if tile == nil {
				continue
			}
			if !biomeCompatible(tile.BiomeID, roleDef.Biomes) {
				continue
			}
			ai.Plan = []string{fmt.Sprintf("%d,%d", nx, ny)}
			aiStore.Set(e, ai)
		} else {
			// Execute plan
			parts := strings.Split(ai.Plan[0], ",")
			nx, _ := strconv.Atoi(parts[0])
			ny, _ := strconv.Atoi(parts[1])
			pos.X = float64(nx)
			pos.Y = float64(ny)
			posStore.Set(e, pos)
			ai.Plan = nil
			aiStore.Set(e, ai)
		}
	}
	return nil
}

// LODSystem updates the LOD component based on Chebyshev distance to the player.
type LODSystem struct {
	playerPos func() (int, int)
}

// NewLODSystem creates an LOD system.
func NewLODSystem(playerPos func() (int, int)) *LODSystem {
	return &LODSystem{playerPos: playerPos}
}

// Name returns the system name.
func (s *LODSystem) Name() string { return "LODSystem" }

// Update recalculates LOD levels.
func (s *LODSystem) Update(w *ecs.World, dt float64) error {
	px, py := s.playerPos()
	posStore := w.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])
	lodStore, _ := w.GetStore(LODID).(*ecs.ComponentStore[LOD])
	if lodStore == nil {
		return nil
	}

	for e, pos := range posStore.All() {
		if !lodStore.Has(e) {
			continue
		}
		dist := chebyshev(int(pos.X), int(pos.Y), px, py)
		var level int
		switch {
		case dist <= 5:
			level = LODLocal
		case dist <= 15:
			level = LODNear
		default:
			level = LODDistant
		}
		lodStore.Set(e, LOD{Level: level})
	}
	return nil
}

func chebyshev(x1, y1, x2, y2 int) int {
	dx := math.Abs(float64(x1 - x2))
	dy := math.Abs(float64(y1 - y2))
	if dx > dy {
		return int(dx)
	}
	return int(dy)
}

// NPCRenderSystem collects render information for NPCs with LOD≥1.
type NPCRenderSystem struct {
	renderInfos []NPCRenderInfo
}

// NewNPCRenderSystem creates a render system.
func NewNPCRenderSystem() *NPCRenderSystem {
	return &NPCRenderSystem{}
}

// Name returns the system name.
func (s *NPCRenderSystem) Name() string { return "NPCRenderSystem" }

// Update gathers renderable NPCs.
func (s *NPCRenderSystem) Update(w *ecs.World, dt float64) error {
	s.renderInfos = nil
	posStore := w.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])
	lodStore, _ := w.GetStore(LODID).(*ecs.ComponentStore[LOD])
	appStore, _ := w.GetStore(AppearanceID).(*ecs.ComponentStore[Appearance])
	if lodStore == nil || appStore == nil {
		return nil
	}

	for e, lod := range lodStore.All() {
		if lod.Level < LODNear {
			continue
		}
		pos, ok := posStore.Get(e)
		if !ok {
			continue
		}
		app, ok := appStore.Get(e)
		if !ok {
			continue
		}
		s.renderInfos = append(s.renderInfos, NPCRenderInfo{
			Entity: e,
			Symbol: app.Symbol,
			Color:  app.Color,
			WorldX: int(pos.X),
			WorldY: int(pos.Y),
		})
	}
	return nil
}

// RenderInfos returns the render information gathered in the last Update.
func (s *NPCRenderSystem) RenderInfos() []NPCRenderInfo {
	return s.renderInfos
}
