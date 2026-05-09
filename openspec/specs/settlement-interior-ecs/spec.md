# settlement-interior-ecs Specification

## Purpose

ECS component extensions needed for the settlement interior view. Adds rewards tracking to AIState, extends NPCRenderInfo with reward and job data, and stores interior rendering fields on the Building component.

## Requirements

### Requirement: AIState Gains LastReward and RewardTick

The `AIState` struct in `internal/simulation/npc/components.go` MUST add two new fields: `LastReward` (float64) holding the most recent Q-learning reward value, and `RewardTick` (int) holding the simulation tick when LastReward was set. Both MUST be zero-valued by default.

#### Scenario: AIState has LastReward field

- GIVEN an NPC entity with AIState component
- WHEN the component is first created
- THEN `LastReward` MUST be 0.0 and `RewardTick` MUST be 0

#### Scenario: LastReward updates after Q-learning

- GIVEN an NPC with AIState where LastReward = 0.0
- WHEN QLearningSystem.Update() computes reward = 0.87 for this NPC
- THEN AIState.LastReward MUST be set to 0.87 and RewardTick MUST be set to the current sim tick

### Requirement: QLearningSystem Writes LastReward

The `QLearningSystem.Update()` method, after calling `ComputeReward()` and before updating the Q-table, MUST write the computed reward value into the NPC's `AIState.LastReward` field and the current tick into `AIState.RewardTick`.

#### Scenario: LastReward written after ComputeReward

- GIVEN a QLearningSystem updating an NPC with AIState.LastReward = 0.0
- WHEN ComputeReward returns 0.87
- THEN AIState.LastReward MUST equal 0.87 and AIState.RewardTick MUST equal the simulation tick

#### Scenario: Reward threshold filter

- GIVEN ComputeReward returns 0.05
- WHEN QLearningSystem writes LastReward
- THEN the system MAY skip writing if abs(reward) < 0.01 to avoid noise

### Requirement: Building Component Gains Interior Fields

The `Building` struct in `internal/simulation/settlement/components.go` MUST add `InteriorSymbol` (string) and `Color` (string) fields, and a `SettlementEntity` (ecs.Entity) field linking the building to its parent settlement entity.

#### Scenario: Building references parent settlement

- GIVEN a building entity spawned for settlement entity ID 7
- WHEN the Building component is created
- THEN `SettlementEntity` MUST be set to 7

#### Scenario: Building without parent settlement

- GIVEN a building entity with no settlement parent
- WHEN the Building component is retrieved
- THEN `SettlementEntity` MUST be 0 (EntityInvalid)

### Requirement: NPCRenderInfo Gains Reward and Role Fields

The `NPCRenderInfo` struct in `internal/simulation/npc/types.go` MUST add `LastReward` (float64), `RewardTick` (int), and `JobRole` (string) fields so the TUI can display reward popups and NPC role info in the interior view.

#### Scenario: NPCRenderInfo carries last reward

- GIVEN an NPC with AIState.LastReward = 0.87 and AIState.RewardTick = 42
- WHEN NPCRenderSystem.Update() gathers render info
- THEN the NPCRenderInfo for that NPC MUST have LastReward = 0.87, RewardTick = 42

#### Scenario: NPCRenderInfo carries job role

- GIVEN an NPC with Job{Role: "farmer"}
- WHEN NPCRenderSystem.Update() gathers render info
- THEN the NPCRenderInfo for that NPC MUST have JobRole = "farmer"

#### Scenario: NPC without AIState has zero reward

- GIVEN an NPC without AIState component
- WHEN NPCRenderSystem.Update() gathers render info
- THEN the resulting NPCRenderInfo MUST have LastReward = 0.0 and RewardTick = 0

### Requirement: NPCRenderSystem Includes Enhanced Info

The `NPCRenderSystem.Update()` method MUST read AIState's LastReward and RewardTick, and Job's Role for each NPC, and populate the corresponding fields in NPCRenderInfo.

#### Scenario: Render system populates new fields

- GIVEN an NPC with Job{Role: "smith"}, AIState{LastReward: 1.2, RewardTick: 50}
- WHEN NPCRenderSystem.Update() runs
- THEN the NPCRenderInfo for entity e MUST contain JobRole = "smith", LastReward = 1.2, RewardTick = 50

### Validation Rules

- PASS: AIState.LastReward persists through tick simulation
- PASS: QLearningSystem writes LastReward after ComputeReward
- PASS: NPCRenderInfo carries correct reward, tick, and role for each NPC
- PASS: Building.SettlementEntity references correct parent
- PASS: Zero values for unset LastReward and RewardTick
- FAIL: LastReward persists across unrelated NPCs (no cross-contamination)
