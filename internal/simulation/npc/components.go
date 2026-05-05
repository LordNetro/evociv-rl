package npc

import (
	"math"
	"math/rand"

	"github.com/charmbracelet/lipgloss"
	"github.com/marco/evociv-rl/internal/ecs"
)

// LOD constants.
const (
	LODDistant = 0
	LODNear    = 1
	LODLocal   = 2
)

// Health represents the vital status of an NPC.
type Health struct {
	Current, Max float64
}

// Personality represents the Big Five traits, each normalized to [0,1].
type Personality struct {
	Openness        float64
	Conscientiousness float64
	Extraversion    float64
	Agreeableness   float64
	Neuroticism     float64
}

// NewPersonality generates deterministic traits using a truncated Gaussian
// distribution with default mean=0.5 and std=0.15 for each trait.
func NewPersonality(rng *rand.Rand) Personality {
	return Personality{
		Openness:        truncatedGaussian(rng, 0.5, 0.15),
		Conscientiousness: truncatedGaussian(rng, 0.5, 0.15),
		Extraversion:    truncatedGaussian(rng, 0.5, 0.15),
		Agreeableness:   truncatedGaussian(rng, 0.5, 0.15),
		Neuroticism:     truncatedGaussian(rng, 0.5, 0.15),
	}
}

func truncatedGaussian(rng *rand.Rand, mean, std float64) float64 {
	v := rng.NormFloat64()*std + mean
	return math.Max(0.0, math.Min(1.0, v))
}

// Job represents an NPC's occupation.
type Job struct {
	Role string
}

// AIState holds cognitive state for GOAP-ready NPCs.
type AIState struct {
	Goals []string
	Plan  []string
	Mood  float64
}

// Appearance defines how an NPC is rendered on the map.
type Appearance struct {
	Symbol rune
	Color  lipgloss.Color
}

// LOD controls the simulation detail level for an NPC.
type LOD struct {
	Level int
}

// Component IDs for the NPC component types.
var (
	HealthID      = ecs.NewComponentID("npc_health")
	PersonalityID = ecs.NewComponentID("npc_personality")
	JobID         = ecs.NewComponentID("npc_job")
	AIStateID     = ecs.NewComponentID("npc_aistate")
	AppearanceID  = ecs.NewComponentID("npc_appearance")
	LODID         = ecs.NewComponentID("npc_lod")
)

// RegisterStores registers the six NPC component stores (plus Position and Name)
// on the given world.
func RegisterStores(w *ecs.World) {
	ecs.RegisterComponentStore[ecs.Position](w, ecs.NewComponentID("position"), ecs.NewComponentStore[ecs.Position]())
	ecs.RegisterComponentStore[ecs.Name](w, ecs.NewComponentID("name"), ecs.NewComponentStore[ecs.Name]())
	ecs.RegisterComponentStore[Health](w, HealthID, ecs.NewComponentStore[Health]())
	ecs.RegisterComponentStore[Personality](w, PersonalityID, ecs.NewComponentStore[Personality]())
	ecs.RegisterComponentStore[Job](w, JobID, ecs.NewComponentStore[Job]())
	ecs.RegisterComponentStore[AIState](w, AIStateID, ecs.NewComponentStore[AIState]())
	ecs.RegisterComponentStore[Appearance](w, AppearanceID, ecs.NewComponentStore[Appearance]())
	ecs.RegisterComponentStore[LOD](w, LODID, ecs.NewComponentStore[LOD]())
}
