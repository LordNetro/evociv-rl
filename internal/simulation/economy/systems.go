package economy

import (
	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/simulation/npc"
	"github.com/marco/evociv-rl/internal/simulation/settlement"
)

// SettlementEconomySystem handles production and consumption for settlements.
type SettlementEconomySystem struct {
	buildingDefs map[string]settlement.BuildingDef
}

// NewSettlementEconomySystem creates an economy system with building definitions.
func NewSettlementEconomySystem(defs []settlement.BuildingDef) *SettlementEconomySystem {
	m := make(map[string]settlement.BuildingDef, len(defs))
	for _, d := range defs {
		m[d.ID] = d
	}
	return &SettlementEconomySystem{buildingDefs: m}
}

// Name returns the system name.
func (s *SettlementEconomySystem) Name() string { return "SettlementEconomySystem" }

// Update processes production and consumption.
func (s *SettlementEconomySystem) Update(w *ecs.World, dt float64) error {
	setStore, ok := w.GetStore(settlement.SettlementID).(*ecs.ComponentStore[settlement.Settlement])
	if !ok {
		return nil
	}
	homeStore, ok := w.GetStore(settlement.HomeRefID).(*ecs.ComponentStore[settlement.HomeReference])
	if !ok {
		return nil
	}
	jobStore, ok := w.GetStore(npc.JobID).(*ecs.ComponentStore[npc.Job])
	if !ok {
		return nil
	}
	resStore, ok := w.GetStore(settlement.ResourceID).(*ecs.ComponentStore[settlement.ResourceStore])
	if !ok {
		return nil
	}

	// Count NPCs per settlement
	npcCounts := make(map[ecs.Entity]int)
	for _, h := range homeStore.All() {
		npcCounts[h.SettlementEntity]++
	}

	// Count workers per role per settlement
	// workers[settlement][role] = count
	workers := make(map[ecs.Entity]map[string]int)
	for e, job := range jobStore.All() {
		home, ok := homeStore.Get(e)
		if !ok {
			continue
		}
		if workers[home.SettlementEntity] == nil {
			workers[home.SettlementEntity] = make(map[string]int)
		}
		workers[home.SettlementEntity][job.Role]++
	}

	for e, set := range setStore.All() {
		// Lazy-init ResourceStore
		var rs settlement.ResourceStore
		if existing, ok := resStore.Get(e); ok {
			rs = existing
		} else {
			rs = settlement.ResourceStore{Resources: map[string]float64{"food": 0, "gold": 0, "tools": 0}}
		}

		// Process buildings
		for _, bID := range set.Buildings {
			def, ok := s.buildingDefs[bID]
			if !ok || len(def.Produces) == 0 {
				continue
			}
			roleWorkers := workers[e][def.Role]
			if roleWorkers > def.MaxWorkers {
				roleWorkers = def.MaxWorkers
			}
			if roleWorkers == 0 {
				continue
			}

			// Check consumption first
			canProduce := true
			for res, rate := range def.Consumes {
				if !rs.Has(res, rate*float64(roleWorkers)*dt) {
					canProduce = false
					break
				}
			}
			if !canProduce {
				continue
			}

			// Consume
			for res, rate := range def.Consumes {
				rs.Remove(res, rate*float64(roleWorkers)*dt)
			}

			// Produce
			for res, rate := range def.Produces {
				rs.Add(res, rate*float64(roleWorkers)*dt)
			}
		}

		// NPC food consumption
		count := npcCounts[e]
		if count > 0 {
			rs.Remove("food", 0.01*float64(count)*dt)
		}

		resStore.Set(e, rs)
	}

	return nil
}

// SettlementGrowthSystem handles settlement leveling.
type SettlementGrowthSystem struct {
	thresholds map[int]settlement.GrowthThreshold
}

// NewSettlementGrowthSystem creates a growth system with thresholds.
func NewSettlementGrowthSystem(thresholds []settlement.GrowthThreshold) *SettlementGrowthSystem {
	m := make(map[int]settlement.GrowthThreshold, len(thresholds))
	for _, t := range thresholds {
		m[t.Level] = t
	}
	return &SettlementGrowthSystem{thresholds: m}
}

// Name returns the system name.
func (s *SettlementGrowthSystem) Name() string { return "SettlementGrowthSystem" }

// Update checks level-up conditions.
func (s *SettlementGrowthSystem) Update(w *ecs.World, dt float64) error {
	setStore, ok := w.GetStore(settlement.SettlementID).(*ecs.ComponentStore[settlement.Settlement])
	if !ok {
		return nil
	}
	resStore, ok := w.GetStore(settlement.ResourceID).(*ecs.ComponentStore[settlement.ResourceStore])
	if !ok {
		return nil
	}

	for e, set := range setStore.All() {
		// Max level cap (MVP level 3)
		if set.Level >= 3 {
			continue
		}

		threshold, ok := s.thresholds[set.Level+1]
		if !ok {
			continue // no threshold defined = max level
		}

		rs, ok := resStore.Get(e)
		if !ok {
			continue
		}

		if rs.Has("food", threshold.Food) && rs.Has("tools", threshold.Tools) && rs.Has("gold", threshold.Gold) {
			rs.Remove("food", threshold.Food)
			rs.Remove("tools", threshold.Tools)
			rs.Remove("gold", threshold.Gold)
			set.Level++
			if threshold.NewRadius > 0 {
				set.Radius = threshold.NewRadius
			}
			setStore.Set(e, set)
			resStore.Set(e, rs)
		}
	}

	return nil
}

// FamineSystem removes NPC home references when food is negative.
type FamineSystem struct{}

// NewFamineSystem creates a famine system.
func NewFamineSystem() *FamineSystem {
	return &FamineSystem{}
}

// Name returns the system name.
func (s *FamineSystem) Name() string { return "FamineSystem" }

// Update processes famine effects.
func (s *FamineSystem) Update(w *ecs.World, dt float64) error {
	resStore, ok := w.GetStore(settlement.ResourceID).(*ecs.ComponentStore[settlement.ResourceStore])
	if !ok {
		return nil
	}
	homeStore, ok := w.GetStore(settlement.HomeRefID).(*ecs.ComponentStore[settlement.HomeReference])
	if !ok {
		return nil
	}

	for e, rs := range resStore.All() {
		food, ok := rs.Resources["food"]
		if !ok || food >= 0 {
			continue
		}

		// Remove HomeReference from one NPC belonging to this settlement
		for ne, h := range homeStore.All() {
			if h.SettlementEntity == e {
				homeStore.Delete(ne)
				break
			}
		}
	}

	return nil
}
