# Tasks: Economy

## Phase 1: Data Layer — YAML extendido + Loaders

- [ ] 1.1 **RED**: tests para `LoadBuildingTypes` con produces/consumes y legacy building; tests para `LoadGrowthThresholds` con archivo válido y faltante
- [ ] 1.2 **GREEN**: extender `BuildingDef` (Role, Produces, Consumes, MaxWorkers); extender `data/buildings.yaml` con campos económicos; crear `data/growth.yaml` con thresholds L1→L2 y L2→L3
- [ ] 1.3 **GREEN**: extender `LoadBuildingTypes` en `data.go` para parsear role/produces/consumes/max_workers con validación de rates ≥ 0; crear `LoadGrowthThresholds()` y `GrowthThreshold` struct en `types.go`

## Phase 2: Core Infrastructure — ResourceStore Helpers

- [ ] 2.1 **RED**: tests para ResourceStore.Add, Remove (suficiente/insuficiente), Has en `components_test.go`
- [ ] 2.2 **GREEN**: implementar `Add(resource, amount)`, `Remove(resource, amount) bool`, `Has(resource, amount) bool` en `components.go`

## Phase 3: ECS Systems — Producción, Crecimiento, Hambruna

- [ ] 3.1 **RED**: tests para SettlementEconomySystem (farm con workers, blacksmith, market produce+consume, NPC food consumption, max_workers cap, lazy-init, house sin producción) en `economy/systems_test.go`
- [ ] 3.2 **GREEN**: crear `economy/systems.go` con `SettlementEconomySystem` (buildingMap lookup, workers por Job.Role, lazy-init ResourceStore, producción/consumo por tick, NPC food consumo)
- [ ] 3.3 **RED**: tests para SettlementGrowthSystem (level-up, partial resources, max level, missing threshold) + FamineSystem (food deficit, 1 NPC/tick, múltiples ticks, recovery, food positivo no-op) en `economy/systems_test.go`
- [ ] 3.4 **GREEN**: implementar `SettlementGrowthSystem` (threshold lookup, level-up con resource deduction, radius increment, max level cap) y `FamineSystem` (remove HomeReference de 1 NPC por tick cuando food < 0)

## Phase 4: TUI — Status Bar + Inspector Extendido

- [ ] 4.1 **RED**: tests para status bar con recursos, sin ResourceStore, inspector con resources/level progress/famine warning en `view_test.go`
- [ ] 4.2 **GREEN**: extender `SettlementRenderInfo` (Food, Gold, Tools, Level, HasResources); extender `SettlementRenderSystem` para leer ResourceStore; modificar status bar en `view.go` a formato `"♦ Aldea | Pop:5 | Food:45 Gold:12 Tools:3"`; extender `renderSettlementInspector` con recursos, progreso de nivel, y advertencia de hambruna

## Phase 5: Integración — main.go + Smoke Test

- [ ] 5.1 **RED**: smoke test que crea world completo con settlements, NPCs, y 3 sistemas económicos; ejecuta múltiples ticks; verifica que recursos, niveles y migración funcionan en conjunto
- [ ] 5.2 **GREEN**: registrar SettlementEconomySystem, SettlementGrowthSystem, FamineSystem en `cmd/evociv/main.go` en orden (economy → growth → famine) después de PopulationSystem y antes de render systems
