# settlement-buildings Specification

## Purpose

Edificios como entidades ECS hijas dentro del radio de cada settlement. Spawn determinado por tipo de asentamiento.

## Requirements

### Requirement: Buildings as ECS Entities

Each building MUST be created as a separate ECS entity with a `Building` component (ID, Name, Level), a `Position` component (relative to settlement center), and NO `Settlement` component. Building entities MUST be spawned during settlement creation, not deferred.

#### Scenario: Building is an independent entity

- GIVEN a settlement entity with ID 7
- WHEN a farm building is created for that settlement
- THEN the farm MUST be a distinct entity with a Building component and a Position component
- AND the farm entity MUST NOT have a Settlement component

### Requirement: Building Composition by Settlement Type

Each settlement type MUST spawn a specific set of buildings:

| Settlement Type | Buildings |
|---------------|-----------|
| village | 3–5 houses, 1 farm |
| town | 5–8 houses, 1 farm, 1 market, 1 tavern |
| city | 10–15 houses, 2 farms, 1 market, 1 tavern, 1 temple, 1 blacksmith |

#### Scenario: Village spawns with houses and farm

- GIVEN a village settlement entity
- WHEN buildings are spawned
- THEN the village MUST have between 3 and 5 house entities and exactly 1 farm entity

#### Scenario: City spawns with all building types

- GIVEN a city settlement entity
- WHEN buildings are spawned
- THEN the city MUST have 10–15 houses, 2 farms, 1 market, 1 tavern, 1 temple, and 1 blacksmith

### Requirement: Relative Position Within Radius

Building positions MUST be within the settlement's radius (distance ≤ Radius from settlement center). Building positions SHOULD be randomly distributed within the radius using the settlement seed for determinism. Buildings MUST NOT overlap (same position) — if a random position collision occurs, the system MUST retry up to 5 times.

#### Scenario: Buildings placed within radius

- GIVEN a village with center (10, 10) and Radius 6
- WHEN a building entity is created for that village
- THEN its Chebyshev distance from (10, 10) MUST be ≤ 6

#### Scenario: Buildings do not overlap

- GIVEN a settlement spawning multiple buildings
- WHEN any two building positions are compared
- THEN they MUST differ in at least one coordinate

### Requirement: Deterministic Building Spawn

Building count, types, and positions MUST be deterministic given the world seed. The system MUST use world seed + 888 as its internal PRNG seed for building placement.

#### Scenario: Same seed produces identical buildings

- GIVEN world seed S
- WHEN spawning buildings for a settlement in two separate runs
- THEN both runs MUST produce identical building counts, types, and positions
