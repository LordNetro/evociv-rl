# Proposal: Settlement Interiors

## Intent

Implementar interiores navegables en buildings. Workers GOAP autonomous entran/salen, producción escala con `workers_inside`. Player como spectator omnisciente con Z-mode view.

## Scope

### In Scope
- Building interior procedural: rooms + corridors + doors (BSP o room placement simple)
- Indoor pathfinding: A* 2D grid dentro del building footprint
- Production: `base_output * workers_inside * efficiency` (modifica settlement-production)
- GOAP actions: EnterBuilding, ExitBuilding, WorkInside — goal "BeProductive"
- TUI: Z-mode viewport mostrando workers inside buildings
- Settlement como mapa jugable con múltiples buildings

### Out of Scope
- Player control de workers
- Building construction por player
- Multi-floor / Z-levels
- Inventarios por building

## Capabilities

### New Capabilities
- `building-interior`: BuildingInterior component con grid 2D de rooms/corridors, puertas, capacidad máxima
- `indoor-pathfinding`: A* 2D sobre interior grid, entrada/salida definidas por door positions
- `production-scaling`: ProductionSystem multiplica por workers_inside count en cada building

### Modified Capabilities
- `settlement-production`: Fórmula extendida de `base_output` → `base_output * workers_inside * efficiency`

## Approach

1. **BuildingInterior ECS component**: grid 2D, rooms[], doors[], max_workers
2. **Procedural layout**: BSP o placement simple determinístico por building seed
3. **IndoorPathfinder**: A* que acepta entrada (door position) y calcula path interno
4. **ProductionSystem update**: counting workers_inside per building, applying multiplier
5. **GOAP integration**: EnterBuilding (path to door → enter), WorkInside (stay), ExitBuilding
6. **TUI Z-mode**: viewport overlay zoomed en interior, muestra workers moviéndose

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/simulation/settlement/components.go` | Modified | BuildingInterior component |
| `internal/simulation/settlement/` | **New** | InteriorGenerator, IndoorPathfinder |
| `internal/simulation/goap/actions/` | **New** | EnterBuilding, ExitBuilding, WorkInside |
| `internal/simulation/production.go` | Modified | workers_inside multiplier |
| `internal/ui/` | Modified | Z-mode viewport overlay |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Pathfinding en interior complejo | Medium | Grid simple (no NavMesh), fallback a door-to-door |
| Performance con muchos workers | Low | Caching de paths, recalcular solo cuando building cambia |
| GOAP infinite loop enter/exit | Medium | Cooldown en GOAP actions, efficiency penalty si cambia mucho |

## Rollback Plan

`git revert` del feature branch. Cambios backwards-compatible en ECS (BuildingInterior es optional).

## Dependencies

- ECS core existente
- SettlementBuildSystem existente
- GOAP planner existente
- TUI model existente

## Success Criteria

- [ ] Worker entra a workshop vacío → producción sube
- [ ] Worker sale del workshop → producción baja
- [ ] Dos workers inside → producción dobla
- [ ] TUI muestra workers moviéndose inside building
- [ ] Tests pasan para interior generation y pathfinding

## Executive Summary

Buildings se vuelven volúmenes navegables con rooms y corredores. Workers GOAP autonomous决定的进入内部工作，生产效率随内部人数缩放。Player spectator view con Z-mode para ver workers dentro de edificios. Cambia settlement-production de fijo a dinámico.