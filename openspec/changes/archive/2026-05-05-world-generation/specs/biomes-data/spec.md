# biomes-data Specification

## Purpose

Biomas data-driven con rangos de altura/humedad/temperatura.

## Requirements

### Requirement: Ranges

Each biome MUST define symbol (rune), min/max of height/humidity/temperature.

#### Scenario: Ranges accessible
- GIVEN biome def with all fields
- WHEN loaded
- THEN fields MUST be accessible

### Requirement: Assignment

MUST assign biome by matching tile values to ranges.

#### Scenario: Plains assigned
- GIVEN tile h=0.3, hu=0.4, t=0.5 within plains ranges
- WHEN assignment runs
- THEN tile MUST be plains

### Requirement: Unknown Fallback

SHOULD have unknown biome. MUST assign when no range matches.

#### Scenario: Extreme → unknown
- GIVEN tile height=100.0
- WHEN no range matches
- THEN biome MUST be unknown
