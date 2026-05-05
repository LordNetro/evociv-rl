# npc-spawner Specification

## Purpose

Generar poblaciones NPC sobre el mundo 256×256 con distribución ponderada por bioma y determinismo por semilla.

## Requirements

### Requirement: Spawn Count and World Size

The spawner MUST create between 50 and 100 NPC entities on a 256×256 world. The exact count MUST be deterministic given the world seed.

#### Scenario: Spawn 50–100 NPCs

- GIVEN a 256×256 world with valid biomes
- WHEN the spawner executes
- THEN the entity count MUST be between 50 and 100 inclusive

### Requirement: Biome-Weighted Placement

Spawn probability MUST be weighted per biome: plains and forest weight 1.0, tundra and desert weight 0.2, ocean and jungle weight 0.0. NPCs MUST NOT spawn in biomes with weight 0.0.

| Biome | Weight | Behavior |
|-------|--------|----------|
| plains, forest | 1.0 | Full probability |
| tundra, desert | 0.2 | Reduced probability |
| ocean, jungle | 0.0 | MUST NOT spawn |

#### Scenario: No NPCs spawn in ocean or jungle

- GIVEN a world tile classified as ocean or jungle
- WHEN the spawner places NPCs
- THEN no NPC SHALL be placed on that tile

#### Scenario: Plains receive more NPCs than tundra statistically

- GIVEN a world with equal plains and tundra tiles
- WHEN spawning 100 NPCs across multiple seeds
- THEN plains MUST accumulate more NPCs than tundra on average

### Requirement: Deterministic Spawning

Spawn placement MUST be deterministic: given the same world seed, the same set of NPC entities (positions, components) MUST be produced. The spawner MUST use world seed + 999 as its internal PRNG seed.

#### Scenario: Same seed produces identical NPCs

- GIVEN a world seed S
- WHEN spawning NPCs twice in separate runs
- THEN both runs MUST produce NPCs at identical positions with identical components

#### Scenario: Different seed produces different placements

- GIVEN two different world seeds S1 and S2
- WHEN spawning NPCs
- THEN the set of positions MUST differ with high probability

### Requirement: YAML Data Definitions

The file `data/npcs.yaml` MUST define races, roles, and trait distributions. Races MUST specify allowed roles. Roles MUST specify allowed biomes and base appearance.

#### Scenario: Load NPC definitions from YAML

- GIVEN a valid `data/npcs.yaml` with at least 2 races and 3 roles
- WHEN loaded via the data loader
- THEN all races, roles, and their relationships MUST be accessible

#### Scenario: Race-role compatibility enforced

- GIVEN a race that does not list "miner" as an allowed role
- WHEN spawning an NPC with that race and role "miner"
- THEN the spawner MUST reject the combination
