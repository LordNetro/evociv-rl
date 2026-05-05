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
