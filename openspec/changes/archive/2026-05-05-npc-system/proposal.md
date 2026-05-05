# Proposal: NPC System

## Intent

Poblar el mundo generado con entidades NPC que lo habiten como agentes ECS autónomos. Transforma el mundo de estático a dinámico, sentando base para futuros sistemas (GOAP, economía, asentamientos) sin implementarlos aún.

## Scope

### In Scope
- 6 componentes ECS: Health, Personality (Big 5), Job, AIState (GOAP-ready), Appearance, LOD
- Spawner biome-weighted: 50–100 NPCs en mundo 256×256, placement seed-based
- 4 sistemas: NPCSpawnSystem, LODSystem, WanderSystem, NPCRenderSystem
- Data-driven: `data/npcs.yaml` (razas, roles, traits, nombres)
- TUI overlay: símbolo '@' sobre biome tiles + panel inspector al presionar 'e'
- Persistencia seed-based (determinista, sin serializar entidades ni tocar SQLite)

### Out of Scope
- GOAP real (solo estado AIState preparado), asentamientos, economía, inventarios, familia/reproducción, interacción jugador-NPC, pathfinding, serialización SQLite

## Capabilities

### New Capabilities
- `npc-data`: Definiciones data-driven de NPCs (razas, roles, traits, nombres) en YAML
- `npc-spawn`: Generación biome-weighted de NPCs con placement seed-based
- `npc-components`: Componentes ECS (Health, Personality, Job, AIState, Appearance, LOD)
- `npc-systems`: Sistemas de simulación (Spawn, Wander, LOD)
- `npc-tui`: Visualización overlay + inspector en TUI Bubbletea

### Modified Capabilities
- None

## Approach

5 fases secuenciales:
1. **Componentes**: tipos ECS + data structs (Health, Personality, Job, AIState, Appearance, LOD)
2. **Spawner + Data**: `data/npcs.yaml`, loader, spawner biome-weighted seed-based
3. **Sistemas**: WanderSystem (movimiento aleatorio), LODSystem (3 niveles de detalle)
4. **TUI overlay**: render '@' sobre mapa + panel inspector con keybinding 'e'
5. **Integración**: wiring en `main.go`, inicialización del spawner, registro de sistemas

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/ecs/component.go` | Modified | +6 nuevos tipos de componente |
| `internal/ecs/world.go` | Modified | RegisterComponentStore para nuevos tipos |
| `internal/simulation/npc/` | New | Paquete spawner + sistemas (spawn, wander, lod) |
| `internal/ui/model.go` | Modified | Estado: selección NPC, overlay flag |
| `internal/ui/view.go` | Modified | Render NPC '@' + inspector panel |
| `internal/ui/update.go` | Modified | Keybinding 'e' para abrir inspector |
| `data/npcs.yaml` | New | Definiciones data-driven (razas, roles, traits) |
| `cmd/evociv/main.go` | Modified | Inicialización spawner + registro sistemas |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Rendimiento TUI con 100+ NPCs renderizados | Medium | LOD 3 niveles con tick rates diferenciales |
| LOD premature optimization | Low | MVP con 2 niveles (visible/no visible); LOD 3 si profiling lo justifica |
| Overlap visual '@' con biome chars | Low | Símbolo '@' con color醒目 sobre biome tile subyacente |

## Rollback Plan

`git revert` del branch completo. Sin migraciones de datos ni cambios de esquema — revertir es limpio y seguro.

## Dependencies

- `ecs-core` spec: sistema ECS funcional (entity, world, component stores, systems)
- `data-loader` spec: cargador YAML con registry pattern
- `biomes-data` spec: biomas con metadata de habitabilidad para weighting

## Success Criteria

- [ ] 50–100 NPCs spawnean biome-weighted en mundo 256×256
- [ ] NPCs visibles como '@' en el mapa TUI con color distintivo
- [ ] Tecla 'e' abre panel inspector con datos del NPC seleccionado
- [ ] NPCs wander aleatorio funcional sin crasheos
- [ ] Misma seed → mismos NPCs (determinismo probado)
- [ ] `go test ./...` pasa sin roturas
