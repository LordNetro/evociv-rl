# settlement-data Specification

## Purpose

Definiciones data-driven de tipos de asentamiento y edificios en YAML. Carga y validación independiente del Registry genérico.

## Requirements

### Requirement: Settlement Type Definitions

El archivo `data/settlements.yaml` MUST definir al menos tres tipos: `village`, `town`, `city`. Cada tipo MUST tener: `id`, `symbol` (rune), `color`, `radius` (int > 0), `biomes` ([]string), `spawn_weight` (float64).

| Type | Symbol | Color | Radius | Biomes | Spawn Weight |
|------|--------|-------|--------|--------|-------------|
| village | ♦ | green | 5-8 | plains, forest | 0.6 |
| town | ▲ | yellow | 10-15 | plains | 0.3 |
| city | ● | red | 18-25 | plains | 0.1 |

#### Scenario: Load valid settlements.yaml

- GIVEN a `data/settlements.yaml` with village, town, city definitions
- WHEN `LoadSettlementTypes()` is called
- THEN all three types MUST be accessible with correct symbol, color, radius, biomes, and spawn_weight

### Requirement: Building Type Definitions

El archivo `data/buildings.yaml` MUST definir al menos: `house`, `farm`, `market`, `tavern`, `temple`, `blacksmith`. Cada edificio MUST tener: `id`, `name`, `symbol` (rune), `color`, `produces` ([]string, MAY be empty).

#### Scenario: Load valid buildings.yaml

- GIVEN a `data/buildings.yaml` with all six building types
- WHEN `LoadBuildingTypes()` is called
- THEN all buildings MUST be accessible with correct fields

### Requirement: Validation — Required Fields

The loader MUST reject any settlement or building type missing a required field (`id`, `symbol`, `color`, `radius` for settlements, `id`, `symbol`, `color` for buildings). Rejection MUST return a typed validation error.

#### Scenario: Reject settlement without id

- GIVEN a settlements.yaml entry missing the `id` field
- WHEN validation runs
- THEN an error MUST be returned describing the missing field

#### Scenario: Reject building without symbol

- GIVEN a buildings.yaml entry missing the `symbol` field
- WHEN validation runs
- THEN an error MUST be returned

### Requirement: Validation — Biome References

The loader MUST reject settlement types whose `biomes` list contains a biome ID not defined in the biomes registry. Unknown biome IDs MUST produce a validation error.

#### Scenario: Reject biome not in registry

- GIVEN a settlement type referencing biome "oceanic" (not in biomes registry)
- WHEN validation runs
- THEN an error MUST be returned identifying the unknown biome
