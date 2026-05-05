package npc

import (
	"testing"
	"testing/fstest"

	"github.com/marco/evociv-rl/internal/data"
)

func TestLoadNpcRaces(t *testing.T) {
	fsys := fstest.MapFS{
		"data/npcs.yaml": &fstest.MapFile{
			Data: []byte(`kind: npc-races
data:
  - id: human
    name: Humano
    description: "Versatil"
    spawn_weight: 1.0
    traits:
      openness: { mean: 0.5, std: 0.15 }
    roles:
      - id: farmer
        weight: 0.4
    name_pool:
      first: ["Aldric"]
      last: ["Torres"]
`),
		},
	}

	loader := data.NewLoader(fsys)
	registry := data.NewRegistry()
	if err := loader.LoadAll("data", registry); err != nil {
		t.Fatalf("load data: %v", err)
	}

	races, err := LoadNpcRaces(registry)
	if err != nil {
		t.Fatalf("LoadNpcRaces error: %v", err)
	}
	if len(races) != 1 {
		t.Fatalf("expected 1 race, got %d", len(races))
	}
	if races[0].ID != "human" {
		t.Errorf("race ID = %q, want human", races[0].ID)
	}
	if races[0].Name != "Humano" {
		t.Errorf("race name = %q, want Humano", races[0].Name)
	}
	if races[0].SpawnWeight != 1.0 {
		t.Errorf("spawn weight = %f, want 1.0", races[0].SpawnWeight)
	}
	if len(races[0].Roles) != 1 || races[0].Roles[0].ID != "farmer" || races[0].Roles[0].Weight != 0.4 {
		t.Errorf("roles not loaded correctly: %+v", races[0].Roles)
	}
	if len(races[0].NamePool.First) != 1 || races[0].NamePool.First[0] != "Aldric" {
		t.Errorf("name pool first not loaded: %+v", races[0].NamePool.First)
	}
	if len(races[0].NamePool.Last) != 1 || races[0].NamePool.Last[0] != "Torres" {
		t.Errorf("name pool last not loaded: %+v", races[0].NamePool.Last)
	}
	if td, ok := races[0].Traits["openness"]; !ok || td.Mean != 0.5 || td.Std != 0.15 {
		t.Errorf("traits not loaded correctly: %+v", races[0].Traits)
	}
}

func TestLoadNpcRoles(t *testing.T) {
	fsys := fstest.MapFS{
		"data/npc-roles.yaml": &fstest.MapFile{
			Data: []byte(`kind: npc-roles
data:
  - id: farmer
    symbol: "@"
    color: "#FFD700"
    biomes: [plains]
`),
		},
	}

	loader := data.NewLoader(fsys)
	registry := data.NewRegistry()
	if err := loader.LoadAll("data", registry); err != nil {
		t.Fatalf("load data: %v", err)
	}

	roles, err := LoadNpcRoles(registry)
	if err != nil {
		t.Fatalf("LoadNpcRoles error: %v", err)
	}
	if len(roles) != 1 {
		t.Fatalf("expected 1 role, got %d", len(roles))
	}
	if roles[0].ID != "farmer" {
		t.Errorf("role ID = %q, want farmer", roles[0].ID)
	}
	if roles[0].Symbol != "@" {
		t.Errorf("symbol = %q, want @", roles[0].Symbol)
	}
	if roles[0].Color != "#FFD700" {
		t.Errorf("color = %q, want #FFD700", roles[0].Color)
	}
	if len(roles[0].Biomes) != 1 || roles[0].Biomes[0] != "plains" {
		t.Errorf("biomes not loaded correctly: %+v", roles[0].Biomes)
	}
}

func TestLoadNpcRacesMissing(t *testing.T) {
	registry := data.NewRegistry()
	_, err := LoadNpcRaces(registry)
	if err == nil {
		t.Error("expected error for missing npc-races")
	}
}

func TestLoadNpcRolesMissing(t *testing.T) {
	registry := data.NewRegistry()
	_, err := LoadNpcRoles(registry)
	if err == nil {
		t.Error("expected error for missing npc-roles")
	}
}

func TestRaceRoleCompatibility(t *testing.T) {
	fsys := fstest.MapFS{
		"data/npcs.yaml": &fstest.MapFile{
			Data: []byte(`kind: npc-races
data:
  - id: dwarf
    name: Enano
    spawn_weight: 0.6
    traits:
      conscientiousness: { mean: 0.7, std: 0.1 }
    roles:
      - id: miner
        weight: 0.6
      - id: smith
        weight: 0.4
    name_pool:
      first: ["Borin"]
      last: ["Hierro"]
`),
		},
		"data/npc-roles.yaml": &fstest.MapFile{
			Data: []byte(`kind: npc-roles
data:
  - id: miner
    symbol: "\u263a"
    color: "#C0C0C0"
    biomes: [plains, desert]
  - id: smith
    symbol: "\u263a"
    color: "#FF4500"
    biomes: [plains]
`),
		},
	}

	loader := data.NewLoader(fsys)
	registry := data.NewRegistry()
	if err := loader.LoadAll("data", registry); err != nil {
		t.Fatalf("load data: %v", err)
	}

	races, err := LoadNpcRaces(registry)
	if err != nil {
		t.Fatalf("LoadNpcRaces error: %v", err)
	}
	roles, err := LoadNpcRoles(registry)
	if err != nil {
		t.Fatalf("LoadNpcRoles error: %v", err)
	}

	// Verify dwarf can be miner or smith
	race := races[0]
	if race.ID != "dwarf" {
		t.Fatalf("expected dwarf, got %s", race.ID)
	}
	if len(race.Roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(race.Roles))
	}

	roleMap := make(map[string]RoleDef)
	for _, r := range roles {
		roleMap[r.ID] = r
	}

	for _, rw := range race.Roles {
		if _, ok := roleMap[rw.ID]; !ok {
			t.Errorf("race references unknown role %s", rw.ID)
		}
	}
}
