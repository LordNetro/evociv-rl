package settlement

import (
	"fmt"
	"math"

	"github.com/marco/evociv-rl/internal/data"
)

// LoadSettlementTypes loads settlement definitions from the data registry.
func LoadSettlementTypes(registry *data.Registry) ([]SettlementDef, error) {
	raw, ok := data.Get[[]any](registry, "settlement-types")
	if !ok {
		return nil, fmt.Errorf("settlement-types not found in registry")
	}

	var defs []SettlementDef
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		d := SettlementDef{}
		if v, ok := m["id"].(string); ok {
			d.ID = v
		}
		if v, ok := m["name"].(string); ok {
			d.Name = v
		}
		if v, ok := m["symbol"].(string); ok {
			d.Symbol = v
		}
		if v, ok := m["color"].(string); ok {
			d.Color = v
		}
		if v, ok := toInt(m["radius"]); ok {
			d.Radius = v
		}
		if v, ok := m["biomes"].([]any); ok {
			for _, b := range v {
				if bs, ok := b.(string); ok {
					d.Biomes = append(d.Biomes, bs)
				}
			}
		}
		if v, ok := m["buildings"].([]any); ok {
			for _, b := range v {
				if bs, ok := b.(string); ok {
					d.Buildings = append(d.Buildings, bs)
				}
			}
		}
		if v, ok := toFloat64(m["spawn_weight"]); ok {
			d.SpawnWeight = v
		}
		defs = append(defs, d)
	}
	return defs, nil
}

// LoadGrowthThresholds loads growth threshold definitions from the data registry.
func LoadGrowthThresholds(registry *data.Registry) ([]GrowthThreshold, error) {
	raw, ok := data.Get[[]any](registry, "growth-thresholds")
	if !ok {
		return nil, nil // missing is not an error
	}

	var defs []GrowthThreshold
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		d := GrowthThreshold{}
		if v, ok := toInt(m["level"]); ok {
			d.Level = v
		}
		if v, ok := toFloat64(m["food"]); ok {
			d.Food = v
		}
		if v, ok := toFloat64(m["tools"]); ok {
			d.Tools = v
		}
		if v, ok := toFloat64(m["gold"]); ok {
			d.Gold = v
		}
		if v, ok := toInt(m["new_radius"]); ok {
			d.NewRadius = v
		}
		if v, ok := toStringSlice(m["new_buildings"]); ok {
			d.NewBuildings = v
		}
		defs = append(defs, d)
	}
	return defs, nil
}

// LoadBuildingTypes loads building definitions from the data registry.
func LoadBuildingTypes(registry *data.Registry) ([]BuildingDef, error) {
	raw, ok := data.Get[[]any](registry, "building-types")
	if !ok {
		return nil, fmt.Errorf("building-types not found in registry")
	}

	var defs []BuildingDef
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		d := BuildingDef{}
		if v, ok := m["id"].(string); ok {
			d.ID = v
		}
		if v, ok := m["name"].(string); ok {
			d.Name = v
		}
		if v, ok := m["role"].(string); ok {
			d.Role = v
		}
		if v, ok := toInt(m["max_workers"]); ok {
			d.MaxWorkers = v
		}
		if v, ok := m["produces"].(map[string]any); ok {
			d.Produces = make(map[string]float64)
			for rk, rv := range v {
				if f, ok := toFloat64(rv); ok {
					if f < 0 {
						return nil, fmt.Errorf("building %q: negative production rate for %q: %f", d.ID, rk, f)
					}
					d.Produces[rk] = f
				}
			}
		}
		if v, ok := m["consumes"].(map[string]any); ok {
			d.Consumes = make(map[string]float64)
			for rk, rv := range v {
				if f, ok := toFloat64(rv); ok {
					if f < 0 {
						return nil, fmt.Errorf("building %q: negative consumption rate for %q: %f", d.ID, rk, f)
					}
					d.Consumes[rk] = f
				}
			}
		}
		defs = append(defs, d)
	}
	return defs, nil
}

// validateSettlementData validates settlement definitions against building definitions.
func validateSettlementData(settlementDefs []SettlementDef, buildingDefs []BuildingDef) error {
	// Build building lookup
	buildingMap := make(map[string]bool)
	for _, b := range buildingDefs {
		buildingMap[b.ID] = true
	}

	var totalWeight float64
	for _, s := range settlementDefs {
		totalWeight += s.SpawnWeight
		for _, b := range s.Buildings {
			if !buildingMap[b] {
				return fmt.Errorf("settlement %q references unknown building %q", s.ID, b)
			}
		}
	}

	if math.Abs(totalWeight-1.0) > 0.01 {
		return fmt.Errorf("spawn weights sum to %f, expected 1.0 ± 0.01", totalWeight)
	}

	return nil
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

func toStringSlice(v any) ([]string, bool) {
	switch val := v.(type) {
	case []any:
		result := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result, true
	default:
		return nil, false
	}
}

func toInt(v any) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	default:
		return 0, false
	}
}
