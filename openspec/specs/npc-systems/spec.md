# npc-systems Specification

## Purpose

Sistemas ECS que simulan NPCs cada tick: spawn inicial, movimiento aleatorio (wander), nivel de detalle (LOD), y renderizado en TUI.

## Requirements

### Requirement: Seven Systems Registered

The World MUST register and execute exactly seven NPC systems per tick: NPCSpawnSystem, NeedsDecaySystem, LODSystem, GOAPSystem, QLearningSystem, WanderSystem, and NPCRenderSystem. The three new GOAP-RL systems (NeedsDecaySystem, GOAPSystem, QLearningSystem) MUST be inserted after NPCSpawnSystem and before WanderSystem.

#### Scenario: All systems execute on Update

- GIVEN a World with all seven NPC systems registered
- WHEN World.Update() is called
- THEN every system MUST execute exactly once

### Requirement: NPCSpawnSystem

NPCSpawnSystem MUST run once on first tick, spawning 50–100 NPCs via the spawner. Subsequent ticks MUST be no-ops. Runs before all other NPC systems.

#### Scenario: Spawn runs only once

- GIVEN a World after first Update tick
- WHEN World.Update() is called a second time
- THEN no new NPC entities MUST be created

### Requirement: NeedsDecaySystem

NeedsDecaySystem MUST run every tick at ALL LOD levels, increasing NPC Hunger (+0.01/tick) and Fatigue (+0.005/tick), clamped to [0, 1]. At LOD 0 (distant), the decay rate MUST be multiplied by 0.5×. Runs after NPCSpawnSystem and before LODSystem.

### Requirement: GOAPSystem

GOAPSystem MUST run every tick for NPCs with LOD ≥ 1, performing GOAP forward-chaining planning. At LOD 2 (local), full plans up to depth 3 are generated. At LOD 1 (near), only a single action is selected without multi-step plans. LOD 0 NPCs are skipped entirely. Runs after LODSystem and before QLearningSystem.

### Requirement: QLearningSystem

QLearningSystem MUST run every tick for NPCs with LOD ≥ 1, applying ε-greedy Q-learning to optimize action selection. It uses GOAP as fallback when all Q-values are below threshold (0.01). LOD 0 NPCs are skipped. Runs after GOAPSystem and before WanderSystem.

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
