# Design: GOAP + Q-Learning Híbrido

## Technical Approach

Sistema híbrido GOAP+RL sobre el ECS existente. 6 acciones data-driven vía YAML (`kind: npc-actions`). Necesidades (Hunger, Fatigue) como componente ECS separado con decaimiento por tick y LOD scaling. GOAP forward-chaining simple selecciona acción según urgencia de necesidad. Q-Learning ε-greedy optimiza selección con tabla discreta ~72 estados/NPC. GOAP es fallback cuando Q < threshold (0.1). Persistencia Q-table en SQLite batch-write.

## Architecture Decisions

### ADR 1: GOAP Forward-Chaining vs Backward-Chaining
| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| Backward (STRIPS-like) | Más expresivo, prof ≤3 es overkill, más complejo | ❌ |
| **Forward-chaining greedy** | Simple, prof≤3 es suficiente para 6 acciones, zero-shot desde tick 1 | ✅ |

El espacio de acciones es pequeño (<10). Backward-chaining añade complejidad sin beneficio. La exploración ya analizó esto en detalle.

### ADR 2: Q-table en Memoria + SQLite vs Solo Memoria
| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| Solo memoria | Rápido, pero Q se pierde al cerrar | ❌ |
| **Memoria + SQLite batch** | ~500KB para 100 NPCs, batch-write cada N ticks, persiste entre sesiones | ✅ |

Tabla Q en memoria (`map[Entity]map[string]map[string]float64`) para acceso O(1). SQLite como respaldo: `SaveQTable` en Close, batch cada 10 ticks si hay cambios.

### ADR 3: ε-Greedy vs UCB / Thompson Sampling
| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| UCB | Mejor convergencia teórica, requiere contadores de visita | ❌ |
| Thompson | Muestreo beta, más preciso pero más caro | ❌ |
| **ε-greedy** | Simple, ε dance 0.5→0.05, GOAP cubre cold start | ✅ |

72 estados × 6 acciones = 432 pares. ε-greedy es suficiente para este espacio y GOAP evita la exploración temprana mala.

### ADR 4: LOD Escalonado para GOAP vs Todos Igual
| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| Todos corren GOAP completo | 100 NPCs planificando cada tick = ~5-10ms | ❌ |
| **LOD: local=full, near=1-step, distant=wander** | 5-15 NPCs locales planifican, resto simplificado o wander | ✅ |

LOD ya existe en el ECS. Reusamos el mismo threshold Chebyshev (5/15). Los NPCs distant siguen con WanderSystem existente.

### ADR 5: Necesidades como Componente Separado vs Dentro de AIState
| Opción | Tradeoff | Decisión |
|--------|----------|----------|
| Dentro de AIState | Menos stores, mezcla concerns cognitivos vs fisiológicos | ❌ |
| **Needs componente separado** | Independiente: NeedsDecaySystem no toca AIState, GOAPSystem lee ambos | ✅ |

Needs cambia cada tick (decaimiento). AIState cambia por planificación. Separarlos permite que NeedsDecaySystem corra en todos los LODs mientras GOAPSystem solo en LOD≥1.

## Data Flow

```
World.Update(dt)
  │
  ├── [1] NeedsDecaySystem (ALL LODs)
  │     └── needs.Hunger += 0.01 × lodMultiplier × dt
  │     └── needs.Fatigue += 0.005 × lodMultiplier × dt
  │     └── clamp [0, 1]
  │
  ├── [2] LODSystem (sin cambios existentes)
  │
  ├── [3] GOAPSystem (LOD≥1)
  │     ├── LOD=2: full plan (prof≤3 forward-chaining)
  │     ├── LOD=1: 1-step action (skip multi-step)
  │     └── LOD=0: skip (WanderSystem maneja)
  │
  ├── [4] QLearningSystem (LOD≥1)
  │     ├── LOD=2: ε-greedy select, execute, reward, update Q
  │     ├── LOD=1: same as local
  │     └── LOD=0: skip
  │
  ├── [5] WanderSystem (LOD≥1, sin cambios existentes)
  │
  └── [6] NPCRenderSystem (LOD≥1, sin cambios)
```

```
Q-Learning Loop:
  state = fmt.Sprintf("%s|%s|%s", needType, biomeID, timeOfDay)
  if maxQ(state) < 0.1 → GOAP fallback
  else → ε-greedy: explore (ε) or exploit (1-ε)
  execute action → compute reward = hungerReduction + fatigueReduction + completionBonus
  Q(s,a) += 0.1 × (reward + 0.9 × maxQ(s') - Q(s,a))
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/simulation/npc/components.go` | Modify | +`Needs{Hunger, Fatigue}` + `NeedsID` + register store |
| `internal/simulation/npc/types.go` | Modify | +`ActionDef`, `ActionRequires`, `ActionEffects`, `ActionReward` |
| `internal/simulation/npc/data.go` | Modify | +`LoadActions(registry)` using YAML kind `npc-actions` |
| `internal/simulation/npc/systems.go` | Modify | +3 systems: NeedsDecaySystem, GOAPSystem, QLearningSystem |
| `internal/simulation/goap/planner.go` | Create | `Plan(Needs, []ActionDef, string) ActionDef` forward-chaining |
| `internal/simulation/rl/qtable.go` | Create | `QTable` struct, `eGreedy()`, `Update()`, load/save methods |
| `data/actions.yaml` | Create | 6 acciones: harvest, forage, rest, socialize, trade, explore |
| `internal/store/store.go` | Modify | +`SaveQTable(npcID, qtable)`, `LoadQTable(npcID)` |
| `internal/store/sqlite.go` | Modify | +`CREATE TABLE qtable` + migrate + Save/Load impl |
| `cmd/evociv/main.go` | Modify | +Register 3 new systems after WanderSystem |
| `internal/simulation/npc/systems_test.go` | Modify | +Tests for 3 new systems |
| `internal/simulation/goap/planner_test.go` | Create | Unit tests for planner |
| `internal/simulation/rl/qtable_test.go` | Create | Unit tests for Q-learning |
| `internal/store/sqlite_test.go` | Modify | +Tests for qtable persistence |

## Interfaces / Contracts

```go
// Needs component — valores [0,1], aumenta con tiempo
type Needs struct {
    Hunger  float64
    Fatigue float64
}

// ActionDef — una acción GOAP data-driven
type ActionDef struct {
    ID       string         `yaml:"id"`
    Name     string         `yaml:"name"`
    Requires ActionRequires `yaml:"requires"`
    Effects  ActionEffects  `yaml:"effects"`
    Reward   ActionReward   `yaml:"reward"`
}

type ActionRequires struct {
    Biomes  []string      `yaml:"biomes"`
    NeedsMin NeedsValues  `yaml:"needs_min"`
    NeedsMax NeedsValues  `yaml:"needs_max"`
}

type ActionEffects struct {
    HungerChange  float64 `yaml:"hunger_change"`
    FatigueChange float64 `yaml:"fatigue_change"`
}

type ActionReward struct {
    Base float64 `yaml:"base"`
}

type NeedsValues struct {
    Hunger  float64 `yaml:"hunger"`
    Fatigue float64 `yaml:"fatigue"`
}

// QTable — tabla Q discreta en memoria
type QTable struct {
    mu     sync.RWMutex
    values map[string]map[string]float64 // state_key → action_id → q_value
}

// eGreedy selecciona acción con ε probabilidad de exploración
func (qt *QTable) eGreedy(state string, actions []ActionDef, rng *rand.Rand) string

// Update ejecuta Q(s,a) += α * (reward + γ * maxQ(s') - Q(s,a))
func (qt *QTable) Update(state, action string, reward float64, nextState string, alpha, gamma float64)

// GOAP planner
func Plan(needs Needs, actions []ActionDef, biome string) ActionDef
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Needs decay | Hunger +0.01/tick, Fatigue +0.005, clamp [0,1], LOD 0.5x distant |
| Unit | LoadActions YAML | 6 actions load, missing fields error, invalid YAML error |
| Unit | GOAP planner | Hunger high→harvest/forage, Fatigue high→rest, low needs→explore/socialize |
| Unit | Q-Learning | Zero-init Q, ε-greedy approx 50% random at ε=0.5, convergence after 1000 iters |
| Unit | Q-table persistence | Save → Load → identical, empty → empty map, open/close cycle |
| Integration | LOD systems | LOD=2 full plan, LOD=1 1-step, LOD=0 no plan/learning, LOD=0 0.5x decay |
| Integration | main.go registration | 3 systems registered in correct order after WanderSystem |

## Migration / Rollout

No migration required. `CREATE TABLE IF NOT EXISTS qtable` se ejecuta en `migrate()`. No hay datos existentes que migrar. Tabla nueva sin impacto en schema previo.
