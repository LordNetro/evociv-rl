# Delta for npc-spawner

## ADDED Requirements

### Requirement: Settlement-Aware NPC Placement

NPCs MUST spawn within the radius of a compatible settlement instead of biome-weighted random positions. The spawner MUST select a settlement for each NPC based on role-to-building matching, then place the NPC at a random position within that settlement's radius. If no compatible settlement exists, the NPC MUST fall back to nomad spawning.

#### Scenario: Farmer spawns in village with farm

- GIVEN a village settlement with a farm building and a farmer NPC to spawn
- WHEN the spawner assigns the NPC
- THEN the NPC MUST be placed within the village's radius
- AND the NPC MUST have a HomeReference pointing to the village entity

#### Scenario: Merchant spawns in town

- GIVEN a town settlement with a market building and a merchant NPC to spawn
- WHEN the spawner assigns the NPC
- THEN the NPC MUST be placed within the town's radius
- AND the NPC MUST have a HomeReference pointing to the town entity

### Requirement: Role-to-Building Matching

Each NPC role MUST map to one or more compatible building types. The spawner MUST prefer settlements containing a building that matches the NPC's role. If the preferred building type is absent, the spawner MAY assign the NPC to any settlement of a compatible type.

| NPC Role | Preferred Building | Compatible Settlement Types |
|----------|-------------------|---------------------------|
| farmer | farm | village, town, city |
| merchant | market | town, city |
| hunter | none (any) | village, town |
| trader | market | town, city |
| miner | none (any) | village, town, city |
| priest | temple | town, city |
| blacksmith | blacksmith | town, city |

#### Scenario: Hunter assigned to nearest village

- GIVEN a hunter NPC and a village settlement (no preferred building needed)
- WHEN the spawner assigns the NPC
- THEN the NPC MUST be placed within the village's radius with HomeReference

### Requirement: Nomad Fallback

If no compatible settlement exists for an NPC (e.g., zero settlements, or all settlements are too far/capacity-exceeded), the NPC MUST spawn at a random biome-weighted position (existing behavior) and MUST NOT have a HomeReference component.

#### Scenario: Nomad NPC without settlement

- GIVEN a world with no settlements
- WHEN spawning an NPC
- THEN the NPC MUST spawn at a biome-weighted random position
- AND the NPC MUST NOT have a HomeReference component (zero value)

### Requirement: Settlement Capacity

Each settlement SHOULD cap its NPC population at `Radius * 2`. If a settlement reaches capacity, the spawner MUST attempt the next compatible settlement before falling back to nomad spawning.

#### Scenario: Settlement capacity overflow becomes nomad

- GIVEN a settlement with Radius 5 (capacity 10) already holding 10 NPCs
- WHEN an 11th compatible NPC attempts to spawn
- THEN the NPC MUST be assigned to a different settlement or become a nomad

## MODIFIED Requirements

### Requirement: Biome-Weighted Placement

Spawn probability weighted per biome NOW applies ONLY to the nomad fallback path. NPCs assigned to a settlement ignore biome weights entirely — they spawn at any tile within the settlement's radius regardless of underlying biome. The original biome weight table (plains/forest 1.0, tundra/desert 0.2, ocean/jungle 0.0) MUST still apply for nomad spawning.
(Previously: biome-weighted placement applied to ALL NPC spawns)

#### Scenario: NPC inside settlement ignores biome weight

- GIVEN a settlement partially overlapping a tundra tile (weight 0.2)
- WHEN an NPC spawns within that settlement's radius on the tundra tile
- THEN the NPC MUST be placed successfully despite the low biome weight

#### Scenario: Nomad still respects biome weights

- GIVEN a nomad NPC (no settlement)
- WHEN the nomad spawns on a tundra tile
- THEN the NPC MUST respect the 0.2 weight (reduced probability compared to plains)

### Requirement: Deterministic Spawning

Spawn placement MUST be deterministic: same world seed MUST produce the same set of NPC entities with the same settlement assignments, HomeReference values, and positions. The spawner MUST continue using world seed + 999 as its primary PRNG seed, with settlement assignment seeded from world seed + 999 + settlement_index.
(Previously: deterministic for positions and components only — no settlement assignment)

#### Scenario: Same seed produces identical settlement assignment

- GIVEN world seed S
- WHEN spawning NPCs twice
- THEN both runs MUST produce NPCs at identical positions AND with identical HomeReference entity IDs

