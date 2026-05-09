# Exploration: Sistema de Asentamientos

## Current State

El proyecto **evociv-rl** es un simulador tipo Dwarf Fortress con las siguientes características actuales:

- **ECS funcional**: 7 sistemas (NPCSpawn, Wander, LOD, NeedsDecay, GOAP, QLearning, NPCRender)
- **Componentes ECS**: Position, Name, Health, Personality, Job, AIState, Appearance, LOD, Needs
- **Mundo**: 256x256 generado con Perlin noise FBM, 6 biomas data-driven (ocean, plains, forest, desert, tundra, jungle)
- **NPCs**: 50-100, spawn biome-weighted, roles (farmer, hunter, merchant, artisan, miner, smith), razas (human, dwarf, elf)
- **Sistemas de comportamiento**: GOAP planifica acciones basadas en necesidades, Q-Learning refuerza con ε-greedy
- **TUI**: Bubbletea, mapa navegable con cursor 🟡, overlay '@' de NPCs, inspector con 'e'
- **Persistencia**: SQLite guarda seed del mundo, Q-table
- **Data-Driven**: YAML para biomas, razas, roles, acciones, config de generación

**NO existe** ningún concepto de asentamiento, ciudad, edificio, economía, o estructura social. Los NPCs spawnan en posiciones aleatorias del mundo (biome-weighted) y deambulan sin un "hogar" o "comunidad".

## Affected Areas

| Archivo | Afectado por |
|---------|-------------|
| `internal/simulation/npc/components.go` | Nuevos componentes: Settlement, Building, Resource |
| `internal/simulation/npc/systems.go` | Nuevos sistemas: SettlementSpawnSystem, SettlementRenderSystem |
| `internal/simulation/npc/spawner.go` | Spawn de NPCs dentro de radios de asentamientos |
| `internal/simulation/npc/types.go` | Tipos SettlementDef, BuildingDef, ResourceDef |
| `internal/simulation/npc/data.go` | Carga de YAML de asentamientos y edificios |
| `internal/world/worldmap.go` | Posible grid separado para asentamientos (probablemente no necesario) |
| `internal/ui/model.go` | Modelo con overlay de settlements, nombres visibles |
| `internal/ui/view.go` | Renderizado de símbolos + nombres de asentamientos |
| `internal/ecs/component_store.go` | Sin cambios (ya soporta cualquier tipo) |
| `internal/ecs/world.go` | Sin cambios (ya soporta RegisterStores adicionales) |
| `data/settlements.yaml` | **NUEVO** — tipos de asentamiento data-driven |
| `data/buildings.yaml` | **NUEVO** — edificios data-driven |
| `data/resources.yaml` | **NUEVO** — recursos data-driven |
| `cmd/evociv/main.go` | Registro de nuevos componentes y sistemas |
| `openspec/changes/settlements/` | Especificaciones, diseño, tareas |

## Approaches

### 1. Asentamientos como Entidades ECS (RECOMENDADO)

Los asentamientos son entidades ECS con un componente `Settlement` que define nombre, tipo, radio, población. Los edificios son entidades ECS hijas con componente `Building`.

**Pros**:
- Reutiliza toda la infraestructura ECS existente (ComponentStore, sistemas, queries)
- Los NPCs pueden tener referencia `SettlementID` a su settlement hogar
- Edificios como entidades permiten Position individual, vida útil, upgrades
- Sin duplicación de lógica — mismo patrón que NPCs
- El LOD system funciona automáticamente para asentamientos distantes
- Coherente con la filosofía del proyecto: ECS-driven

**Cons**:
- Más entidades en el World (100+ si hay 10 asentamientos × 10 edificios)
- El render overlay necesita orden: settlements debajo de NPCs
- No hay "grid" separado para búsqueda espacial rápida (~O(n) por tick)

**Effort**: Medium

### 2. Asentamientos como Grid Separado (NO RECOMENDADO)

Los asentamientos se almacenan en una estructura separada dentro de `world.WorldMap` o como `SettlementMap` análogo.

**Pros**:
- Búsqueda espacial O(1) por tile
- Separación de concerns: settlement logic fuera del ECS
- Render más simple (overlay directo del grid)

**Cons**:
- Duplica la gestión de posiciones (ECS Position + SettlementMap)
- Los NPCs no tienen referencia directa a su settlement
- Rompe el patrón ECS del proyecto
- Añade complejidad: ahora hay dos "worlds" que sincronizar
- No escala a edificios individuales fácilmente

**Effort**: Medium (pero introduce deuda técnica)

### 3. Asentamientos ECS con Aceleración Espacial (OPCIONAL para futuro)

Igual que Approach 1, pero se añade un `SpatialHash` opcional para búsquedas rápidas.

**Pros**: Todos los de Approach 1 + búsqueda O(1) por tile
**Cons**: Complejidad extra en MVP, el mundo es pequeño (256×256, ~10 settlements)
**Effort**: High (no necesario para MVP)

## Recommendation

**Approach 1: Asentamientos como Entidades ECS.**

Razones:
1. Es el camino que ya tiene el proyecto — ECS puro, sin inventar nuevos patrones
2. El mundo es 256×256, con ~5-10 asentamientos y ~5-10 edificios cada uno. O(n) iteración es trivial
3. Los NPCs ya tienen Position + LOD; los settlements heredan eso GRATIS
4. El LOD system ya existe — settlements distantes no se renderizan ni simulan
5. Para el MVP, la simplicidad de "todas las entidades en el mismo World" es inmejorable

### Arquitectura de Componentes Propuesta

```go
// SettlementType define la categoría del asentamiento
type SettlementType string
const (
    SettlementVillage SettlementType = "village"
    SettlementTown    SettlementType = "town"
    SettlementCity    SettlementType = "city"
)

// Settlement component — attach a entities que son asentamientos
type Settlement struct {
    Name       string
    Type       SettlementType
    Radius     int            // tiles que abarca
    Population int            // NPCs asignados
    Level      int            // nivel de desarrollo (1-5)
}

// Building component — attach a entities que son edificios
type Building struct {
    ID       string          // "farm", "market", "tavern", "temple"
    Name     string
    Level    int             // nivel del edificio (1-3)
}

// ResourceStore component — recursos producidos/almacenados
type ResourceStore struct {
    Resources map[string]float64  // "food": 100, "metal": 50, "gold": 20
}

// Production component — define qué produce un edificio
type Production struct {
    Output     map[string]float64  // "food": 1.0 por tick
    Input      map[string]float64  // qué consume
    Workers    int
    MaxWorkers int
}

// HomeReference — NPC referencia a su settlement
type HomeReference struct {
    SettlementEntity ecs.Entity
}
```

### Settlement Types (Data-Driven YAML)

```yaml
kind: settlement-types
data:
  - id: village
    name: Aldea
    symbol: "♦"
    color: "#8B7355"
    min_pop: 5
    max_pop: 30
    radius: 3
    biomes: [plains, forest]
    buildings: [house, farm]
    spawn_weight: 0.6

  - id: town
    name: Pueblo
    symbol: "▲"
    color: "#B8860B"
    min_pop: 20
    max_pop: 100
    radius: 5
    biomes: [plains]
    buildings: [house, market, tavern, blacksmith, farm]
    spawn_weight: 0.3

  - id: city
    name: Ciudad
    symbol: "●"
    color: "#DAA520"
    min_pop: 80
    max_pop: 500
    radius: 8
    biomes: [plains]
    buildings: [house, market, temple, tavern, blacksmith, farm, library]
    spawn_weight: 0.1
```

### Buildings (Data-Driven YAML)

```yaml
kind: building-types
data:
  - id: house
    name: Casa
    symbol: "☐"
    color: "#CD853F"
    provides: [housing]
    capacity: 4

  - id: farm
    name: Granja
    symbol: "⌂"
    color: "#90EE90"
    produces: {food: 2.0}
    consumes: {}
    max_workers: 3
    biomes: [plains]

  - id: market
    name: Mercado
    symbol: "⌂"
    color: "#00CED1"
    produces: {gold: 1.0}
    consumes: {food: 0.5}
    max_workers: 2

  - id: tavern
    name: Taberna
    symbol: "☼"
    color: "#FF4500"
    produces: {}
    consumes: {food: 0.2}
    max_workers: 1

  - id: temple
    name: Templo
    symbol: "☥"
    color: "#9370DB"
    produces: {}
    consumes: {gold: 0.1}

  - id: blacksmith
    name: Herreria
    symbol: "☖"
    color: "#FF6347"
    produces: {metal: 1.0}
    consumes: {}
    max_workers: 2
```

### Visualización TUI

Los settlements se renderizan en este orden:
1. **Base del mapa** (bioma: ., T, ~, etc.)
2. **Settlement center** (símbolo especial: ♦ ▲ ●) — con nombre al lado si el cursor está cerca
3. **NPC overlay** (@, etc.) — encima del settlement si están en el mismo tile
4. **Cursor** (🟡) — siempre encima de todo

```go
// SettlementRenderInfo para el overlay
type SettlementRenderInfo struct {
    Entity         ecs.Entity
    Symbol         rune
    Color          lipgloss.Color
    WorldX, WorldY int
    Name           string
    Type           SettlementType
}
```

La lógica de `renderOverlay` debe priorizar NPC sobre settlement:

```
renderOverlay(worldX, worldY):
  if NPC at (x,y) → render NPC symbol
  if Settlement center at (x,y) → render Settlement symbol
  else → "" (render biome tile)
```

### Spawn de NPCs en Asentamientos

El spawn actual (`internal/simulation/npc/spawner.go`) se modifica para spawnear NPCs dentro del radio de asentamientos en lugar de posiciones mundiales aleatorias:

```
Para cada NPC a spawnear:
  1. Elegir raza y rol (igual que ahora)
  2. Encontrar settlement compatible:
     - Rol "farmer" → preferir settlements con farm
     - Rol "merchant" → preferir towns/cities con market
     - Rol "hunter" → forest-edge settlements
  3. Posición: random tile dentro del settlement.radius
  4. Asignar HomeReference al NPC
```

Los settlements sin NPCs asignados serían "fantasmas" visuales. Los NPCs son la población real.

### Integración con GOAP

Nuevas acciones GOAP que usan settlements (post-MVP):

```yaml
- id: go_home
  requires:
    needs_min: {hunger: 0.0, fatigue: 0.5}
    needs_max: {hunger: 1.0, fatigue: 1.0}
  effects:
    fatigue_change: -0.3
  reward:
    base: 1.0

- id: trade_at_market
  requires:
    needs_min: {hunger: 0.0, fatigue: 0.0}
    needs_max: {hunger: 0.6, fatigue: 0.6}
    biomes: [plains]
  effects:
    hunger_change: -0.1
    fatigue_change: 0.0
  reward:
    base: 0.5

- id: socialize_at_tavern
  requires:
    needs_min: {hunger: 0.0, fatigue: 0.0}
    needs_max: {hunger: 0.5, fatigue: 0.5}
  effects:
    fatigue_change: -0.2
  reward:
    base: 0.4
```

### Economía Básica (post-MVP)

Para el MVP, la economía NO está incluida. Pero se diseña pensando en ella:

1. **Producción**: Cada tick, los edificios con workers producen recursos
   - Farm: food = workers × 0.5
   - Market: gold = workers × 0.3 (consume food)
   - Blacksmith: metal = workers × 0.2

2. **Consumo**: Los NPCs consumen recursos del settlement:
   - 0.01 food por tick por NPC
   - Si food < 0 → hambruna → NPCs migran o mueren

3. **SettlementEconomySystem**: Sistema ECS que cada tick actualiza producción/consumo

```go
type SettlementEconomySystem struct {}

func (s *SettlementEconomySystem) Update(w *ecs.World, dt float64) error {
    for each entity with Settlement + ResourceStore:
        for each entity with Building + Production within radius:
            produce resources
        for each NPC with this settlement as home:
            consume resources from store
}
```

## Alcance MVP

```
✅ Fase 1 — Infraestructura de datos
  - settlements.yaml y buildings.yaml con definiciones data-driven
  - Validadores YAML para settlement-types y building-types
  - Tests de carga y validación

✅ Fase 2 — Componentes ECS
  - Settlement component {Name, Type, Radius, Population, Level}
  - Building component {ID, Name, Level}
  - ResourceStore component {Resources map}
  - HomeReference component {SettlementEntity}
  - RegisterStoresSettlements(w) — registro de nuevos stores
  - Tests de componentes

✅ Fase 3 — Spawn de Asentamientos en el Mundo
  - SettlementSpawnSystem: genera settlements en biomas compatibles
  - Algoritmo: sampling basado en bioma con distancia mínima entre settlements
  - 5-10 settlements por mundo 256×256
  - Tests: settlements spawn en plains, no en ocean

✅ Fase 4 — Spawn de NPCs en Asentamientos
  - Modificar spawner.go: NPCs spawn dentro de settlements
  - Rol-to-building matching (farmers → farms, merchants → market)
  - Asignar HomeReference a cada NPC
  - Tests: NPCs están dentro de radio de settlements

✅ Fase 5 — Visualización TUI
  - Símbolos especiales por tipo (♦ ▲ ●)
  - Nombre del settlement: "♦ Aldea del Norte"
  - Overlay ordenado: NPC > Settlement > Biome
  - Inspector muestra datos del settlement (población, edificios)
  - Tests de render

✅ Fase 6 — Edificios en Asentamientos
  - Spawn de buildings como entidades dentro del settlement
  - Cada settlement tiene buildings según su tipo
  - Tests

🔲 Extras (post-MVP)
  - Economía: producción/consumo de recursos
  - Acciones GOAP vinculadas a buildings (trade, socialize)
  - Crecimiento de población
  - Nombres procedurales de settlements
  - Persistencia SQLite de settlements
```

## Risks

1. **Rendimiento de Overlay** — Si settlements y NPCs están en el mismo tile, el overlay necesita priorización. **Mitigación**: La función `renderOverlay` ya existe, solo añadir chequeo de settlements después de NPC.

2. **Spawn de settlements en biome incorrecto** — Si no hay suficientes tiles de plains para towns/cities, puede spawnear fuera. **Mitigación**: Validar biomas en YAML + fallback a biome más cercano con distancia Manhattan limitada.

3. **NPCs spawn sin settlement** — Si hay más NPCs que capacidad de settlements. **Mitigación**: Si no hay settlement compatible, NPC spawnea como "nómada" (sin HomeReference, comportamiento actual).

4. **GOAP necesita consultas espaciales** — "Estoy cerca de un mercado?" requiere buscar settlements cercanos. **Mitigación**: Para MVP, GOAP no usa settlements; las acciones sociales/trade se añaden en post-MVP.

5. **Inflación de entidades ECS** — 256×256 mundo con 10 settlements × 10 edificios = 100 entidades extra. **Impacto**: Mínimo. El ECS ya maneja 100 NPCs sin problema. 200 entidades totales siguen siendo triviales para Go.

6. **Nombres duplicados** — Si dos settlements reciben el mismo nombre procedural. **Mitigación**: Pool de nombres con deduplicación, o añadir sufijo (Norte, Sur, Este, Oeste).

## Ready for Proposal

**Sí.** Esta exploración tiene suficiente profundidad para proceder con:
1. **sdd-propose**: Definir alcance exacto del MVP, fases, rollback plan
2. **sdd-spec**: Especificaciones Given/When/Then para cada fase
3. **sdd-design**: Diseño técnico detallado con componentes, sistemas, y YAML
4. **sdd-tasks**: Breakdown en tareas implementables individualmente

## Anexo: Pseudocódigo de SettlementSpawnSystem

```
SettlementSpawnSystem.Update(w, dt):
    if already spawned → return

    types = load settlement types from registry
    for each type:
        count = calculate count based on world size and spawn_weight
        for i in 0..count:
            attempts = 0
            while attempts < 100:
                pick random tile matching biome requirements
                if tile has settlement within min_distance: retry
                create entity with Position + Settlement + Name
                spawn buildings as child entities with Building + Position
                break
```

## Anexo: Mapa de Archivos Nuevos

| Archivo | Propósito |
|---------|-----------|
| `data/settlements.yaml` | Settlement types data-driven |
| `data/buildings.yaml` | Building types data-driven |
| `data/resources.yaml` | Resource definitions (MVP opcional) |
| `internal/simulation/settlement/` | **Paquete nuevo** con componentes, sistemas, datos |
| `internal/simulation/settlement/components.go` | Settlement, Building, ResourceStore, HomeReference, Production |
| `internal/simulation/settlement/types.go` | SettlementDef, BuildingDef, SettlementRenderInfo |
| `internal/simulation/settlement/data.go` | LoadSettlementTypes, LoadBuildingTypes |
| `internal/simulation/settlement/systems.go` | SettlementSpawnSystem, SettlementEconomySystem |
| `internal/simulation/settlement/systems_test.go` | Tests |
| `openspec/changes/settlements/` | SDD artifacts completos |
