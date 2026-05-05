# Verification Report: economy

**Change**: economy
**Version**: N/A
**Mode**: Strict TDD
**Date**: 2026-05-05

---

## Completeness

| Metric | Value |
|--------|-------|
| Tasks total | 14 |
| Tasks complete | 14 |
| Tasks incomplete | 0 |

All tasks from `tasks.md` are reported complete in the apply-progress artifact. Implementation evidence exists in the codebase for every task.

---

## Build & Tests Execution

**Build**: ✅ Passed
```
go build ./cmd/evociv
(exit code 0, no output)
```

**Tests**: ✅ All passed
```
ok  github.com/marco/evociv-rl/cmd/evociv                  0.579s
ok  github.com/marco/evociv-rl/internal/data                0.509s
ok  github.com/marco/evociv-rl/internal/ecs                 0.454s
ok  github.com/marco/evociv-rl/internal/simulation/economy  0.495s
ok  github.com/marco/evociv-rl/internal/simulation/goap     0.476s
ok  github.com/marco/evociv-rl/internal/simulation/npc      0.535s
ok  github.com/marco/evociv-rl/internal/simulation/rl       0.431s
ok  github.com/marco/evociv-rl/internal/simulation/settlement 0.471s
ok  github.com/marco/evociv-rl/internal/store               0.749s
ok  github.com/marco/evociv-rl/internal/ui                  0.563s
ok  github.com/marco/evociv-rl/internal/world               0.552s
ok  github.com/marco/evociv-rl/internal/world/gen           0.545s
```

**Coverage**: ➖ Not available (no coverage tool configured per `openspec/config.yaml`)

---

## TDD Compliance

| Check | Result | Details |
|-------|--------|---------|
| TDD Evidence reported | ✅ | Found in apply-progress (sdd/economy/apply-progress) |
| All tasks have tests | ✅ | 6 task rows with test files listed |
| RED confirmed (tests exist) | ✅ | All reported test files exist in codebase |
| GREEN confirmed (tests pass) | ✅ | All tests pass on execution |
| Triangulation adequate | ✅ | 4–9 cases per task; no single-case tasks for multi-scenario specs |
| Safety Net for modified files | ✅ | Safety net reported for all modified files |

**TDD Compliance**: 6/6 checks passed

---

## Test Layer Distribution

| Layer | Tests | Files | Tools |
|-------|-------|-------|-------|
| Unit | ~32 | 5 | go test |
| Integration | 1 | 1 | go test |
| E2E | 0 | 0 | not installed |
| **Total** | **~33** | **6** | |

---

## Changed File Coverage

Coverage analysis skipped — no coverage tool detected (`openspec/config.yaml` reports `coverage.available: false`).

---

## Assertion Quality

**Assertion quality**: ✅ All assertions verify real behavior

All test assertions assert concrete values (e.g., `rs.Resources["food"] != 3.98`, `set.Level != 2`) rather than tautologies, type-only checks, or smoke-only renders. No ghost loops, no mock-heavy tests, and no implementation-detail coupling detected.

---

## Quality Metrics

**Linter**: ➖ Not available (`golangci-lint` not installed per config)
**Type Checker / Vet**: ✅ No errors (`go vet ./...` passed silently)

---

## Spec Compliance Matrix

### economy-data

| Requirement | Scenario | Test | Result |
|-------------|----------|------|--------|
| Economic Building Fields | Load buildings.yaml with produces and consumes | `data_test.go > TestLoadBuildingTypesWithEconomicFields` | ✅ COMPLIANT |
| Economic Building Fields | Legacy building without economic fields still loads | `data_test.go > TestLoadBuildingTypesWithEconomicFields` | ✅ COMPLIANT |
| Growth Threshold Definitions | Load valid growth.yaml | `data_test.go > TestLoadGrowthThresholdsValid` | ✅ COMPLIANT |
| Growth Threshold Definitions | Missing growth.yaml returns empty | `data_test.go > TestLoadGrowthThresholdsMissing` | ✅ COMPLIANT |
| BuildingDef Struct Extension | BuildingDef struct has new fields | `data_test.go > TestLoadBuildingTypesWithEconomicFields` | ✅ COMPLIANT |
| Validation — Produces Format | Reject negative production rate | `data_test.go > TestLoadBuildingTypesRejectsNegativeRate` | ✅ COMPLIANT |
| Building Type Definitions (MOD) | Load valid buildings.yaml with economic data | `data_test.go > TestLoadBuildingTypesWithEconomicFields` | ✅ COMPLIANT |

### economy-system

| Requirement | Scenario | Test | Result |
|-------------|----------|------|--------|
| SettlementEconomySystem Execution | Produce food from farm with workers | `economy/systems_test.go > TestEconomySystemFarmProducesFood` | ✅ COMPLIANT |
| SettlementEconomySystem Execution | Blacksmith produces tools | `economy/systems_test.go > TestEconomySystemBlacksmithProducesTools` | ✅ COMPLIANT |
| SettlementEconomySystem Execution | Market produces gold and consumes food | `economy/systems_test.go > TestEconomySystemMarketProducesGoldAndConsumesFood` | ✅ COMPLIANT |
| NPC Food Consumption | NPC consumes food from settlement | `economy/systems_test.go > TestEconomySystemNPCFoodConsumption` | ✅ COMPLIANT |
| NPC Food Consumption | No NPCs, no consumption | `economy/systems_test.go > TestEconomySystemNoNPCsNoConsumption` | ✅ COMPLIANT |
| Worker Assignment | Workers capped by max_workers | `economy/systems_test.go > TestEconomySystemMaxWorkersCap` | ✅ COMPLIANT |
| Worker Assignment | No workers, no production | `economy/systems_test.go > TestEconomySystemNoWorkersNoProduction` | ✅ COMPLIANT |
| ResourceStore Helpers | Add increments resource | `settlement/components_test.go > TestResourceStoreAdd` | ✅ COMPLIANT |
| ResourceStore Helpers | Remove decrements when sufficient | `settlement/components_test.go > TestResourceStoreRemoveSufficient` | ✅ COMPLIANT |
| ResourceStore Helpers | Remove returns false when insufficient | `settlement/components_test.go > TestResourceStoreRemoveInsufficient` | ✅ COMPLIANT |
| Lazy-Init ResourceStore | Lazy-init on first tick | `economy/systems_test.go > TestEconomySystemLazyInitResourceStore` | ✅ COMPLIANT |
| Building Without Production | House does not produce | `economy/systems_test.go > TestEconomySystemHouseIgnored` | ✅ COMPLIANT |

### settlement-growth

| Requirement | Scenario | Test | Result |
|-------------|----------|------|--------|
| SettlementGrowthSystem Execution | Settlement levels up when thresholds met | `economy/systems_test.go > TestGrowthSystemLevelUp` | ✅ COMPLIANT |
| SettlementGrowthSystem Execution | Settlement stays at same level without enough resources | `economy/systems_test.go > TestGrowthSystemPartialResources` | ✅ COMPLIANT |
| Level Thresholds | Level 2 to 3 requires more resources | `economy/systems_test.go > TestGrowthSystemLevel2To3` | ✅ COMPLIANT |
| Level Thresholds | Partial resources for next level | `economy/systems_test.go > TestGrowthSystemPartialResources` | ✅ COMPLIANT |
| Resource Deduction on Level-Up | Resources are deducted on level up | `economy/systems_test.go > TestGrowthSystemLevelUp` | ✅ COMPLIANT |
| Max Level Cap | Max level settlement does not check thresholds | `economy/systems_test.go > TestGrowthSystemMaxLevel` | ✅ COMPLIANT |
| Missing Threshold Returns Error | No threshold for next level | `economy/systems_test.go > TestGrowthSystemMissingThreshold` | ✅ COMPLIANT |

### famine-system

| Requirement | Scenario | Test | Result |
|-------------|----------|------|--------|
| FamineSystem Execution | Food deficit detected | `economy/systems_test.go > TestFamineSystemRemovesOneNPC` | ✅ COMPLIANT |
| FamineSystem Execution | Positive food, no action | `economy/systems_test.go > TestFamineSystemPositiveFoodNoAction` | ✅ COMPLIANT |
| NPC Migration on Food Deficit | NPC loses HomeReference during famine | `economy/systems_test.go > TestFamineSystemRemovesOneNPC` | ✅ COMPLIANT |
| NPC Migration on Food Deficit | Multiple ticks remove multiple NPCs | `economy/systems_test.go > TestFamineSystemMultipleTicks` | ✅ COMPLIANT |
| NPC Migration on Food Deficit | All NPCs migrate when deficit persists | `economy/systems_test.go > TestFamineSystemAllMigrate` | ✅ COMPLIANT |
| Nomad Behavior | Nomad wanders like homeless NPC | (none found) | ❌ UNTESTED |
| Recovery After Famine | Famine stops when food recovers | `economy/systems_test.go > TestFamineSystemPositiveFoodNoAction` | ⚠️ PARTIAL |

### economy-tui

| Requirement | Scenario | Test | Result |
|-------------|----------|------|--------|
| Status Bar with Resources | Status bar shows resources on cursor hover | `ui/view_test.go > TestRenderMapStatusBarWithResources` | ✅ COMPLIANT |
| Status Bar with Resources | Status bar shows only name and pop when no ResourceStore | `ui/view_test.go > TestRenderMapStatusBarNoResources` | ✅ COMPLIANT |
| Status Bar with Resources | No status bar for empty tiles | (none found) | ❌ UNTESTED |
| Inspector with Economic Data | Inspector shows resources and level progress | `ui/view_test.go > TestRenderInspectorSettlementResources` | ✅ COMPLIANT |
| Inspector with Economic Data | Inspector at max level shows "MAX" | `ui/view_test.go > TestRenderInspectorSettlementMaxLevel` | ✅ COMPLIANT |
| Inspector with Economic Data | Inspector shows famine warning | `ui/view_test.go > TestRenderInspectorFamineWarning` | ✅ COMPLIANT |

### Integration / Smoke

| Requirement | Scenario | Test | Result |
|-------------|----------|------|--------|
| Integration — main.go + Smoke Test | Smoke test full world with economy systems | `cmd/evociv/main_test.go > TestSmokeEconomyIntegration` | ✅ COMPLIANT |

**Compliance summary**: 36/39 scenarios compliant, 1 partial, 2 untested

---

## Correctness (Static — Structural Evidence)

| Requirement | Status | Notes |
|------------|--------|-------|
| Economic Building Fields | ✅ Implemented | `BuildingDef` extended; `data/buildings.yaml` updated; loader validates rates ≥ 0 |
| Growth Threshold Definitions | ⚠️ Partial | Loader and YAML exist, but `GrowthThreshold.NewBuildings` is `int` instead of spec-required `[]string` |
| BuildingDef Struct Extension | ✅ Implemented | `Role`, `Produces`, `Consumes`, `MaxWorkers` added to struct |
| Validation — Produces Format | ✅ Implemented | `LoadBuildingTypes` returns error for negative rates |
| Building Type Definitions (MOD) | ⚠️ Partial | Spec says every building MUST have `symbol` and `color`; neither `BuildingDef` struct nor `data/buildings.yaml` includes them |
| SettlementEconomySystem | ✅ Implemented | `economy/systems.go` with production, consumption, lazy-init |
| NPC Food Consumption | ✅ Implemented | 0.01 food/NPC/tick deducted in `SettlementEconomySystem.Update` |
| Worker Assignment | ✅ Implemented | Capped by `max_workers`; matched by `Job.Role` |
| ResourceStore Helpers | ✅ Implemented | `Add`, `Remove`, `Has` on `ResourceStore` |
| Lazy-Init ResourceStore | ✅ Implemented | Created with zeroed defaults if missing |
| Building Without Production | ✅ Implemented | Buildings with empty `Produces` are skipped |
| SettlementGrowthSystem | ✅ Implemented | Threshold lookup, level-up, resource deduction, max-level cap |
| Level Thresholds | ✅ Implemented | Loaded from `data/growth.yaml` |
| Resource Deduction on Level-Up | ✅ Implemented | `rs.Remove` called for food, tools, gold |
| Max Level Cap | ✅ Implemented | Hard cap at Level 3 |
| Missing Threshold | ✅ Implemented | Treated as max level (skip) |
| FamineSystem | ✅ Implemented | Removes one `HomeReference` per tick when food < 0 |
| Status Bar with Resources | ⚠️ Partial | Resources shown, but format deviates from spec (spacing and decimals) |
| Inspector with Economic Data | ⚠️ Partial | Resources, level progress, and famine warning shown; max-level outputs duplicate "Level" line |

---

## Coherence (Design)

| Decision | Followed? | Notes |
|----------|-----------|-------|
| Producción en memoria del sistema vs ECS component | ✅ Yes | `buildingDefs` map stored in system; no per-entity production component |
| Hambruna = migración vs muerte | ✅ Yes | `HomeReference` removed; NPC becomes nomad |
| GOAP no acoplado en MVP | ✅ Yes | Economy systems operate at settlement level; no NPC GOAP coupling |
| Thresholds YAML vs hardcodeados | ✅ Yes | `data/growth.yaml` loaded via `LoadGrowthThresholds` |
| File Changes table | ✅ Yes | All 14 files created/modified as designed |
| `GrowthThreshold.NewBuildings` type | ❌ No | Design and spec specify `[]string`; implementation uses `int` |

---

## Issues Found

### CRITICAL (must fix before archive)

1. **`GrowthThreshold.NewBuildings` type mismatch**
   - **Spec**: `NewBuildings []string` (array of building IDs)
   - **Design**: `NewBuildings []string`
   - **Implementation**: `NewBuildings int` in `internal/simulation/settlement/types.go:32`
   - **Impact**: `data/growth.yaml` cannot represent `new_buildings: ["house"]` correctly. `LoadGrowthThresholds` parses the field with `toInt`, silently ignoring array values.
   - **Fix**: Change struct field to `[]string`, update loader to parse `[]any` into `[]string`, and update YAML if needed.

### WARNING (should fix)

1. **`data/buildings.yaml` missing `symbol` and `color`**
   - The economy-data spec states every building MUST have `symbol` (rune) and `color`. Neither the YAML nor `BuildingDef` includes these fields. This appears to be a pre-existing schema gap, but it is explicitly stated in the delta spec.

2. **Status bar format deviates from spec**
   - Spec expects `"♦ Aldea | Pop:5 | Food:45.0 Gold:12.0 Tools:3.0"`
   - Actual output: `" ♦ Aldea | Pop: 5 | Food: 45 Gold: 12 Tools: 3 "` (extra spaces, missing decimals)

3. **`renderSettlementInspector` duplicates Level line at max level**
   - Line 145 always prints `Level: %d`; lines 159–160 then print `Level: 3 (MAX)`. Result contains both lines for max-level settlements.

4. **Famine spec scenario "Nomad wanders like homeless NPC" untested**
   - No test in this change verifies that an NPC without `HomeReference` is handled by `WanderSystem`. Relies entirely on pre-existing wander behavior.

5. **Famine recovery scenario only partially tested**
   - `TestFamineSystemPositiveFoodNoAction` verifies no removals when food is positive, but does not test the transition from negative to positive food (recovery after famine).

### SUGGESTION (nice to have)

1. **Add test for empty-tile status bar**
   - The scenario "No status bar for empty tiles" has no explicit test in `view_test.go`.

---

## Verdict

**FAIL**

One CRITICAL issue blocks archive: `GrowthThreshold.NewBuildings` is typed as `int` instead of the spec-mandated `[]string`. This breaks the data contract for growth thresholds and prevents future use of the `new_buildings` field as designed. All other functional requirements are implemented and tested correctly.
