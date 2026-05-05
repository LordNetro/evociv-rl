# goap-actions Specification

## Purpose

Acciones GOAP cargadas desde YAML, data-driven, cada una con requisitos de bioma, rangos de necesidad, efectos y recompensa base.

## Requirements

### Requirement: Action Definitions

The system MUST define exactly six actions loaded from `data/actions.yaml`: harvest, forage, rest, socialize, trade, and explore.

Each action entry MUST contain:

| Field | Type | Required |
|-------|------|----------|
| `id` | string | MUST |
| `name` | string | MUST |
| `requires.biomes` | []string | MAY |
| `requires.needs_min` | {hunger, fatigue float64} | MUST |
| `requires.needs_max` | {hunger, fatigue float64} | MUST |
| `effects.hunger_change` | float64 | MUST |
| `effects.fatigue_change` | float64 | MUST |
| `reward.base` | float64 | MUST |

#### Scenario: Load valid YAML

- GIVEN a valid `data/actions.yaml` with all six actions
- WHEN LoadActions() is called
- THEN all six actions MUST be returned with correct field values

#### Scenario: Biome-restricted action

- GIVEN action "harvest" with requires.biomes=["plains", "forest"]
- WHEN checking availability for an NPC on an "ocean" tile
- THEN harvest MUST be marked as unavailable for that tile

#### Scenario: Actions available without biome restriction

- GIVEN actions forage, rest, socialize, trade, explore without requires.biomes
- WHEN checking availability
- THEN they MUST be available on any biome

### Requirement: Action Validation

The system MUST reject action definitions with missing required fields, returning a descriptive error.

#### Scenario: Missing required field returns error

- GIVEN an actions.yaml entry missing `effects.hunger_change`
- WHEN LoadActions() is called
- THEN a validation error MUST be returned identifying the missing field and action ID

#### Scenario: Invalid YAML structure returns error

- GIVEN an actions.yaml with invalid syntax
- WHEN LoadActions() is called
- THEN a parse error MUST be returned
