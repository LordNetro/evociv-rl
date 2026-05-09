# Plan de refactor: Mover hacia estilo Dwarf Fortress

Objetivo: Reorganizar la arquitectura y UX para asemejarse a Dwarf Fortress (mapa por azulejos ASCII, edificios multi-azulejo y workflows de oficio), manteniendo los módulos de RL y GOAP actuales e integrándolos con las nuevas abstracciones.

Fases (iterativo, PRs pequeños):

1) Mapear subsistemas actuales
- ECS core: `internal/ecs`
- Mundo y generación: `internal/world`, `internal/world/gen`
- Simulación: `internal/simulation` (subpaquetes: `npc`, `settlement`, `goap`, `rl`)
- UI/TUI: `internal/ui` (Bubbletea)
- Datos/Store: `internal/store`, `data/` YAML
- CLI/bootstrap: `cmd/evociv`

2) Diseño de abstracciones Dwarf-Fortress-like
- Tilemap por Z-levels (XZ + Z)
- Multi-tile buildings con interiores enlazados a entidades
- Jobs y professions (cola de tasks por edificio/manager)
- Inventario por agente/stackable items
- Pathfinding y zonas (expandir sistema existente)

3) Refactor incremental (PRs recomendados)
- PR A: Añadir tipos `Inventory`, `Item`, `Job` (solo tipos, sin integraciones) — esto facilita cambios siguientes.
- PR B: Añadir `JobSystem` mínimo y `Job` component; adaptar `npc` para aceptar y ejecutar jobs via GOAP/QLearning rewards.
- PR C: Convertir renderer a ASCII tilemap (TUI) y añadir layers (terrain, buildings, entities, UI)
- PR D: Refactor world-gen a tiles y multi-tile buildings, mantener generadores actuales como adaptadores
- PR E: Telemetría/visualización de Q-table + training overlays

4) Pruebas y compatibilidad
- Mantener y ampliar tests unitarios: `internal/simulation/*` y `internal/ecs`
- Activar TDD para cambios críticos (Pathfinding, JobSystem)

5) Entregables
- `openspec/changes/df-refactor-plan.md` (este documento)
- PRs pequeños con tests verdes
- Documentación en `openspec/specs/df-refactor/`

Riesgos y mitigaciones
- Romper la integración RL+GOAP → Mitigar con PRs pequeños y adaptadores; mantener interfaces estables (`AIState`, `ActionDef`).
- Sobrecarga de cambios en UI → mantener renderer antiguo y construir nuevo en paralelo.

Siguiente paso automático: crear tipos iniciales `Inventory`, `Item`, `Job` en `internal/simulation/df/components.go` y ejecutar la batería de tests.
