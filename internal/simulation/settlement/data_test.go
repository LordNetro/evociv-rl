package settlement

import (
	"testing"
	"testing/fstest"

	"github.com/marco/evociv-rl/internal/data"
)

func TestLoadSettlementTypes(t *testing.T) {
	fsys := fstest.MapFS{
		"data/settlements.yaml": &fstest.MapFile{
			Data: []byte(`kind: settlement-types
data:
  - id: village
    name: Aldea
    symbol: "♦"
    color: "#8B7355"
    radius: 3
    biomes: [plains, forest]
    buildings: [house, farm]
    spawn_weight: 0.6
  - id: town
    name: Pueblo
    symbol: "▲"
    color: "#B8860B"
    radius: 5
    biomes: [plains]
    buildings: [house, market, tavern, blacksmith, farm]
    spawn_weight: 0.3
  - id: city
    name: Ciudad
    symbol: "●"
    color: "#DAA520"
    radius: 8
    biomes: [plains]
    buildings: [house, market, temple, tavern, blacksmith, farm]
    spawn_weight: 0.1
`),
		},
		"data/buildings.yaml": &fstest.MapFile{
			Data: []byte(`kind: building-types
data:
  - id: house
    name: Casa
  - id: farm
    name: Granja
  - id: market
    name: Mercado
  - id: tavern
    name: Taberna
  - id: temple
    name: Templo
  - id: blacksmith
    name: Herreria
`),
		},
	}

	loader := data.NewLoader(fsys)
	registry := data.NewRegistry()
	if err := loader.LoadAll("data", registry); err != nil {
		t.Fatalf("load data: %v", err)
	}

	settlementDefs, err := LoadSettlementTypes(registry)
	if err != nil {
		t.Fatalf("LoadSettlementTypes error: %v", err)
	}
	if len(settlementDefs) != 3 {
		t.Fatalf("expected 3 settlement types, got %d", len(settlementDefs))
	}
	if settlementDefs[0].ID != "village" {
		t.Errorf("first settlement ID = %q, want village", settlementDefs[0].ID)
	}
	if settlementDefs[0].Radius != 3 {
		t.Errorf("village radius = %d, want 3", settlementDefs[0].Radius)
	}
	if len(settlementDefs[0].Biomes) != 2 {
		t.Errorf("village biomes = %v, want 2", settlementDefs[0].Biomes)
	}
	if len(settlementDefs[0].Buildings) != 2 {
		t.Errorf("village buildings = %v, want 2", settlementDefs[0].Buildings)
	}
	if settlementDefs[0].SpawnWeight != 0.6 {
		t.Errorf("village spawn_weight = %f, want 0.6", settlementDefs[0].SpawnWeight)
	}

	buildingDefs, err := LoadBuildingTypes(registry)
	if err != nil {
		t.Fatalf("LoadBuildingTypes error: %v", err)
	}
	if len(buildingDefs) != 6 {
		t.Fatalf("expected 6 building types, got %d", len(buildingDefs))
	}
	if buildingDefs[0].ID != "house" {
		t.Errorf("first building ID = %q, want house", buildingDefs[0].ID)
	}
	if buildingDefs[0].Name != "Casa" {
		t.Errorf("first building Name = %q, want Casa", buildingDefs[0].Name)
	}
}

func TestLoadSettlementTypesMissing(t *testing.T) {
	registry := data.NewRegistry()
	_, err := LoadSettlementTypes(registry)
	if err == nil {
		t.Error("expected error for missing settlement-types")
	}
}

func TestLoadBuildingTypesMissing(t *testing.T) {
	registry := data.NewRegistry()
	_, err := LoadBuildingTypes(registry)
	if err == nil {
		t.Error("expected error for missing building-types")
	}
}

func TestValidateSettlementDataWeights(t *testing.T) {
	fsys := fstest.MapFS{
		"data/settlements.yaml": &fstest.MapFile{
			Data: []byte(`kind: settlement-types
data:
  - id: village
    name: Aldea
    symbol: "♦"
    color: "#8B7355"
    radius: 3
    biomes: [plains]
    buildings: [house]
    spawn_weight: 0.7
  - id: town
    name: Pueblo
    symbol: "▲"
    color: "#B8860B"
    radius: 5
    biomes: [plains]
    buildings: [house]
    spawn_weight: 0.2
`),
		},
		"data/buildings.yaml": &fstest.MapFile{
			Data: []byte(`kind: building-types
data:
  - id: house
    name: Casa
`),
		},
	}

	loader := data.NewLoader(fsys)
	registry := data.NewRegistry()
	if err := loader.LoadAll("data", registry); err != nil {
		t.Fatalf("load data: %v", err)
	}

	settlementDefs, err := LoadSettlementTypes(registry)
	if err != nil {
		t.Fatalf("LoadSettlementTypes error: %v", err)
	}
	buildingDefs, err := LoadBuildingTypes(registry)
	if err != nil {
		t.Fatalf("LoadBuildingTypes error: %v", err)
	}

	// Weights sum to 0.9, should fail validation (±0.01 tolerance)
	err = validateSettlementData(settlementDefs, buildingDefs)
	if err == nil {
		t.Error("expected validation error for weights not summing to 1.0")
	}
}

func TestValidateSettlementDataUnknownBuilding(t *testing.T) {
	fsys := fstest.MapFS{
		"data/settlements.yaml": &fstest.MapFile{
			Data: []byte(`kind: settlement-types
data:
  - id: village
    name: Aldea
    symbol: "♦"
    color: "#8B7355"
    radius: 3
    biomes: [plains]
    buildings: [house, forge]
    spawn_weight: 1.0
`),
		},
		"data/buildings.yaml": &fstest.MapFile{
			Data: []byte(`kind: building-types
data:
  - id: house
    name: Casa
`),
		},
	}

	loader := data.NewLoader(fsys)
	registry := data.NewRegistry()
	if err := loader.LoadAll("data", registry); err != nil {
		t.Fatalf("load data: %v", err)
	}

	settlementDefs, err := LoadSettlementTypes(registry)
	if err != nil {
		t.Fatalf("LoadSettlementTypes error: %v", err)
	}
	buildingDefs, err := LoadBuildingTypes(registry)
	if err != nil {
		t.Fatalf("LoadBuildingTypes error: %v", err)
	}

	err = validateSettlementData(settlementDefs, buildingDefs)
	if err == nil {
		t.Error("expected validation error for unknown building reference")
	}
}

func TestValidateSettlementDataValid(t *testing.T) {
	fsys := fstest.MapFS{
		"data/settlements.yaml": &fstest.MapFile{
			Data: []byte(`kind: settlement-types
data:
  - id: village
    name: Aldea
    symbol: "♦"
    color: "#8B7355"
    radius: 3
    biomes: [plains]
    buildings: [house]
    spawn_weight: 1.0
`),
		},
		"data/buildings.yaml": &fstest.MapFile{
			Data: []byte(`kind: building-types
data:
  - id: house
    name: Casa
`),
		},
	}

	loader := data.NewLoader(fsys)
	registry := data.NewRegistry()
	if err := loader.LoadAll("data", registry); err != nil {
		t.Fatalf("load data: %v", err)
	}

	settlementDefs, err := LoadSettlementTypes(registry)
	if err != nil {
		t.Fatalf("LoadSettlementTypes error: %v", err)
	}
	buildingDefs, err := LoadBuildingTypes(registry)
	if err != nil {
		t.Fatalf("LoadBuildingTypes error: %v", err)
	}

	err = validateSettlementData(settlementDefs, buildingDefs)
	if err != nil {
		t.Errorf("expected validation to pass, got error: %v", err)
	}
}
