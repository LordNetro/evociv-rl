# Design: Settlement Interior View

## Technical Approach

Add a `"settlement"` screen to the existing Bubbletea TUI. Reuse the ECS render-systems pattern to collect building render info each tick. Track LastReward through AIState → NPCRenderInfo for reward popups. Override LOD in settlement view — show all NPCs with matching HomeReference. Dynamic role-matching for worker aggregation (no new WorkplaceReference).

## Architecture Decisions

### AD1: Screen Architecture

| Option | Tradeoff | Decision |
|--------|----------|----------|
| Add `"settlement"` string + dedicated state struct | Simple switch; isolated state | **Chosen** — matches existing pattern |
| Separate sub-model | More modular but more wiring | Rejected — overkill for MVP |

### AD2: Building Render Info

| Option | Tradeoff | Decision |
|--------|----------|----------|
| New `BuildingRenderSystem` | Consistent with NPC/Settlement render pattern | **Chosen** — clean ECS separation |
| Filter on-the-fly in view.go | Less code but couples ECS with rendering | Rejected — breaks existing pattern |

### AD3: Reward Display

| Option | Tradeoff | Decision |
|--------|----------|----------|
| LastReward on AIState → NPCRenderInfo + RewardPopup in Model | Clean pipeline, popup lifecycle in Model | **Chosen** |
| Read AIState directly in view.go | Simpler but couples view to ECS | Rejected |

### AD4: NPC LOD Override

| Option | Tradeoff | Decision |
|--------|----------|----------|
| Override in NPCRenderSystem: include NPCs with HomeReference matching settlement | No global behavior change | **Chosen** — `RenderInfosForSettlement()` method |
| Change global LOD thresholds | Affects map view | Rejected |

### AD5: Worker Aggregation

| Option | Tradeoff | Decision |
|--------|----------|----------|
| Dynamic role-matching (Job.Role == BuildingDef.Role) | Consistent with economy, no new components | **Chosen** for MVP |
| WorkplaceReference component | More accurate but new system required | Rejected for now |

## Data Flow

```
tickMsg → ecsWorld.Update(dt)
  SettlementSpawnSystem → PopulationSystem
  → SettlementEconomySystem → SettlementGrowthSystem → FamineSystem
  → GOAPSystem → QLearningSystem (stores LastReward on AIState)
  → NPCRenderSystem (reads AIState → NPCRenderInfo.LastReward)
  → SettlementRenderSystem
  → BuildingRenderSystem (NEW)

refreshOverlay():
  if screen == "settlement":
    filter NPCs by HomeReference == settlementEntity
    filter Buildings by SettlementEntity == settlementEntity
    process RewardPopups (decrement, create, remove expired)
  else:
    standard overlay (NPC + settlement)

renderSettlementView():
  for each (x,y) in viewport:
    if building at (x,y) → render building symbol+color
    if NPC at (x,y) → render NPC symbol+color
    if rewardPopup at (x,y) → render "+N.NN"
    else → render biome tile
  if inspectorOpen → render inspector; else → render status bar
```

## Type Changes

### BuildingDef (types.go — modified)
```go
type BuildingDef struct {
    ID, Name, Role string
    InteriorSymbol string `yaml:"interior_symbol"` // NEW
    Color          string `yaml:"color"`           // NEW
    Produces, Consumes map[string]float64
    MaxWorkers int
}
```

### Building (components.go — modified)
```go
type Building struct {
    ID, Name         string
    Level            int
    Symbol           rune       // NEW — copied from BuildingDef on spawn
    Color            string     // NEW — copied from BuildingDef on spawn
    SettlementEntity ecs.Entity // NEW — set during spawn
}
```

### AIState (components.go — modified)
```go
type AIState struct {
    Goals         []string
    Plan          []string
    Mood          float64
    CurrentAction string
    LastReward    float64 // NEW — set after ComputeReward() in QLearningSystem
}
```

### NPCRenderInfo (types.go — modified)
```go
type NPCRenderInfo struct {
    Entity         ecs.Entity
    Symbol         rune
    Color          lipgloss.Color
    WorldX, WorldY int
    LastReward     float64  // NEW — populated from AIState.LastReward
}
```

### New: BuildingRenderInfo (types.go — new)
```go
type BuildingRenderInfo struct {
    Entity         ecs.Entity
    Symbol         rune
    Color          string
    Name           string
    ID             string
    Level          int
    WorldX, WorldY int
    SettlementEntity ecs.Entity
    Role           string
    MaxWorkers     int
}
```

### New: RewardPopup (ui/model.go — new)
```go
type RewardPopup struct {
    WorldX, WorldY int
    Text           string // e.g. "+0.87"
    TicksLeft      int    // starts at 5, decrements each tick
}
```

### New: SettlementViewState (ui/model.go — new)
```go
type SettlementViewState struct {
    SettlementEntity   ecs.Entity
    SettlementCenterX  int
    SettlementCenterY  int
    CursorX, CursorY   int       // viewport-local coordinates
    ViewportRadius     int       // 3, 5, or 8
}
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `data/buildings.yaml` | Modify | Add `interior_symbol` and `color` to each building entry |
| `internal/simulation/settlement/types.go` | Modify | Add `InteriorSymbol`, `Color` to BuildingDef; add `LastReward` to NPCRenderInfo; new `BuildingRenderInfo` struct |
| `internal/simulation/settlement/components.go` | Modify | Add `Symbol rune`, `Color string`, `SettlementEntity ecs.Entity` to Building |
| `internal/simulation/settlement/data.go` | Modify | Parse `interior_symbol` and `color` in `LoadBuildingTypes()` |
| `internal/simulation/settlement/systems.go` | Modify | Copy Symbol+Color+SettlementEntity when spawning buildings; new `BuildingRenderSystem` |
| `internal/simulation/npc/components.go` | Modify | Add `LastReward float64` to AIState |
| `internal/simulation/npc/types.go` | Modify | Add `LastReward float64` to NPCRenderInfo |
| `internal/simulation/npc/systems.go` | Modify | Store `ai.LastReward` after `ComputeReward()`; populate in NPCRenderSystem; add `RenderInfosForSettlement()` method |
| `internal/ui/model.go` | Modify | Add `"settlement"` screen, `settlementViewState`, `settlementNPCs`, `settlementBuildings`, `rewardPopups`; keyboard handling for settlement nav |
| `internal/ui/view.go` | Modify | Add `renderSettlementView()`, building inspector, NPC inspector with Q-stats, reward popup rendering |
| `internal/ui/model_test.go` | Modify | Table-driven tests for settlement screen transitions, keyboard handling |
| `internal/ui/view_test.go` | Modify | Tests for settlement view rendering, building symbols, reward popups, inspectors |
| `internal/simulation/settlement/systems_test.go` | Modify | Tests for BuildingRenderSystem |
| `internal/simulation/npc/systems_test.go` | Modify | Tests for LastReward flow |

## Interfaces / Contracts

### New methods

```go
// BuildingRenderSystem (settlement/systems.go)
type BuildingRenderSystem struct {
    renderInfos []BuildingRenderInfo
    buildingDefs map[string]BuildingDef // lookup for Role, MaxWorkers
}
func NewBuildingRenderSystem(defs []BuildingDef) *BuildingRenderSystem
func (s *BuildingRenderSystem) Name() string
func (s *BuildingRenderSystem) Update(w *ecs.World, dt float64) error
func (s *BuildingRenderSystem) RenderInfos() []BuildingRenderInfo

// NPCRenderSystem extension (npc/systems.go)
// New method to return NPCs filtered by settlement
func (s *NPCRenderSystem) RenderInfosForSettlement(w *ecs.World, settlementEntity ecs.Entity) []NPCRenderInfo

// Model extensions (ui/model.go)
func (m *Model) openSettlementView(entity ecs.Entity, cx, cy, radius int)
func (m *Model) closeSettlementView()
func (m *Model) processRewardPopups() // create/decrement/expire
```

### New view functions (ui/view.go)

```go
func renderSettlementView(m Model) string
func renderBuildingInspector(m Model) string
func renderSettlementNPCInspector(m Model) string
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit (ECS) | BuildingRenderSystem collects correct entities | Create ECS world, add building entities + Position + Building, assert BuildingRenderSystem output |
| Unit (ECS) | LastReward flows from QLearningSystem → AIState → NPCRenderInfo | Wire up real Q-learning, verify stored reward appears in render info |
| Unit (UI) | Settlement screen transitions (`e` opens view, `esc` closes) | Direct `Model.Update()` calls, assert `m.screen == "settlement"` |
| Unit (UI) | Building inspector shows name, level, role, workers | Setup settlement + building + NPC data, assert inspector output contains fields |
| Unit (UI) | Reward popups appear and expire | Inject NPCRenderInfo with LastReward, assert RewardPopup created and decremented |
| Golden | Settlement view rendering with buildings + NPCs | Output compared to `testdata/settlement-*.golden`, `--update-golden` flag |
| Table-driven | All keyboard navigation within settlement view | `[]struct{name string; keys []tea.Msg; expected func(Model) bool}` pattern |

## Migration

**No migration required.** Existing buildings (no Symbol/Color/SettlementEntity) render silently:

- **Building.Symbol**: zero value `0` → `BuildingRenderSystem` skips entities with Symbol == 0 → building not drawn in interior view (acceptable — buildings exist but aren't visible until next spawn)
- **AIState.LastReward**: zero value `0.0` → reward popups show `+0.00` → use threshold check (`abs(reward) > 0.01` in NPCRenderSystem) to skip display
- **Building.SettlementEntity**: zero value `0` → filtered out by `RenderInfosForSettlement()` → only affects rendering in settlement view, not simulation

All new fields have sensible zero defaults. Newly spawned buildings (after code change) will have all fields populated correctly.
