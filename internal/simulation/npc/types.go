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

// NPCRenderInfo carries everything the TUI needs to draw an NPC overlay.
type NPCRenderInfo struct {
	Entity         ecs.Entity
	Symbol         rune
	Color          lipgloss.Color
	WorldX, WorldY int
}
