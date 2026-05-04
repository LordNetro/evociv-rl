package gen

import (
	"fmt"

	"github.com/marco/evociv-rl/internal/data"
	"github.com/marco/evociv-rl/internal/world"
)

// BiomeDef defines a biome with ranges for height, humidity, and temperature.
type BiomeDef struct {
	ID             string  `yaml:"id"`
	Name           string  `yaml:"name"`
	Symbol         string  `yaml:"symbol"`
	Color          string  `yaml:"color"`
	MinHeight      float64 `yaml:"minHeight"`
	MaxHeight      float64 `yaml:"maxHeight"`
	MinHumidity    float64 `yaml:"minHumidity"`
	MaxHumidity    float64 `yaml:"maxHumidity"`
	MinTemperature float64 `yaml:"minTemperature"`
	MaxTemperature float64 `yaml:"maxTemperature"`
}

// MatchBiome finds the first biome whose ranges contain the tile's values.
// Returns "unknown" if no biome matches.
func MatchBiome(tile world.Tile, biomes []BiomeDef) string {
	for _, b := range biomes {
		if tile.Height >= b.MinHeight && tile.Height <= b.MaxHeight &&
			tile.Humidity >= b.MinHumidity && tile.Humidity <= b.MaxHumidity &&
			tile.Temperature >= b.MinTemperature && tile.Temperature <= b.MaxTemperature {
			return b.ID
		}
	}
	return "unknown"
}

// LoadBiomes loads biome definitions from the data registry.
func LoadBiomes(registry *data.Registry) ([]BiomeDef, error) {
	raw, ok := data.Get[[]any](registry, "biomes")
	if !ok {
		return nil, fmt.Errorf("biomes not found in registry")
	}

	var biomes []BiomeDef
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		b := BiomeDef{}
		if v, ok := m["id"].(string); ok {
			b.ID = v
		}
		if v, ok := m["name"].(string); ok {
			b.Name = v
		}
		if v, ok := m["symbol"].(string); ok {
			b.Symbol = v
		}
		if v, ok := m["color"].(string); ok {
			b.Color = v
		}
		if v, ok := toFloat64(m["minHeight"]); ok {
			b.MinHeight = v
		}
		if v, ok := toFloat64(m["maxHeight"]); ok {
			b.MaxHeight = v
		}
		if v, ok := toFloat64(m["minHumidity"]); ok {
			b.MinHumidity = v
		}
		if v, ok := toFloat64(m["maxHumidity"]); ok {
			b.MaxHumidity = v
		}
		if v, ok := toFloat64(m["minTemperature"]); ok {
			b.MinTemperature = v
		}
		if v, ok := toFloat64(m["maxTemperature"]); ok {
			b.MaxTemperature = v
		}
		biomes = append(biomes, b)
	}
	return biomes, nil
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
