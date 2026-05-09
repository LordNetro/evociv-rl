# npc-systems Specification

## Purpose

Sistemas ECS que simulan NPCs cada tick: spawn inicial, movimiento aleatorio (wander), nivel de detalle (LOD), y renderizado en TUI.

## Requirements

### Requirement: Four Systems Registered

The World MUST register and execute exactly four NPC systems per tick: NPCSpawnSystem, WanderSystem, LODSystem, and NPCRenderSystem.

#### Scenario: All systems execute on Update

- GIVEN a World with all four NPC systems registered
- WHEN World.Update() is called
- THEN every system MUST execute exactly once

### Requirement: NPCSpawnSystem

NPCSpawnSystem MUST run once on first tick, spawning 50–100 NPCs via the spawner. Subsequent ticks MUST be no-ops.

#### Scenario: Spawn runs only once

- GIVEN a World after first Update tick
- WHEN World.Update() is called a second time
- THEN no new NPC entities MUST be created

### Requirement: WanderSystem

NPCs with a Job component MUST move to an adjacent tile (8-direction) within a biome compatible with their role. Movement MUST be random but bounded to world edges.

#### Scenario: Wander within world bounds

- GIVEN an NPC at position (x, y) with Job "farmer" on a plains tile
- WHEN WanderSystem processes this NPC
- THEN the NPC MUST move to an adjacent tile (x±1, y±1) that is also plains or forest
- AND the new position MUST be within [0, 0] to [255, 255]

#### Scenario: NPC stays if no compatible neighbor

- GIVEN an NPC surrounded entirely by ocean tiles
- WHEN WanderSystem processes this NPC
- THEN the NPC MUST remain at its current position

### Requirement: LODSystem

LODSystem MUST assign LOD level per NPC based on Chebyshev distance to the player avatar: distance ≤ 5 → level 2, distance ≤ 15 → level 1, distance > 15 → level 0.

#### Scenario: LOD changes as player moves

- GIVEN an NPC at distance 3 from the player
- WHEN LODSystem runs
- THEN the NPC MUST have LOD level 2
- WHEN the player moves to distance 10
- THEN the LOD level MUST drop to 1

#### Scenario: Far NPCs are LOD 0

- GIVEN an NPC at distance 20 from the player
- WHEN LODSystem runs
- THEN the NPC MUST have LOD level 0

### Requirement: NPCRenderSystem

NPCRenderSystem MUST render NPCs with LOD level ≥ 1 as their assigned symbol (default '@') on the TUI map overlay, positioned at their world coordinates offset by camera.

#### Scenario: Render visible NPCs

- GIVEN an NPC at world position (10, 15) with LOD level 2 and symbol '@'
- WHEN NPCRenderSystem executes with camera at (5, 5)
- THEN the TUI MUST display '@' at screen position (5, 10)

#### Scenario: Skip LOD 0 NPCs

- GIVEN an NPC with LOD level 0
- WHEN NPCRenderSystem executes
- THEN the NPC MUST NOT be rendered
