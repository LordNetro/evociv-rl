# Exploration: Sistema de Economía y Producción

## Current State

El proyecto **evociv-rl** tiene una base ECS sólida con 10+ sistemas en producción. Los asentamientos ya existen como entidades ECS (`Settlement` con Buildings, Population, Level, Radius) y los NPCs spawn dentro de settlements con `HomeReference`. Sin embargo:

- **ResourceStore component** (`map[string]float64`) existe registrado pero **nunca se escribe ni se lee** en ningún sistema
- **Production component** mencionado en la exploración original de settlements pero **nunca implementado** — no existe en `components.go`
- **No hay economía**: ningún sistema produce recursos, ningún NPC consume, los settlements no crecen
- **GOAP actions** (harvest, forage, rest, socialize, trade, explore) solo afectan Hunger/Fatigue, no recursos
- **Settlement.Level** siempre es 1 — no hay mecánica de crecimiento
- **PopulationSystem** cuenta NPCs pero no hay límite de población basado en recursos

### Componentes ECS Relevantes

| Componente | Archivo | Estado |
|---|---|---|
| `ResourceStore{Resources map[string]float64}` | `settlement/components.go` | ✅ Registrado, NUNCA usado |
| `Settlement{Level int}` | `settlement/components.go` | ✅ Level=1 siempre |
| `Building{ID, Name, Level}` | `settlement/components.go` | ✅ Sin datos de producción |
| `Job{Role string}` | `npc/components.go` | ✅ "farmer", "merchant", "smith", etc. |
| `HomeReference{SettlementEntity}` | `settlement/components.go` | ✅ Vincula NPC→settlement |
| `Needs{Hunger, Fatigue}` | `npc/components.go` | ✅ Solo para GOAP/RL |

### YAML Relevantes

| Archivo | Contenido | Carencia |
|---|---|---|
| `data/buildings.yaml` | 6 buildings (house, farm, market, tavern, temple, blacksmith) | Sin `produces`, `consumes`, `max_workers` |
| `data/actions.yaml` | 6 acciones GOAP (harvest, forage, rest, socialize, trade, explore) | `trade` existe pero no produce gold |
| `data/npc-roles.yaml` | farmer, hunter, merchant, artisan, miner, smith | Sin vinculación a producción |

### Datos Clave del Código Fuente

- **Tick rate**: 200ms (5 ticks/segundo)
- **Needs decay**: Hunger +0.01/tick, Fatigue +0.005/tick
- **NPCs**: 50-100 por mundo
- **Settlements**: 5-10 por mundo, 3-8 de radio
- **Buildings por settlement**: 2 (village) a 6 (city)

---

## Affected Areas

| Área | Impacto | Archivos |
|---|---|---|
| `internal/simulation/economy/` | **NUEVO** — Paquete completo con SettlementEconomySystem | `systems.go`, `components.go` (si se extiende Production), `types.go` |
| `internal/simulation/settlement/components.go` | **Modificado** — Production component, ResourceStore usage, Settlement growth fields | `components.go`, `types.go` |
| `internal/simulation/settlement/systems.go` | **Modificado** — PopulationSystem con food-check, SettlementEconomySystem | `systems.go` |
| `internal/simulation/npc/systems.go` | **Modificado** — QLearningSystem/GOAPSystem para integrar efectos económicos | `systems.go` |
| `data/buildings.yaml` | **Modificado** — Añadir produces, consumes, max_workers por building | `buildings.yaml` |
| `data/growth.yaml` | **NUEVO** — Thresholds de crecimiento por nivel | `growth.yaml` |
| `data/actions.yaml` | **Modificado** — work_at_farm, rest_at_home, trade económico | `actions.yaml` |
| `internal/ui/view.go` | **Modificado** — Status bar con recursos, inspector económico | `view.go` |
| `internal/ui/model.go` | **Modificado** — Campos para recursos en overlay/inspector | `model.go` |
| `cmd/evociv/main.go` | **Modificado** — Registrar SettlementEconomySystem, GrowthSystem | `main.go` |
| `internal/simulation/settlement/data.go` | **Modificado** — LoadBuildingTypes con produces/consumes | `data.go` |

---

## Approaches

### 1. Sistema Económico Integrado en settlement/ (RECOMENDADO)

La economía vive dentro del paquete `internal/simulation/economy/` como sistemas ECS separados que operan sobre los componentes existentes (`ResourceStore`, `Settlement`, `Building`, `Job`, `HomeReference`).

**Producción**: `SettlementEconomySystem` cada tick itera settlements, encuentra NPCs con HomeReference a cada settlement, agrupa por Job, mira qué buildings hay en el settlement, y produce recursos según `workers × building_productivity`.

**Consumo**: Cada NPC consume 0.01 food/tick de su settlement. SettlementEconomySystem descuenta de ResourceStore.

**Crecimiento**: `SettlementGrowthSystem` separado que verifica si el settlement acumuló suficientes recursos para subir de nivel.

**Pros**:
- Separación limpia de concerns (producción ≠ crecimiento ≠ visualización)
- Reutiliza componentes existentes (ResourceStore, HomeReference, Job, Building)
- SettlementEconomySystem puede desactivarse individualmente para debugging
- Sigue el mismo patrón que todos los demás sistemas ECS del proyecto
- Fácil de testear: mock ResourceStore, verificar cambios después de Update()

**Cons**:
- Más sistemas en el Update loop (2 adicionales)
- Necesita acceso a cross-store queries (settlements + NPCs + buildings)
- La producción por building requiere extender BuildingDef en YAML

**Effort**: Medium (estimar 3-4 fases de implementación)

### 2. Sistema Embebido en SettlementSystems (NO RECOMENDADO)

Añadir la lógica económica directamente dentro de los sistemas existentes de settlement (`PopulationSystem`, `SettlementRenderSystem`).

**Pros**:
- Sin nuevo paquete
- Sin nuevos archivos
- Aprovecha bucles existentes

**Cons**:
- Mezcla responsabilidades (PopulationSystem ahora también produce recursos)
- Viola Single Responsibility — difícil de testear, mantener, desactivar
- Rompe el patrón de delegación del proyecto (cada system hace UNA cosa)
- El testing se vuelve complejo: un test de population ahora depende de economía

**Effort**: Low (rápido, pero deuda técnica inmediata)

### 3. Sistema Data-Driven Puro (FUTURO)

Toda la economía se define en YAML: buildings definen recipes (inputs + outputs + rate), NPC roles definen qué recipes ejecutan, y un único `EconomySystem` genérico procesa recipes.

**Pros**:
- Extremadamente flexible (cambiar economía tocando solo YAML)
- Sin cambios de código para nuevos recursos o recipes
- Extensible a cadenas de producción complejas

**Cons**:
- Requiere un engine de recipes — sobreingeniería para MVP
- Dependencia de validación YAML compleja (ciclos, balances)
- Dificulta testing: la lógica está en YAML, no en Go
- Go no es dinámico — mapear recipes a structs requiere reflection o type switches

**Effort**: High (no recomendado para MVP, excelente para v2)

---

## Arquitectura del Sistema Económico

### Componentes

#### 1. ResourceStore (EXISTE — expandir uso)

```go
type ResourceStore struct {
    Resources map[string]float64 // "food": 100.0, "gold": 50.0, "tools": 20.0
}
```

Ya existe en `settlement/components.go`. Se añadirá un helper:
```go
func (rs *ResourceStore) Add(resource string, amount float64)
func (rs *ResourceStore) Remove(resource string, amount float64) bool // false si insuficiente
func (rs *ResourceStore) Has(resource string, amount float64) bool
```

#### 2. Production (NUEVO — componente opcional en buildings)

```go
type Production struct {
    Outputs    map[string]float64 // {"food": 2.0}  — por worker por tick
    Inputs     map[string]float64 // {"food": 0.5}  — por worker por tick (opcional)
    MaxWorkers int                // 3
    Workers    int                // contador actual
}
```

Se attacha a entidades **building** que producen. SettlementEconomySystem itera buildings con Production, busca NPCs con Job compatible, asigna workers, y produce/consume.

**Alternativa**: En lugar de un componente separado, los datos de producción se cargan desde YAML extendido y se cachean en el sistema. Esto evita añadir otro registro de store.

**Decisión**: Datos de producción en memoria del sistema (no en componente ECS). Los buildings solo son entidades con `Building{ID, Name, Level}`. La producción se resuelve por ID de building contra un map cargado de YAML. Esto mantiene liviano el ECS.

### SettlementEconomySystem

```go
type SettlementEconomySystem struct {
    buildingProd map[string]BuildingProd // cargado de YAML
}

type BuildingProd struct {
    Produces    map[string]float64 // food: 2.0
    Consumes    map[string]float64 // food: 0.5 (opcional)
    MaxWorkers  int
    Role        string             // "farmer" produce en "farm"
}
```

#### Algoritmo cada tick:

```
SettlementEconomySystem.Update(w, dt):
  1. Obtener ResourceStore, Settlement, Building stores
  2. Obtener Job store y HomeReference store (NPCs)
  3. Para cada entidad con Settlement:
     a. Inicializar ResourceStore si no existe (Resources = {food:0, gold:0, tools:0})
     b. PRODUCCIÓN: Para cada building en el settlement:
        - Obtener definición de producción desde buildingProd[building.ID]
        - Si el building tiene MaxWorkers > 0:
          * Contar NPCs con HomeReference a este settlement Y Job.Role == buildingProd.Role
          * workers = min(count, building.MaxWorkers)
          * Para cada output: rs.Add(resource, rate * workers * dt)
          * Para cada input: rs.Remove(resource, rate * workers * dt) — si no hay, skip
     c. CONSUMO: Contar NPCs con HomeReference a este settlement
        - food_needed = npc_count * 0.01 * dt
        - rs.Remove("food", food_needed)
        - Si food < 0: marcar hambruna en el settlement (flag o meter en otro componente)
     d. Guardar ResourceStore actualizado
```

#### Balance de Recursos (por tick, dt=1.0)

| Building | Role | Output/worker/tick | Input/worker/tick | MaxWorkers | Producción total |
|---|---|---|---|---|---|
| farm | farmer | +2.0 food | — | 3 | +6.0 food |
| blacksmith | smith | +1.0 tools | — | 2 | +2.0 tools |
| market | merchant | +1.0 gold | -0.5 food | 2 | +2.0 gold, -1.0 food |

**Consumo NPC**: 0.01 food/tick por NPC.

**Ejemplo**: Aldea con 1 farm (2 farmers), 1 market (1 merchant), 5 NPCs total:
- Producción: +4.0 food (farm) +1.0 gold (market) -0.5 food (market input)
- Consumo: -0.05 food (5 NPCs × 0.01)
- Neto: +3.45 food/tick, +1.0 gold/tick

A 5 ticks/segundo, en 10 segundos (50 ticks):
- +172.5 food, +50 gold
- Para subir a Level 2 (100 food + 10 tools + 5 gold): ~29 segundos de producción estable

### SettlementGrowthSystem

```go
type SettlementGrowthSystem struct {
    thresholds map[int]GrowthThreshold // cargado de YAML
}

type GrowthThreshold struct {
    Level    int
    Food     float64
    Tools    float64
    Gold     float64
    Radius   int    // nuevo radio al subir
    MaxPop   int    // nueva población máxima
}
```

#### Algoritmo:

```
SettlementGrowthSystem.Update(w, dt):
  1. Obtener ResourceStore y Settlement stores
  2. Para cada entidad con Settlement:
     a. Si Level >= maxLevel → skip
     b. threshold = thresholds[Level + 1]
     c. Si rs.Has("food", threshold.Food) 
        AND rs.Has("tools", threshold.Tools) 
        AND rs.Has("gold", threshold.Gold):
        - Deducir recursos: rs.Remove(...)
        - set.Level++
        - set.Radius = threshold.Radius
        - (Opcional) spawnear nuevos buildings según SettlementDef
        - (Opcional) aumentar capacidad de población
```

#### Thresholds YAML (data/growth.yaml):

```yaml
kind: growth-thresholds
data:
  - level: 2
    food: 100
    tools: 10
    gold: 5
    radius: 4
    max_pop: 15
    buildings: [house, farm]  # nuevos buildings al subir
  - level: 3
    food: 500
    tools: 50
    gold: 25
    radius: 6
    max_pop: 30
    buildings: [market, tavern]
```

### Hambruna (Famine System)

Si un settlement tiene food < 0:
1. NPCs sin food suficiente empiezan a perder Health
2. Opción A: NPCs migran (pierden HomeReference, se vuelven nómadas)
3. Opción B: NPCs mueren (Health decrece hasta 0 → RemoveEntity)

**Recomendación MVP**: Opción A (migración). Es más indulgente, permite recuperación, no pierde entidades permanentemente. La muerte (Opción B) se añade en post-MVP.

```go
type FamineSystem struct {
    famineThreshold float64 // -10 food acumulado
}
```

Si el settlement tiene déficit de food acumulado > threshold, NPCs empiezan a migrar (pierden HomeReference uno por tick).

---

## Visualización TUI

### Status Bar

Cuando el cursor está sobre un settlement, la barra de estado muestra:

```
♦ Aldea del Norte | Pop:5 | Level:1 | 🍞 45.0 | ⛏️ 3.0 |  Gold 12.0
```

Formato: `{símbolo} {nombre} | Pop:{n} | Level:{l} | 🍞 {food} | ⛏️ {tools} |  {gold}`

Implementación en `renderMap()` de `view.go`:
```go
for _, info := range m.settlementOverlay {
    if info.WorldX == wx && info.WorldY == wy {
        // Obtener ResourceStore del settlement
        rs, ok := ecs.GetComponent[settlement.ResourceStore](m.ecsWorld, ecs.Entity(info.Entity))
        if ok {
            status = fmt.Sprintf(" %s %s | Pop:%d | Lv:%d | 🍞 %.1f | ⛏️ %.1f |  %.1f ",
                string(info.Symbol), info.Name, info.Population, 
                rs.Resources["food"], rs.Resources["tools"], rs.Resources["gold"])
        }
    }
}
```

### Inspector de Settlement (expandido)

Añadir recursos y estado de hambruna al inspector existente:

```
=== Settlement ===
Name: Aldea del Norte
Type: village
Population: 5/15
Level: 2
Radius: 4
Buildings: house, farm, market
Resources:
  🍞 Food:  45.0 / 100  (to Lv.3)
  ⛏️ Tools:  3.0 / 50
   Gold: 12.0 / 25
Status: ✅ Estable
```

### Símbolo de Edificio Ocupado vs Vacío

Para el overlay de edificios (futuro, no MVP):
- Building ocupado: símbolo normal + indicador (e.g., granja con '@' cerca)
- Building vacío: símbolo atenuado

**MVP**: los edificios no se renderizan individualmente en el mapa (solo cuentan como parte del settlement). Se muestran solo en el inspector.

---

## Integración GOAP

### Nuevas Acciones Económicas

Las acciones GOAP actuales solo modifican Hunger/Fatigue. Para integrar economía, hay dos caminos:

#### Camino A: GOAP solo gestiona necesidades (RECOMENDADO MVP)

GOAP sigue igual. La economía es un sistema ECS aparte que no necesita planificación — es automática. El NPC trabaja si tiene Job y hay un building compatible. No necesita "decidir" trabajar.

- `SettlementEconomySystem` asigna workers automáticamente basado en Job + building disponible
- GOAP sigue planificando harvest/forage/rest para necesidades individuales
- **No hay nuevas acciones GOAP económicas en MVP**

#### Camino B: Acciones GOAP vinculadas a buildings (post-MVP)

```yaml
- id: work_at_farm
  name: Trabajar en Granja
  requires:
    building: farm
    needs_min: {hunger: 0.0, fatigue: 0.0}
    needs_max: {hunger: 0.8, fatigue: 0.8}
  effects:
    hunger_change: -0.1
    fatigue_change: 0.05
  economy:
    produces: {food: 2.0}
  reward:
    base: 1.0

- id: rest_at_home
  name: Descansar en Casa
  requires:
    building: house
    needs_min: {hunger: 0.0, fatigue: 0.3}
    needs_max: {hunger: 1.0, fatigue: 1.0}
  effects:
    fatigue_change: -0.3
  reward:
    base: 0.8

- id: trade_at_market
  name: Comerciar
  requires:
    building: market
    needs_min: {hunger: 0.0, fatigue: 0.0}
    needs_max: {hunger: 0.6, fatigue: 0.6}
  effects:
    hunger_change: -0.05
    fatigue_change: 0.0
  economy:
    consumes: {food: 0.5}
    produces: {gold: 1.0}
  reward:
    base: 0.6
```

Para que GOAP funcione con buildings, el planner necesita saber:
1. En qué biome está el NPC
2. Qué buildings hay en su settlement
3. Su estado actual de necesidades

Esto requiere modificar `goap.Plan()` para aceptar un contexto de buildings disponibles, o bien filtrar acciones por building antes de planificar.

**Decisión MVP**: GOAP no cambia. La economía es un sistema separado. GOAP sigue gestionando solo necesidades. En post-MVP se vinculan.

---

## Almacenamiento de Producción en YAML

### Extensión de buildings.yaml

```yaml
kind: building-types
data:
  - id: house
    name: Casa
    # No produce, no consume
    max_workers: 0

  - id: farm
    name: Granja
    role: farmer
    produces:
      food: 2.0
    max_workers: 3

  - id: market
    name: Mercado
    role: merchant
    produces:
      gold: 1.0
    consumes:
      food: 0.5
    max_workers: 2

  - id: tavern
    name: Taberna
    role: merchant
    consumes:
      food: 0.2
    max_workers: 1
    # No produce (es punto de consumo social)

  - id: temple
    name: Templo
    role: priest
    consumes:
      gold: 0.1
    max_workers: 1

  - id: blacksmith
    name: Herreria
    role: smith
    produces:
      tools: 1.0
    max_workers: 2
```

### Estructura Interna (carga en data.go)

```go
type BuildingDef struct {
    ID         string            `yaml:"id"`
    Name       string            `yaml:"name"`
    Role       string            `yaml:"role"`
    Produces   map[string]float64 `yaml:"produces"`
    Consumes   map[string]float64 `yaml:"consumes"`
    MaxWorkers int               `yaml:"max_workers"`
}
```

Los datos existentes (`buildings.yaml` actual sin `produces`) siguen siendo válidos — campos zero-value significan "no produce".

---

## Alcance MVP

### ✅ Fase 1: Infraestructura económica
- Extender `BuildingDef` con `Role`, `Produces`, `Consumes`, `MaxWorkers`
- Extender `buildings.yaml` con datos de producción
- Cargar y validar nuevos campos en `LoadBuildingTypes`
- Tests: YAML con produces, consumes se carga correctamente

### ✅ Fase 2: SettlementEconomySystem
- Sistema ECS que produce recursos según workers asignados y buildings
- Consumo automático de food por NPC
- ResourceStore ahora se escribe y se lee realmente
- Tests: farm con 2 farmers produce +4.0 food/tick, NPCs consumen 0.01 food/tick

### ✅ Fase 3: SettlementGrowthSystem
- Sistema ECS que verifica thresholds de recursos para subir de nivel
- `data/growth.yaml` con thresholds por nivel
- Al subir de nivel: +Radius, nuevos buildings spawnean, +max_population
- Tests: settlement con recursos suficientes sube de nivel, sin recursos no sube

### ✅ Fase 4: Hambruna y Migración
- Si settlement tiene food deficit, NPCs pierden HomeReference (migran)
- FamineSystem: threshold de -10 food acumulado → migración
- Tests: settlement sin food por 10+ ticks → NPCs migran

### ✅ Fase 5: Visualización TUI
- Status bar: recursos del settlement bajo cursor
- Inspector expandido con recursos, level progress, estado
- Tests: status bar muestra food/gold/tools correctamente

### 🔲 Post-MVP
- GOAP con acciones vinculadas a buildings (work_at_farm, rest_at_home)
- Muerte de NPCs por hambruna (no solo migración)
- Interfaz comercial entre settlements
- Data-driven completo con recipes en YAML
- Balance tuning: rates de producción, consumo, growth

---

## Recomendación

**Approach 1: SettlementEconomySystem como sistema ECS independiente en `internal/simulation/economy/`.**

Razones:
1. Separación limpia — economía no se mezcla con spawn, wander, o render
2. Sigue el patrón ECS establecido del proyecto
3. ResourceStore ya existe — solo necesita ser escrito
4. La producción data-driven desde YAML extendido evita código duplicado
5. Cada sistema es testeable independientemente
6. Fácil desactivar (quitar del SystemManager) para debugging

### Orden de Implementación

```
Fase 1: YAML extendido (BuildingDef + growth.yaml) + loaders + tests
Fase 2: ResourceStore helpers (Add/Remove/Has) + tests
Fase 3: SettlementEconomySystem + tests
Fase 4: SettlementGrowthSystem + tests
Fase 5: FamineSystem + tests  
Fase 6: TUI status bar + inspector expandido + tests
Fase 7: Main wiring + smoke test
```

### ADRs Clave

- **ADR-1**: Production data vive en el sistema (map cargado de YAML), NO como componente ECS. Los buildings siguen siendo ligeros (`Building{ID, Name, Level}`).
- **ADR-2**: GOAP no se modifica en MVP. La economía es automática (sistema ECS), no planificada (GOAP). Se vinculan en post-MVP.
- **ADR-3**: Hambruna inicialmente causa migración (pérdida de HomeReference), no muerte. La muerte se añade cuando el sistema está maduro.
- **ADR-4**: Los thresholds de crecimiento son data-driven (YAML) desde el día 1. No hardcodeados.

---

## Riesgos

| Riesgo | Probabilidad | Mitigación |
|---|---|---|
| **Desequilibrio económico**: producción >> consumo o viceversa | Alta | Datos en YAML → tuning sin recompilar. Tests de integración con rates fijos detectan desbalance |
| **ResourceStore sin inicializar** en settlements existentes (save games) | Media | SettlementEconomySystem inicializa ResourceStore si no existe (lazy init) |
| **Settlement sin buildings productivos** (solo houses) | Media | No produce nada, solo consume → NPCs migran naturalmente (comportamiento correcto) |
| **Overflow numérico**: recursos acumulados sin límite | Baja | Sin límite para MVP (como Dwarf Fortress clásico). En post-MVP: decay o límite por nivel |
| **Multiplicación de workers**: mismo NPC cuenta para múltiples buildings | Baja | Cada NPC tiene un solo Job → se asigna al building que corresponde a su Role |
| **GOAP y economía compiten por Hunger/Fatigue** | Media | GOAP gestiona necesidades individuales, economía gestiona recursos del settlement. No compiten porque operan en niveles distintos |

---

## Ready for Proposal

**Sí.** Esta exploración tiene suficiente profundidad para proceder con:

1. **sdd-propose**: Definir alcance exacto del MVP económico con fases y rollback plan
2. **sdd-spec**: Especificaciones Given/When/Then para: producción, consumo, crecimiento, hambruna, visualización
3. **sdd-design**: Diseño técnico detallado de SettlementEconomySystem, SettlementGrowthSystem, FamineSystem
4. **sdd-tasks**: Breakdown en tareas implementables individualmente con Strict TDD

### Archivos a Crear/Modificar

| Archivo | Acción |
|---|---|
| `internal/simulation/economy/systems.go` | **CREATE** — SettlementEconomySystem, SettlementGrowthSystem, FamineSystem |
| `internal/simulation/economy/systems_test.go` | **CREATE** — Tests completos |
| `internal/simulation/settlement/components.go` | **MODIFY** — Helpers Add/Remove/Has en ResourceStore |
| `internal/simulation/settlement/types.go` | **MODIFY** — BuildingDef con Role, Produces, Consumes, MaxWorkers |
| `internal/simulation/settlement/data.go` | **MODIFY** — LoadBuildingTypes con nuevos campos |
| `internal/simulation/settlement/data_test.go` | **MODIFY** — Tests con produces/consumes en YAML |
| `data/buildings.yaml` | **MODIFY** — Añadir role, produces, consumes, max_workers |
| `data/growth.yaml` | **CREATE** — Thresholds de crecimiento |
| `internal/ui/view.go` | **MODIFY** — Status bar con recursos, inspector expandido |
| `internal/ui/model.go` | **MODIFY** — Si necesita nuevos campos para recursos |
| `internal/ui/view_test.go` | **MODIFY** — Tests de visualización económica |
| `cmd/evociv/main.go` | **MODIFY** — Registrar SettlementEconomySystem, SettlementGrowthSystem |
| `openspec/changes/economy/` | SDD artifacts completos |
