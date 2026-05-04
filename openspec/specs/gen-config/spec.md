# gen-config Specification

## Purpose

Config YAML con parámetros de ruido y dimensiones.

## Requirements

### Requirement: Fields

MUST define octaves (int), lacunarity (float64), gain (float64), scale (float64), width (int), height (int), seed (int64).

#### Scenario: Valid loads
- GIVEN gen-config.yaml with valid values
- WHEN loaded
- THEN all fields populated

### Requirement: Validation

MUST reject width≤0 or height≤0. SHOULD default octaves=4, lacunarity=2.0, gain=0.5.

#### Scenario: Zero rejected
- GIVEN width=0
- WHEN Validate()
- THEN MUST return error

#### Scenario: Defaults
- GIVEN config with width=64, height=64, seed=42 only
- WHEN loaded
- THEN octaves SHOULD default to 4
