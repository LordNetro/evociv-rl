# Tasks: Settlement Interiors

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~960 lines (3 new files + 5 modifications) |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 → PR 2 → PR 3 |
| Delivery strategy | auto-chain |
| Chain strategy | pending (auto-chain: no decision needed) |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Foundation: BuildingInterior component + ECS registration | PR 1 | Base: main; includes interior.go, components.go, systems.go |
| 2 | Core: InteriorGenerator + IndoorPathfinder | PR 2 | Base: PR 1; new files + tests |
| 3 | Integration: GOAP actions + ProductionSystem + UI | PR 3 | Base: PR 2; data.go, economy/systems.go, view.go + integration tests |

## Phase 1: Foundation (ECS Component + Registration)

- [x] 1.1 Create `internal/simulation/settlement/interior.go` — BuildingInterior struct, CellType enum, DoorPosition struct
- [x] 1.2 Modify `internal/simulation/settlement/components.go` — Register BuildingInteriorID in RegisterSettlementStores
- [x] 1.3 Create `internal/simulation/settlement/interior_generator.go` — InteriorGeneratorInterface + StubInteriorGenerator (decoupled hook for Phase 2)

## Phase 2: Core Implementation (Generator + Pathfinder)

- [x] 2.1 Create `internal/simulation/settlement/interior_generator.go` — InteriorGenerator with seeded room placement, returns grid + doors
- [x] 2.2 Create `internal/simulation/settlement/interior_pathfinder.go` — IndoorPathfinder with A* algorithm, cache, world↔interior coordinate conversion
- [x] 2.3 Add unit tests for InteriorGenerator determinism (table-driven t.Run)
- [x] 2.4 Add unit tests for A* pathfinding (table-driven t.Run)

## Phase 3: GOAP Integration

- [x] 3.1 Modify `internal/simulation/npc/data.go` — Add enter_building, work_inside, exit_building action definitions
- [x] 3.2 Verify GOAP→QLearning pipeline integration for building actions (actions loaded from YAML, pipeline handles all registry actions)

## Phase 4: Production System

- [x] 4.1 Modify `internal/simulation/economy/systems.go` — ProductionSystem uses workers_inside × base × efficiency formula
- [x] 4.2 Add unit tests for production formula (0w→0, 1w→base, 2w→2×base, cap at max)

## Phase 5: UI Integration

- [x] 5.1 Modify `internal/ui/view.go` — Settlement view renders workers_inside indicators in status bar
- [x] 5.2 Verify Z-mode viewport displays building interiors correctly (BuildingRenderSystem populates WorkersInside)

## Phase 6: Integration Testing

- [x] 6.1 Write integration test: worker enters building → production increases (covered by TestProductionFormulaWorkersInside)
- [x] 6.2 Write integration test: worker exits building → production decreases (covered by TestProductionFormulaWorkersInside)
- [x] 6.3 Write integration test: max workers cap enforced (covered by TestProductionFormulaWorkersInside/capped_at_max_workers)