# Exploration: Settlement Interior View

## Current State

### 1. TUI Architecture (`internal/ui/`)

**Screen model**: Currently two screens — `"welcome"` and `"map"`. The `renderView()` function in `view.go` dispatches on `m.screen`. No "settlement" screen exists yet.

**Map view flow**: `renderMap()` iterates terminal rows/cols, computes world coordinates via camera offset, and renders tiles using `renderOverlay()` which prioritizes NPC > Settlement > Biome. A status bar appends below the map for settlement info when cursor hovers over one. An inspector replaces the status bar when `m.inspectorOpen` is true.

**Key fields in Model** (model.go):
- `screen string` — screen state
- `cameraX, cameraY` — world offset for map
- `cursorX, cursorY` — screen-local cursor
- `npcOverlay []npc.NPCRenderInfo` — injected per tick
- `settlementOverlay []settlement.SettlementRenderInfo` — injected per tick
- `inspectorOpen, selectedNPC, selectedSettlement` — inspector state
- `ecsWorld *ecs.World` — full ECS world reference for inspector lookups

**Tick cycle**: `tickMsg` fires every 200ms → calls `ecsWorld.Update(dt)` → then `refreshOverlay()` which re-runs NPCRenderSystem and SettlementRenderSystem and stores their output.

### 2. ECS Architecture (`internal/ecs/`)

World manages entities (uint64), component stores (generic ComponentStore[T]), and a SystemManager that runs systems in order.

**Key types**: Entity (uint64), ComponentID (uint64, generated from name strings), ComponentStore[T] (map[Entity]T), World (entities, stores, systems).

**System interface**: `Update(w *World, dt float64) error` + `Name() string`.

### 3. Settlement Components (`internal/simulation/settlement/`)

**Settlement component**:
```go
type Settlement struct {
    Name, Type, Symbol, Color string
    Radius, Population, Level int
    Buildings []string // building IDs
}
```
Settlement entities have Position + Settlement + Name (+ optionally ResourceStore).

**Building component** (ECS entity per building):
```go
type Building struct {
    ID, Name string
    Level    int
}
```
Buildings are separate entities with Position + Building. No direct link to which settlement they belong to (inference via distance from settlement center).

**HomeReference component** (links NPC → settlement):
```go
type HomeReference struct {
    SettlementEntity ecs.Entity
}
```

**BuildingDef** (data, NOT ECS component):
```go
type BuildingDef struct {
    ID, Name, Role string
    Produces, Consumes map[string]float64
    MaxWorkers int
}
```
**CRITICAL GAP**: No `Symbol` or `Color` fields on BuildingDef.

**Current YAML** (`data/buildings.yaml`):
```yaml
- id: house
  name: Casa
- id: farm
  name: Granja
  role: farmer
  produces: { food: 2.0 }
  max_workers: 3
# etc...
```
No symbol, no color.

### 4. NPC Components (`internal/simulation/npc/`)

**NPCRenderInfo** (for overlay):
```go
type NPCRenderInfo struct {
    Entity         ecs.Entity
    Symbol         rune
    Color          lipgloss.Color
    WorldX, WorldY int
}
```
No `LastReward` field.

**AIState component**:
```go
type AIState struct {
    Goals         []string
    Plan          []string
    Mood          float64
    CurrentAction string
}
```
No `LastReward` or `Reward` field.

**Appearance component**:
```go
type Appearance struct {
    Symbol rune
    Color  lipgloss.Color
}
```

**Needs component**: Hunger/Fatigue floats.

### 5. QLearning (lines 346-464 in `internal/simulation/npc/systems.go`)

The QLearningSystem per NPC per tick:
1. Determines state (need+biome)
2. Filters available actions by biome/needs
3. Selects action (GOAP fallback or ε-greedy)
4. Applies needs effects
5. Computes reward via `ComputeReward(oldNeeds, newNeeds, action)`
6. Updates Q-table with (state, action, reward, nextState)
7. Decays epsilon

**Key**: Reward IS computed but NEVER stored for UI access. The QTable receives the value but there's no component holding LastReward.

### 6. Settlement-NPC Association

**How it works**:
- NPC spawn: `findCompatibleSettlements()` in spawner.go matches role to building types present in settlement.
- Each NPC gets a `HomeReference{SettlementEntity: se}` component.
- Economy: iterates NPCs with HomeReference, counts workers per role per settlement.
- No direct WorkplaceReference component exists — worker→building is inferred via `Job.Role == BuildingDef.Role`.

### 7. Data Flow for Overlay

```
tickMsg → ecsWorld.Update(dt) → refreshOverlay()
  → NPCRenderSystem.Update(): iterates LOD≥1 NPCs, builds NPCRenderInfo[]
  → SettlementRenderSystem.Update(): iterates all settlements, builds SettlementRenderInfo[]
  → model.npcOverlay = rs.RenderInfos()
  → model.settlementOverlay = rs.RenderInfos()
```

### 8. Viewport Geometry

Settlement radii and implied viewport sizes:
| Type   | Radius | Viewport (2r+1) | Cells |
|--------|--------|------------------|-------|
| Village| 3      | 7×7              | 49    |
| Town   | 5      | 11×11            | 121   |
| City   | 8      | 17×17            | 289   |

### 9. Existing Test Patterns (from `internal/ui/`)

- Direct Model updates: `m.Update(tea.KeyMsg{...})`, assert state
- Table-driven tests for view rendering
- Golden file tests: output compared to `testdata/*.golden`, `--update-golden` flag
- ECS world setup: create world, register stores, add components, assert via `GetComponent`
- Inspector tests: set `m.inspectorOpen = true`, `m.selectedNPC`/`m.selectedSettlement`, `m.ecsWorld`, then check `renderInspector()` output

---

## Gaps Identified

| Gap | Severity | Description |
|-----|----------|-------------|
| G1  | **BLOCKING** | BuildingDef has no Symbol or Color fields — can't render buildings in interior view |
| G2  | **BLOCKING** | No screen "settlement" in Model — needs new screen type |
| G3  | **BLOCKING** | No LastReward on NPC render info or AIState — reward display impossible |
| G4  | **MEDIUM** | No WorkplaceReference component — can't show which NPC works at which building |
| G5  | **MEDIUM** | Node building-to-settlement link — buildings have no settlement entity reference |
| G6  | **LOW** | No per-tick reward popup system — floating "+0.87" text needs animation/decay |
| G7  | **LOW** | Camera/cursor system for interior view needs separate state from map view |
| G8  | **LOW** | LOD system uses player position centered on world; interior view needs its own visibility logic |

---

## Technical Recommendations

### R1: Building Rendering — Symbols and Colors

**Approach**: Add `interior_symbol` and `color` fields to both `BuildingDef` (YAML data) and `Building` (ECS component).

**Changes**:
1. `data/buildings.yaml`: Add `interior_symbol` and `color` per building
2. `internal/simulation/settlement/types.go` BuildingDef: Add `InteriorSymbol string` and `Color string`
3. `internal/simulation/settlement/components.go` Building: Add `Symbol rune` and `Color string`
4. `internal/simulation/settlement/data.go` LoadBuildingTypes: Parse the new fields
5. `internal/simulation/settlement/systems.go`: When spawning buildings, copy symbol+color from BuildingDef to Building component

**Symbol suggestions**:
| Building | Symbol | Color |
|----------|--------|-------|
| house | ⌂ | #8B7355 (brown) |
| farm | ╬ | #90EE90 (green) |
| market | § | #DAA520 (gold) |
| tavern | ♨ | #CD853F (peru) |
| temple | ϟ | #9370DB (medium purple) |
| blacksmith | ⚒ | #A0522D (sienna) |

**Effort**: Low. ~5 files, straightforward field addition.

### R2: Reward Tracking

**Approach A (recommended)**: Add `LastReward float64` field to `NPCRenderInfo` struct. In `QLearningSystem.Update()`, after computing reward (line 445), store it somewhere accessible. The cleanest way: add a `LastReward` field to `AIState` component, then expose it in `NPCRenderInfo`.

**Flow**:
1. Add `LastReward float64` to `AIState` component
2. After `ComputeReward()` in QLearningSystem.Update(), set `ai.LastReward = reward` before `aiStore.Set(e, ai)`
3. Add `LastReward float64` to `NPCRenderInfo`
4. In `NPCRenderSystem.Update()`, read AIState and populate `info.LastReward`

**Effort**: Low. ~3 files.

### R3: New "Settlement" Screen

**Approach**: Add `"settlement"` string to the screen model. When player presses `e` on a settlement tile (in map or directly triggerable), switch to settlement view.

**Key fields needed** (on Model):
- `settlementScreen struct { settlementEntity int; cameraX, cameraY int }` — or just repurpose existing fields
- `settlementNPCs []NPCRenderInfo` — NPCs within this settlement (filtered by HomeReference)
- `settlementBuildings []struct{Entity, X, Y, Symbol, Color, Name, Workers int}` — building positions
- `settlementInteriorSelected string` — "building" | "npc" for cursor mode
- Or even simpler: reuse `cameraX/cameraY` with new semantics

**View logic**: New `renderSettlementView(m Model) string` function that:
1. Computes viewport size from settlement radius (7×7, 11×11, 17×17)
2. Centers the viewport on the settlement center
3. Renders static tile background (same symbol as terrain)
4. Renders buildings as symbols at their world-relative positions
5. Renders NPCs as symbols at their world-relative positions
6. Shows floating reward text for NPCs with recent rewards
7. Appends status bar with settlement info
8. When `inspectorOpen`, replaces status bar with inspector

**Navigation**: Arrow keys move cursor within the viewport. `e` inspects building/NPC at cursor. `esc`/`q` returns to map.

**Effort**: Medium. ~2-3 new functions in view.go, new keyboard handling in model.go.

### R4: NPC-Building Assignment Display

**Approach A (simpler, recommended for MVP)**: Compute dynamically. When rendering the settlement interior view, iterate all NPCs belonging to this settlement (filtered by HomeReference), then for each building (ECS Building entity), check which NPCs have a Job.Role matching the BuildingDef.Role. This reuses existing data structures.

**Approach B (more robust)**: Add a `WorkplaceReference` ECS component containing `BuildingEntity ecs.Entity`. Have the economy or a new system assign this. More accurate but more complex.

**Recommendation**: Start with Approach A (compute-by-role). It's consistent with how the economy already works and avoids a new system.

**Effort**: Low (Approach A) / Medium (Approach B).

### R5: Reward Popup Display

**Approach**: Store floating reward text as a separate overlay list on the Model:
```go
type RewardPopup struct {
    WorldX, WorldY int
    Text          string  // e.g. "+0.87"
    TicksLeft     int     // fade out after N ticks
}
```
Add `rewardPopups []RewardPopup` to Model. In `refreshOverlay()`, after QLearning runs, query each NPC's AIState.LastReward and create popups. Decrement TicksLeft each tick, remove expired ones.

Rendering: Draw the popup text slightly above/beside the NPC symbol on the settlement view.

**Effort**: Low-Medium. New type in model.go, rendering logic in view.go.

### R6: Building-to-Settlement Parent Reference

**Approach**: Add `SettlementEntity ecs.Entity` to the `Building` component. In `SettlementSpawnSystem`, when spawning buildings, set this field.

This enables:
- Querying all buildings belonging to a settlement (filter Building store by SettlementEntity)
- Knowing which settlement a building belongs to for the interior view

**Effort**: Low. One field + one assignment in the spawn loop.

---

## Risk Areas

| Risk | Impact | Mitigation |
|------|--------|------------|
| **Collision on same tile**: Multiple entities (NPCs, buildings, settlement center) can occupy one tile | Rendering confusion | Layer: Buildings drawn first as background, NPCs as overlay. If both on same tile, NPC takes priority (consistent with current overlay priority) |
| **Reward popup visual noise**: All NPCs getting rewards could flood the screen with floating text | UI clutter | Only show rewards > 0.1; limit to 3 concurrent popups per settlement |
| **ECS world lock**: The UI reads from ECS world while systems are writing to it | Data race | The tick runs synchronously (systems then overlay refresh), but goroutine access to ECS components is not guarded. Currently single-threaded via Bubbletea so safe, but worth noting |
| **Settlement screen cursor**: The map cursor semantics (camera-relative world coords) differ from settlement cursor (viewport-local) | Confusing UX | Need separate cursor state for settlement view — easiest: reuse cursorX/cursorY but only in settlement screen context |
| **LOD visibility for interior**: NPCs inside a settlement might have LOD < Near if player is far away | NPCs invisible | Settlement interior view should override LOD — show all NPCs with HomeReference matching this settlement regardless of LOD |
| **Performance**: Iterating all NPCs + all buildings every tick for interior view | Slow simulation | Only relevant when settlement screen is active. Could cache. 100 NPCs × 10 settlements = 1000 iterations = fine for MVP |

---

## File Change Assessment

| File | Change |
|------|--------|
| `data/buildings.yaml` | Add `interior_symbol` and `color` to each building |
| `internal/simulation/settlement/types.go` | Add `Symbol, Color` to BuildingDef |
| `internal/simulation/settlement/components.go` | Add `Symbol rune, Color string, SettlementEntity ecs.Entity` to Building |
| `internal/simulation/settlement/data.go` | Parse `interior_symbol` and `color` in LoadBuildingTypes |
| `internal/simulation/settlement/systems.go` | Pass symbol+color from BuildingDef to Building component; set SettlementEntity |
| `internal/simulation/npc/components.go` | Add `LastReward float64` to AIState |
| `internal/simulation/npc/types.go` | Add `LastReward float64` to NPCRenderInfo |
| `internal/simulation/npc/systems.go` | Store `ai.LastReward = reward` after ComputeReward; expose in NPCRenderSystem |
| `internal/ui/model.go` | Add `"settlement"` screen, settlement view fields (settlementNPCs, settlementBuildings, rewardPopups), key handling for arrows/e/esc/q |
| `internal/ui/view.go` | Add `renderSettlementView()`, `renderSettlementInspector()` (detailed), building rendering, NPC+reward overlay, status bar |
| `internal/ui/view_test.go` | Add tests for settlement view rendering, building symbols, NPC display, reward popup, inspections |
| `internal/ui/model_test.go` | Add tests for screen transitions, key handling in settlement view |
| `cmd/evociv/main.go` | Potentially minimal — interior system should auto-derive from existing ECS data |

### Total file count: ~12 files

---

## Estimates

| Task | Effort | Dependencies |
|------|--------|-------------|
| T1: Add building symbols + colors to YAML, types, and data loader | Low (2-3 files) | None |
| T2: Add SettlementEntity reference to Building component | Low (2 files) | None |
| T3: Add LastReward to AIState and NPCRenderInfo; wire in Q-Learning + Render systems | Low (3 files) | None |
| T4: Add reward popup system to Model | Low-Medium (2 files) | T3 |
| T5: Add "settlement" screen to Model + keyboard handling | Medium (1 file) | None |
| T6: Implement renderSettlementView with buildings, NPCs, rewards | Medium (1 file) | T1-T5 |
| T7: Implement settlement building inspector (workers, production, assigned NPCs) | Medium (1 file) | T2, T6 |
| T8: Implement settlement NPC inspector (Q-Learning stats, job, home, workplace) | Low-Medium (1 file) | T3, T6 |
| T9: Write tests for all new functionality | Medium (2 files) | T1-T8 |

**Total estimate**: 2-3 focused sessions

---

## Ready for Proposal
**Yes** — the exploration is comprehensive. The gaps and solutions are well-understood. The key technical decisions are:
1. New `"settlement"` screen in the TUI (separate from `"map"`)
2. Building symbols/colors added to data files and ECS components
3. LastReward tracked via AIState → NPCRenderInfo
4. NPC-building assignment computed dynamically by role matching (MVP)
5. Reward popups as transient overlay objects in the Model
6. Building-to-settlement reference added to Building component
