# Proposal: Settlement Interior View

## Intent

Full-screen settlement interior view for the TUI — inspect buildings, NPCs, and RL rewards inside any settlement. Closes the core "god simulator" feedback loop.

## Scope

### In Scope
- `"settlement"` screen: dynamic viewport (7×7/11×11/17×17), arrow nav, `e` inspect, `esc`/`q` exit
- Building symbols `⌂╬§♨ϟ⚒` + colors at world-relative positions
- NPCs visible inside settlement (all LOD, same tick rate)
- Reward popups `+0.87` near NPCs, fade after 5 ticks
- Building inspector: name, level, role, workers, produces/consumes, assigned NPCs
- NPC inspector: name, role, health, home, workplace, Q-learning stats
- Status bar: tile info; replaced by inspector when open
- Data: `interior_symbol` + `color` on BuildingDef/Building; `LastReward` on AIState; `SettlementEntity` on Building

### Out of Scope
- `WorkplaceReference` component (dynamic role-matching for MVP)
- World-travel speed changes, entities outside radius, animated transitions

## Capabilities

### New
- `settlement-interior-view`: Full interior screen with buildings, NPCs, popups, dual inspectors

### Modified
- `settlement-components`: Add `Symbol, Color, SettlementEntity` to Building; `LastReward` to AIState
- `settlement-buildings`: BuildingDef gains `InteriorSymbol, Color`; loader/spawn updated
- `settlement-tui`: Extended with interior render, inspectors, popups, navigation
- `npc-tui`: Extended with reward display, Q-learning stats (last reward, policy, epsilon, state)
- `qlearning-engine`: Expose `LastReward` after `ComputeReward()` for UI

## Approach

6 phases: (1) YAML+Go data types, (2) ECS wiring (LastReward, SettlementEntity), (3) TUI Model + key handling, (4) renderSettlementView, (5) tests, (6) integration.

## Affected Areas

| Area | Change |
|------|--------|
| `data/buildings.yaml` | Add `interior_symbol`, `color` |
| `internal/simulation/settlement/{types,components,data,systems}.go` | New fields on BuildingDef/Building; parse+spawn |
| `internal/simulation/npc/{components,types,systems}.go` | LastReward on AIState → NPCRenderInfo |
| `internal/ui/model.go` | `"settlement"` screen state + key handling |
| `internal/ui/view.go` | `renderSettlementView()`, inspectors, popups, status bar |
| `internal/ui/*_test.go` | Table-driven + golden file tests |
| `cmd/evociv/main.go` | Minimal wiring if needed |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Tile collision: NPCs+buildings overlap | Med | Layer: buildings Bg, NPCs overlay (consistent with map) |
| Reward popup noise | Low | Min threshold 0.1, max 3 concurrent |
| LOD hides interior NPCs | Low | Override: show all NPCs with matching HomeReference |

## Rollback Plan

Per-phase commits. Map code path guarded by `screen == "map"` — unaffected by interior view. Revert in reverse phase order.

## Dependencies

Existing settlement ECS entities, Building entities, NPC entities with HomeReference + AIState. QLearningSystem already computes reward (not exposed).

## Success Criteria

- [ ] `go test ./...` passes
- [ ] `e` on a settlement opens interior view
- [ ] Buildings render with correct symbols/colors at expected positions
- [ ] NPCs move within viewport; reward popups appear and fade
- [ ] Building inspector shows name, level, role, workers, production
- [ ] NPC inspector shows role, health, home, workplace, Q stats
- [ ] Arrow keys navigate; `e` inspects; `esc`/`q` returns to map
- [ ] Status bar shows tile info; inspector replaces it when open
