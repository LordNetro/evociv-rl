# Design: Settlements

## Technical Approach

Asentamientos como entidades ECS en un nuevo paquete `internal/simulation/settlement/`, reutilizando el patrón existente de componentes, stores, y sistemas del paquete `npc`. Los settlements son entidades con `Position` + `Settlement`; los buildings son entidades hijas con `Position` + `Building` dentro del radio del settlement. `HomeReference` vincula NPCs a su settlement. Los datos son YAML data-driven cargados vía `data.Loader` (kind: `settlement-types`, `building-types`). El TUI overlay extiende `renderOverlay()` para priorizar NPC > Settlement > Biome.

## Package Architecture

```
internal/simulation/settlement/
├── components.go    // Settlement, Building, ResourceStore, HomeReference + IDs + RegisterStores
├── types.go         // SettlementDef, BuildingDef, SettlementRenderInfo, SettlementType constants
├── data.go          // LoadSettlementTypes(), LoadBuildingTypes() + validación
├── data_test.go     // Tests de carga YAML y validación
├── systems.go       // SettlementSpawnSystem, SettlementRenderSystem
├── systems_test.go  // Tests de spawn y render
```

Sigue el mismo patrón que `internal/simulation/npc/`: components define structs + IDs + `RegisterStores()`, types define YAML defs, data define loaders, systems define sistemas ECS.

## Componentes ECS

```go
type Settlement struct {
    Name       string
    Type       string   // "village" | "town" | "city"
    Radius     int      // tiles que abarca (3, 5, 8)
    Population int      // NPCs asignados (counter)
    Level      int      // nivel de desarrollo (1-5, MVP siempre 1)
}

type Building struct {
    ID    string   // "house" | "farm" | "market" | "tavern" | "temple" | "blacksmith"
    Name  string
    Level int      // 1-3, MVP siempre 1
}

type HomeReference struct {
    SettlementEntity ecs.Entity   // entity ID del settlement hogar
}

type ResourceStore struct {
    Resources map[string]float64  // {"food": 100, "gold": 50} — para futuro
}
```

**Component IDs** (naming consistente con `npc/`):
```go
SettlementID = ecs.NewComponentID("settlement")
BuildingID    = ecs.NewComponentID("building")
HomeRefID     = ecs.NewComponentID("home_reference")
ResourceID    = ecs.NewComponentID("resource_store")
```

**Store Registration**: `RegisterSettlementStores(w *ecs.World)` registra los 4 stores. Settlement, Building, ResourceStore usan `Position` ya registrado. Se llama en `main.go` tras `npc.RegisterStores()`.

## Data-Driven

### data/settlements.yaml (kind: settlement-types)

3 tipos con spawn_weight acumulativo. Compatible con el patrón `kind/data` de `data.Loader`.

```yaml
kind: settlement-types
data:
  - id: village
    name: Aldea
    symbol: "♦"
    color: "#8B7355"
    radius: 3
    biomes: [plains, forest]
    buildings: [house, farm]
    spawn_weight: 0.6

  - id: town
    name: Pueblo
    symbol: "▲"
    color: "#B8860B"
    radius: 5
    biomes: [plains]
    buildings: [house, market, tavern, blacksmith, farm]
    spawn_weight: 0.3

  - id: city
    name: Ciudad
    symbol: "●"
    color: "#DAA520"
    radius: 8
    biomes: [plains]
    buildings: [house, market, temple, tavern, blacksmith, farm]
    spawn_weight: 0.1
```

### data/buildings.yaml (kind: building-types)

6 edificios. Cada settlement type define qué edificios spawnear (referencia por ID).

```yaml
kind: building-types
data:
  - id: house
    name: Casa
  - id: farm
    name: Granja
  - id: market
    name: Mercado
  - id: tavern
    name: Taberna
  - id: temple
    name: Templo
  - id: blacksmith
    name: Herreria
```

### Loaders

```go
type SettlementDef struct {
    ID          string   `yaml:"id"`
    Name        string   `yaml:"name"`
    Symbol      string   `yaml:"symbol"`
    Color       string   `yaml:"color"`
    Radius      int      `yaml:"radius"`
    Biomes      []string `yaml:"biomes"`
    Buildings   []string `yaml:"buildings"`
    SpawnWeight float64  `yaml:"spawn_weight"`
}

type BuildingDef struct {
    ID   string `yaml:"id"`
    Name string `yaml:"name"`
}
```

`LoadSettlementTypes(registry)` y `LoadBuildingTypes(registry)` usan `data.Get[[]any]()`, parsean con type assertions igual que `LoadNpcRaces()`. **Validación**: spawn_weights suman ~1.0 (±0.01), buildings referenced existen, biomes son válidos.

## SettlementSpawnSystem

```go
type SettlementSpawnSystem struct {
    spawned        bool
    wm             *world.WorldMap
    seed           int64
    settlementDefs []SettlementDef
    buildingDefs   []BuildingDef
}
```

### Algoritmo

```
1. Si ya spawned → return
2. settlementRNG = rand.New(rand.NewSource(seed + 777))  // deterministic
3. buildingRNG = rand.New(rand.NewSource(seed + 888))    // deterministic
4. Calcular targetCount = 5 + settlementRNG.Intn(6)  // 5-10
5. Para i = 0..targetCount:
   a. Elegir tipo de settlement por spawn_weight (weighted random, igual que pickRace)
   b. intentos = 0, max_attempts = 200 (mundo 256×256 da espacio)
   c. Loop:
      - tile aleatorio (settlementRNG)
      - biome compatible? (tile.BiomeID in type.Biomes)
      - distancia mínima ≥ 20 tiles de otros settlements? (Chebyshev)
      - Sí → crear entidad con Position, Settlement (Name procedural), LOD
      - No → reintentar
   d. Si no encontró tile → skip
   e. Spawn buildings: para cada buildingID en type.Buildings:
      - Posición aleatoria dentro del radio del settlement (buildingRNG)
      - Crear entidad hija con Position + Building
      - entity ID tracking opcional para el settlement
6. spawned = true
```

### Nombres procedurales

Pool fijo en types.go (no YAML para MVP, extensión futura):
```go
var settlementNamePrefixes = []string{
    "Norte", "Sur", "Este", "Oeste", "Alto", "Bajo",
    "Nuevo", "Viejo", "Valle", "Monte", "Río", "Lago",
}
var settlementNameSuffixes = []string{
    "del Valle", "de la Colina", "del Bosque", "del Río",
    "Dorado", "Plateado", "Verde", "del Sol",
}
```
Combinación: `prefix[i] + " " + type.Name + " " + suffix[j]` → "Norte Aldea del Valle". Se usa el settlementRNG para seleccionar. Si hay colisión, se regenera (max 5 intentos). Truncar a 15 chars para TUI con `truncateName()`.

### Determinismo

`seed + 777` para settlements, `seed + 888` para buildings. Mismo seed global → mismos settlements, mismas posiciones, mismos nombres.

## NPC Spawn Modificado

La función `Spawn()` en `spawner.go` recibe un nuevo parámetro opcional `settlementDefs` o se detectan settlements del mundo. Estrategia:

1. **Antes de spawnear NPCs**: recolectar todos los settlements del mundo via SettlementStore.
2. **Para cada NPC** (loop existente):
   - Elegir raza y rol (sin cambios)
   - **Nuevo paso**: buscar settlement compatible:
     - farmer → settlement con farm building en sus buildings
     - merchant → town/city con market
     - hunter → settlement en biome forest (forest edge settlement)
     - priest → settlement con temple
     - blacksmith → settlement con blacksmith
     - miner → cualquier settlement en bioma compatible
     - artisan → cualquier settlement
   - Si encuentra: position aleatoria dentro del settlement radius, asignar HomeReference
   - Si no encuentra: biome-weighted random position actual (nomad fallback)
3. **Capacity check**: cada settlement tiene cap `Radius * 2` NPCs. Si alcanzó, probar siguiente settlement.

### Cambios en spawner.go

- `Spawn()` acepta `settlements []settlement.SpawnInfo` (o query desde el mundo)
- Rol-to-building mapping como `map[string][]string` (rol → building IDs)
- `HomeReference` se añade vía `ecs.AddComponent(w, e, settlement.HomeReference{SettlementEntity: settlementEntity})`
- Seed: settlement_index usado dentro del seed+999 (determinismo existente)

## TUI Overlay

### SettlementRenderInfo

```go
type SettlementRenderInfo struct {
    Entity         ecs.Entity
    Symbol         rune        // '♦', '▲', '●'
    Color          lipgloss.Color
    Name           string      // truncado a 15 caracteres
    WorldX, WorldY int
}
```

### SettlementRenderSystem

Igual que `NPCRenderSystem` pero filtra entidades con `Settlement` en vez de `Appearance`+`LOD`. Opcional: filtrar por LOD (solo settlements con LOD≥Near). Expone `RenderInfos() []SettlementRenderInfo`.

### renderOverlay order

```go
func renderOverlay(m Model, worldX, worldY int) string {
    // 1. NPC overlay (highest priority)
    for _, info := range m.npcOverlay {
        if info.WorldX == worldX && info.WorldY == worldY {
            return styledNPC(info)
        }
    }
    // 2. Settlement overlay (middle priority)
    for _, info := range m.settlementOverlay {
        if info.WorldX == worldX && info.WorldY == worldY {
            return styledSettlement(info)
        }
    }
    // 3. Biome tile (default — handled in renderMap)
    return ""
}
```

### Inspector

`tryOpenInspector()` actual: busca NPC en cursor. Se extiende: si no hay NPC, busca Settlement en cursor → abre inspector con datos del settlement (Name, Type, Radius, Population, Level, Buildings).

### Model changes

```go
type Model struct {
    // ... campos existentes
    settlementOverlay []settlement.SettlementRenderInfo
}

func (m *Model) SetSettlementOverlay(overlay []settlement.SettlementRenderInfo) {
    m.settlementOverlay = overlay
}
```

En `refreshOverlay()`: además de NPCRenderSystem, buscar SettlementRenderSystem y recolectar overlay.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/simulation/settlement/components.go` | **Create** | Settlement, Building, ResourceStore, HomeReference + IDs + RegisterSettlementStores |
| `internal/simulation/settlement/types.go` | **Create** | SettlementDef, BuildingDef, SettlementRenderInfo, SettlementType constants |
| `internal/simulation/settlement/data.go` | **Create** | LoadSettlementTypes(), LoadBuildingTypes() + validateSettlementData() |
| `internal/simulation/settlement/data_test.go` | **Create** | Tests de carga YAML con fstest.MapFS |
| `internal/simulation/settlement/systems.go` | **Create** | SettlementSpawnSystem, SettlementRenderSystem |
| `internal/simulation/settlement/systems_test.go` | **Create** | Tests de spawn (5-10 en plains, 0 en ocean) y render |
| `data/settlements.yaml` | **Create** | 3 settlement types (village, town, city) con símbolos y biomas |
| `data/buildings.yaml` | **Create** | 6 building types (house, farm, market, tavern, temple, blacksmith) |
| `internal/simulation/npc/spawner.go` | **Modify** | Settlement-aware NPC placement, HomeReference assignment, capacity check |
| `internal/simulation/npc/spawner_test.go` | **Modify** | Tests para settlement spawn, nomad fallback, capacity overflow |
| `internal/simulation/npc/systems.go` | **Modify** | NPCSpawnSystem pasa settlement data a Spawn() |
| `internal/ui/model.go` | **Modify** | settlementOverlay field, SetSettlementOverlay(), refreshOverlay() extended |
| `internal/ui/model_test.go` | **Modify** | Tests para inspector de settlements |
| `internal/ui/view.go` | **Modify** | renderOverlay() extended with settlement priority, renderInspector() extended |
| `internal/ui/view_test.go` | **Modify** | Tests para settlement overlay rendering |
| `cmd/evociv/main.go` | **Modify** | RegisterSettlementStores, LoadSettlementTypes, LoadBuildingTypes, add sistemas |
| `openspec/changes/settlements/design.md` | **Create** | Este documento |

## ADRs

### ADR-1: Asentamientos como entidades ECS vs grid separado

| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| **ECS entities** | Reutiliza World, ComponentStore, LOD, queries. Sin duplicación. ~100 entidades extra trivial. | ✅ **Elegido** |
| Grid separado | O(1) lookup por tile, pero duplica gestión de posiciones, rompe patrón ECS | ❌ Rechazado |

**Rationale**: El mundo es 256×256, ~10 settlements. O(n) iteración es trivial. Los settlements heredan LOD, Position, y sistemas de render GRATIS. Consistente con la filosofía ECS del proyecto.

### ADR-2: Edificios como entidades ECS hijas vs slots del settlement

| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| **Entidades ECS hijas** | Position individual, extensibles (vida útil, upgrades), misma API que NPCs | ✅ **Elegido** |
| Slots []Building | Sin entidades extra, pero no tienen posición, no escalan a interacciones futuras | ❌ Rechazado |

**Rationale**: Los buildings como entidades permiten que en el futuro tengan HP, workers asignados, producción individual. El patrón de "entidad hija dentro de radio" es natural en ECS. El costo es ~60 entidades extras (10 settlements × 6 buildings), irrelevante.

### ADR-3: Nombres procedurales con pool + sufijos

| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| **Pool fijo + sufijos** | Controlable, sin dependencies externas, determinista | ✅ **Elegido** |
| Markov chain / LLM | Más variedad, pero indeterminista, complejidad extra | ❌ Rechazado |

**Rationale**: Para MVP, un pool de ~12 prefijos + ~8 sufijos combinados con el tipo genera nombres variados (12×3×8 = 288 combinaciones). Suficiente para 5-10 settlements. Truncar a 15 chars para TUI. En el futuro se puede externalizar a YAML.

### ADR-4: Overlay orden NPC > Settlement > Biome

| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| **NPC > Settlement > Biome** | Lectura clara: NPCs son entidades activas encima de asentamientos | ✅ **Elegido** |
| Settlement > NPC > Biome | Asentamientos visibles siempre, NPCs ocultos si coinciden | ❌ Rechazado |
| Solo settlement si hay NPC | Confunde al usuario (desaparece NPC) | ❌ Rechazado |

**Rationale**: Los NPCs son la unidad interactiva principal (inspector, GOAP, Q-Learning). Un NPC en el mismo tile que un settlement center debe ser visible. El settlement se ve cuando no hay NPC. Coincidencias son raras (el settlement center es 1 tile, el NPC puede estar en cualquiera de los Radius×Radius tiles).

## Data Flow

```
main.go
  │
  ├── data.Loader.LoadAll("data", registry)
  │   ├── settlements.yaml → kind:"settlement-types" → registry
  │   └── buildings.yaml   → kind:"building-types"   → registry
  │
  ├── settlement.LoadSettlementTypes(registry)
  ├── settlement.LoadBuildingTypes(registry)
  ├── settlement.RegisterSettlementStores(ecsWorld)
  │
  └── ecsWorld.AddSystem(settlement.NewSettlementSpawnSystem(...))
      └── SettlementSpawnSystem.Update()
          ├── Crea entidades con Position + Settlement + LOD
          └── Crea entidades hijas con Position + Building
              │
              ▼
          SettlementRenderSystem.Update()
              └── Recolecta []SettlementRenderInfo
                  │
                  ▼
              Model.SetSettlementOverlay()
                  │
                  ▼
              renderOverlay() → NPC > Settlement > Biome
```

```
npc.Spawn(w, wm, config, seed, raceDefs, roleDefs, settlementEntities)
  │
  ├── Para cada NPC:
  │   ├── Elegir raza + rol (sin cambios)
  │   ├── Buscar settlement compatible (NUEVO):
  │   │   ├── farmer → settlement con farm building
  │   │   ├── merchant → town/city con market
  │   │   ├── hunter → settlement en forest biome
  │   │   ├── priest → settlement con temple
  │   │   ├── blacksmith → settlement con blacksmith
  │   │   ├── miner → cualquier settlement compatible
  │   │   └── artisan → cualquier settlement
  │   ├── Encontrado? → Position aleatoria dentro del radio + HomeReference
  │   └── No encontrado? → biome-weighted random (fallback nómada)
  │
  └── Retorna error (sin cambios)
```

## Interfaces / Contracts

```go
// SettlementSpawnSystem — ECS system, spawns once
type SettlementSpawnSystem struct {
    spawned        bool
    wm             *world.WorldMap
    seed           int64
    settlementDefs []SettlementDef
    buildingDefs   []BuildingDef
}

func NewSettlementSpawnSystem(wm *world.WorldMap, seed int64, settlementDefs []SettlementDef, buildingDefs []BuildingDef) *SettlementSpawnSystem
func (s *SettlementSpawnSystem) Name() string                          // "SettlementSpawnSystem"
func (s *SettlementSpawnSystem) Update(w *ecs.World, dt float64) error // spawns once

// SettlementRenderSystem — ECS system, collects render info each tick
type SettlementRenderSystem struct { renderInfos []SettlementRenderInfo }
func NewSettlementRenderSystem() *SettlementRenderSystem
func (s *SettlementRenderSystem) Name() string                              // "SettlementRenderSystem"
func (s *SettlementRenderSystem) Update(w *ecs.World, dt float64) error     // collects renderables
func (s *SettlementRenderSystem) RenderInfos() []SettlementRenderInfo

// Modified Spawn signature (spawner.go)
func Spawn(w *ecs.World, wm *world.WorldMap, config SpawnConfig, seed int64, raceDefs []RaceDef, roleDefs []RoleDef, settlementEntities []ecs.Entity) error
```

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit | `LoadSettlementTypes` / `LoadBuildingTypes` | `fstest.MapFS` con YAML inline, verify structs, test missing kind error |
| Unit | `SettlementSpawnSystem` spawns 5-10 | World 256×256, seed fijo, verify len(Settlement store) ∈ [5,10], all on compatible biomes |
| Unit | No settlements in ocean | World 100% ocean, expect 0 settlements spawned |
| Unit | Buildings spawn | Verify each settlement has correct building types per def, within radius |
| Unit | NPC spawn in settlement | Create world with settlements, spawn NPC, verify Position inside settlement radius AND HomeReference != 0 |
| Unit | Nomad fallback | World with 0 settlements, spawn NPC, verify no HomeReference (zero value) |
| Unit | Capacity overflow | Settlement Radius=3 (cap 6), spawn 10 compatible NPCs, verify ≤6 have HomeReference to it |
| Unit | Determinism | Same seed×2 → identical settlement positions, names, building positions |
| Unit | renderOverlay priority | Tile with NPC + settlement → NPC returned, Tile with settlement only → settlement symbol |
| Unit | renderInspector settlement | Cursor on settlement tile, press 'e' → inspector shows settlement data |
| Unit | Name truncation | Name > 15 chars → truncated to 15 |
| Unit | Spawn weight sum | `validateSettlementData()` rejects weights not summing to 1.0 ± 0.01 |
| Unit | Building reference valid | `validateSettlementData()` rejects settlement type referencing unknown building ID |

## Open Questions

- [ ] ¿Pool de nombres en YAML (settlement_name_pool) para futuro o hardcoded en types.go para MVP?
- [ ] Best: hardcoded en MVP, YAML en post-MVP. **Decisión**: hardcoded `types.go`.

- [ ] ¿LOD en settlements? Los settlements heredan LOD del sistema existente si tienen componente LOD, o se renderizan siempre?
- [ ] Best: settlements se renderizan siempre (son ~10, la query visual es importante incluso lejos). No llevar componente LOD. **Decisión**: Sin LOD, render siempre.

- [ ] ¿Capacity tracking en Population field del componente o en estructura separada?
- [ ] Best: Population en Settlement component se incrementa al asignar NPCs. **Decisión**: Campo Population en Settlement.

- [ ] Inspector de settlement: ¿qué data mostrar exactamente?
- [ ] Best: Name, Type, Population/Radius, Buildings list. **Decisión**: Mostrar Name, Type, Population, Radius, Level, Building count.

- [ ] ¿SettlementRenderSystem se ejecuta siempre o solo cuando hay cambios?
- [ ] Best: se ejecuta cada tick (mismo que NPCRenderSystem, recolección O(n) es trivial para ~10 entidades). **Decisión**: Cada tick, en refreshOverlay().

## Migration / Rollout

No migration required. Se añaden nuevos YAML y registros ECS. El spawner modificado mantiene compatibilidad hacia atrás: si no hay settlements, todos los NPCs son nómadas (comportamiento actual). Rollback: `git revert` del branch.
