package npc

import (
	"fmt"
	"math/rand"

	"github.com/charmbracelet/lipgloss"
	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/world"
)

// biomeWeight returns the spawn suitability for a biome.
func biomeWeight(biome string) float64 {
	switch biome {
	case "plains", "forest":
		return 1.0
	case "tundra", "desert":
		return 0.2
	case "ocean", "jungle":
		return 0.0
	default:
		return 0.0
	}
}

// Spawn creates NPCs in the ECS world based on the provided configuration.
func Spawn(w *ecs.World, wm *world.WorldMap, config SpawnConfig, seed int64, raceDefs []RaceDef, roleDefs []RoleDef) error {
	count := config.Count
	if count == 0 {
		area := float64(wm.Width * wm.Height)
		count = int(area * config.Density)
		if count < 50 {
			count = 50
		}
		if count > 100 {
			count = 100
		}
	}

	rng := rand.New(rand.NewSource(seed + 999))

	// Build role lookup map
	roleMap := make(map[string]RoleDef)
	for _, rd := range roleDefs {
		roleMap[rd.ID] = rd
	}

	// Precompute total race weight
	var totalRaceWeight float64
	for _, r := range raceDefs {
		totalRaceWeight += r.SpawnWeight
	}

	posStore, ok := w.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])
	if !ok {
		return fmt.Errorf("position store not registered")
	}

	for i := 0; i < count; i++ {
		// 1. Choose race
		race := pickRace(raceDefs, totalRaceWeight, rng)
		if race == nil {
			continue
		}

		// 2. Choose role within race
		roleWeight := pickRoleWeight(race.Roles, rng)
		if roleWeight == nil {
			continue
		}

		// 3. Verify role exists
		roleDef, ok := roleMap[roleWeight.ID]
		if !ok {
			continue // race-role rejection
		}

		// 4. Find valid position
		var posX, posY int
		found := false
		for attempt := 0; attempt < 100; attempt++ {
			x := rng.Intn(wm.Width)
			y := rng.Intn(wm.Height)
			tile := wm.TileAt(x, y)
			if tile == nil {
				continue
			}
			wgt := biomeWeight(tile.BiomeID)
			if wgt <= 0 {
				continue
			}
			if !biomeCompatible(tile.BiomeID, roleDef.Biomes) {
				continue
			}
			if rng.Float64() > wgt {
				continue
			}
			if isOccupied(posStore, x, y) {
				continue
			}
			posX, posY = x, y
			found = true
			break
		}
		if !found {
			continue
		}

		// 5. Generate name
		name := generateName(race.NamePool, rng)

		// 6. Create entity and components
		e := w.NewEntity()
		ecs.AddComponent(w, e, ecs.Position{X: float64(posX), Y: float64(posY)})
		ecs.AddComponent(w, e, ecs.Name{Name: name})
		ecs.AddComponent(w, e, Health{Current: 100, Max: 100})
		ecs.AddComponent(w, e, NewPersonality(rng))
		ecs.AddComponent(w, e, Job{Role: roleDef.ID})
		ecs.AddComponent(w, e, AIState{Goals: []string{"wander"}, Mood: 0})
		ecs.AddComponent(w, e, Appearance{
			Symbol: []rune(roleDef.Symbol)[0],
			Color:  lipgloss.Color(roleDef.Color),
		})
		ecs.AddComponent(w, e, LOD{Level: LODLocal})
	}

	return nil
}

func pickRace(races []RaceDef, totalWeight float64, rng *rand.Rand) *RaceDef {
	if totalWeight <= 0 {
		return nil
	}
	target := rng.Float64() * totalWeight
	var cumulative float64
	for i := range races {
		cumulative += races[i].SpawnWeight
		if target < cumulative {
			return &races[i]
		}
	}
	return &races[len(races)-1]
}

func pickRoleWeight(roles []RoleWeight, rng *rand.Rand) *RoleWeight {
	var total float64
	for _, r := range roles {
		total += r.Weight
	}
	if total <= 0 {
		return nil
	}
	target := rng.Float64() * total
	var cumulative float64
	for i := range roles {
		cumulative += roles[i].Weight
		if target < cumulative {
			return &roles[i]
		}
	}
	return &roles[len(roles)-1]
}

func biomeCompatible(biome string, allowed []string) bool {
	for _, b := range allowed {
		if b == biome {
			return true
		}
	}
	return false
}

func isOccupied(store *ecs.ComponentStore[ecs.Position], x, y int) bool {
	for _, p := range store.All() {
		if int(p.X) == x && int(p.Y) == y {
			return true
		}
	}
	return false
}

func generateName(pool NamePool, rng *rand.Rand) string {
	first := "Unknown"
	if len(pool.First) > 0 {
		first = pool.First[rng.Intn(len(pool.First))]
	}
	last := ""
	if len(pool.Last) > 0 {
		last = pool.Last[rng.Intn(len(pool.Last))]
	}
	if last != "" {
		return first + " " + last
	}
	return first
}
