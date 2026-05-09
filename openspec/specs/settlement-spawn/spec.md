# settlement-spawn Specification

## Purpose

Generar asentamientos (settlements) sobre el mundo 256×256 con distribución por bioma y determinismo por semilla.

## Requirements

### Requirement: Spawn Count and World Size

The SettlementSpawnSystem MUST create between 5 and 10 settlement entities on a 256×256 world. The exact count MUST be deterministic given the world seed.

#### Scenario: Spawn 5–10 settlements

- GIVEN a 256×256 world with valid biomes
- WHEN SettlementSpawnSystem.Update() executes
- THEN the settlement entity count MUST be between 5 and 10 inclusive

#### Scenario: Settlement count deterministic by seed

- GIVEN world seed S
- WHEN spawning settlements twice
- THEN both runs MUST produce the same number of settlement entities

### Requirement: Biome-Compatibility by Settlement Type

Each settlement type MUST spawn only in its declared compatible biomes: `village` MUST spawn in `plains` or `forest`, `town` MUST spawn only in `plains`, `city` MUST spawn only in `plains`. The spawn position MUST be a tile whose biome matches the settlement type's `biomes` list.

#### Scenario: Village spawns in plains or forest

- GIVEN a world with plains, forest, and ocean tiles
- WHEN a village settlement is spawned
- THEN its position MUST be on a plains or forest tile

#### Scenario: No settlement spawns in ocean

- GIVEN a world with ocean tiles
- WHEN settlements are spawned
- THEN no settlement SHALL be placed on an ocean tile

#### Scenario: Town spawns only in plains

- GIVEN a world with plains, forest, tundra, and desert tiles
- WHEN a town settlement is spawned
- THEN its position MUST be on a plains tile

### Requirement: Minimum Distance Between Settlements

The distance (Chebyshev/metric max) between any two settlement centers MUST be greater than 10 tiles. If no valid position is available after N attempts respecting minimum distance, the system MAY skip that settlement attempt.

#### Scenario: Settlements respect minimum distance

- GIVEN two settlements at positions (5, 5) and (16, 5)
- WHEN measuring distance
- THEN the distance SHALL be 11 (> 10), satisfying the constraint

### Requirement: Spawn Weight Distribution

Settlement type selection MUST use weighted random sampling based on each type's `spawn_weight`. Village (weight 0.6) MUST appear more frequently than town (0.3), which MUST appear more than city (0.1), statistically over multiple seeds.

#### Scenario: Villages outnumber towns and cities

- GIVEN 100 seeds generating settlements on identical terrain
- WHEN counting settlement types
- THEN villages MUST outnumber towns, and towns MUST outnumber cities on average

### Requirement: Deterministic Spawning

Spawn placement MUST be deterministic: same world seed MUST produce the same settlement positions, types, and entity components. The system MUST use world seed + 777 as its internal PRNG seed for settlement generation.

#### Scenario: Same seed produces identical settlements

- GIVEN world seed S
- WHEN spawning settlements in two separate runs
- THEN both runs MUST produce identical positions, types, and component values

#### Scenario: Different seed produces different placements

- GIVEN two world seeds S1 and S2
- WHEN spawning settlements
- THEN the set of positions and types MUST differ with high probability
