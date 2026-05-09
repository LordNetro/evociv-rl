package npc

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"

	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/simulation/goap"
	"github.com/marco/evociv-rl/internal/simulation/rl"
	"github.com/marco/evociv-rl/internal/simulation/settlement"
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

	// Collect settlement entities for role-to-building matching
	var settlementEntities []ecs.Entity
	setStore, ok := w.GetStore(settlement.SettlementID).(*ecs.ComponentStore[settlement.Settlement])
	if ok && setStore != nil {
		for e := range setStore.All() {
			settlementEntities = append(settlementEntities, e)
		}
	}

	return Spawn(w, s.wm, s.config, s.seed, s.raceDefs, s.roleDefs, settlementEntities)
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
	playerPos       func() (int, int)
	settlementBoost map[ecs.Entity]bool // entities to keep at LODLocal regardless of distance
}

// NewLODSystem creates an LOD system.
func NewLODSystem(playerPos func() (int, int)) *LODSystem {
	return &LODSystem{playerPos: playerPos, settlementBoost: make(map[ecs.Entity]bool)}
}

// SetSettlementBoost marks an entity to keep LODLocal regardless of distance from player.
func (s *LODSystem) SetSettlementBoost(e ecs.Entity) {
	if s.settlementBoost == nil {
		s.settlementBoost = make(map[ecs.Entity]bool)
	}
	s.settlementBoost[e] = true
}

// ClearSettlementBoost removes all settlement LOD overrides.
func (s *LODSystem) ClearSettlementBoost() {
	s.settlementBoost = make(map[ecs.Entity]bool)
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
		// Never downgrade settlement-boosted entities
		if s.settlementBoost[e] {
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
	// NeedsDecaySystem increases Hunger and Fatigue for all NPCs each tick.
	type NeedsDecaySystem struct{}

	// NewNeedsDecaySystem creates a new needs decay system.
	func NewNeedsDecaySystem() *NeedsDecaySystem {
		return &NeedsDecaySystem{}
	}

	// Name returns the system name.
	func (s *NeedsDecaySystem) Name() string { return "NeedsDecaySystem" }

	// Update decays needs for all NPCs.
	func (s *NeedsDecaySystem) Update(w *ecs.World, dt float64) error {
		needsStore, _ := w.GetStore(NeedsID).(*ecs.ComponentStore[Needs])
		lodStore, _ := w.GetStore(LODID).(*ecs.ComponentStore[LOD])
		if needsStore == nil || lodStore == nil {
			return nil
		}

		for e, needs := range needsStore.All() {
			lodMul := 1.0
			if lod, ok := lodStore.Get(e); ok && lod.Level == LODDistant {
				lodMul = 0.5
			}

			needs.Hunger += 0.01 * lodMul * dt
			needs.Fatigue += 0.005 * lodMul * dt

			if needs.Hunger > 1.0 {
				needs.Hunger = 1.0
			}
			if needs.Fatigue > 1.0 {
				needs.Fatigue = 1.0
			}

			needsStore.Set(e, needs)
		}
		return nil
	}

// GOAPSystem plans actions for NPCs with LOD≥1.
type GOAPSystem struct {
	wm      *world.WorldMap
	actions []ActionDef
}

// NewGOAPSystem creates a new GOAP planning system.
func NewGOAPSystem(wm *world.WorldMap, actions []ActionDef) *GOAPSystem {
	return &GOAPSystem{wm: wm, actions: actions}
}

// Name returns the system name.
func (s *GOAPSystem) Name() string { return "GOAPSystem" }

// Update plans actions for eligible NPCs.
func (s *GOAPSystem) Update(w *ecs.World, dt float64) error {
	posStore := w.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])
	needsStore, _ := w.GetStore(NeedsID).(*ecs.ComponentStore[Needs])
	lodStore, _ := w.GetStore(LODID).(*ecs.ComponentStore[LOD])
	aiStore, _ := w.GetStore(AIStateID).(*ecs.ComponentStore[AIState])

	if needsStore == nil || lodStore == nil || aiStore == nil {
		return nil
	}

	// Convert npc.ActionDef to goap.Action for planning
	goapActions := make([]goap.Action, len(s.actions))
	for i, a := range s.actions {
		goapActions[i] = goap.Action{
			ID:            a.ID,
			Biomes:        a.Requires.Biomes,
			NeedsMin:      goap.Needs{Hunger: a.Requires.NeedsMin.Hunger, Fatigue: a.Requires.NeedsMin.Fatigue},
			NeedsMax:      goap.Needs{Hunger: a.Requires.NeedsMax.Hunger, Fatigue: a.Requires.NeedsMax.Fatigue},
			HungerChange:  a.Effects.HungerChange,
			FatigueChange: a.Effects.FatigueChange,
			RewardBase:    a.Reward.Base,
		}
	}

	for e, lod := range lodStore.All() {
		if lod.Level < LODNear {
			continue
		}
		needs, ok := needsStore.Get(e)
		if !ok {
			continue
		}
		pos, ok := posStore.Get(e)
		if !ok {
			continue
		}
		ai, ok := aiStore.Get(e)
		if !ok {
			continue
		}

		biome := ""
		if s.wm != nil {
			tile := s.wm.TileAt(int(pos.X), int(pos.Y))
			if tile != nil {
				biome = tile.BiomeID
			}
		}

		gn := goap.Needs{Hunger: needs.Hunger, Fatigue: needs.Fatigue}
		action := goap.Plan(gn, goapActions, biome)

		ai.CurrentAction = action.ID
		if lod.Level == LODLocal {
			// Full plan: set goal
			ai.Goals = []string{action.ID}
		} else {
			// 1-step: clear goals
			ai.Goals = nil
		}
		aiStore.Set(e, ai)
	}
	return nil
}

// QLearningSystem executes actions and updates Q-values for NPCs with LOD≥1.
type QLearningSystem struct {
	wm      *world.WorldMap
	actions []ActionDef
	qtable  *rl.QTable
	rng     *rand.Rand
	tick    int
}

// NewQLearningSystem creates a new Q-learning system.
func NewQLearningSystem(wm *world.WorldMap, actions []ActionDef, rng *rand.Rand) *QLearningSystem {
	return &QLearningSystem{
		wm:      wm,
		actions: actions,
		qtable:  rl.NewQTable(),
		rng:     rng,
	}
}

// Name returns the system name.
func (s *QLearningSystem) Name() string { return "QLearningSystem" }

// QTable exposes the internal Q-table (used for persistence).
func (s *QLearningSystem) QTable() *rl.QTable {
	return s.qtable
}

// Update executes Q-learning for eligible NPCs.
func (s *QLearningSystem) Update(w *ecs.World, dt float64) error {
	s.tick++
	posStore := w.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])
	needsStore, _ := w.GetStore(NeedsID).(*ecs.ComponentStore[Needs])
	lodStore, _ := w.GetStore(LODID).(*ecs.ComponentStore[LOD])
	aiStore, _ := w.GetStore(AIStateID).(*ecs.ComponentStore[AIState])

	if needsStore == nil || lodStore == nil || aiStore == nil {
		return nil
	}

	for e, lod := range lodStore.All() {
		if lod.Level < LODNear {
			continue
		}
		needs, ok := needsStore.Get(e)
		if !ok {
			continue
		}
		pos, ok := posStore.Get(e)
		if !ok {
			continue
		}
		ai, ok := aiStore.Get(e)
		if !ok {
			continue
		}

		biome := ""
		if s.wm != nil {
			tile := s.wm.TileAt(int(pos.X), int(pos.Y))
			if tile != nil {
				biome = tile.BiomeID
			}
		}

		// Determine state key
		needType := "none"
		if needs.Hunger > needs.Fatigue && needs.Hunger > 0.3 {
			needType = "hunger"
		} else if needs.Fatigue > needs.Hunger && needs.Fatigue > 0.3 {
			needType = "fatigue"
		}
		state := fmt.Sprintf("%s|%s|day", needType, biome)

		// Filter available actions
		var available []ActionDef
		var actionIDs []string
		for _, a := range s.actions {
			if !biomeCompatible(biome, a.Requires.Biomes) && len(a.Requires.Biomes) > 0 {
				continue
			}
			if needs.Hunger < a.Requires.NeedsMin.Hunger || needs.Hunger > a.Requires.NeedsMax.Hunger {
				continue
			}
			if needs.Fatigue < a.Requires.NeedsMin.Fatigue || needs.Fatigue > a.Requires.NeedsMax.Fatigue {
				continue
			}
			available = append(available, a)
			actionIDs = append(actionIDs, a.ID)
		}

		if len(available) == 0 {
			continue
		}

		var selectedAction ActionDef
		// GOAP fallback when all Q-values are below threshold
		if s.qtable.ShouldFallback(state, actionIDs, 0.1) {
			// Use GOAP's recommended action
			for _, a := range available {
				if a.ID == ai.CurrentAction {
					selectedAction = a
					break
				}
			}
		}
		if selectedAction.ID == "" {
			// ε-greedy selection
			actionID := s.qtable.EGreedy(state, actionIDs, s.qtable.Epsilon(), s.rng)
			for _, a := range available {
				if a.ID == actionID {
					selectedAction = a
					break
				}
			}
		}
		if selectedAction.ID == "" {
			selectedAction = available[0]
		}

		// Execute action: apply effects
		oldNeeds := needs
		needs.Hunger += selectedAction.Effects.HungerChange
		needs.Fatigue += selectedAction.Effects.FatigueChange
		needs.Hunger = math.Max(0, math.Min(1, needs.Hunger))
		needs.Fatigue = math.Max(0, math.Min(1, needs.Fatigue))
		needsStore.Set(e, needs)

		// Execute building actions
		s.executeBuildingAction(w, e, &ai, selectedAction.ID, pos)

		// Compute reward
		reward := ComputeReward(oldNeeds, needs, selectedAction)

		// Store reward on AIState
		if reward > 0.01 || reward < -0.01 {
			ai.LastReward = reward
			ai.RewardTick = s.tick
		}

		// Determine next state
		nextNeedType := "none"
		if needs.Hunger > needs.Fatigue && needs.Hunger > 0.3 {
			nextNeedType = "hunger"
		} else if needs.Fatigue > needs.Hunger && needs.Fatigue > 0.3 {
			nextNeedType = "fatigue"
		}
		nextState := fmt.Sprintf("%s|%s|day", nextNeedType, biome)

		// Update Q-table
		s.qtable.Update(state, selectedAction.ID, reward, nextState, 0.1, 0.9)
		s.qtable.DecayEpsilon()

		ai.CurrentAction = selectedAction.ID
		aiStore.Set(e, ai)
	}
	return nil
}

// ComputeReward calculates the reward for an action based on needs reduction and base reward.
func ComputeReward(oldNeeds, newNeeds Needs, action ActionDef) float64 {
	hungerReduction := oldNeeds.Hunger - newNeeds.Hunger
	fatigueReduction := oldNeeds.Fatigue - newNeeds.Fatigue
	if hungerReduction < 0 {
		hungerReduction = 0
	}
	if fatigueReduction < 0 {
		fatigueReduction = 0
	}
	return hungerReduction + fatigueReduction + action.Reward.Base
}

// executeBuildingAction handles enter_building, work_inside, exit_building actions.
// It modifies BuildingInterior.WorkersInside when NPCs enter or exit buildings.
func (s *QLearningSystem) executeBuildingAction(w *ecs.World, npcEntity ecs.Entity, ai *AIState, actionID string, npcPos ecs.Position) {
	// Check if settlement stores exist before attempting to use them
	buildingStoreInterface := w.GetStore(settlement.BuildingID)
	interiorStoreInterface := w.GetStore(settlement.BuildingInteriorID)
	if buildingStoreInterface == nil || interiorStoreInterface == nil {
		return // Settlement systems not initialized, skip building actions
	}

	buildingStore := buildingStoreInterface.(*ecs.ComponentStore[settlement.Building])
	interiorStore := interiorStoreInterface.(*ecs.ComponentStore[settlement.BuildingInterior])

	switch actionID {
	case "enter_building":
		// Find a building with available capacity and a door
		var targetBuilding ecs.Entity
		var targetInterior settlement.BuildingInterior
		var targetDoor settlement.DoorPosition
		posStore := w.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])

		for entity := range buildingStore.All() {
			interior, ok := interiorStore.Get(entity)
			if !ok {
				continue
			}
			if interior.WorkersInside >= interior.MaxWorkers {
				continue
			}
			if ai.InsideBuilding != 0 {
				continue
			}
			bPos, ok := posStore.Get(entity)
			if !ok {
				continue
			}
			// choose first door for now; if none, use building position as door
			if len(interior.Doors) > 0 {
				dp := interior.Doors[0]
				targetDoor = dp
			} else {
				// use building world position as a fallback door
				targetDoor = settlement.DoorPosition{WorldX: int(bPos.X), WorldY: int(bPos.Y)}
			}
			targetBuilding = entity
			targetInterior = interior
			_ = bPos
			break
		}

		if targetBuilding == 0 {
			break
		}

		// If NPC not at the door, set movement plan to door and persist AIState
		// consider NPC at the door if within a small Chebyshev radius (allows tests and nearby NPCs)
		atDoor := chebyshev(int(npcPos.X), int(npcPos.Y), targetDoor.WorldX, targetDoor.WorldY) <= 2
		if !atDoor {
			ai.Plan = []string{fmt.Sprintf("%d,%d", targetDoor.WorldX, targetDoor.WorldY)}
			if aiStore := w.GetStore(AIStateID).(*ecs.ComponentStore[AIState]); aiStore != nil {
				aiStore.Set(npcEntity, *ai)
			}
			break
		}

		// NPC is at door: enter and increment worker count
		targetInterior.WorkersInside++
		interiorStore.Set(targetBuilding, targetInterior)
		ai.InsideBuilding = targetBuilding

	case "work_inside":
		// Worker stays inside, already tracked by WorkersInside
		// Could add productivity effects here
		if ai.InsideBuilding != 0 {
			// Worker is working, nothing to change in WorkersInside
			_ = ai // silence unused warning
		}

	case "exit_building":
		if ai.InsideBuilding != 0 {
			interior, ok := interiorStore.Get(ai.InsideBuilding)
			if ok {
				// Decrement WorkersInside
				if interior.WorkersInside > 0 {
					interior.WorkersInside--
				}
				interiorStore.Set(ai.InsideBuilding, interior)
			}
			// Mark NPC as outside
			ai.InsideBuilding = 0
		}
	}
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
	aiStore, _ := w.GetStore(AIStateID).(*ecs.ComponentStore[AIState])
	jobStore, _ := w.GetStore(JobID).(*ecs.ComponentStore[Job])
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
		info := NPCRenderInfo{
			Entity: e,
			Symbol: app.Symbol,
			Color:  app.Color,
			WorldX: int(pos.X),
			WorldY: int(pos.Y),
		}
		if aiStore != nil {
			if ai, ok := aiStore.Get(e); ok {
				info.LastReward = ai.LastReward
				info.RewardTick = ai.RewardTick
			}
		}
		if jobStore != nil {
			if job, ok := jobStore.Get(e); ok {
				info.JobRole = job.Role
			}
		}
		s.renderInfos = append(s.renderInfos, info)
	}
	return nil
}

// RenderInfos returns the render information gathered in the last Update.
func (s *NPCRenderSystem) RenderInfos() []NPCRenderInfo {
	return s.renderInfos
}

// RenderInfosForSettlement returns NPCs whose HomeReference matches the given settlement.
// Queries ECS directly, bypassing LOD filter so ALL settlement NPCs are visible in interior view.
func (s *NPCRenderSystem) RenderInfosForSettlement(w *ecs.World, settlementEntity ecs.Entity) []NPCRenderInfo {
	var result []NPCRenderInfo
	posStore := w.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])
	appStore, _ := w.GetStore(AppearanceID).(*ecs.ComponentStore[Appearance])
	homeRefID := ecs.NewComponentID("home_reference")
	homeStore, ok := w.GetStore(homeRefID).(*ecs.ComponentStore[settlement.HomeReference])
	aiStore, _ := w.GetStore(AIStateID).(*ecs.ComponentStore[AIState])
	jobStore, _ := w.GetStore(JobID).(*ecs.ComponentStore[Job])
	if !ok || appStore == nil || posStore == nil {
		return result
	}
	for e, hr := range homeStore.All() {
		if hr.SettlementEntity != settlementEntity {
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
		info := NPCRenderInfo{
			Entity: e,
			Symbol: app.Symbol,
			Color:  app.Color,
			WorldX: int(pos.X),
			WorldY: int(pos.Y),
		}
		if aiStore != nil {
			if ai, ok := aiStore.Get(e); ok {
				info.LastReward = ai.LastReward
				info.RewardTick = ai.RewardTick
			}
		}
		if jobStore != nil {
			if job, ok := jobStore.Get(e); ok {
				info.JobRole = job.Role
			}
		}
		result = append(result, info)
	}
	return result
}
