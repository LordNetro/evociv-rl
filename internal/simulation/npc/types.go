package npc

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/marco/evociv-rl/internal/ecs"
)

// RaceDef defines an NPC race with spawn weights, traits, roles, and name pools.
type RaceDef struct {
	ID          string             `yaml:"id"`
	Name        string             `yaml:"name"`
	Description string             `yaml:"description"`
	SpawnWeight float64            `yaml:"spawn_weight"`
	Traits      map[string]TraitDef `yaml:"traits"`
	Roles       []RoleWeight       `yaml:"roles"`
	NamePool    NamePool           `yaml:"name_pool"`
}

// TraitDef defines the distribution parameters for a personality trait.
type TraitDef struct {
	Mean float64 `yaml:"mean"`
	Std  float64 `yaml:"std"`
}

// RoleWeight pairs a role ID with a selection weight within a race.
type RoleWeight struct {
	ID     string  `yaml:"id"`
	Weight float64 `yaml:"weight"`
}

// NamePool holds first and last name lists for a race.
type NamePool struct {
	First []string `yaml:"first"`
	Last  []string `yaml:"last"`
}

// RoleDef defines a role with its visual symbol, color, and compatible biomes.
type RoleDef struct {
	ID     string   `yaml:"id"`
	Symbol string   `yaml:"symbol"`
	Color  string   `yaml:"color"`
	Biomes []string `yaml:"biomes"`
}

// SpawnConfig controls how many NPCs to spawn.
type SpawnConfig struct {
	Count   int
	Density float64
}

// NeedsValues represents hunger and fatigue as a pair, used in action requirements.
type NeedsValues struct {
	Hunger  float64 `yaml:"hunger"`
	Fatigue float64 `yaml:"fatigue"`
}

// ActionDef defines a GOAP action loaded from YAML.
type ActionDef struct {
	ID       string         `yaml:"id"`
	Name     string         `yaml:"name"`
	Requires ActionRequires `yaml:"requires"`
	Effects  ActionEffects  `yaml:"effects"`
	Reward   ActionReward   `yaml:"reward"`
}

// ActionRequires defines biome and need constraints for an action.
type ActionRequires struct {
	Biomes   []string    `yaml:"biomes"`
	NeedsMin NeedsValues `yaml:"needs_min"`
	NeedsMax NeedsValues `yaml:"needs_max"`
}

// ActionEffects defines how an action changes needs.
type ActionEffects struct {
	HungerChange  float64 `yaml:"hunger_change"`
	FatigueChange float64 `yaml:"fatigue_change"`
}

// ActionReward defines the base reward for completing an action.
type ActionReward struct {
	Base float64 `yaml:"base"`
}

// NPCRenderInfo carries everything the TUI needs to draw an NPC overlay.
type NPCRenderInfo struct {
	Entity         ecs.Entity
	Symbol         rune
	Color          lipgloss.Color
	WorldX, WorldY int
}
