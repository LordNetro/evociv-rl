# Proposal: Settlements

## Intent

Implementar asentamientos (ciudades, pueblos, aldeas) como entidades ECS con componentes, spawn bioma-weight, edificios, y visualización TUI. Los NPCs pasan a spawn dentro de asentamientos según su rol, reemplazando el spawn aleatorio actual.

## Scope

### In Scope
- Componentes ECS: Settlement, Building, ResourceStore, HomeReference
- YAML data-driven: `settlements.yaml`, `buildings.yaml`
- SettlementSpawnSystem: genera 5-10 asentamientos en biomas compatibles
- Spawn de NPCs dentro de radios de asentamientos con rol-to-building matching
- Visualización TUI: símbolos ♦▲● + nombre del settlement
- Edificios como entidades ECS hijas dentro de cada settlement

### Out of Scope
- Economía: producción/consumo de recursos
- Acciones GOAP vinculadas a settlements o buildings
- Crecimiento poblacional automatizado
- Persistencia SQLite de settlements

## Capabilities

### New Capabilities
- `settlement-spawn`: generación de asentamientos en el mundo por bioma
- `settlement-components`: componentes ECS para asentamientos y edificios
- `settlement-tui`: renderizado de settlements en el mapa TUI

### Modified Capabilities
- `npc-spawner`: NPCs spawn dentro de settlements, no en posiciones aleatorias
- `npc-components`: nuevo componente HomeReference en NPCs

## Approach

6 fases secuenciales, cada una con tests (Strict TDD):

1. **Infra Data**: `settlements.yaml`, `buildings.yaml` + carga/validación YAML
2. **Componentes ECS**: Settlement, Building, ResourceStore, HomeReference + registro de stores
3. **Spawn Settlements**: SettlementSpawnSystem con sampling bioma-weight y distancia mínima
4. **Spawn NPCs en Settlements**: modificar spawner.go, rol-to-building matching, asignar HomeReference
5. **TUI**: símbolos ♦▲●, nombres en overlay, orden NPC > Settlement > Biome, inspector
6. **Edificios**: buildings como entidades hijas dentro del settlement

Decisiones clave: asentamientos como entidades ECS (no grid separado), NPCs tienen HomeReference, overlay prioriza NPC > Settlement > Biome.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/simulation/settlement/` | **New** | Paquete: componentes, sistemas, tipos, datos |
| `data/settlements.yaml` | **New** | Settlement types data-driven |
| `data/buildings.yaml` | **New** | Building types data-driven |
| `internal/simulation/npc/spawner.go` | Modified | Spawn dentro de settlements |
| `internal/simulation/npc/components.go` | Modified | +HomeReference component |
| `internal/ui/view.go` | Modified | Settlement overlay rendering |
| `internal/ui/model.go` | Modified | Settlement model data |
| `cmd/evociv/main.go` | Modified | Registrar nuevos stores y sistemas |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Spawn en biome incorrecto si faltan tiles | Low | Fallback a biome más cercano por Manhattan |
| Más NPCs que capacidad de settlements | Low | NPCs nómadas (sin HomeReference) como fallback |
| Overlay conflict NPC ↔ Settlement | Low | Orden explícito: NPC > Settlement > Biome |

## Rollback Plan

`git revert` del branch de settlements. Si ya mergeado a main, revert commit y crear PR de revert.

## Dependencies

Ninguna. El ECS core ya soporta RegisterStores adicionales.

## Success Criteria

- [ ] 5-10 settlements spawn en biomas válidos en mundo 256×256
- [ ] NPCs spawn dentro de radios de settlements con HomeReference asignado
- [ ] Símbolos ♦▲● visibles en TUI con nombres al hover
- [ ] Edificios spawn como entidades hijas dentro de settlements
- [ ] Todos los tests existentes siguen pasando
