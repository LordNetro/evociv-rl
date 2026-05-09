# Tasks: Settlements

*strict_tdd: true*

## Phase 1: Foundation

- [x] 1.1 **YAML data & loaders** — `data/settlements.yaml` + `data/buildings.yaml` + `data.go` (LoadSettlementTypes, LoadBuildingTypes, validateSpawnWeights) + `data_test.go` (fstest.MapFS, weight sum ±0.01, unknown building ref, missing kind error)
  RED: test load errors on bad kind/weights/refs. GREEN: implement loaders + validation.

- [x] 1.2 **ECS components & types** — `components.go` (Settlement, Building, HomeReference, ResourceStore + IDs + RegisterSettlementStores) + `types.go` (SettlementDef, BuildingDef, SettlementRenderInfo, name pool) + `components_test.go` (RegisterSettlementStores registers 4 stores)
  RED: test all 4 stores registered. GREEN: implement structs, IDs, RegisterSettlementStores.

## Phase 2: Core Systems

- [x] 2.1 **SettlementSpawnSystem** — `systems.go` (spawn 5–10, weighted type pick, min Chebyshev 20, biome filter, seed+777 determinista, procedural names, building children) + `systems_test.go` (spawn count [5,10] on plains, 0 on ocean, determinism, buildings inside radius)
  RED: test plains yield 5–10 settlements, ocean yields 0, same seed same positions. GREEN: implement system.

- [x] 2.2 **NPC spawn en settlements** — `spawner.go` (rol→building map, settlement query, HomeReference add, capacity Radius×2, nomad fallback) + `spawner_test.go` (farmer in farm settlement has HomeRef, nomad zero HomeRef, capacity overflow)
  RED: test farmer HomeRef non-zero, no-settlement nomad has zero HomeRef. GREEN: implement matching + fallback.

## Phase 3: TUI

- [x] 3.1 **TUI overlay** — `model.go` (settlementOverlay field, SetSettlementOverlay) + `view.go` (renderOverlay: NPC > settlement > biome, styledSettlement) + tests (NPC+Settlement tile returns NPC, settlement-only tile returns symbol)
  RED: test overlay priority. GREEN: implement renderSettlementOverlay + styledSettlement.

- [x] 3.2 **TUI inspector** — `model.go` (tryOpenInspector settlement fallback) + `view.go` (renderInspector settlement section: Name, Type, Pop, Radius, Level, Buildings) + tests (cursor on settlement + 'e' shows fields)
  RED: test inspector opens on settlement tile. GREEN: implement settlement branch in tryOpenInspector + renderInspector.

## Phase 4: Integration

- [x] 4.1 **Main wiring** — `main.go` (RegisterSettlementStores, LoadSettlementTypes/BuildingTypes, add SettlementSpawnSystem + SettlementRenderSystem, wire overlay refresh) + smoke test (world init yields settlement entities, overlay renders)
  RED: smoke test asserts settlement entities after init. GREEN: wire all systems in main, verify end-to-end.
