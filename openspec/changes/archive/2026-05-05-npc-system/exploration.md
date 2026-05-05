# Exploration: Sistema de NPCs

## Current State

El proyecto tiene una base sólida pero sin NPCs:

- **ECS map-based**: `Entity(uint64)`, `ComponentStore[T]`, `World` con `NewEntity()`/`AddComponent`/`GetComponent`/`RemoveEntity`, `System` interface, `SystemManager`
- **Componentes actuales**: `Position{X,Y float64; Z int}`, `Name{string}`, `Tags{[]string}`
- **WorldMap**: Grid 2D de `Tile{Height,Humidity,Temperature,BiomeID}`, 256×256
- **Data-Driven**: `Loader` YAML + `Registry` genérico, `biomes.yaml`, `gen-config.yaml`
- **TUI**: Bubbletea con pantalla welcome + map screen (cámara wasd, biome symbols)
- **Store SQLite**: `worlds` table con seed/width/height
- **Simulación**: Directorios `economy/`, `goap/`, `rl/` vacíos — no hay NPCs, ni asentamientos, ni IA

## Affected Areas

- `internal/ecs/component.go` — nuevos tipos de componente (Health, Personality, Job, Inventory, Family, AIState, Appearance, LOD)
- `internal/ecs/world.go` — RegisterComponentStore para nuevos tipos, posible mejora RemoveEntity
- `internal/simulation/npc/` — **NUEVO** paquete: spawner, systems (AI, wander, LOD)
- `internal/ui/model.go` — estado de selección cursor, lista NPCs visibles, inspector
- `internal/ui/view.go` — overlay NPC en mapa, panel inspector, leyenda
- `internal/store/store.go` — extensión de interface para NPC data
- `internal/store/sqlite.go` — migraciones y queries NPC
- `data/npcs.yaml` — **NUEVO** definiciones de razas, roles, traits, nombres
- `internal/data/validator.go` — validadores para NPC data (dependencias circulares, referencias)
- `cmd/evociv/main.go` — inicialización del spawner de NPCs

---

## 1. Componentes ECS para NPCs

### Propuesta de Componentes Base

```go
// === Existentes (reutilizar) ===
// Position  { X, Y float64; Z int }
// Name      { Name string }
// Tags      { Tags []string }

// === Nuevos ===

// Health — salud actual/máxima para mecánicas de supervivencia
type Health struct {
    Current, Max int
}

// Personality — Big Five traits normalizados [0.0, 1.0]
// Permite derivar comportamiento: agresión, sociabilidad, etc.
type Personality struct {
    Openness        float64 // apertura a experiencia
    Conscientiousness float64 // escrupulosidad
    Extraversion    float64 // extraversión
    Agreeableness   float64 // amabilidad
    Neuroticism     float64 // neuroticismo
}

// Job — rol productivo del NPC
type Job struct {
    Role      string       // "farmer", "blacksmith", "merchant", etc.
    Workplace Entity       // entidad del lugar de trabajo (futuro: building)
    SkillLevel float64
}

// Inventory — items que el NPC lleva consigo
type Inventory struct {
    Items []ItemStack      // ItemStack { ID string, Quantity int }
    Capacity int
}

// Family — relaciones sociales base
type Family struct {
    Spouse   Entity
    Children []Entity
    Dynasty  string // apellido / nombre de familia
}

// AIState — estado interno para GOAP + comportamientos
type AIState struct {
    Goals       []GOAPGoal        // prioridad: 0..1
    CurrentPlan []GOAPAction
    Memory      map[string]any    // blackboard: "last_seen_food_at", "home_location", etc.
    Mood        float64           // ánimo agregado [-1, 1]
}

// GOAPGoal — una meta con prioridad
type GOAPGoal struct {
    Type     string  // "find_food", "rest", "socialize", "work"
    Priority float64 // 0..1
}

// GOAPAction — una acción planificada
type GOAPAction struct {
    Type     string  // "move_to", "gather", "craft", "talk"
    TargetID string  // id del target
    Cost     float64
    Duration int     // ticks
}

// Appearance — representación visual en el mapa
type Appearance struct {
    Symbol rune
    Color  string  // hex color, ej: "#FFD700"
}

// NPCFlag — flags binarias para estado rápido
type NPCFlag uint64
const (
    NPCFlagAlive     NPCFlag = 1 << iota
    NPCFlagSleeping
    NPCFlagHungry
    NPCFlagBusy
    NPCFlagTrading
)

// LOD — Level of Detail para simulación diferencial
type LOD struct {
    Level    LODLevel  // Distant, Near, Local
    LastTick int       // último tick en que se simuló
}

type LODLevel int
const (
    LODDistant LODLevel = iota // solo existe, no se simula
    LODNear                    // wandering básico, economía simplificada
    LODLocal                   // GOAP completo, diálogos, eventos
)
```

### Tabla de Prioridades

| Componente | MVP | Post-MVP | Justificación |
|---|---|---|---|
| Health | ✅ | | Sin salud no hay simulación de riesgo |
| Personality | ✅ (baseline) | | Traits aleatorios desde YAML, usados por GOAP después |
| Job | ✅ | | Roles para economía y wandering contextual |
| Inventory | ❌ | ✅ | Requiere sistema de items/economía |
| Family | ❌ | ✅ | Requiere reproducción/herencia |
| AIState | ✅ (GOAP-ready) | | Necesario para wandering, aunque GOAP real sea post-MVP |
| Appearance | ✅ | | Necesario para visualización TUI |
| LOD | ✅ | | Crítico para performance con muchos NPCs |
| NPCFlag | ✅ | | Flags baratos para queries frecuentes |

---

## 2. Spawner

### Enfoque Recomendado: Híbrido Template + Biome Cluster

**Fase 1 — MVP:** Spawn aleatorio en biomas aptos con templates planos.

**Fase 2 — Post-MVP:** Spawn en asentamientos con distribución social.

### Algoritmo MVP

```
1. Definir NPC_COUNT = f(worldArea) → ej: 100 para 256×256
2. Para cada NPC:
   a. Elegir raza desde YAML (distribución ponderada)
   b. Elegir ubicación: muestreo aleatorio en biomas "habitables"
      - Habitable: plains, forest (weight=1.0)
      - Escaso: tundra, desert (weight=0.2)
      - Inhabitable: ocean, jungle (weight=0.0)
   c. Generar traits de personalidad con ruido gaussiano
   d. Asignar rol según biome/bioma (farmer en plains, hunter en forest)
   e. Crear Entity + attach componentes
3. Registrar en World
```

### YAML — `data/npcs.yaml`

```yaml
kind: npc-races
data:
  - id: human
    name: Humano
    description: "Versátil y adaptable"
    weights:
      spawn: 1.0
    traits:
      openness:        { mean: 0.5, std: 0.15 }
      conscientiousness: { mean: 0.5, std: 0.15 }
      extraversion:    { mean: 0.5, std: 0.15 }
      agreeableness:   { mean: 0.5, std: 0.15 }
      neuroticism:     { mean: 0.5, std: 0.15 }
    roles:
      - id: farmer
        weight: 0.4
        biomes: [plains]
      - id: hunter
        weight: 0.3
        biomes: [forest, jungle]
      - id: merchant
        weight: 0.2
        biomes: [plains]
      - id: artisan
        weight: 0.1
        biomes: [plains]
    name_pool:
      first: ["Aldric", "Bran", "Cedric", "Doran", "Elara", "Fina", "Greta", "Hilda"]
      last:  ["Torres", "Río", "Piedra", "Valle", "Bosque", "Puente"]
    symbol: "@"
    color: "#FFD700"

  - id: dwarf
    name: Enano
    weights:
      spawn: 0.3
    traits:
      conscientiousness: { mean: 0.7, std: 0.1 }
      extraversion:      { mean: 0.3, std: 0.1 }
    roles:
      - id: miner
        weight: 0.6
        biomes: [plains, desert]
      - id: smith
        weight: 0.4
        biomes: [plains]
    name_pool:
      first: ["Borin", "Durm", "Gimli", "Kili", "Thrain"]
      last:  ["Hierro", "Roca", "Martillo", "Horno"]
    symbol: "☺"
    color: "#C0C0C0"
```

### Esquema de Roles

```yaml
kind: npc-roles
data:
  - id: farmer
    name: Granjero
    description: "Trabaja la tierra"
    skill_base: 0.3
    skill_variance: 0.2
    biome_preference: [plains]
  - id: hunter
    name: Cazador
    description: "Caza animales"
    skill_base: 0.4
    skill_variance: 0.25
    biome_preference: [forest, jungle]
  - id: merchant
    name: Mercader
    description: "Comercia bienes"
    skill_base: 0.5
    skill_variance: 0.15
    biome_preference: [plains]
  - id: artisan
    name: Artesano
    description: "Crea objetos"
    skill_base: 0.4
    skill_variance: 0.2
    biome_preference: [plains]
  - id: miner  
    name: Minero
    description: "Extrae minerales"
    skill_base: 0.3
    skill_variance: 0.2
    biome_preference: [plains, desert]
```

### Arquitectura del Spawner (MVP)

```go
// internal/simulation/npc/spawner.go

type SpawnConfig struct {
    Count    int     // NPCs a generar (0 = auto: area * density)
    Density  float64 // NPCs por tile (default 0.002 → ~130 para 256²)
}

type NPCSpawner struct {
    world  *ecs.World
    wm     *world.WorldMap
    config SpawnConfig
}

func (s *NPCSpawner) Spawn(raceDefs []RaceDef, roleDefs []RoleDef) error {
    // 1. Calcular count si es auto
    // 2. Iterar count, generar NPCs con placement en biomas aptos
    // 3. Retornar error si algún paso falla
    return nil
}
```

---

## 3. Visualización TUI

### Estrategia MVP: Overlay + Inspector

**Overlay en mapa:** Cuando un tile tiene NPC(s), se muestra el símbolo del NPC (ej: `@`) en lugar del símbolo del bioma. Si hay múltiples NPCs en un tile, se muestra un contador (ej: `2`).

```go
// Modificación en renderMap() de view.go

func renderNPCOverlay(m Model, worldX, worldY int) string {
    npcs := m.npcsAt(worldX, worldY)
    if len(npcs) == 0 {
        return "" // sin overlay, mostrar biome normal
    }
    if len(npcs) == 1 {
        app, _ := ecs.GetComponent[Appearance](m.ecsWorld, npcs[0])
        return renderStyled(string(app.Symbol), app.Color)
    }
    // múltiples NPCs: mostrar contador
    count := len(npcs)
    return renderStyled(fmt.Sprintf("%d", min(count, 9)), "#FFD700")
}
```

**Inspector lateral:** Al seleccionar un tile (con cursor/tecla de selección), mostrar info del NPC en panel derecho.

```
Pantalla TUI (esquema):
┌──────────────────────┬──────────────┐
│  Mapa con NPCs @     │  Inspector   │
│  ~~~~~....TTT        │             │
│  ~~~...TTTddd        │  Aldric     │
│  ...TTTdddddd        │  ♥ 15/20   │
│  @..TTT.dddd         │  Granjero   │
│  .......dddd         │  [plains]   │
│                      │             │
│  [q] quit [wasd]     │  O: 0.45    │
│  move [e] inspect    │  C: 0.62    │
│                      │  E: 0.31    │
│                      │  A: 0.55    │
│                      │  N: 0.48    │
└──────────────────────┴──────────────┘
```

### Nuevas teclas

| Tecla | Acción |
|-------|--------|
| `e` | Inspeccionar NPC en tile bajo cursor |
| `TAB` | Ciclar NPC si hay múltiples en el tile |
| `i` | Abrir/cerrar panel inspector |

### Cambios en Model

```go
// Nuevos campos en Model
type Model struct {
    // ... existentes ...
    ecsWorld    *ecs.World       // referencia al ECS world
    cursorX     int              // cursor de selección (inicial = camera center)
    cursorY     int
    selectedNPC Entity           // NPC seleccionado actual
    npcVisible  map[Coord][]Entity // cache: NPCs visibles por tile
}
```

---

## 4. LOD Simulation

### Arquitectura de 3 Niveles

```
Local (radio < 10 tiles)
├── GOAP completo: planificar, ejecutar acciones
├── Economía detallada: producción/consumo
├── Eventos y diálogos
├── Check de colisiones con otros NPCs Local
└── Frecuencia: cada tick

Near (radio 10-30 tiles)
├── Wandering básico: move_to aleatorio
├── Economía simplificada: balance producción/consumo por tick
├── Sin colisiones detalladas
└── Frecuencia: cada 5 ticks

Distant (radio > 30 tiles)
├── No se simula IA
├── Solo existe: posición + estado congelado
├── Ocasional: chequeo de muerte por hambre (cada 50 ticks)
└── Frecuencia: cada 20 ticks
```

### Sistema LOD

```go
// internal/simulation/npc/lod.go

type LODSystem struct {
    playerPos func() (int, int) // callback a posición del jugador/cursor
}

func (s *LODSystem) Update(w *ecs.World, dt float64) error {
    px, py := s.playerPos()
    lodStore := w.GetStore(ecs.NewComponentID("lod")).(*ecs.ComponentStore[ecs.LOD])
    
    for e, lod := range lodStore.All() {
        pos, ok := ecs.GetComponent[ecs.Position](w, e)
        if !ok {
            continue
        }
        dist := manhattanDistance(int(pos.X), int(pos.Y), px, py)
        
        newLevel := classifyLOD(dist)
        if newLevel != lod.Level {
            lod.Level = newLevel
            lod.LastTick = currentTick
            lodStore.Set(e, lod)
        }
    }
    return nil
}

func classifyLOD(dist int) LODLevel {
    switch {
    case dist < 10:
        return LODLocal
    case dist < 30:
        return LODNear
    default:
        return LODDistant
    }
}
```

### Sistema de Wandering (MVP)

```go
// internal/simulation/npc/wander.go

type WanderSystem struct {
    wm *world.WorldMap
}

func (s *WanderSystem) Update(w *ecs.World, dt float64) error {
    for _, e := range w.Entities() {
        lod, ok := ecs.GetComponent[ecs.LOD](w, e)
        if !ok || lod.Level == ecs.LODDistant {
            continue
        }
        
        state, ok := ecs.GetComponent[ecs.AIState](w, e)
        if !ok {
            continue
        }
        
        if len(state.CurrentPlan) == 0 {
            state.CurrentPlan = s.planWander(w, e)
            ecs.AddComponent(w, e, state)
        }
        
        s.executePlan(w, e, dt)
    }
    return nil
}
```

### Métrica: NPCs por tick

| Nivel | Costo relativo | NPCs (100 total) | Tiempo estimado |
|-------|---------------|-------------------|-----------------|
| Local | 1.0x | ~10 | 10 unidades |
| Near | 0.3x | ~30 | 9 unidades |
| Distant | 0.02x | ~60 | 1.2 unidades |
| **Total** | | **100** | **~20.2 unidades** |

Esto significa que con LOD, **simular 100 NPCs cuesta ~20% de lo que costaría** si todos estuvieran en modo Local. La relación escala bien: 1000 NPCs → ~200 unidades (vs 1000 sin LOD).

---

## 5. Data-Driven

### Archivos a Crear

| Archivo | Propósito |
|---------|-----------|
| `data/npcs.yaml` | Definiciones de razas: traits, roles, name_pool, símbolo, color |
| `data/npc-roles.yaml` | Definiciones de roles: skill_base, biome_preference |
| `data/npc-names.yaml` | Pool de nombres (first/last) por raza (opcional, puede ir en npcs.yaml) |

### Validación Data-Driven

- Cada raza MUST tener al menos un role definido
- Cada role reference en raza MUST existir en npc-roles.yaml
- `spawn_weights` MUST sumar a un valor coherente
- `traits` MUST tener mean en [0,1] y std > 0

---

## 6. Persistencia

### Enfoque MVP: Seed-Based + Snapshot Opcional

**Opción A — Seed-based (recomendada para MVP):**
```
NPCs son deterministas desde:
  seed_mundo = genConfig.Seed
  seed_npcs  = seed_mundo + 999  // offset dedicado
```
Ventaja: No hay que serializar nada. Todos los NPCs se regeneran idénticos.
Desventaja: No se preserva estado dinámico (health cambiante, inventario).

**Opción B — Full state (post-MVP):**
```
Tabla SQLite:
  npc_entities:
    entity_id  INTEGER PRIMARY KEY
    world_id   INTEGER REFERENCES worlds(id)
    race       TEXT
    role       TEXT
    seed       INTEGER  // seed individual para traits
    
  npc_components:
    entity_id  INTEGER
    component_type TEXT    // "health", "position", etc.
    data       BLOB   // JSON o binary
```
Ventaja: Estado completo preservado.
Desventaja: Complejidad alta, migrations, sync entre ECS y DB.

**Recomendación híbrida para MVP:**
1. Seed-based para generación de NPCs (posición, traits, nombre, rol)
2. SQLite guarda solo `npc_count` y `npc_seed_offset` en la tabla `worlds`
3. Post-MVP: migrar a tabla `npc_entities` para preservar estado dinámico

### Schema SQLite extendido (post-MVP)

```sql
CREATE TABLE IF NOT EXISTS npc_entities (
    entity_id INTEGER PRIMARY KEY,
    world_id INTEGER NOT NULL,
    race_id TEXT NOT NULL,
    role_id TEXT NOT NULL,
    seed INTEGER NOT NULL,
    FOREIGN KEY (world_id) REFERENCES worlds(id)
);

CREATE TABLE IF NOT EXISTS npc_component_data (
    entity_id INTEGER NOT NULL,
    component_type TEXT NOT NULL,
    data BLOB NOT NULL,
    PRIMARY KEY (entity_id, component_type)
);
```

---

## 7. Alcance MVP

### ¿Cuántos NPCs?

- **MVP**: 50-100 NPCs para un mundo de 256×256 (densidad ~0.0015 NPCs/tile)
- **Post-MVP**: Escalable a 500-1000+ con LOD

### Componentes Esenciales para MVP

| Componente | Prioridad | Razón |
|-----------|-----------|-------|
| Health | 🔴 Crítico | Sin salud no hay riesgo ni muerte |
| Personality (baseline) | 🟡 Necesario | Traits definen comportamiento futuro; en MVP solo random baseline |
| Job | 🟡 Necesario | Rol determina wandering y visualización |
| AIState | 🟡 Necesario | Necesario para wandering aunque GOAP sea post-MVP |
| Appearance | 🟡 Necesario | Sin esto no se ven en el mapa |
| LOD | 🟡 Necesario | Sin LOD no escalamos a 100 NPCs |
| Inventory | ⚪ Futuro | Requiere sistema de items |
| Family | ⚪ Futuro | Requiere reproducción |

### Sistemas Base para MVP

| Sistema | Prioridad | Descripción |
|---------|-----------|-------------|
| NPCSpawnSystem | 🔴 Crítico | Generar NPCs en world creation |
| LODSystem | 🟡 Necesario | Clasificar NPCs por distancia |
| WanderSystem (básico) | 🟡 Necesario | Movimiento aleatorio contextual |
| NPCRenderSystem | 🟡 Necesario | Mostrar NPCs en TUI |

### Lo que NO incluye MVP

- ❌ GOAP real (solo wandering básico con AIState placeholder)
- ❌ Asentamientos / buildings / workplaces
- ❌ Economía (production/consumo)
- ❌ Interacción jugador→NPC (diálogos, combate, comercio)
- ❌ Familias y reproducción
- ❌ Inventarios
- ❌ Eventos dinámicos (enfermedades, desastres)

---

## 8. Análisis de Opciones

### Spawner

| Enfoque | Pros | Cons | Esfuerzo |
|---------|------|------|----------|
| **A. Random uniforme** | Simple, rápido de implementar | NPCs en océano, sin cohesión social | Bajo |
| **B. Biome-weighted** | Coherente con el mundo (granjeros en llanuras) | Sigue siendo aleatorio, sin asentamientos | Medio |
| **C. Settlement-based** | Realista, forma comunidades | Requiere sistema de asentamientos (post-MVP) | Alto |

**Recomendación: B para MVP**, migrar a C post-MVP.

### Persistencia

| Enfoque | Pros | Cons | Esfuerzo |
|---------|------|------|----------|
| **A. Seed-based** | Cero serialización, determinista | No preserva estado dinámico | Bajo |
| **B. Full state SQLite** | Estado completo preservado | Complejidad alta, sync ECS↔DB | Alto |
| **C. Híbrido** | Balance: seed para generación, snapshot para health | Dos sistemas de persistencia | Medio |

**Recomendación: A para MVP**, migrar a C post-MVP.

### Visualización

| Enfoque | Pros | Cons | Esfuerzo |
|---------|------|------|----------|
| **A. Overlay directo** | Simple, un símbolo reemplaza biome | Pierdes info del biome en tiles con NPCs | Bajo |
| **B. Overlay + color blend** | Ves biome + NPC (composición) | Más complejo de renderizar | Medio |
| **C. Panel lateral** | Rica información | Consume espacio en terminal | Medio |

**Recomendación: A + C para MVP**. Overlay de símbolo, con panel inspector al seleccionar.

---

## 9. Riesgos

| Riesgo | Impacto | Probabilidad | Mitigación |
|--------|---------|-------------|------------|
| Performance TUI con >100 NPCs | Alto | Media | LOD desde MVP, solo renderizar NPCs en viewport |
| ComponentStore crece sin bound | Medio | Baja | LOD limita entidades simuladas; seed-based evita acumulación |
| Wandering NPCs se salen del mapa | Bajo | Alta | WanderSystem verifica InBounds siempre |
| Spawner lento en mundo grande | Medio | Media | Paralelizar generación, muestreo aleatorio no iterativo |
| Types de componentes olvidados en RemoveEntity | Medio | Alta | Refactorizar RemoveEntity para usar reflection genérica (ya existe) |

---

## 10. Recomendación Final y Ready for Proposal

### Orden de Implementación Sugerido

| Fase | Tarea | Dependencias |
|------|-------|-------------|
| 0 | Crear `internal/simulation/npc/` y `data/npcs.yaml` | — |
| 1 | Implementar componentes NPC (Health, Personality, Job, AIState, Appearance, LOD) | Fase 0 |
| 2 | Implementar NPCSpawnSystem (biome-weighted) | Fase 1 |
| 3 | Implementar LODSystem | Fase 1 |
| 4 | Implementar WanderSystem (básico) | Fase 1 |
| 5 | Integrar render NPC en TUI (overlay + cursor + inspector) | Fase 1 |
| 6 | Tests de integración (spawn + wander + render + persist) | Fases 1-5 |

### Ready for Proposal

**Sí.** Esta exploración tiene toda la información necesaria para proceder con la propuesta (sdd-propose), especificaciones (sdd-spec) y diseño técnico (sdd-design).

Resumen para la propuesta:
- **MVP**: 50-100 NPCs, 6 componentes nuevos, 3 sistemas ECS (Spawn, LOD, Wander), overlay en TUI + inspector básico
- **Data-driven**: Razas y roles desde YAML con biome-weighted placement
- **Persistencia**: Seed-based (determinista), extensible a full state post-MVP
- **Sin GOAP real en MVP** — solo AIState placeholder para wandering
- **LOD desde el día 1** — 3 niveles para escalar sin problemas
