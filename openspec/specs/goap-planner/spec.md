# goap-planner Specification

## Purpose

Planificador GOAP forward-chaining que evalúa necesidades NPC, selecciona la acción con mayor reward potencial, y genera secuencias {move_to_biome → execute_action}. Soporta LOD escalonado.

## Requirements

### Requirement: Need Evaluation

The planner MUST evaluate all needs (Hunger, Fatigue) and select the most urgent one, defined as the need with the highest current value.

#### Scenario: High hunger prioritizes harvest/forage

- GIVEN an NPC with Hunger=0.8 and Fatigue=0.1
- WHEN the planner evaluates needs
- THEN the selected need MUST be Hunger
- AND the recommended action type MUST be harvest or forage

#### Scenario: High fatigue prioritizes rest

- GIVEN an NPC with Fatigue=0.8 and Hunger=0.1
- WHEN the planner evaluates needs
- THEN the selected need MUST be Fatigue
- AND the recommended action MUST be rest

#### Scenario: Low needs allow exploration

- GIVEN an NPC with Hunger=0.1 and Fatigue=0.1
- WHEN the planner evaluates needs
- THEN the selected action MAY be explore or socialize

### Requirement: Plan Generation

The planner MUST generate a plan of depth ≤ 3 as a sequence: {move_to_biome → execute_action}. The plan MUST include the target biome if the action requires one.

#### Scenario: Plan includes biome movement

- GIVEN action harvest requires biome "plains"
- WHEN generating a plan for an NPC on a "forest" tile
- THEN the plan MUST include a move_to_biome step targeting "plains" followed by execute_action "harvest"

### Requirement: Replanning

The planner MUST replan when: (a) the highest need changes, (b) an action execution fails, or (c) the environment (NPC position / biome) changes.

#### Scenario: Replan on need change

- GIVEN an NPC executing a plan for Fatigue (rest)
- WHEN Hunger rises above Fatigue mid-plan
- THEN the planner MUST invalidate the current plan and generate a new one for Hunger

### Requirement: LOD Scaling

The planner MUST execute full GOAP (depth ≤ 3) at LOD 2 (local), simplified (1 action ahead) at LOD 1 (near), and MUST NOT plan at LOD 0 (distant — wander only).

#### Scenario: Full planning local

- GIVEN an NPC with LOD=2
- WHEN the planner runs
- THEN it MUST evaluate all needs and generate a plan up to depth 3

#### Scenario: Simplified planning near

- GIVEN an NPC with LOD=1
- WHEN the planner runs
- THEN it MUST select the best action but MUST NOT generate multi-step plans

#### Scenario: No planning distant

- GIVEN an NPC with LOD=0
- WHEN the planner runs
- THEN it MUST NOT generate any plan
