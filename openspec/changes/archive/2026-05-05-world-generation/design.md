# Design: World Generation

## Technical Approach

Generación procedural de mapas con Perlin Noise 2D + fBm en Go puro, grid de tiles 2D como slice contiguo (separado de ECS), pipeline en 4 fases con asignación data-driven de biomas por rangos, visualización TUI navegable con cámara 2D simple, y persistencia seed-based en SQLite.

## Architecture Decisions

### Decision: Grid separado de ECS
| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| Tiles como entidades ECS | Overhead enorme (65536+ entities), Position component innecesario | ❌ |
| Slice contiguo `[]Tile` | Acceso O(1), cache-friendly, sin indirección | ✅ |

WorldMap es un struct de datos puro. Si una entidad necesita su tile, referencia vía `Coord`.

### Decision: Perlin implementation propia
| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| Librería externa (e.g. `github.com/aquilax/go-perlin`) | Dependencia más, control limitado, licencia | ❌ |
| Implementación propia | Sin dependencias, control total, ~60 líneas | ✅ |

Basada en algoritmo clásico de Perlin (2002). Permutation table de 256 derivada del seed vía `rand.NewSource(seed)`.

### Decision: Biomas data-driven por rangos
| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| Reglas hardcodeadas | Rígido, requiere recompilar para ajustar | ❌ |
| Rangos en YAML desde Registry | Fácil de ajustar, consistente con data-loader existente | ✅ |

Cada bioma define `minHeight, maxHeight, minHumidity, maxHumidity, minTemperature, maxTemperature`. Asignación: primer match en orden de definición. Fallback: biome `unknown`.

### Decision: Cámara 2D simple
| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| Bubbletea viewport + lipgloss scrolling | Overkill, más superficie de bugs | ❌ |
| Camera{ X, Y int } + offset manual | Control total, ~20 líneas | ✅ |

Cámara limita tiles visibles según terminal size. Render por línea completa (prerenderizada como ANSI string) para evitar overhead de lipgloss por tile.

### Decision: Persistencia seed-based (no mapa completo)
| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| Guardar grid completo en SQLite | ~1MB para 256x256, lento, no escala | ❌ |
| Guardar solo seed + dimensiones | Rápido, reproducible (determinista) | ✅ |

El mapa se regenera siempre desde seed + GenConfig. Si en futuro se necesita modificación del mundo, se agregará delta persistence.

## Data Flow

```
main.go
  │
  ├── data.Loader("data/biomes.yaml") ──► Registry["biomes"]
  ├── data.Loader("data/gen-config.yaml") ──► GenConfig
  │
  └── world.Generate(w, h, config, registry)
        │
        ├── Fase 1: noise.FBM2D(height) ──► Tile.Height
        ├── Fase 2: noise.FBM2D(humidity) ──► Tile.Humidity
        ├── Fase 3: noise.FBM2D(temp) * modulate(height) ──► Tile.Temperature
        └── Fase 4: biomeRegistry.Match(tile) ──► Tile.BiomeID
              │
              └── WorldMap{tiles[], registry}
                    │
                    └── ui.Model.screen="map"
                          ├── camera{X,Y}
                          └── renderMap() → ANSI lines
```

```
Store (SQLite)
  ├── SaveWorld(seed, w, h) ──► INSERT INTO worlds
  └── LoadLatestWorld() ──► SELECT ... ORDER BY id DESC LIMIT 1
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/world/types.go` | Create | Tile, Coord, WorldMap structs |
| `internal/world/worldmap.go` | Create | NewWorldMap, TileAt, InBounds |
| `internal/world/gen/noise.go` | Create | Perlin2D, FBM2D, permutation table |
| `internal/world/gen/generate.go` | Create | Generate pipeline 4 fases |
| `internal/world/gen/genconfig.go` | Create | GenConfig struct + LoadGenConfig + Validate |
| `data/biomes.yaml` | Modify | Add symbol, min/max height/humidity/temperature |
| `data/gen-config.yaml` | Create | Default gen params (octaves, scale, etc.) |
| `internal/ui/model.go` | Modify | Add screen, camera, worldMap fields; handle wasd/m/q |
| `internal/ui/view.go` | Modify | Add renderMap(), conditional dispatch en renderView |
| `internal/store/store.go` | Modify | Add SaveWorld, LoadLatestWorld to interface |
| `internal/store/sqlite.go` | Modify | Implement SaveWorld, LoadLatestWorld, auto-migrate worlds table |
| `cmd/evociv/main.go` | Modify | Load gen-config, call Generate, pass WorldMap to Model |
| `internal/world/types_test.go` | Create | Test TileAt index, InBounds |
| `internal/world/gen/noise_test.go` | Create | Golden tests for Perlin2D + FBM2D |
| `internal/world/gen/generate_test.go` | Create | Test 4-fase pipeline, seed reproducibility |
| `internal/data/loader_test.go` | Extend | Test biomes range fields loading |
| `internal/ui/model_test.go` | Extend | Test wasd, 'm' toggle, edge bounds |
| `internal/ui/view_test.go` | Extend | Test map rendering output |
| `internal/store/sqlite_test.go` | Extend | Test SaveWorld + LoadLatestWorld |

## Interfaces / Contracts

### World Types
```go
// internal/world/types.go
type Coord struct{ X, Y int }

type Tile struct {
    Height, Humidity, Temperature float64
    BiomeID                       string
}

type WorldMap struct {
    Width, Height int
    Tiles         []Tile
    BiomeRegistry *data.Registry // for biome lookup
}
func NewWorldMap(w, h int) *WorldMap
func (m *WorldMap) TileAt(x, y int) *Tile
func (m *WorldMap) InBounds(x, y int) bool
```

### Gen Config
```go
// internal/world/gen/genconfig.go
type GenConfig struct {
    Seed      int64   `yaml:"seed"`
    Width     int     `yaml:"width"`
    Height    int     `yaml:"height"`
    Octaves   int     `yaml:"octaves"`
    Lacunarity float64 `yaml:"lacunarity"`
    Gain      float64 `yaml:"gain"`
    Scale     float64 `yaml:"scale"`
}
func LoadGenConfig(path string, fsys fs.FS) (GenConfig, error)
func (c GenConfig) Validate() error
```

Default vals: width=256, height=256, octaves=6, lacunarity=2.0, gain=0.5, scale=100.0, seed=0.

### Noise
```go
// internal/world/gen/noise.go
func Perlin2D(x, y, scale float64, seed int64) float64
func FBM2D(x, y float64, octaves int, lacunarity, gain, scale float64, seed int64) float64
```

### Generate Pipeline
```go
// internal/world/gen/generate.go
func Generate(w, h int, config GenConfig, biomeRegistry *data.Registry) (*WorldMap, error)
```

### Store Extensions
```go
// internal/store/store.go (modified)
type Store interface {
    Open(path string) error
    Close() error
    Health() error
    Path() string
    SaveWorld(seed int64, width, height int) error
    LoadLatestWorld() (seed int64, width, height int, err error)
}
```

### UI Model Fields (new)
```go
// internal/ui/model.go (modified)
type Model struct {
    ready, quitting bool
    width, height   int
    screen          string // "welcome" | "map"
    cameraX, cameraY int
    worldMap        *world.WorldMap
}
```

### Biomas YAML (extended)
```yaml
kind: biomes
data:
  - id: ocean
    name: Océano
    symbol: '~'
    color: "#1E90FF"
    minHeight: -1.0
    maxHeight: 0.1
    minHumidity: 0.0
    maxHumidity: 1.0
    minTemperature: 0.0
    maxTemperature: 1.0
  # ... more biomes with same range fields
```

### Gen Config YAML
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

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Perlin2D + FBM2D | Golden values: `Perlin2D(0,0,100,42)` debe ser consistente. fBm con octavas distintas produce valores distintos. Mismo seed → mismo output. |
| Unit | WorldMap/TileAt | Índice correcto `y*width+x`. InBounds rechaza negativos y out-of-range. |
| Unit | Generate pipeline | 4 fases completan sin error. Tiles tienen BiomeID no vacío. Seeds distintos producen mapas distintos. Width≤0 retorna error. |
| Unit | GenConfig | Validación rechaza width≤0. Defaults se aplican correctamente. |
| Unit | Biome assignment | Tile con valores dentro de rango asigna biome correcto. Tile sin match asigna "unknown". |
| Integration | Store SaveWorld/LoadLatestWorld | Guardar y recuperar seed. Empty store retorna error. |
| Integration | TUI model | Msg wasd mueve cámara. Msg 'm' cambia screen. Borde no se pasa. |
| E2E | Visual map rendering | TestViewContainsMapContent verifica que renderMap() produce salida. |

## Migration / Rollout

No migration required. El cambio es aditivo — las tablas existentes (`evociv.db`) no se modifican. La tabla `worlds` se crea vía `CREATE TABLE IF NOT EXISTS` en el primer `SaveWorld`.

## Open Questions

Ninguna. Todas las decisiones están cubiertas por los ADRs y specs. El diseño es implementable secuencialmente por fase.
