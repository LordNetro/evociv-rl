# Verification Report: Settlements

**Change**: settlements  
**Version**: N/A  
**Mode**: Strict TDD  
**Date**: 2026-05-05  
**Verifier**: sdd-verify  

---

## Completeness

| Metric | Value |
|--------|-------|
| Tasks total | 7 |
| Tasks complete | 7 |
| Tasks incomplete | 0 |

All tasks from `tasks.md` are marked complete:
- [x] 1.1 YAML data & loaders
- [x] 1.2 ECS components & types
- [x] 2.1 SettlementSpawnSystem
- [x] 2.2 NPC spawn en settlements
- [x] 3.1 TUI overlay
- [x] 3.2 TUI inspector
- [x] 4.1 Main wiring

---

## Build & Tests Execution

**Build**: ✅ Passed
```
go build ./cmd/evociv
# exit 0, no errors
```

**Tests**: ✅ 44 passed / ❌ 0 failed / ⚠️ 0 skipped
```
go test ./... -count=1
ok  	github.com/marco/evociv-rl/cmd/evociv		0.563s
ok  	github.com/marco/evociv-rl/internal/data		0.468s
ok  	github.com/marco/evociv-rl/internal/ecs		0.496s
ok  	github.com/marco/evociv-rl/internal/simulation/goap	0.487s
ok  	github.com/marco/evociv-rl/internal/simulation/npc	0.545s
ok  	github.com/marco/evociv-rl/internal/simulation/rl	0.476s
ok  	github.com/marco/evociv-rl/internal/simulation/settlement	0.445s
ok  	github.com/marco/evociv-rl/internal/store		0.793s
ok  	github.com/marco/evociv-rl/internal/ui		0.521s
ok  	github.com/marco/evociv-rl/internal/world		0.454s
ok  	github.com/marco/evociv-rl/internal/world/gen	0.571s
```

**Coverage**:
| Package | Coverage |
|---------|----------|
| `internal/simulation/settlement` | 82.7% |
| `internal/simulation/npc` | 86.3% |
| `internal/ui` | 83.2% |
| `cmd/evociv` | 0.0% |

---

## TDD Compliance

| Check | Result | Details |
|-------|--------|---------|
| TDD Evidence reported | ✅ | Found in `sdd/settlements/apply-progress` |
| All tasks have tests | ✅ | 7/7 tasks have test files |
| RED confirmed (tests exist) | ✅ | All reported test files verified present in codebase |
| GREEN confirmed (tests pass) | ✅ | All tests pass on execution |
| Triangulation adequate | ⚠️ | Task 1.1 reports 4 cases but file contains 6 tests; counts generally match for other tasks |
| Safety Net for modified files | ⚠️ | Reported safety-net counts are approximate (e.g., spawner had ~8 original tests; progress reports 9/9) |

**TDD Compliance**: 4/6 checks passed (2 warnings)

---

## Test Layer Distribution

| Layer | Tests | Files | Tools |
|-------|-------|-------|-------|
| Unit | 43 | 7 | `go test` |
| Integration | 1 | 1 | `go test` |
| E2E | 0 | 0 | not installed |
| **Total** | **44** | **8** | |

**Unit test files** (`internal/simulation/settlement/*_test.go`, `internal/simulation/npc/spawner_test.go`, `internal/ui/*_test.go`) contain isolated assertions with no HTTP calls, browser contexts, or render-to-DOM checks.  
**Integration test file** (`cmd/evociv/main_test.go`) wires full ECS world + data loader + NPC/settlement systems end-to-end.

---

## Changed File Coverage

| File | Line % | Notes |
|------|--------|-------|
| `internal/simulation/settlement/data.go` | ~82% | `toFloat64`/`toInt` helpers partially exercised |
| `internal/simulation/settlement/components.go` | 100% | |
| `internal/simulation/settlement/systems.go` | ~78% | `Name()`, `NewSettlementRenderSystem`, `SettlementRenderSystem.Update/RenderInfos` at 0% |
| `internal/simulation/settlement/types.go` | — | declarative (no executable lines) |
| `internal/simulation/npc/spawner.go` | ~86% | `findCompatibleSettlements` partially exercised (hunter/miner paths untested) |
| `internal/simulation/npc/systems.go` | ~86% | package-level average |
| `internal/ui/model.go` | ~83% | `refreshOverlay`, `SetNPCOverlay`, `SetECSWorld` not directly hit |
| `internal/ui/view.go` | ~83% | `renderInspector` branches partially hit |
| `cmd/evociv/main.go` | 0.0% | smoke test builds its own world; `run()` unexercised |

**Average changed-file coverage (packages)**: 82.7% / 86.3% / 83.2% / 0.0%  
> `cmd/evociv` 0.0% is expected for a `main` package when the integration test instantiates components manually rather than calling `run()`.

---

## Assertion Quality

✅ **All assertions verify real behavior**

Scanned test files:
- `settlement/data_test.go`
- `settlement/components_test.go`
- `settlement/systems_test.go`
- `npc/spawner_test.go`
- `ui/view_test.go`
- `ui/model_test.go`
- `cmd/evociv/main_test.go`

No tautologies, ghost loops, empty-collection-only checks, or type-only assertions were found. Every test exercises production code and asserts meaningful outcomes (counts, positions, component values, string contents).

---

## Quality Metrics

**Linter**: ➖ Not available (`golangci-lint` not installed)  
**Type Checker**: ✅ No errors (`go vet ./...` passes cleanly)

---

## Spec Compliance Matrix

Source spec: `openspec/changes/settlements/specs/npc-spawner/spec.md`

| Requirement | Scenario | Test | Result |
|-------------|----------|------|--------|
| **REQ-1: Settlement-Aware NPC Placement** | Farmer spawns in village with farm | `npc/spawner_test.go > TestSpawnFarmerInSettlement` | ✅ COMPLIANT |
| **REQ-1: Settlement-Aware NPC Placement** | Merchant spawns in town | (none found) | ❌ UNTESTED |
| **REQ-2: Role-to-Building Matching** | Hunter assigned to nearest village | (none found) | ❌ UNTESTED |
| **REQ-3: Nomad Fallback** | Nomad NPC without settlement | `npc/spawner_test.go > TestSpawnNomadNoHomeReference` | ✅ COMPLIANT |
| **REQ-4: Settlement Capacity** | Settlement capacity overflow becomes nomad | `npc/spawner_test.go > TestSpawnCapacityOverflow` | ✅ COMPLIANT |
| **REQ-5: Biome-Weighted Placement** | NPC inside settlement ignores biome weight | (none found) | ❌ UNTESTED |
| **REQ-5: Biome-Weighted Placement** | Nomad still respects biome weights | `npc/spawner_test.go > TestSpawnPlainsGreaterThanTundra` | ✅ COMPLIANT |
| **REQ-6: Deterministic Spawning** | Same seed produces identical settlement assignment | (none found) | ❌ UNTESTED |

**Compliance summary**: **3/8 scenarios compliant** (37.5%)

### Additional Behavioral Checks (from design/tasks)

| Behavior | Test | Result |
|----------|------|--------|
| Settlement spawn count 5–10 on plains | `settlement/systems_test.go > TestSettlementSpawnSystemPlains` | ✅ COMPLIANT |
| 0 settlements on ocean | `settlement/systems_test.go > TestSettlementSpawnSystemOcean` | ✅ COMPLIANT |
| Min distance ≥ 20 (Chebyshev) | `settlement/systems_test.go > TestSettlementSpawnSystemMinDistance` | ✅ COMPLIANT |
| Deterministic settlement positions | `settlement/systems_test.go > TestSettlementSpawnSystemDeterminism` | ✅ COMPLIANT |
| Buildings inside radius | `settlement/systems_test.go > TestSettlementSpawnSystemBuildingsInsideRadius` | ✅ COMPLIANT |
| Settlement spawn runs once | `settlement/systems_test.go > TestSettlementSpawnSystemRunsOnce` | ✅ COMPLIANT |
| NPC > Settlement > Biome overlay | `ui/view_test.go > TestRenderOverlaySettlementPriority` | ✅ COMPLIANT |
| Settlement-only overlay symbol | `ui/view_test.go > TestRenderOverlaySettlementOnly` | ✅ COMPLIANT |
| Inspector opens on settlement | `ui/model_test.go > TestInspectorOpenOnSettlement` | ✅ COMPLIANT |
| Inspector shows settlement data | `ui/view_test.go > TestRenderInspectorSettlementData` | ✅ COMPLIANT |
| End-to-end smoke test | `cmd/evociv/main_test.go > TestSmokeSettlementIntegration` | ✅ COMPLIANT |
| Weighted settlement pick | (no direct distribution test) | ⚠️ PARTIAL |
| Procedural name generation | (no direct test) | ⚠️ PARTIAL |
| Role-to-building matrix (merchant, hunter, priest, blacksmith) | (only farmer tested) | ⚠️ PARTIAL |

---

## Correctness (Static — Structural Evidence)

| Requirement | Status | Notes |
|------------|--------|-------|
| YAML data & loaders | ✅ Implemented | `data.go` + `data_test.go` |
| ECS components & types | ✅ Implemented | `components.go`, `types.go`, `components_test.go` |
| SettlementSpawnSystem | ✅ Implemented | `systems.go`, `systems_test.go` |
| NPC spawn in settlements | ✅ Implemented | `npc/spawner.go` modified, settlement-aware logic present |
| HomeReference component | ✅ Implemented | Added to `components.go`, assigned in `spawner.go` |
| Capacity check (Radius×2) | ✅ Implemented | `findCompatibleSettlements` enforces cap |
| Nomad fallback | ✅ Implemented | Falls back to biome-weighted random when no settlement found |
| TUI overlay priority | ✅ Implemented | `renderOverlay`: NPC → Settlement → Biome |
| TUI inspector settlement | ✅ Implemented | `renderSettlementInspector` + `tryOpenInspector` fallback |
| Main wiring | ✅ Implemented | `main.go` registers stores, loaders, systems |

---

## Coherence (Design)

| Decision | Followed? | Notes |
|----------|-----------|-------|
| ADR-1: ECS entities (not separate grid) | ✅ Yes | Settlements and buildings are ECS entities |
| ADR-2: Buildings as ECS child entities (not slots) | ✅ Yes | Building entities with Position + Building component |
| ADR-3: Procedural names via fixed pool | ✅ Yes | `types.go` defines prefixes/suffixes; `systems.go` generates names |
| ADR-4: Overlay order NPC > Settlement > Biome | ✅ Yes | `view.go` implements exact priority |
| No LOD component on settlements | ✅ Yes | Settlements do not receive LOD; render always |
| Population field in Settlement component | ✅ Yes | Present; used for capacity tracking in spawner |
| Inspector shows Name/Type/Pop/Radius/Level/Buildings | ⚠️ Deviated | Shows building **list** instead of **count**; UX improvement, design says "Building count" |
| SettlementRenderSystem every tick | ✅ Yes | `refreshOverlay()` calls it each tick |
| File changes match design table | ✅ Yes | All expected files created/modified |

---

## Issues Found

### CRITICAL (must fix before archive)

1. **Spec scenario UNTESTED: "Merchant spawns in town"**  
   `npc-spawner/spec.md` REQ-1 scenario 2 has no corresponding test. The `findCompatibleSettlements` path for `merchant` → `market` is present in code but not exercised at runtime.

2. **Spec scenario UNTESTED: "Hunter assigned to nearest village"**  
   `npc-spawner/spec.md` REQ-2 scenario has no test. Hunter logic (village/town preference, no building requirement) is implemented but not proven by a passing test.

3. **Spec scenario UNTESTED: "NPC inside settlement ignores biome weight"**  
   `npc-spawner/spec.md` REQ-5 scenario 1 has no test. Code bypasses `biomeWeight` for settlement-assigned NPCs, but this behavior is not explicitly validated.

4. **Spec scenario UNTESTED: "Same seed produces identical settlement assignment"**  
   `npc-spawner/spec.md` REQ-6 scenario has no test. `TestSpawnDeterminism` runs with `nil` settlements; no test verifies deterministic `HomeReference` values when settlements exist.

### WARNING (should fix)

1. **Apply-progress triangulation count mismatch for Task 1.1**  
   Progress reports 4 cases; `data_test.go` contains 6 test functions. Counts are directionally correct but not exact.

2. **Apply-progress safety-net count is approximate**  
   Task 2.2 reports 9/9 passing before modification; original `spawner_test.go` contained ~8 tests (additional tests may have been in other files of the same package). The safety-net claim is plausible but not strictly auditable.

3. **Settlement inspector deviation from design**  
   Design ADR specifies "Building count" in inspector; implementation renders the full building list. This is an improvement but a documented deviation.

4. **`SettlementRenderSystem` has 0% coverage in its own unit**  
   `NewSettlementRenderSystem`, `Name`, `Update`, and `RenderInfos` are hit only via integration (`main_test.go` and `model.go` refresh), not by a dedicated unit test in `settlement/systems_test.go`.

### SUGGESTION (nice to have)

1. Add dedicated unit tests for `SettlementRenderSystem.Update` and `RenderInfos`.
2. Add a test for procedural name edge cases (collision retry, truncation at 15 chars).
3. Expand role-to-building tests to cover `merchant`, `hunter`, `priest`, `blacksmith`, `miner`, `trader`.
4. Add a determinism test that runs `Spawn` twice with the same seed and identical settlement entities, asserting identical `HomeReference` values.
5. Add a test placing a settlement over a tundra tile and verifying that an assigned NPC spawns successfully despite the low biome weight.

---

## Verdict

**FAIL**

The implementation is structurally complete, builds cleanly, and all existing tests pass. However, **Strict TDD compliance requires every spec scenario to be proven by a passing test**. Four spec scenarios from the formal `npc-spawner/spec.md` are currently UNTESTED:

- Merchant spawns in town
- Hunter assigned to nearest village
- NPC inside settlement ignores biome weight
- Same seed produces identical settlement assignment

These gaps must be closed (tests written and passing) before the change can be archived.
