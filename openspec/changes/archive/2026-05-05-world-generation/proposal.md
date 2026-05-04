# Proposal: World Generation

## Intent

Implementar generación procedural de mapas mundiales para Evociv-RL usando Perlin Noise 2D con fBm, grid de tiles 2D separado de ECS, asignación data-driven de biomas por rangos, y visualización navegable en TUI.

## Scope

### In Scope
- Algoritmo Perlin 2D + fBm (implementación propia, sin dependencias externas)
- WorldMap grid 2D (slice contiguo, acceso por TileAt(x,y), separado de ECS)
- Generación en 4 fases: altura → humedad → temperatura → bioma
- Biomas data-driven con rangos (symbol, min/max de height/humidity/temperature)
- Config de generación en YAML (octaves, lacunarity, gain, scale, seed)
- Pantalla de mapa TUI con navegación wasd/flechas y colores por bioma
- Tecla 'm' para alternar entre bienvenida y mapa
- Persistencia seed-based en SQLite (tabla worlds)
- Tests: ruido, generación, grid, biomas, UI (mapa + navegación)

### Out of Scope
- Zoom in/out del mapa
- Mini-mapa
- Ríos y cuerpos de agua procedurales
- Mundo toroidal (wrapping)
- Inspector de tile (info detallada al seleccionar)
- LOD (Level-of-Detail) para simulación
- Mundo esférico o proyección cartográfica

## Capabilities

### New Capabilities
- `world-gen`: Generación procedural de mundo — Perlin 2D + fBm, grid WorldMap, asignación de biomas por rangos, configuración data-driven en YAML
- `tui-map`: Pantalla de mapa TUI con navegación wasd/flechas, renderizado de tiles coloreados por bioma, cambio de pantalla con tecla 'm'

### Modified Capabilities
- `store-sqlite`: Extender la interfaz Store con métodos SaveWorld/LoadLatestWorld para persistencia seed-based de mundos generados

## Approach

Seis fases secuenciales: (1) implementar Perlin 2D + fBm en `internal/world/gen/noise.go`, (2) construir WorldMap grid en `internal/world/` (types.go, worldmap.go), (3) sistema de biomas data-driven + gen-config.yaml, (4) TUI map screen con navegación, (5) persistencia seed en SQLite, (6) integración en `cmd/evociv/main.go`. Cada fase incluye sus tests.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/world/gen/noise.go` | New | Perlin 2D + fBm |
| `internal/world/gen/generate.go` | New | Pipeline 4 fases |
| `internal/world/types.go` | New | Tile, WorldMap, Coord |
| `internal/world/worldmap.go` | New | Grid + TileAt/InBounds |
| `internal/world/gen/genconfig.go` | New | Config types |
| `internal/ui/model.go` | Modify | Screen state, worldMap, cámara |
| `internal/ui/view.go` | Modify | Renderizado de mapa |
| `internal/ui/update.go` | Modify | Manejo de wasd/flechas/m |
| `data/biomes.yaml` | Modify | Extender con symbol y rangos |
| `data/gen-config.yaml` | New | Config de generación |
| `internal/store/store.go` | Modify | Extender interface |
| `internal/store/sqlite.go` | Modify | SaveWorld/LoadWorld |
| `cmd/evociv/main.go` | Modify | Pipeline init + generación |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Perlin Go puro lento (>100ms en 1024x1024) | Medium | Default 256x256 (~6ms); benchmark y optimizar si necesario |
| Artefactos en Perlin casero | Medium | Golden tests contra valores conocidos |
| Lipgloss per-tile lento en terminal | Low | Pre-calcular ANSI sequences por tile |
| Terminal <80x24 no usable | Low | Adaptar al WindowSizeMsg |

## Rollback Plan

`git revert` del cambio completo. Restaurar `openspec/specs/store-sqlite/spec.md` si fue modificado. Eliminar `data/gen-config.yaml`. Restaurar `data/biomes.yaml` original.

## Dependencies

Ninguna externa. Requiere `ecs-core` (Position de entidades), `data-loader` (cargar YAML), `store-sqlite` (persistencia), `tui-welcome` (base modelo TUI).

## Success Criteria

- [ ] Perlin 2D produce output reproducible (mismo seed = mismo mapa)
- [ ] WorldMap genera grid NxN con tiles accesibles por coordenada
- [ ] Biomas se asignan correctamente según rangos de altura/humedad/temperatura
- [ ] Mapa TUI navegable con wasd/flechas, tiles coloreados por bioma
- [ ] Tecla 'm' alterna entre welcome y mapa
- [ ] Seed persiste en SQLite y permite regenerar el mismo mundo
- [ ] Todos los tests pasan: `go test ./...`
