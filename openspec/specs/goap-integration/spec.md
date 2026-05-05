# goap-integration Specification

## Purpose

Integración de los tres nuevos sistemas ECS (NeedsDecaySystem, GOAPSystem, QLearningSystem) en el pipeline de simulación, respetando orden de ejecución y LOD.

## Requirements

### Requirement: Three New Systems

The system MUST register and execute exactly three new ECS systems: NeedsDecaySystem, GOAPSystem, and QLearningSystem. Execution order MUST be: NeedsDecaySystem → GOAPSystem → QLearningSystem.

#### Scenario: Systems build without errors

- GIVEN the project source code
- WHEN `go build ./...` is run
- THEN no compilation errors MUST be reported

#### Scenario: Correct execution order

- GIVEN a World with all three systems registered in order
- WHEN World.Update() is called
- THEN NeedsDecaySystem MUST execute first (updates needs)
- AND GOAPSystem MUST execute second (generates plans using updated needs)
- AND QLearningSystem MUST execute third (learns from executed actions)

### Requirement: System Registration in main.go

The `cmd/evociv/main.go` file MUST register NeedsDecaySystem, GOAPSystem, and QLearningSystem in that exact order, after the existing NPC systems (NPCSpawnSystem, WanderSystem, LODSystem).

#### Scenario: main.go registers systems

- GIVEN the main.go file
- WHEN inspected
- THEN it MUST contain calls to register NeedsDecaySystem, GOAPSystem, and QLearningSystem in the specified order
- AND the registration MUST occur after WanderSystem and before NPCRenderSystem

### Requirement: LOD Respect

All three systems MUST respect the NPC's LOD level:

| System | LOD 2 (local) | LOD 1 (near) | LOD 0 (distant) |
|--------|---------------|--------------|-----------------|
| NeedsDecaySystem | Full decay | Full decay | 0.5× decay |
| GOAPSystem | Full planning | Simplified (1 action) | No planning |
| QLearningSystem | Full learning | Full learning | No learning |

#### Scenario: LOD distant skips GOAP and QL

- GIVEN an NPC with LOD=0
- WHEN World.Update() executes
- THEN NeedsDecaySystem MUST apply reduced decay (0.5×)
- AND GOAPSystem MUST NOT generate a plan
- AND QLearningSystem MUST NOT update Q-values
