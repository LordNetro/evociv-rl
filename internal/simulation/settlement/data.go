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
