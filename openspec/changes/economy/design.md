# Design: Economy

## Technical Approach

Tres sistemas ECS secuenciales en un nuevo paquete `internal/simulation/economy/` que gestionan producción, consumo, crecimiento y hambruna. Los datos económicos (producción por edificio, thresholds de crecimiento) son **data-driven via YAML** — se cargan en memoria como lookup maps dentro de los sistemas, no como componentes ECS. SettlementEconomySystem produce recursos según workers asignados por Job.Role, SettlementGrowthSystem verifica thresholds cada tick, y FamineSystem expulsa NPCs (remueve HomeReference) cuando el food del settlement es negativo. La TUI extiende SettlementRenderSystem para transportar resource data al overlay.

## Data Flow (single tick)

```
SettlementEconomySystem (primero)
  │
  ├─ for each settlement:
  │   ├─ Lazy-init ResourceStore si no existe
  │   ├─ Contar NPCs con HomeReference a este settlement
  │   ├─ for each building type in settlement.Buildings:
  │   │   ├─ Lookup BuildingDef (map en memoria)
  │   │   ├─ Contar NPCs con Job.Role == building.Role (cap max_workers)
  │   │   ├─ Produces: ResourceStore.Add(resource, rate * workers * dt)
  │   │   └─ Consumes: ResourceStore.Remove(resource, rate * workers * dt)
  │   └─ Consumo NPC: ResourceStore.Remove("food", 0.01 * npcCount * dt)
  │
SettlementGrowthSystem (segundo)
  │
  ├─ for each settlement with ResourceStore:
  │   ├─ Buscar GrowthThreshold para Level+1
  │   ├─ Si resource >= threshold → Level++, Radius++, resources -= threshold
  │   └─ Si no hay threshold → skip (max level)
  │
FamineSystem (tercero)
  │
  └─ for each settlement with ResourceStore["food"] < 0:
      └─ Remover HomeReference de 1 NPC → food deficit se reduce porque ese NPC ya no consume
```

## Architecture Decisions

### 1. Producción en memoria del sistema vs ECS component

| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| **In-memory map** en el sistema (BuildingDef → produces/consumes) | Sin estado ECS extra, más rápido lookup, YAML como única fuente de verdad | ✅ **Elegido** |
| Componente `ProductionRate` en cada building entity | Más ECS-puro, pero sobrecarga de componentes para datos estáticos que no cambian por entidad | ❌ Rechazado — los rates son globales por tipo de building, no por instancia |

**Rationale**: Los datos de producción son estáticos (definidos por tipo de building en YAML). Meterlos en ECS component para cada building entity añadiría complejidad sin beneficio. El sistema mantiene `map[string]BuildingDef` que construye en New. Si en el futuro cada building necesita rates distintos (mejoras, etc.), se migra a componente.

### 2. Hambruna = migración vs muerte

| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| **Migración** — remover HomeReference, NPC se vuelve nómada | Simple, reversible si el NPC vuelve, consistente con wander existente | ✅ **Elegido** |
| Muerte — eliminar entidad NPC | Más realista, pero irreversible, requiere resurrección/re-spawn, añade logística | ❌ Rechazado — para MVP la migración es suficiente y menos disruptiva |
| Mixto — migración primero, muerte si persiste | Más matizado, pero complejidad extra para MVP | ❌ Rechazado |

**Rationale**: La migración (remover HomeReference) reusa el sistema WanderSystem existente — los NPCs sin HomeReference ya deambulan como nómadas. No necesita nuevo comportamiento. La muerte se considerará post-MVP si el déficit persiste por N ticks.

### 3. GOAP no acoplado en MVP

| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| **GOAP separado** — economía opera a nivel settlement, GOAP a nivel NPC individual | Sistemas ortogonales, no se acoplan | ✅ **Elegido** |
| GOAP vinculado — NPCs eligen buildings según hunger/fatigue | Más realista pero acopla economía con planificación individual | ❌ Rechazado — añade complejidad y dependencias |

**Rationale**: Los sistemas de economía operan a nivel settlement (recursos agregados). GOAP/RL opera a nivel NPC individual (necesidades). Son ortogonales. En el futuro se puede conectar: un NPC con hambre podría priorizar trabajar en una farm versus una blacksmith. Pero para MVP, los NPCs producen según su Job.Role independientemente de sus necesidades.

### 4. Thresholds YAML vs hardcodeados

| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| **data/growth.yaml** — thresholds en YAML con loader | Data-driven, ajustable sin recompilar, consistente con arquitectura existente | ✅ **Elegido** |
| Constantes en Go — hardcodeadas en el sistema | Simple, sin loader, sin archivo extra | ❌ Rechazado — rompe el patrón data-driven del proyecto |

**Rationale**: Todo el contenido del juego es data-driven via YAML (buildings, npcs, biomes, acciones). Los thresholds de crecimiento deben seguir el mismo patrón. Además permite balanceo sin recompilar.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `data/growth.yaml` | **Create** | Thresholds de crecimiento: level, food, tools, gold, new_radius, new_buildings |
| `data/buildings.yaml` | Modify | Añadir `produces`, `consumes`, `max_workers`, `role` a cada building productivo |
| `internal/simulation/economy/systems.go` | **Create** | SettlementEconomySystem, SettlementGrowthSystem, FamineSystem + constructoras |
| `internal/simulation/economy/systems_test.go` | **Create** | Tests TDD para los 3 sistemas |
| `internal/simulation/settlement/types.go` | Modify | BuildingDef: +Role, Produces, Consumes, MaxWorkers. New: GrowthThreshold. SettlementRenderInfo: +Food, Gold, Tools, Level |
| `internal/simulation/settlement/components.go` | Modify | ResourceStore: +Add(), +Remove(), +Has() |
| `internal/simulation/settlement/components_test.go` | Modify | Tests para Add/Remove/Has |
| `internal/simulation/settlement/data.go` | Modify | LoadBuildingTypes: parsear produces/consumes/role/max_workers. New: LoadGrowthThresholds() |
| `internal/simulation/settlement/data_test.go` | Modify | Tests para nuevos campos building + growth thresholds |
| `internal/simulation/settlement/systems.go` | Modify | SettlementRenderSystem: leer ResourceStore y poblar nuevos campos de SettlementRenderInfo |
| `internal/simulation/settlement/systems_test.go` | Modify | Tests para render system con recursos (si aplica) |
| `internal/ui/view.go` | Modify | Status bar: "♦ Aldea \| Pop:5 \| Food:45 Gold:12 Tools:3". Inspector: recursos + level progress + famine warning |
| `cmd/evociv/main.go` | Modify | Crear y registrar SettlementEconomySystem, SettlementGrowthSystem, FamineSystem en orden correcto |

## Interfaces / Contracts

### SettlementEconomySystem

```go
package economy

type SettlementEconomySystem struct {
    buildingMap map[string]settlement.BuildingDef
    // consumo interno: 0.01 food/tick por NPC
}

func NewSettlementEconomySystem(buildingDefs []settlement.BuildingDef) *SettlementEconomySystem
func (s *SettlementEconomySystem) Name() string
func (s *SettlementEconomySystem) Update(w *ecs.World, dt float64) error
```

### SettlementGrowthSystem

```go
type SettlementGrowthSystem struct {
    thresholds map[int]settlement.GrowthThreshold // keyed by level
    maxLevel   int
}

func NewSettlementGrowthSystem(thresholds []settlement.GrowthThreshold) *SettlementGrowthSystem
func (s *SettlementGrowthSystem) Name() string
func (s *SettlementGrowthSystem) Update(w *ecs.World, dt float64) error
```

### FamineSystem

```go
type FamineSystem struct{}

func NewFamineSystem() *FamineSystem
func (s *FamineSystem) Name() string
func (s *FamineSystem) Update(w *ecs.World, dt float64) error
```

### ResourceStore Helpers (en settlement/components.go)

```go
func (rs *ResourceStore) Add(resource string, amount float64)
func (rs *ResourceStore) Remove(resource string, amount float64) bool
func (rs *ResourceStore) Has(resource string, amount float64) bool
```

### GrowthThreshold (en settlement/types.go)

```go
type GrowthThreshold struct {
    Level        int
    Food         float64
    Tools        float64
    Gold         float64
    NewRadius    int
    NewBuildings []string
}
```

### SettlementRenderInfo extended (en settlement/types.go)

```go
type SettlementRenderInfo struct {
    Entity         int
    Symbol         rune
    Color          string
    Name           string
    WorldX, WorldY int
    Population     int
    Food           float64 // new
    Gold           float64 // new
    Tools          float64 // new
    Level          int     // new
    HasResources   bool    // new — false si ResourceStore no existe
}
```

### YAML: growth.yaml

```yaml
kind: growth-thresholds
data:
  - level: 2
    food: 100
    tools: 10
    gold: 5
    new_radius: 4
    new_buildings: []
  - level: 3
    food: 500
    tools: 50
    gold: 25
    new_radius: 6
    new_buildings: []
```

### YAML: buildings.yaml (extendido)

```yaml
kind: building-types
data:
  - id: house
    name: Casa
  - id: farm
    name: Granja
    role: farmer
    max_workers: 3
    produces: {food: 2.0}
  - id: market
    name: Mercado
    role: merchant
    max_workers: 2
    produces: {gold: 1.0}
    consumes: {food: 0.5}
  - id: tavern
    name: Taberna
  - id: temple
    name: Templo
  - id: blacksmith
    name: Herreria
    role: smith
    max_workers: 2
    produces: {tools: 1.0}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit — ResourceStore | Add, Remove, Has, edge cases (remove insuficiente) | Crear ResourceStore directamente, llamar helpers, verificar estado. `components_test.go` |
| Unit — Data loading | BuildingDef con nuevos campos, legacy buildings siguen funcionando, validación rates negativos, growth.yaml loading, missing file | `fstest.MapFS` + `data.Loader` + loaders. `data_test.go` |
| Unit — EconomySystem | Producción farm con workers, blacksmith, market (produce + consume), NPC food consumption, max_workers cap, lazy-init ResourceStore, house sin producción | Crear ECS world con settlement + NPCs + buildings, ejecutar Update, verificar ResourceStore. `economy/systems_test.go` |
| Unit — GrowthSystem | Level-up con threshold exacto, level-up con exceso, partial resources (no level-up), max level, missing threshold | Crear settlement con ResourceStore + Level, ejecutar Update, verificar Level/Radius/resources. `economy/systems_test.go` |
| Unit — FamineSystem | Déficit detectado, 1 NPC removido por tick, múltiples ticks, recovery (deja de remover), food positivo no hace nada | Crear settlement con ResourceStore food negativo + NPCs con HomeReference, ejecutar Update, verificar HomeReferences. `economy/systems_test.go` |
| Integration | Sistema completo: spawn → economy → growth → famine cycle | World completo con NPCs, settlements, buildings; ejecutar múltiples ticks; verificar evolución de recursos y niveles |
| TUI | Status bar rendering con recursos, inspector resources, inspector famine warning | Pruebas de renderizado de cadenas. `view_test.go` |

## Migration / Rollout

No migration required. Todos los settlements existentes se crean con Level=1 y sin ResourceStore. SettlementEconomySystem hace lazy-init del ResourceStore en el primer tick. Los buildings sin campos económicos (legacy) se cargan con Produces=nil, Consumes=nil, Role="", MaxWorkers=0 — el sistema los ignora durante producción.

Rollback: `git revert` del branch completo. Si mergeado, revert commit + PR de revert.

## Open Questions

- [ ] ¿La TUI necesita mostrar famine warning en tiempo real o solo en inspector? Spec dice ambas: status bar podría mostrar "⚠" si food < 0
- [ ] ¿new_buildings en GrowthThreshold debe spawnear edificios automáticamente? Spec no lo especifica — MVP lo deja como slice vacío, no se spawnan
- [ ] ¿El FamineSystem debe remover NPCs basado en algún orden (LOD, orden de creación, aleatorio)? Spec no especifica — se usará orden de iteración del store (determinístico por entidad ID)
