# Tasks: World Generation

## Task 1 — WorldMap Grid ✅

**Archivos**: `internal/world/types.go`, `internal/world/worldmap.go`, `internal/world/types_test.go`

**Qué hace**: Define `Coord{X,Y int}`, `Tile{Height,Humidity,Temperature float64; BiomeID string}`, `WorldMap{Width,Height int; Tiles []Tile}`. Implementa `NewWorldMap(w,h)` con slice contiguo `make([]Tile, w*h)`, `TileAt(x,y) *Tile` con index `y*width+x`, `InBounds(x,y) bool`.

**Tests** (TDD RED→GREEN):
- `TestNewWorldMap`: size correcto, tiles no nil
- `TestTileAt`: acceso index `y*w+x`, zero value por defecto, escritura+lectura consistente
- `TestTileAtNil`: TileAt devuelve nil si out-of-bounds (o panic recovery)
- `TestInBounds`: negativos y >= dim retornan false; esquinas retornan true

## Task 2 — Perlin Noise 2D ✅

**Archivos**: `internal/world/gen/noise.go`, `internal/world/gen/noise_test.go`

**Qué hace**: Implementa Perlin 2D clásico (2002) con permutation table de 256 derivada del seed vía `rand.NewSource(seed)`. `Perlin2D(x, y, scale float64, seed int64) float64` → valor [-1,1]. `FBM2D(x, y float64, octaves int, lacunarity, gain, scale float64, seed int64) float64` → suma ponderada de octavas.

**Tests** (TDD RED→GREEN + golden):
- `TestPerlinDeterministic`: mismo seed → misma salida en (0,0), (0.5, 1.5), (-3, 7)
- `TestPerlinSeedsDiffer`: seeds 42 vs 99 producen valores distintos
- `TestFBMDeterministic`: mismo seed + params → mismo output
- `TestFBMOctavesIncreaseDetail`: 1 octava vs 4 octavas producen valores distintos
- Golden test: valores conocidos para `Perlin2D(0,0,100,42)` en `testdata/perlin.golden` — comparar con `go test -update`

## Task 3 — GenConfig ✅

**Archivos**: `internal/world/gen/genconfig.go`, `data/gen-config.yaml`, `internal/world/gen/genconfig_test.go`

**Qué hace**: `GenConfig` struct con Seed, Width, Height, Octaves, Lacunarity, Gain, Scale. `LoadGenConfig(path string, fsys fs.FS) (GenConfig, error)` — carga YAML con kind: gen-config. `Validate() error` — rechaza width≤0, height≤0; defaults: octaves=6, lacunarity=2.0, gain=0.5, scale=100.0.

`data/gen-config.yaml`:
```yaml
kind: gen-config
data:
  seed: 42
  width: 256
  height: 256
  octaves: 6
  lacunarity: 2.0
  gain: 0.5
  scale: 100.0
```

**Tests** (TDD RED→GREEN):
- `TestGenConfigDefaults`: config parcial con solo width=64,height=64 → defaults aplicados
- `TestGenConfigValidate`: width=0 retorna error; valores válidos retorna nil
- `TestLoadGenConfig`: YAML válido se carga correctamente
- `TestLoadGenConfigInvalidYAML`: YAML malformado retorna error

## Task 4 — Biomas Extendidos ✅

**Archivos**: `data/biomes.yaml`, `internal/world/gen/generate.go` (lógica de matching), `internal/world/gen/generate_test.go`

**Qué hace**: Extiende `data/biomes.yaml` añadiendo `symbol` (rune), `minHeight`, `maxHeight`, `minHumidity`, `maxHumidity`, `minTemperature`, `maxTemperature` por bioma. Biomas: ocean, plains, forest, desert, tundra, jungle. Lógica `MatchBiome(tile Tile, biomas []BiomeDef) string`: primer match en orden de definición, fallback `"unknown"`.

**Tests** (TDD RED→GREEN):
- `TestBiomeMatchPlains`: tile con h=0.3, hu=0.4, t=0.5 dentro de rangos plains → assigns plains
- `TestBiomeMatchUnknown`: tile height=100.0 sin match → unknown
- `TestBiomeMatchOcean`: tile con height negativa dentro de rango ocean → assigns ocean
- `TestBiomeLoadRanges`: biomas cargados desde YAML contienen campos symbol/min/max

## Task 5 — Pipeline de Generación ✅

**Archivos**: `internal/world/gen/generate.go`, `internal/world/gen/generate_test.go`

**Qué hace**: `Generate(w, h int, config GenConfig, biomeRegistry *data.Registry) (*WorldMap, error)`. 4 fases secuenciales:
1. **Altura**: `FBM2D(x,y, octaves, lacunarity, gain, scale, seed)` range [-1,1] → `tile.Height`
2. **Humedad**: `FBM2D(x,y, octaves, lacunarity, gain, scale, seed+1)` → `tile.Humidity`
3. **Temperatura**: `FBM2D(...seed+2) * ((tile.Height+1)/2)` (modula con altura) → `tile.Temperature`
4. **Bioma**: MatchBiome(tile, loadedBiomes) → `tile.BiomeID`

**Tests** (TDD RED→GREEN):
- `TestGenerateFillsAllTiles`: Generate(3,3) produce 9 tiles con BiomeID no vacío
- `TestGenerateSeedsDiffer`: mismo tamaño, seeds distintos → arrays Height diferentes
- `TestGenerateInvalidParams`: width=0 retorna error
- `TestGenerateReproducible`: mismo seed + params → idéntico WorldMap (tile a tile)
- `TestGenerateTemperatureModulated`: temperatura varía con altura (no es independiente)

## Task 6 — Store SQLite Persistencia ✅

**Archivos**: `internal/store/store.go`, `internal/store/sqlite.go`, `internal/store/sqlite_test.go`

**Qué hace**: Extiende `Store` interface con `SaveWorld(seed int64, width, height int) error` y `LoadLatestWorld() (seed int64, width, height int, error)`. En `SQLiteStore`: migración automática con `CREATE TABLE IF NOT EXISTS worlds (id INTEGER PRIMARY KEY AUTOINCREMENT, seed INTEGER NOT NULL, width INTEGER NOT NULL, height INTEGER NOT NULL, created_at TEXT DEFAULT (datetime('now')))`. `SaveWorld` hace `INSERT`, `LoadLatestWorld` hace `SELECT ... ORDER BY id DESC LIMIT 1`.

**Tests** (TDD RED→GREEN):
- `TestSaveAndLoadWorld`: SaveWorld(42,64,64) → LoadLatestWorld() retorna (42,64,64)
- `TestLoadEmptyStore`: store nuevo sin mundos → error
- `TestSaveMultipleLoadLatest`: guardar 3 mundos → LoadLatestWorld retorna el último
- `TestIdempotentMigration`: Open dos veces no falla (IF NOT EXISTS)

## Task 7 — TUI Model Extendido ✅

**Archivos**: `internal/ui/model.go`, `internal/ui/model_test.go`

**Qué hace**: Añade campos a `Model`: `screen string` ("welcome" | "map"), `cameraX, cameraY int`, `worldMap *world.WorldMap`. En `Update`: tecla `'m'` togglea screen; `'w'/'a'/'s'/'d'` y flechas mueven cámara (con bounds clamping: cameraX ≥ 0, cameraY ≥ 0, cameraX < map.Width, cameraY < map.Height); `'q'` sigue saliendo. Requiere injectar WorldMap vía `SetWorldMap(wm *world.WorldMap)` o campo exportado.

**Tests** (TDD RED→GREEN):
- `TestToggleToMap`: screen="welcome", press 'm' → screen="map"
- `TestToggleToWelcome`: screen="map", press 'm' → screen="welcome"
- `TestCameraMoveWASD`: camera(0,0), press 'd' → cameraX=1; press 's' → cameraY=1
- `TestCameraBounds`: camera en borde derecho, press 'd' → no cambia, no panic
- `TestQuitStillWorks`: press 'q' → quitting=true, tea.Quit
- `TestArrowKeysWork`: press Right → cameraX+1; press Down → cameraY+1

## Task 8 — TUI View Mapa ✅

**Archivos**: `internal/ui/view.go`, `internal/ui/view_test.go`, `internal/ui/testdata/map.golden`

**Qué hace**: Modifica `renderView` para despachar según `m.screen`: "welcome" → render existente; "map" → `renderMap(m)`. `renderMap` genera líneas ANSI: itera filas visibles (desde cameraY hasta cameraY+termHeight), por cada fila itera columnas (cameraX a cameraX+termWidth), renderiza `symbol` del bioma con color lipgloss. Borde del mundo: tiles fuera de rango muestran espacio. Pre-renderiza líneas completas como strings ANSI para performance.

**Tests** (TDD RED→GREEN + golden):
- `TestRenderMapShowsTiles`: view con worldMap 3x3 y screen="map" produce salida con símbolos
- `TestRenderScreenToggle`: renderView con screen="welcome" muestra título; screen="map" no
- `TestRenderMapNoWorldMap`: screen="map" sin worldMap → no panic, mensaje "Generando mundo..."
- Golden test: `testdata/map.golden` con mapa 5x5 conocido, comparar con `go test -update`
- `TestRenderMapWindowAdapt`: WindowSizeMsg(80,24) → renderMap produce ~24 líneas

## Task 9 — main.go Integración ✅

**Archivos**: `cmd/evociv/main.go`, `cmd/evociv/main_test.go`

**Qué hace**: En `run()` después de cargar data: (1) cargar `data/gen-config.yaml` con `gen.LoadGenConfig`, (2) obtener biomas del registry con `data.Get[[]any]`, (3) llamar `world.Generate(w, h, config, registry)`, (4) crear `ui.NewModel()` con worldMap inyectado, (5) iniciar TUI. Manejar errores: si Generate falla, log.Warn y continuar sin mapa (nil worldMap).

**Tests**:
- `TestMainBuild`: `go build ./cmd/evociv/` produce binario sin error
- `TestMainInitRuns`: smoke test que run() no panic con data mínima (usando MapFS mock)
- `TestGenerateWithConfig`: integración rápida: cargar gen-config de prueba, Generate, verificar tiles no vacíos

---

## Resumen de Dependencias

```
Task 1 (WorldMap) ──────┐
                         ├──→ Task 5 (Pipeline) ──→ Task 9 (main.go)
Task 2 (Perlin) ────────┘         │
                         ┌────────┘
Task 3 (GenConfig) ──────┤
                         ├──→ Task 5 (Pipeline)
Task 4 (Biomas) ─────────┘

Task 6 (Store) ─── independiente, depende de store.go existente

Task 7 (TUI Model) ←── depende de Task 1 (WorldMap)
Task 8 (TUI View)  ←── depende de Task 7 (Model con screen/worldMap)
Task 9 (main.go)   ←── depende de Tasks 3, 4, 5, 7
```

**Orden recomendado**: 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8 → 9
(6 puede correr en paralelo con 1-5 por ser independiente)
