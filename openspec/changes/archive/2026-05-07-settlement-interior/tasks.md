# Tasks: settlement-interior

## Phase 1: Data Layer — Building Symbols + Colors

- [x] 1.1 **RED**: tests for `LoadBuildingTypes` parsing `interior_symbol` and `color`; missing fields fallback to zero values; backward compat with legacy entries
- [x] 1.2 **GREEN**: add `InteriorSymbol` + `Color` to `BuildingDef` in `types.go`; extend `LoadBuildingTypes` in `data.go` to parse both fields; update `data/buildings.yaml` with symbols and colors for all 6 buildings

## Phase 2: ECS — AIState Rewards + Building + NPCRenderInfo Enhancements

- [x] 2.1 **RED**: tests for AIState.LastReward/RewardTick zero defaults; QLearningSystem writes reward after ComputeReward; NPCRenderInfo carries LastReward/RewardTick/JobRole; Building gets Symbol/Color/SettlementEntity
- [x] 2.2 **GREEN**: add `LastReward` + `RewardTick` to `AIState` in `npc/components.go`; write reward in `QLearningSystem.Update` after `ComputeReward()`
- [x] 2.3 **GREEN**: add `LastReward` + `RewardTick` + `JobRole` to `NPCRenderInfo` in `npc/types.go`; populate in `NPCRenderSystem.Update` from AIState and Job; add `RenderInfosForSettlement(w, settlementEntity)` method
- [x] 2.4 **GREEN**: add `Symbol rune` + `Color string` + `SettlementEntity ecs.Entity` to `Building` in `settlement/components.go`; copy from BuildingDef on spawn in `SettlementSpawnSystem`
- [x] 2.5 **GREEN**: create `BuildingRenderSystem` in `settlement/systems.go` with `BuildingRenderInfo` struct in `types.go`; collect all buildings with non-zero Symbol per tick; expose `RenderInfos()` and `RenderInfosForSettlement(w, entity)` methods

## Phase 3: TUI Model — Settlement Screen State

- [x] 3.1 **RED**: tests for `"settlement"` screen transition (`'e'` on settlement tile), exit (`'q'`/`esc`), cursor navigation with edge clamping, reward popup lifecycle (create/decrement/expire), max 3 concurrent
- [x] 3.2 **GREEN**: add `SettlementViewState` (SettlementEntity, center coords, cursor, viewport radius), `RewardPopup` (WorldX/Y, Text, TicksLeft), and fields `settlementBuildings`/`settlementNPCs`/`rewardPopups` to Model
- [x] 3.3 **GREEN**: implement `openSettlementView(entity, cx, cy, radius)` — sets screen + state; `closeSettlementView()` — clears and returns to map; `processRewardPopups()` — create from NPC LastReward, decrement ticks, remove expired
- [x] 3.4 **GREEN**: keyboard handling in `Update` for `"settlement"` screen: arrows move cursor (clamped), `'e'` opens inspector (NPC priority), `'q'`/`esc` close inspector or exit view

## Phase 4: TUI View — Settlement Interior Rendering

- [x] 4.1 **RED**: tests for `renderSettlementView` with buildings at correct offsets, NPC overlay on buildings, cursor highlight, reward popups, building inspector (name/level/role/workers), NPC inspector (name/role/health/reward/epsilon/state), status bar (cursor pos + entity), stale rewards filtered, truncation
- [x] 4.2 **GREEN**: implement `renderSettlementView()` — dynamic viewport grid, building layer (symbol+color), NPC overlay on top, cursor highlight, reward popup floats; dispatch to `renderView` for `"settlement"` screen
- [x] 4.3 **GREEN**: implement `renderBuildingInspector()` — name, level, role, workers current/max via role-matching, assigned NPCs with reward; `renderSettlementNPCInspector()` — name, role, health, home, workplace, LastReward, epsilon, current state
- [x] 4.4 **GREEN**: implement status bar — dark background, cursor position + entity name (NPC priority), reward activity summary (last 3 ticks), truncation to terminal width; replaced by inspector when open

## Phase 5: Integration — main.go + Tests

- [x] 5.1 **GREEN**: wire `BuildingRenderSystem` in `cmd/evociv/main.go`; pass `[]BuildingRenderInfo` to Model; add `SetSettlementBuildings` method; override NPC LOD in settlement view via `RenderInfosForSettlement()`
- [x] 5.2 **GREEN**: golden file tests for settlement view (village 7×7, buildings + NPCs + reward popup); table-driven keyboard nav tests; smoke test with full world + multi-tick settlement rendering

(End of file - total 33 lines)
