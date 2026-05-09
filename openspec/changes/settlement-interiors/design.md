# Design: Settlement Interiors

## Technical Approach

New `BuildingInterior` ECS component (runtime state) + deterministic `InteriorGenerator` (seeded room placement) + `IndoorPathfinder` (A* on grid) + GOAP actions for enter/exit. Production dynamically counts `WorkersInside` per building.

## Architecture Decisions

| Decision | Choice | Alternatives | Rationale |
|----------|--------|-------------|-----------|
| Interior generation | Seeded room placement | BSP tree (rejected) | Deterministic, simple to test; BSP overkill for 1–3 rooms |
| ECS component | New `BuildingInterior` (not on `Building`) | Extend `Building` struct (rejected) | `Building` = spawn-time config; `BuildingInterior` = runtime state. Separate = backward-compatible |
| Grid resolution | 1 cell = 1 world tile | Sub-tile (rejected) | Trivial coordinate transform: interior pos == world pos |
| Pathfinding | A* + cache | NavMesh (rejected), door-to-door (rejected) | <100 cells, <1ms. Cache invalidates on layout change |
| GOAP integration | New `ActionDef`s only | Separate system (rejected) | GOAPSystem (planning) + QLearningSystem (execution) pipeline already exists |

## Data Flow

```
SettlementSpawnSystem
  └─ spawn building entity
        └─ InteriorGenerator.Generate(seed, buildingID, w, h)
              └─ returns InteriorGrid + Doors[]

ECS World:
  Building (spawn config: InteriorSymbol, Color, SettlementEntity)
  BuildingInterior (runtime: grid, workers, doors, maxWorkers)

ProductionSystem (per tick):
  for each building with Produces:
    workers = min(len(WorkersInside), MaxWorkers)
    output = base × workers × efficiency

GOAP → QLearning → Action: enter_building|work_inside|exit_building
  └─ updates BuildingInterior.WorkersInside
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/simulation/settlement/interior.go` | **Create** | `BuildingInterior`, `CellType`, `DoorPosition`, `BuildingInteriorID` |
| `internal/simulation/settlement/interior_generator.go` | **Create** | Seeded room placement → grid + doors |
| `internal/simulation/settlement/interior_pathfinder.go` | **Create** | A* on grid, cache, world↔interior coordinate math |
| `internal/simulation/settlement/systems.go` | Modify | SpawnSystem calls generator, attaches `BuildingInterior` |
| `internal/simulation/settlement/components.go` | Modify | `RegisterSettlementStores` registers `BuildingInteriorID` |
| `internal/simulation/npc/data.go` | Modify | Add `enter_building`, `work_inside`, `exit_building` action defs |
| `internal/simulation/economy/systems.go` | Modify | `ProductionSystem`: `workers × base × efficiency` |
| `internal/ui/view.go` | Modify | Settlement view renders grid + workers inside |

## Interfaces

```go
// interior.go
type CellType int // CellFloor, CellWall, CellDoor, CellCorridor

type DoorPosition struct{ InteriorX, InteriorY int }

type BuildingInterior struct {
    Grid           [][]CellType
    Width, Height  int
    Doors          []DoorPosition
    WorkersInside []ecs.Entity
    MaxWorkers    int
    Seed           int64
}

// interior_generator.go
type InteriorGenerator struct{}
func (g *InteriorGenerator) Generate(seed int64, buildingID string, w, h int) BuildingInterior

// interior_pathfinder.go
type IndoorPathfinder struct{}
func (pf *IndoorPathfinder) FindPath(grid [][]CellType, from, to DoorPosition) ([]DoorPosition, error)
func (pf *IndoorPathfinder) WorldEntryPos(buildingPos ecs.Position, door DoorPosition) ecs.Position
```

## Testing Strategy

| Layer | What | How |
|-------|------|-----|
| Unit | InteriorGenerator determinism | `t.Run()`: same seed → same grid; diff seed → different |
| Unit | A* pathfinding | `t.Run()`: door→room, door→door, isolated→error |
| Unit | Production formula | `t.Run()`: 0w→0, 1w→base, 2w→2×base, cap at max |
| Unit | WorkersInside tracking | `t.Run()`: enter empty, reject when full, exit |
| Integration | SpawnSystem produces interior | Existing test + interior assertions |
| Integration | GOAP enter/exit transitions | QLearningSystem test |

Follow `go-testing`: table-driven `t.Run()`, test behavior not implementation.

## Migration / Rollout

No migration. `BuildingInterior` is optional — buildings without it produce 0 (graceful default). Feature-flag-free.

## Open Questions

- [ ] **TUI zoom**: 1:1 with world tiles or zoomed in? Current settlement view uses world scale.
- [ ] **Efficiency**: Global per-building multiplier (recommend) or per-worker stat?
- [ ] **WorkInside duration**: Every tick via NeedsDecay (recommend) or explicit tick count?
