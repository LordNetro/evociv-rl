# Exploration: GOAP + RL híbrido para NPCs

## Current State

El sistema actual tiene NPCs con componentes ECS preparados para GOAP pero sin comportamiento inteligente:

- **AIState** componente ya existe: `Goals []string`, `Plan []string`, `Mood float64` — el scaffold está listo
- **WanderSystem** usa `ai.Plan` como cola de un solo paso (coordenada destino) — es un proto-GOAP simplificado
- **4 sistemas funcionando**: NPCSpawnSystem, WanderSystem, LODSystem, NPCRenderSystem
- **50-100 NPCs** con razas, roles, personalidad Big 5, biomas
- **Directorios `goap/`, `rl/`, `economy/` existen pero están vacíos** — esperando implementación
- **Persistencia SQLite** solo guarda metadatos del mundo (seed), no hay tabla para Q-learning ni estado NPC
- **YAML data-driven** con loader maduro (loader.go + registry.go + kind-based routing)
- **LOD sistémico**: 3 niveles (distant/near/local) basado en Chebyshev distance (5/15 thresholds)

## Affected Areas

| Path | Por qué |
|------|---------|
| `internal/simulation/goap/` | Directorio vacío — implementación GOAP aquí |
| `internal/simulation/rl/` | Directorio vacío — implementación RL aquí |
| `internal/simulation/economy/` | Directorio vacío — recursos/inventario futuro |
| `internal/simulation/npc/components.go` | Nuevos componentes: Needs, Inventory, QTable (o ID ref) |
| `internal/simulation/npc/systems.go` | Nuevo GOAPSystem + NeedsDecaySystem; refactor WanderSystem |
| `internal/simulation/npc/types.go` | Nuevos tipos: ActionDef, NeedType, QState |
| `internal/simulation/npc/data.go` | LoadActions() — cargar acciones desde YAML |
| `data/actions.yaml` | **NUEVO** — acciones data-driven |
| `internal/store/sqlite.go` | Nueva tabla `qtable` para persistencia Q |
| `internal/store/store.go` | Interfaz Store extender con Q-table methods |
| `internal/ecs/world.go` | Sin cambios — el ECS ya soporta sistemas arbitrarios |
| `cmd/evociv/main.go` | Registrar nuevo GOAPSystem + NeedsDecaySystem |
| `openspec/specs/npc-systems/spec.md` | Actualizar con nuevos requirements |

## Approaches

### 1. GOAP Puro (sin RL)

**Forward-chaining**: desde el estado actual, explora acciones que reducen necesidades hasta encontrar un plan válido.

```
Estado: {hunger: 0.8, fatigue: 0.3, biome: plains}
→ Acciones factibles: [harvest_wheat, forage_berries, rest]
→ Evalúa efectos: harvest_wheat reduce hunger -0.3, fatigue +0.1
→ Plan: [move_to(plains_farm), harvest_wheat, eat]
```

**Pros**:
- Simple, determinista, fácil de testear y depurar
- Sin estado que persisitir (zero-shot planning)
- Comportamiento explicable: "NPC tiene hambre → busca comida"

**Cons**:
- Cada NPC replanifica desde cero cada tick → O(n*m) con n=NPCs, m=acciones
- No aprende de experiencia: mismo error siempre
- Difícil de sintonizar: pesos manuales para cada situación
- Sin adaptación a cambios del mundo (estaciones, escasez)

**Esfuerzo**: Medium (~3 días)

### 2. GOAP + Q-Learning Híbrido (RECOMENDADO)

GOAP planifica a corto plazo; Q-Learning optimiza decisiones a largo plazo.

**Arquitectura en 2 capas**:
1. **GOAP layer** (planificación táctical): dado un objetivo (ej. "reducir hambre"), genera secuencia de acciones
2. **Q layer** (selección estratégica): elige QUÉ objetivo perseguir basado en experiencia

```
Cada tick:
1. NeedsDecaySystem: incrementa necesidades (hunger +0.01/tick, etc.)
2. GOAPSystem:
   a. Si no tiene plan activo o plan falló:
      - Evalúa todas las necesidades, calcula "urgencia" = need_value * (1 + Neuroticism*0.5)
      - Q-table consulta: dado estado S = (need_highest, biome, hour), mejor acción A
      - Si Q-value > threshold: ejecuta acción aprendida
      - Si no: GOAP planifica forward desde estado actual
   b. Si tiene plan activo: ejecuta siguiente paso
3. Al completar acción: calcula reward = reducción de necesidad, actualiza Q(s,a)
```

**Q-State discreto**: `(highest_need_idx 0..3, biome_id 0..5, time_of_day 0..2)` = ~4*6*3 = 72 estados
**Q-Actions**: mismas que acciones GOAP + "wander" por defecto
**Reward**: `Δhunger*0.4 + Δfatigue*0.3 + Δsocial*0.2 + Δwealth*0.1` (pesos por personalidad)

**Pros**:
- NPCs aprenden con el tiempo qué funciona en cada situación
- GOAP da comportamiento sensato desde el día 1 (zero-shot)
- Q-Learning refina con experiencia (cold start manejado por GOAP)
- 72 estados es tiny — tabla Q cabe en ~2KB por NPC
- Personalidad influye en pesos de reward → NPCs únicos

**Cons**:
- Más complejo que GOAP puro
- Dos sistemas que mantener (GOAP + RL)
- Q-table necesita persistencia o se pierde entre sesiones
- Curva de aprendizaje: NPCs pueden tomar decisiones subóptimas early

**Esfuerzo**: Medium-High (~5-7 días)

### 3. RL End-to-End (Deep Q-Network)

Red neuronal pequeña (3 capas fully connected) en vez de tabla Q.

**Pros**:
- Estado continuo (no discretizar)
- Generaliza mejor a estados no vistos
- Sin tabla Q que explote en size

**Cons**:
- Dependencia externa (Golang + tensores = painful)
- Cold start terrible sin GOAP de respaldo
- Overkill para 72 estados
- Depuración muy difícil
- Sin GPU, entrenamiento lento

**Esfuerzo**: Very High (~2-3 semanas)

## Recommendation

**GOAP + Q-Learning híbrido** (opción 2). Razones:

1. **Arquitectura ya preparada**: AIState con Goals/Plan, directorios goap/rl vacíos
2. **Comportamiento inmediato**: GOAP funciona desde el primer commit sin entrenamiento
3. **Espacio de estado manejable**: 72 estados discretos → tabla Q en memoria ~2KB/NPC
4. **Sin dependencias externas**: tabla Q es un `map[QState]map[ActionID]float64` — Go puro
5. **Data-driven desde el inicio**: acciones en YAML, separación concerns
6. **MVP alcanzable en ~1 semana**: necesidades básicas + 6 acciones + GOAP + Q-table en SQLite

## Risks

| Riesgo | Impacto | Mitigación |
|--------|---------|------------|
| **R1: Perf con 100 NPCs planificando** | Alto — cada NPC corriendo GOAP forward-chaining cada tick | LOD: local = full GOAP, near = GOAP simplificado cada 5 ticks, distant = wander-only |
| **R2: Q-table no converge** | Medio — rewards mal diseñados | Usar GOAP como fallback; si Q(s,a) < threshold, delegar a GOAP |
| **R3: Acciones YAML muy verbose** | Bajo — mantenible con validación | Reusa patrón validator.go existente |
| **R4: NPCs no se comportan como su rol** | Medio — farmer debería preferir harvest sobre hunt | Role bias en Q-learning initial values; personality weights |
| **R5: Persistencia Q-table muy grande** | Bajo — 72 estados × 10 acciones × 8 bytes = ~5KB/NPC = 500KB para 100 NPCs | SQLite tabla única con (npc_id, state_hash, action_id, q_value) |

## MVP Scope

### In Scope (Primer entregable)

| Componente | Descripción |
|------------|-------------|
| **Needs component** | Hunger (0..1), Fatigue (0..1) — decae cada tick |
| **NeedsDecaySystem** | Incrementa necesidades según tiempo + biome conditions |
| **4-6 acciones data-driven** | harvest_wheat, forage_berries, hunt_animal, rest, wander, mine_stone |
| **GOAP planner** | Forward-chaining simple: desde estado actual, aplica acciones hasta reducir necesidad más urgente |
| **Q-table in-memory** | `map[QState]map[ActionID]float64` por NPC |
| **Q-Learning loop** | ε-greedy selection, reward = Δneed, α=0.1, γ=0.9 |
| **SQLite persistence** | Tabla `qtable(npc_id, state_hash, action_id, q_value)` |
| **GOAPSystem** | System ECS que itera NPCs con AIState + Needs |
| **Comportamiento visible** | NPCs con hambre se mueven a biomas con comida |

### Out of Scope (MVP+)

| Componente | Cuándo |
|------------|--------|
| Social need | MVP+1 |
| Wealth / economy system | MVP+2 |
| Personality affecting rewards | MVP+1 |
| Tool/inventory system | MVP+2 |
| Multi-step GOAP plans (>3 pasos) | MVP+1 |
| DQN / Deep RL | Post-MVP |
| LLM dialogue integration | Post-MVP |
| NPC-NPC interactions | Post-MVP |

## Detailed Architecture

### 1. Nuevos Componentes ECS

```go
// Needs — valores 0..1, 1 = máxima necesidad
type Needs struct {
    Hunger  float64
    Fatigue float64
}

// NeedsID = ecs.NewComponentID("npc_needs")

// QTableRef — referencia a tabla Q del NPC (evita tener map en ECS)
type QTableRef struct {
    NPCID uint64 // NPC entity ID para lookup en QTableStore global
}

// QTableRefID = ecs.NewComponentID("npc_qtable")
```

### 2. GOAP como ECS System

```go
type GOAPSystem struct {
    wm        *world.WorldMap
    actions   []ActionDef
    qtable    *QTableStore  // global Q-table manager
    rng       *rand.Rand
    tickCount int
}

func (s *GOAPSystem) Update(w *ecs.World, dt float64) error {
    // Por LOD:
    // - Local: planificar cada tick
    // - Near: planificar cada 5 ticks
    // - Distant: skip (WanderSystem se encarga)
    
    for e, lod := range lodStore.All() {
        if lod.Level < LODLocal {
            // near: planificar cada 5 ticks
            if s.tickCount % 5 != 0 { continue }
        }
        
        needs := needsStore.Get(e)
        ai := aiStore.Get(e)
        pos := posStore.Get(e)
        tile := s.wm.TileAt(int(pos.X), int(pos.Y))
        
        // 1. Evaluar necesidad más urgente
        targetNeed := argmax(needs.Hunger, needs.Fatigue)
        
        // 2. Consultar Q-table para mejor acción en este estado
        state := QState{
            HighestNeed: targetNeed,
            BiomeID:     tile.BiomeID,
        }
        bestAction := s.qtable.BestAction(state, e)
        
        // 3. Si Q-value > threshold, ejecutar acción aprendida
        //    Si no, GOAP planifica
        if s.qtable.Value(state, bestAction, e) > 0.3 {
            executeAction(e, bestAction, w, s.wm)
        } else {
            plan := s.goapPlan(state, needs, ai.Goals)
            ai.Plan = plan
            aiStore.Set(e, ai)
        }
    }
}
```

### 3. Acciones YAML

```yaml
kind: npc-actions
data:
  - id: harvest_wheat
    name: "Cosechar trigo"
    type: gather
    cooldown: 3
    requires:
      biome: [plains]
      need_above: { hunger: 0.3 }  # solo si hambre > 0.3
    effects:
      hunger: -0.3
      fatigue: +0.15
    reward:
      base: 0.5
      tags: [work, outdoor]
    animation: harvest

  - id: forage_berries
    name: "Recolectar bayas"
    type: gather
    cooldown: 1
    requires:
      biome: [forest]
    effects:
      hunger: -0.15
      fatigue: +0.05
    reward:
      base: 0.3
      tags: [forage, outdoor]

  - id: hunt_animal
    name: "Cazar animal"
    type: hunt
    cooldown: 5
    requires:
      biome: [forest, plains]
      skill: { hunting: 1 }
    effects:
      hunger: -0.5
      fatigue: +0.25
    reward:
      base: 0.8
      tags: [work, outdoor, dangerous]

  - id: rest
    name: "Descansar"
    type: rest
    cooldown: 0
    requires:
      need_above: { fatigue: 0.2 }
    effects:
      hunger: +0.02  # descansar gasta un poco de energía
      fatigue: -0.4
    reward:
      base: 0.4
      tags: [rest, idle]

  - id: wander
    name: "Vagar"
    type: move
    cooldown: 0
    requires: {}
    effects: {}
    reward:
      base: 0.0
      tags: [idle]

  - id: mine_stone
    name: "Minar piedra"
    type: gather
    cooldown: 4
    requires:
      biome: [desert, plains]
      need_above: { fatigue: 0.0 }
    effects:
      hunger: -0.1
      fatigue: +0.3
    reward:
      base: 0.4
      tags: [work, outdoor]
```

### 4. Almacenamiento Q-Table

**En memoria** (por performance):
```go
type QState struct {
    HighestNeed int    // 0=hunger, 1=fatigue (encoded como int)
    BiomeID     string // "plains", "forest", etc.
}

type QTableStore struct {
    mu     sync.RWMutex
    tables map[ecs.Entity]map[QState]map[string]float64
    // tables[entity][state][actionID] = q_value
}
```

**Persistencia SQLite** (cada N ticks o al cerrar):

```sql
CREATE TABLE IF NOT EXISTS qtable (
    npc_id     INTEGER NOT NULL,
    state_hash TEXT    NOT NULL,  -- fmt.Sprintf("%d|%s", highestNeed, biomeID)
    action_id  TEXT    NOT NULL,
    q_value    REAL    NOT NULL DEFAULT 0.0,
    visits     INTEGER NOT NULL DEFAULT 0,
    updated_at TEXT    DEFAULT (datetime('now')),
    PRIMARY KEY (npc_id, state_hash, action_id)
);

CREATE INDEX idx_qtable_npc ON qtable(npc_id);
```

### 5. LOD Escalabilidad

| LOD | Ticks entre plan | GOAP | Q-Learning | Sistema |
|-----|-----------------|------|------------|---------|
| Local (≤5 tiles) | 1 tick | Full | ε=0.1 | GOAPSystem + NeedsDecay |
| Near (≤15 tiles) | 5 ticks | Simplified | ε=0.05 | GOAPSystem (skip si plan OK) |
| Distant (>15) | ∞ | None (Wander) | Frozen | WanderSystem existente |

### 6. Ciclo de Actualización por Tick

```
1. NeedsDecaySystem (todos los NPCs):
   - hunger += 0.01 * dt
   - fatigue += 0.005 * dt
   - Clamp a [0, 1]

2. LODSystem (todos los NPCs):
   - Calcula distancia al player
   - Asigna LOD

3. GOAPSystem (LOD ≥ 1):
   - Si LOD=Local: full tick
   - Si LOD=Near: cada 5 ticks
   - Si plan vacío o completado:
     a. Determinar necesidad más urgente
     b. Consultar Q-table o GOAP planner
     c. Generar/ejecutar plan
   - Si plan activo: ejecutar siguiente paso del plan
   - Reward al completar acción: Δneed, update Q
   - Persist Q-table a SQLite cada N actualizaciones

4. WanderSystem (LOD=0):
   - Movimiento aleatorio (sin cambios)

5. NPCRenderSystem (LOD ≥ 1):
   - Recoger infos de render (sin cambios)
```

### 7. Forward-Chaining GOAP (Simplificado)

```
function plan(state, needs, goal) -> []ActionID:
    1. Si goal está satisfecho (need < threshold): return []
    2. Filtrar acciones factibles (biome, necesidades)
    3. Ordenar por mejor ratio efecto/costo
    4. Para cada acción candidata:
       a. Simular efecto en estado
       b. Si acerca al goal: seleccionar
       c. Si no: backtrack (máx 3 pasos de profundidad)
    5. Si no hay acción que mejore: return [wander]
```

No es A* ni STRIPS completo — es una heurística greedy con profundidad limitada. Suficiente para MVP. Forward-chaining porque el espacio de acciones es pequeño (<10).

## Testing Strategy

### Test de planificación (unitarios)

```go
func TestGOAP_PlannerSelectsActionForHunger(t *testing.T) {
    state := SimulationState{
        Needs: Needs{Hunger: 0.8, Fatigue: 0.1},
        Biome: "plains",
    }
    actions := []ActionDef{
        harvestWheat, // hunger -0.3
        rest,         // fatigue -0.4
        wander,       // no effect
    }
    planner := NewGOAPPlanner(actions)
    plan := planner.Plan(state, "satisfy_hunger")
    
    // Debería elegir harvest_wheat sobre rest/wander
    if len(plan) == 0 || plan[0] != "harvest_wheat" {
        t.Errorf("expected harvest_wheat, got %v", plan)
    }
}
```

### Test de Q-Learning (convergencia)

```go
func TestQLearning_ConvergesAfterRepetition(t *testing.T) {
    agent := NewQLearningAgent(0.1, 0.9, 0.1) // alpha, gamma, epsilon
    state := QState{HighestNeed: 0, BiomeID: "plains"}
    
    for i := 0; i < 1000; i++ {
        action := agent.SelectAction(state)
        reward := simulateAction(action, state)
        agent.Learn(state, action, reward, state) // same state (no transition)
    }
    
    // Después de 1000 iteraciones, harvest_wheat debería tener Q > wander
    qHarvest := agent.Q(state, "harvest_wheat")
    qWander := agent.Q(state, "wander")
    if qHarvest <= qWander {
        t.Errorf("harvest_wheat (%f) should beat wander (%f) after 1000 iterations", qHarvest, qWander)
    }
}
```

### Test de integración (NPC con hambre → busca comida)

```go
func TestGOAPSystem_NPCMovesToFoodSource(t *testing.T) {
    // GIVEN un NPC en biome desert con hambre=0.9
    // WHEN GOAPSystem ejecuta 10 ticks
    // THEN NPC debería moverse hacia biome vecino con comida (plains/forest)
    
    mibieraSetup(t, "desert", npcPos{5,5})
    AddComponent(w, npc, Needs{Hunger: 0.9, Fatigue: 0.1})
    
    goapSys := NewGOAPSystem(wm, actions, qtable, rng)
    for i := 0; i < 10; i++ {
        goapSys.Update(w, 1.0)
    }
    
    pos := GetComponent[Position](w, npc)
    tile := wm.TileAt(int(pos.X), int(pos.Y))
    if tile.BiomeID == "desert" {
        t.Error("NPC should have moved toward a food biome")
    }
    // Verificar que necesidad bajó
    needs := GetComponent[Needs](w, npc)
    if needs.Hunger >= 0.9 {
        t.Error("Hunger should have decreased after moving to food biome")
    }
}
```

### Test de persistencia Q-table

```go
func TestQTablePersistence(t *testing.T) {
    // GIVEN Q-table con valores aprendidos
    // WHEN guardar a SQLite y recargar
    // THEN valores deben coincidir
}
```

## Ready for Proposal

**Yes**. La exploración es completa y hay suficiente entendimiento para crear el proposal, specs, y design.

Los directorios `internal/simulation/goap/`, `internal/simulation/rl/` ya existen pero están vacíos — la implementación no comenzó, todo está por hacer.

**Próximo paso recomendado**: `sdd-propose` para formalizar el cambio "goap-rl" con alcance MVP, seguido de `sdd-spec` para las delta specs sobre npc-systems y la nueva definición de acciones.
