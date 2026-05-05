# Proposal: Economy

## Intent

Implementar economía de producción y consumo. Edificios producen recursos con workers asignados por Role, NPCs consumen food del settlement, y los settlements suben de nivel al acumular recursos suficientes.

## Scope

### In Scope
- YAML extendido: `buildings.yaml` con `produces`/`consumes`/`max_workers`/`role`; `growth.yaml` con thresholds por nivel
- `ResourceStore` helpers: `Add()`, `Remove()`, `Has()` — el componente existe pero nunca se escribe
- `SettlementEconomySystem` (nuevo paquete `internal/simulation/economy/`): produce recursos por building+role, consume food por NPC
- `SettlementGrowthSystem`: verifica thresholds, sube de nivel, expande radio, spawn buildings
- `FamineSystem`: déficit de food → migración (HomeReference removal)
- TUI: status bar con recursos + inspector expandido con level progress y estado

### Out of Scope
- GOAP vinculado a edificios (post-MVP)
- Comercio entre settlements
- Monedas, precios dinámicos o market
- Muerte de NPCs por hambruna (solo migración en MVP)

## Capabilities

### New Capabilities
- `economy-system`: producción automática por building+role, consumo de food por NPC, lazy-init de ResourceStore
- `settlement-growth`: level-up con thresholds YAML, expansión de radio y buildings al subir
- `famine-system`: déficit acumulado → migración (HomeReference removal)
- `economy-tui`: status bar con recursos, inspector con level progress y estado

### Modified Capabilities
- `settlement-buildings`: BuildingDef extendido con Role, Produces, Consumes, MaxWorkers
- `settlement-data`: `data/growth.yaml` nuevo, `LoadBuildingTypes` con carga de producción
- `settlement-components`: ResourceStore helpers Add/Remove/Has + lazy-init

## Approach

7 fases secuenciales con Strict TDD (cada fase = código + tests):

1. **YAML extendido**: `buildings.yaml` + `growth.yaml` + loaders + validación
2. **ResourceStore helpers**: Add/Remove/Has en `components.go` + tests
3. **SettlementEconomySystem**: produce+consume cada tick, lazy-init ResourceStore
4. **SettlementGrowthSystem**: thresholds YAML, level-up con deducción de recursos
5. **FamineSystem**: déficit tracking → migración de NPCs
6. **TUI**: status bar + inspector expandido con recursos y estado
7. **Main wiring**: registrar sistemas en `cmd/evociv/main.go` + smoke test

Decisiones clave: producción en memoria del sistema (no ECS component), hambruna = migración (no muerte), GOAP no acoplado en MVP.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/simulation/economy/` | **New** | Paquete: SettlementEconomySystem, SettlementGrowthSystem, FamineSystem |
| `internal/simulation/settlement/components.go` | Modified | ResourceStore helpers Add/Remove/Has |
| `internal/simulation/settlement/types.go` | Modified | BuildingDef con Produces, Consumes, MaxWorkers, Role |
| `internal/simulation/settlement/data.go` | Modified | LoadBuildingTypes extendido, LoadGrowthThresholds |
| `data/buildings.yaml` | Modified | +produces, consumes, max_workers, role |
| `data/growth.yaml` | **New** | Thresholds de crecimiento por nivel |
| `internal/ui/view.go` | Modified | Status bar + inspector con recursos |
| `internal/ui/model.go` | Modified | Campos para economía (si necesario) |
| `cmd/evociv/main.go` | Modified | Registrar EconomySystem, GrowthSystem |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Desbalance producción/consumo | Medium | Datos en YAML → tuning sin recompilar |
| ResourceStore nil en saves existentes | Medium | Lazy-init en SettlementEconomySystem |
| GOAP y economía compiten por Hunger/Fatigue | Low | Operan en niveles distintos (individual vs settlement) |

## Rollback Plan

`git revert` del branch de economía. Si mergeado, revert commit + PR de revert.

## Dependencies

Ninguna externa. SettlementEconomySystem depende de componentes existentes (ResourceStore, Settlement, Building, HomeReference, Job).

## Success Criteria

- [ ] Farm con 2 farmers produce +4.0 food/tick en tests
- [ ] NPCs consumen 0.01 food/tick de su settlement
- [ ] Settlement sube de nivel al acumular recursos suficientes
- [ ] Settlement sin food por 10+ ticks → NPCs migran (pierden HomeReference)
- [ ] Status bar muestra recursos del settlement bajo cursor
- [ ] Inspector muestra food/gold/tools y level progress
- [ ] Todos los tests existentes siguen pasando
