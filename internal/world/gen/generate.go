package gen

import (
	"fmt"

	"github.com/marco/evociv-rl/internal/world"
)

// Generate creates a new WorldMap using the 4-phase generation pipeline:
// 1. Height (FBM noise)
// 2. Humidity (FBM noise with seed+1)
// 3. Temperature (FBM noise with seed+2, modulated by height)
// 4. Biome assignment via range matching.
func Generate(w, h int, config GenConfig, biomes []BiomeDef) (*world.WorldMap, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	wm := world.NewWorldMap(w, h)
	if wm == nil {
		return nil, fmt.Errorf("failed to create world map")
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			tile := wm.TileAt(x, y)
			if tile == nil {
				continue
			}

			// Phase 1: Height
			tile.Height = FBM2D(float64(x), float64(y), config.Octaves, config.Lacunarity, config.Gain, config.Scale, config.Seed)

			// Phase 2: Humidity
			tile.Humidity = FBM2D(float64(x), float64(y), config.Octaves, config.Lacunarity, config.Gain, config.Scale, config.Seed+1)

			// Phase 3: Temperature (modulated by height)
			rawTemp := FBM2D(float64(x), float64(y), config.Octaves, config.Lacunarity, config.Gain, config.Scale, config.Seed+2)
			heightFactor := (tile.Height + 1.0) / 2.0 // normalize height from [-1,1] to [0,1]
			tile.Temperature = rawTemp * heightFactor

			// Phase 4: Biome
			tile.BiomeID = MatchBiome(*tile, biomes)
		}
	}

	return wm, nil
}
