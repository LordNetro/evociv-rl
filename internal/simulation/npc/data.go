package npc

import (
	"fmt"

	"github.com/marco/evociv-rl/internal/data"
)

// LoadNpcRaces loads race definitions from the data registry.
func LoadNpcRaces(registry *data.Registry) ([]RaceDef, error) {
	raw, ok := data.Get[[]any](registry, "npc-races")
	if !ok {
		return nil, fmt.Errorf("npc-races not found in registry")
	}

	var races []RaceDef
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		r := RaceDef{}
		if v, ok := m["id"].(string); ok {
			r.ID = v
		}
		if v, ok := m["name"].(string); ok {
			r.Name = v
		}
		if v, ok := m["description"].(string); ok {
			r.Description = v
		}
		if v, ok := toFloat64(m["spawn_weight"]); ok {
			r.SpawnWeight = v
		}
		if v, ok := m["traits"].(map[string]any); ok {
			r.Traits = parseTraits(v)
		}
		if v, ok := m["roles"].([]any); ok {
			r.Roles = parseRoleWeights(v)
		}
		if v, ok := m["name_pool"].(map[string]any); ok {
			r.NamePool = parseNamePool(v)
		}
		races = append(races, r)
	}
	return races, nil
}

// LoadNpcRoles loads role definitions from the data registry.
func LoadNpcRoles(registry *data.Registry) ([]RoleDef, error) {
	raw, ok := data.Get[[]any](registry, "npc-roles")
	if !ok {
		return nil, fmt.Errorf("npc-roles not found in registry")
	}

	var roles []RoleDef
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		r := RoleDef{}
		if v, ok := m["id"].(string); ok {
			r.ID = v
		}
		if v, ok := m["symbol"].(string); ok {
			r.Symbol = v
		}
		if v, ok := m["color"].(string); ok {
			r.Color = v
		}
		if v, ok := m["biomes"].([]any); ok {
			for _, b := range v {
				if bs, ok := b.(string); ok {
					r.Biomes = append(r.Biomes, bs)
				}
			}
		}
		roles = append(roles, r)
	}
	return roles, nil
}

func parseTraits(m map[string]any) map[string]TraitDef {
	out := make(map[string]TraitDef)
	for key, val := range m {
		tm, ok := val.(map[string]any)
		if !ok {
			continue
		}
		td := TraitDef{}
		if v, ok := toFloat64(tm["mean"]); ok {
			td.Mean = v
		}
		if v, ok := toFloat64(tm["std"]); ok {
			td.Std = v
		}
		out[key] = td
	}
	return out
}

func parseRoleWeights(arr []any) []RoleWeight {
	var out []RoleWeight
	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		rw := RoleWeight{}
		if v, ok := m["id"].(string); ok {
			rw.ID = v
		}
		if v, ok := toFloat64(m["weight"]); ok {
			rw.Weight = v
		}
		out = append(out, rw)
	}
	return out
}

func parseNamePool(m map[string]any) NamePool {
	np := NamePool{}
	if v, ok := m["first"].([]any); ok {
		for _, s := range v {
			if str, ok := s.(string); ok {
				np.First = append(np.First, str)
			}
		}
	}
	if v, ok := m["last"].([]any); ok {
		for _, s := range v {
			if str, ok := s.(string); ok {
				np.Last = append(np.Last, str)
			}
		}
	}
	return np
}

// LoadActions loads action definitions from the data registry.
func LoadActions(registry *data.Registry) ([]ActionDef, error) {
	raw, ok := data.Get[[]any](registry, "npc-actions")
	if !ok {
		return nil, fmt.Errorf("npc-actions not found in registry")
	}

	var actions []ActionDef
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		a := ActionDef{}
		if v, ok := m["id"].(string); ok {
			a.ID = v
		}
		if v, ok := m["name"].(string); ok {
			a.Name = v
		}
		if v, ok := m["requires"].(map[string]any); ok {
			a.Requires = parseActionRequires(v)
		}
		if v, ok := m["effects"].(map[string]any); ok {
			a.Effects = parseActionEffects(v)
		}
		if v, ok := m["reward"].(map[string]any); ok {
			a.Reward = parseActionReward(v)
		}
		actions = append(actions, a)
	}
	return actions, nil
}

func parseActionRequires(m map[string]any) ActionRequires {
	r := ActionRequires{}
	if v, ok := m["biomes"].([]any); ok {
		for _, b := range v {
			if bs, ok := b.(string); ok {
				r.Biomes = append(r.Biomes, bs)
			}
		}
	}
	if v, ok := m["needs_min"].(map[string]any); ok {
		r.NeedsMin = parseNeedsValues(v)
	}
	if v, ok := m["needs_max"].(map[string]any); ok {
		r.NeedsMax = parseNeedsValues(v)
	}
	return r
}

func parseActionEffects(m map[string]any) ActionEffects {
	e := ActionEffects{}
	if v, ok := toFloat64(m["hunger_change"]); ok {
		e.HungerChange = v
	}
	if v, ok := toFloat64(m["fatigue_change"]); ok {
		e.FatigueChange = v
	}
	return e
}

func parseActionReward(m map[string]any) ActionReward {
	r := ActionReward{}
	if v, ok := toFloat64(m["base"]); ok {
		r.Base = v
	}
	return r
}

func parseNeedsValues(m map[string]any) NeedsValues {
	nv := NeedsValues{}
	if v, ok := toFloat64(m["hunger"]); ok {
		nv.Hunger = v
	}
	if v, ok := toFloat64(m["fatigue"]); ok {
		nv.Fatigue = v
	}
	return nv
}

func toFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}
