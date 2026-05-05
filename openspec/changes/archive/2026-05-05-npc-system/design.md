# Design: NPC System

## Technical Approach

Sistema ECS de NPCs con 6 componentes, spawner biome-weighted seed-based (offset +999), 4 sistemas ECS (Spawn, Wander, LOD, Render), overlay TUI con símbolo '@' + inspector modal, persistencia seed-only (nada de SQLite para entidades). Data-driven vía YAML cargado con el Registry existente. LOD 3 niveles (0=off, 1=near, 2=local) para escalar a 100+ NPCs.

## Architecture Decisions

### Decision: 6 componentes vs menos en MVP
| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| 4 componentes (Health, Job, Appearance, LOD) | Menos código pero AIState necesario para Wander, Personality para futuro GOAP | ❌ |
| **6 componentes** (Health, Personality, Job, AIState, Appearance, LOD) | Cada componente es ~5-10 líneas; la estructura completa evita refactors futuros | ✅ |

Los 6 componentes se alinean con spec `npc-components`. Inventory y Family quedan post-MVP.

### Decision: Spawner biome-weighted vs uniforme
| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| Uniforme | Simple pero NPCs en océano, sin sentido lúdico | ❌ |
| **Biome-weighted** | Coherente con el mundo: farmers en plains, hunters en forest | ✅ |

Pesos: plains/forest=1.0, tundra/desert=0.2, ocean/jungle=0.0. Seed determinista: `rand.New(rand.NewSource(seed + 999))`.

### Decision: LOD 3 niveles vs 2
| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| 2 niveles (visible/no visible) | Menos código, pero sin escalabilidad futura | ❌ |
| **3 niveles** (local/near/distant) | Diferencial tick rate: local cada tick, near cada 5, distant cada 20 | ✅ |

Distancias Chebyshev: ≤5 → local, ≤15 → near, >15 → distant (spec `npc-systems`). Wander solo en local/near.

### Decision: Seed-based vs serialización de entidades
| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| Full state SQLite | Estado completo preservado, pero alta complejidad sync ECS↔DB | ❌ |
| **Seed-based** | Cero serialización, determinista, regeneración idéntica | ✅ |

Solo se persiste `npc_seed_offset` en tabla worlds. Post-MVP se migrará a tablas `npc_entities` + `npc_component_data`.

### Decision: Overlay '@' vs tile replacement
| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| Reemplazar biome symbol | Se pierde info del biome en tiles con NPCs | ❌ |
| **Overlay '@' sobre biome** | Se ve NPC + biome de fondo vía lipgloss layering | ✅ |

## Data Flow

```
main.go
  │
  ├── data.Loader("data/npcs.yaml") ──► Registry["npc-races"]
  ├── data.Loader("data/npc-roles.yaml") ──► Registry["npc-roles"]
  │
  └── ecs.World.Update(dt)
        │
        ├── [1st tick] NPCSpawnSystem
        │     └── spawner.Spawn(raceDefs, roleDefs, seed+999)
        │           ├── Para cada NPC: elegir raza→rol→posición biome-weighted
        │           ├── Crear Entity + 6 componentes
        │           └── Registrar en World
        │
        ├── [every tick] LODSystem
        │     └── Calcular Chebyshev distance a playerPos → LOD.Level
        │
        ├── [if LOD≥1] WanderSystem
        │     └── Mover NPC a tile adyacente compatible con rol
        │
        └── NPCRenderSystem
              └── Generar []NPCRenderInfo para TUI (símbolo + posición)
                    │
                    └── ui.Model.npcOverlay → renderOverlay() dibuja '@'
```

```
Persistencia:
  SaveWorld(seed, w, h, npcSeedOffset) ──► INSERT INTO worlds
  LoadLatestWorld() ──► SELECT seed, width, height, npc_seed_offset
```

## Componentes ECS

### `internal/simulation/npc/components.go`

```go
package npc

import "math/rand"

// Health — vitalidad del NPC. float64 para compatibilidad con futuras
// mecánicas de daño porcentual y regeneración.
type Health struct {
    Current, Max float64
}

// Personality — Big Five traits normalizados [0.0, 1.0].
// Generación Gaussiana truncada: rand.NormFloat64() * std + mean, clamp [0,1].
type Personality struct {
    Openness        float64 // apertura a experiencia
    Conscientiousness float64 // escrupulosidad
    Extraversion    float64 // extraversión
    Agreeableness   float64 // amabilidad
    Neuroticism     float64 // neuroticismo (inverso a estabilidad)
}

// NewPersonality genera traits determinísticos desde un RNG seed por entidad.
func NewPersonality(rng *rand.Rand) Personality { /* ... */ }

// Job — ocupación del NPC. Role string para data-driven lookup.
type Job struct {
    Role string // "farmer", "hunter", "merchant", etc.
}

// AIState — estado cognitivo GOAP-ready. En MVP solo almacena wandering plan.
type AIState struct {
    Goals       []string   // IDs de metas ("wander", "rest")
    Plan        []string   // acciones planificadas
    Mood        float64    // ánimo agregado [-1, 1]
}

// Appearance — representación visual en el mapa.
// Depende de race+role: race define base symbol, role puede override color.
type Appearance struct {
    Symbol rune              // '@', '☺', etc.
    Color  lipgloss.Color    // color.Attribute o hex string
}

// LOD — Level of Detail para simulación diferencial.
type LOD struct {
    Level int  // 0=distant(off), 1=near, 2=local
}
```

Constants:
```go
const (
    LODDistant = 0
    LODNear    = 1
    LODLocal   = 2
)
```

## Data-Driven YAML

### `data/npcs.yaml`

```yaml
kind: npc-races
data:
  - id: human
    name: Humano
    description: "Versátil y adaptable"
    spawn_weight: 1.0
    traits:
      openness:        { mean: 0.5, std: 0.15 }
      conscientiousness: { mean: 0.5, std: 0.15 }
      extraversion:    { mean: 0.5, std: 0.15 }
      agreeableness:   { mean: 0.5, std: 0.15 }
      neuroticism:     { mean: 0.5, std: 0.15 }
    roles:
      - id: farmer
        weight: 0.4
      - id: hunter
        weight: 0.3
      - id: merchant
        weight: 0.2
      - id: artisan
        weight: 0.1
    name_pool:
      first: ["Aldric", "Bran", "Cedric", "Doran", "Elara", "Fina", "Greta"]
      last:  ["Torres", "Río", "Piedra", "Valle", "Bosque"]
  - id: dwarf
    name: Enano
    spawn_weight: 0.6
    traits:
      conscientiousness: { mean: 0.7, std: 0.1 }
      extraversion:      { mean: 0.3, std: 0.1 }
    roles:
      - id: miner
        weight: 0.6
      - id: smith
        weight: 0.4
    name_pool:
      first: ["Borin", "Durm", "Gimli", "Kili"]
      last:  ["Hierro", "Roca", "Martillo"]
  - id: elf
    name: Elfo
    spawn_weight: 0.3
    traits:
      openness:   { mean: 0.7, std: 0.1 }
      neuroticism: { mean: 0.3, std: 0.1 }
    roles:
      - id: hunter
        weight: 0.5
      - id: artisan
        weight: 0.5
    name_pool:
      first: ["Aeris", "Cael", "Elros", "Lúthien", "Thael"]
      last:  ["Bosque", "Luna", "Viento", "Arco"]
```

### `data/npc-roles.yaml`

```yaml
kind: npc-roles
data:
  - id: farmer
    symbol: "@"
    color: "#FFD700"
    biomes: [plains]
  - id: hunter
    symbol: "@"
    color: "#8B4513"
    biomes: [forest, jungle]
  - id: merchant
    symbol: "@"
    color: "#00CED1"
    biomes: [plains]
  - id: artisan
    symbol: "@"
    color: "#FF69B4"
    biomes: [plains]
  - id: miner
    symbol: "☺"
    color: "#C0C0C0"
    biomes: [plains, desert]
  - id: smith
    symbol: "☺"
    color: "#FF4500"
    biomes: [plains]
```

El loader existente `data.Loader.LoadAll("data", registry)` carga ambos archivos automáticamente. Se define tipo `NpcRaceDef` y `NpcRoleDef` con unmarshalling directo (similar a `BiomeDef` en `internal/world/gen/biomes.go`).

## Spawner

### `internal/simulation/npc/spawner.go`

```go
package npc

type SpawnConfig struct {
    Count    int     // 0 = auto (Density * worldArea)
    Density  float64 // default 0.0015 (~98 NPCs en 256×256)
}

func Spawn(w *ecs.World, wm *world.WorldMap, config SpawnConfig, seed int64,
    raceDefs []RaceDef, roleDefs []RoleDef) error
```

Algoritmo:
1. `count := config.Count`; si es 0 → `count = int(float64(wm.Width*wm.Height) * config.Density)` clamp [50, 100]
2. `rng := rand.New(rand.NewSource(seed + 999))`
3. Por cada NPC a spawnear:
   a. Elegir raza por `spawn_weight` ponderado
   b. Elegir rol dentro de raza por `weight` ponderado
   c. Samplear posición: random tile `(rng.Intn(wm.Width), rng.Intn(wm.Height))`, verificar biome weight > 0, verificar no ocupado (via ComponentStore[Position]), reintentar hasta 100 fallos
   d. Generar Personality con `NewPersonality(rng)` — Gaussiana truncada [0,1] por trait
   e. Crear Entity → AddComponent para Position, Name (from name_pool), Health, Personality, Job, AIState, Appearance (symbol/color desde roleDef), LOD{Level: LODLocal}

## Sistemas

### `internal/simulation/npc/systems.go`

**NPCSpawnSystem**: `spawned bool` interna. En `Update()`: si `!spawned`, llama `Spawn(...)`, setea `spawned=true`. No-op en ticks subsiguientes.

**WanderSystem**: Itera entidades con LOD≥1. Para cada NPC:
- Obtener Position, Job, AIState
- Si no tiene plan, generar 1 wander goal: elegir tile aleatorio adyacente (8-dir) cuyo biome esté en `roleDef.Biomes`
- Si hay plan, ejecutar: mover Position paso a paso
- Si no hay tile compatible, skip (NPC se queda)

**LODSystem**: Obtiene `playerPos()` callback. Para cada entidad con LOD+Position:
- `dist := chebyshev(posX, posY, playerX, playerY)`
- Asignar: ≤5→LODLocal, ≤15→LODNear, >15→LODDistant

**NPCRenderSystem**: Itera entidades con LOD≥1. Genera `[]NPCRenderInfo{Symbol, Color, WorldX, WorldY}` para el TUI. No renderiza directamente.

## TUI Overlay + Inspector

### `internal/ui/model.go` — nuevos campos:
```go
type Model struct {
    // ... existentes ...
    ecsWorld      *ecs.World
    cursorX       int
    cursorY       int
    npcOverlay    []NPCRenderInfo  // desde NPCRenderSystem
    inspectorOpen bool
    selectedNPC   ecs.Entity
}
```

### `internal/ui/view.go` — nuevas funciones:
```go
func renderOverlay(m Model, worldX, worldY int) string {
    // Busca en m.npcOverlay si hay NPC en (worldX, worldY)
    // Si sí: renderiza symbol con color (lipgloss.NewStyle().Foreground())
    // Si no: retorna el biome symbol normal
    // Nota: Overlay se renderiza sobre el tile biome subyacente
}

func renderInspector(m Model) string {
    // Si !inspectorOpen → retorna ""
    // Si selectedNPC válido → muestra panel derecho con:
    //   Name, Health (Current/Max), Job, Personality (O/C/E/A/N truncados a 2 decimales), Biome
}
```

RenderMap modificado: cada tile llama `renderOverlay()` que devuelve el NPC symbol + color si hay NPC, o vacío para que el biome symbol se muestre normal. El overlay se pinta `encima` del biome.

### Update — nuevas teclas:
| Tecla | Acción |
|-------|--------|
| `e` | Abre inspector en tile bajo cursor (cursorX, cursorY) |
| `q` o `esc` | Cierra inspector (si está abierto) |
| Flechas | Mueven cursor cuando inspector abierto (cursorX/Y, no camera) |

Inspección: busca entidad en `npcOverlay` que coincida con `(cursorX+cameraX, cursorY+cameraY)`. Si encuentra, setea `selectedNPC` y `inspectorOpen=true`.

## Persistencia

### Tabla worlds extendida:

```sql
CREATE TABLE IF NOT EXISTS worlds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    seed INTEGER NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    npc_seed_offset INTEGER DEFAULT 999,
    created_at TEXT DEFAULT (datetime('now'))
);
```

### `internal/store/store.go` — interfaz modificada:
```go
type Store interface {
    // ... existentes ...
    SaveWorld(seed int64, width, height int, npcSeedOffset int64) error
    LoadLatestWorld() (seed int64, width, height int, npcSeedOffset int64, err error)
}
```

Migration: `ALTER TABLE worlds ADD COLUMN npc_seed_offset INTEGER DEFAULT 999` si no existe (vía `PRAGMA table_info` o `CREATE TABLE IF NOT EXISTS` con columna incluida — como es CREATE siempre se ejecuta, la columna ya está en el CREATE).

**Enfoque**: `worlds` se crea con `npc_seed_offset` desde el inicio en el migrate. Si la tabla ya existe sin la columna, se agrega con ALTER TABLE. Esto maneja bases existentes.

### `internal/store/sqlite.go` — cambios:
- `SaveWorld`: INSERT incluye `npc_seed_offset`
- `LoadLatestWorld`: SELECT incluye `npc_seed_offset`
- `migrate`: CREATE TABLE con columna + migración condicional para bases existentes

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/simulation/npc/components.go` | Create | 6 structs ECS + NewPersonality() |
| `internal/simulation/npc/types.go` | Create | RaceDef, RoleDef, SpawnConfig, NPCRenderInfo |
| `internal/simulation/npc/spawner.go` | Create | Spawn() biome-weighted seed-based |
| `internal/simulation/npc/systems.go` | Create | NPCSpawnSystem, WanderSystem, LODSystem, NPCRenderSystem |
| `data/npcs.yaml` | Create | Razas con traits, roles, name_pool |
| `data/npc-roles.yaml` | Create | Roles con símbolo, color, biomas compatibles |
| `internal/ecs/component.go` | Modify | Registrar ComponentID para los 6 nuevos tipos |
| `internal/ecs/world.go` | Modify | RemoveEntity extensible para nuevos types |
| `internal/ui/model.go` | Modify | Añadir ecsWorld, cursorX/Y, npcOverlay, inspectorOpen, selectedNPC |
| `internal/ui/view.go` | Modify | renderOverlay(), renderInspector(), renderMap extendido |
| `internal/ui/update.go` | Modify | Teclas 'e', flechas en modo inspector, q/esc close |
| `internal/store/store.go` | Modify | Interfaz con npcSeedOffset |
| `internal/store/sqlite.go` | Modify | Migración columna + impl nuevo Save/Load |
| `cmd/evociv/main.go` | Modify | Inicializar spawner, registrar sistemas, pasar ecsWorld a Model |
| `internal/simulation/npc/components_test.go` | Create | Tests creación + determinismo Personality |
| `internal/simulation/npc/spawner_test.go` | Create | Tests spawn count, biome weights, determinismo |
| `internal/simulation/npc/systems_test.go` | Create | Tests LOD, Wander bounds, Spawn once |
| `internal/ui/model_test.go` | Extend | Tests cursor, inspector toggle, overlay |
| `internal/ui/view_test.go` | Extend | Tests renderOverlay, renderInspector output |
| `internal/store/sqlite_test.go` | Extend | Tests npcSeedOffset save/load |
| `data/npcs.yaml` aliases | Validate | Registrar validadores en validator.go |

## Interfaces / Contracts

### RaceDef / RoleDef (loaded from YAML)
```go
type RaceDef struct {
    ID          string             `yaml:"id"`
    Name        string             `yaml:"name"`
    SpawnWeight float64            `yaml:"spawn_weight"`
    Traits      map[string]TraitDef `yaml:"traits"` // mean+std per trait
    Roles       []RoleWeight       `yaml:"roles"`
    NamePool    NamePool           `yaml:"name_pool"`
}

type RoleDef struct {
    ID     string   `yaml:"id"`
    Symbol string   `yaml:"symbol"`
    Color  string   `yaml:"color"`
    Biomes []string `yaml:"biomes"`
}
```

### NPCRenderInfo (data transfer para TUI)
```go
type NPCRenderInfo struct {
    Entity  ecs.Entity
    Symbol  rune
    Color   lipgloss.Color
    WorldX, WorldY int
}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Personality generation | Misma seed+entityID → mismos 5 traits. Traits diferentes entre entidades. Valores en [0,1]. |
| Unit | Spawner count | 256×256 → entre 50 y 100 NPCs. Misma seed → identical set. |
| Unit | Biome-weighted placement | Ocean/jungle: 0 NPCs. Plains > tundra estadísticamente. |
| Unit | LOD classification | Chebyshev dist 3→local, 10→near, 20→distant. Cambia con player move. |
| Unit | Wander bounds | NPC no sale de [0,0]→[255,255]. NPC rodeado de océano no se mueve. |
| Integration | Spawn → Render pipeline | NPCs spawn → NPCRenderSystem produce []NPCRenderInfo con posiciones correctas. |
| Integration | TUI overlay | '@' aparece en tile correcto con camera offset. Inspector muestra datos del NPC. |
| Integration | Store npcSeedOffset | SaveWorld(42,64,64,999) → LoadLatestWorld() retorna offset=999. |
| E2E | Determinismo completo | Dos corridas con misma seed producen exactamente mismos NPCs (posición, traits, rol). |

## Migration / Rollout

Migración de schema: columna `npc_seed_offset` se añade vía ALTER TABLE condicional en `migrate()`. No hay migración de datos — las bases existentes ganan la columna con DEFAULT 999. No requiere feature flags.

## Open Questions

None. Todas las decisiones están cubiertas por los ADRs y specs.
