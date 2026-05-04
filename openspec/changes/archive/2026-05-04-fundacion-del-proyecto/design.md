# Design: Fundación del proyecto

## 1. Arquitectura General

### Diagrama de Paquetes

```
cmd/evociv/
  main.go                  ← Entry point: inicializa World, carga datos, lanza TUI
           │
           ▼
┌─────────────────────┐    ┌──────────────────────────┐
│   internal/ecs      │    │    internal/data          │
│                     │    │                          │
│  Entity (uint64)    │    │  Registry                 │
│  World              │───►│  Loader (YAML → Registry) │
│  ComponentStore[T]  │    │  Validator[T]             │
│  System interface   │    │                          │
│  SystemManager      │    └──────────────────────────┘
└─────────────────────┘              │
         │                           │ data/biomes.yaml
         │                           ▼
         │                 ┌──────────────────────────┐
         │                 │    internal/store          │
         │                 │                          │
         └────────────────►│  Store interface          │
                           │  SQLiteStore (impl)       │
                           └──────────────────────────┘
                                      │
                                      ▼
                           ┌──────────────────────────┐
                           │    internal/ui             │
                           │                          │
                           │  Model (Bubbletea)        │
                           │  View (lipgloss)          │
                           │  Update (eventos tecla)   │
                           └──────────────────────────┘
```

### Flujo de Datos entre Paquetes

```
arranque:
  main.go
    ├── data.Loader{}.Load("data/") → registry poblado
    ├── ecs.NewWorld() → world vacío
    ├── store.NewSQLiteStore("evociv.db").Open() → db conectado
    └── ui.NewModel(registry, world, store) → modelo TUI

ciclo TUI:
  tea.NewProgram(model).Run()
    ├── Update(msg):
    │     ├── tea.KeyMsg{'}q'}' → tea.Quit
    │     └── otro → no-op
    └── View():
          └── lipgloss.JoinVertical(
                título, versión, instrucciones
              )
```

### Ciclo de Vida de la Aplicación

```
1. INIT    → main(): parse flags, crear dependencias
2. BOOT    → data.Loader.Load(): carga YAML a Registry
3. CONNECT → store.Open(): abre SQLite
4. RUN     → tea.NewProgram(model).Run(): bloquea hasta quit
5. SHUTDOWN→ store.Close(): cierra SQLite
```

---

## 2. Diseño Detallado por Capability

### ecs-core

**Tipos clave:**

```go
// Entity: identificador único numérico
type Entity uint64

// ComponentID: identificador de tipo de componente (derivado de reflect)
type ComponentID string

// World: contenedor central de entidades y componentes
type World struct {
    nextID     Entity
    components map[ComponentID]map[Entity]any
    systems    []System
    mu         sync.RWMutex   // safety concurrente para reads
}

func NewWorld() *World

func (w *World) NewEntity() Entity
func (w *World) AddComponent(e Entity, id ComponentID, val any)
func (w *World) GetComponent(e Entity, id ComponentID) (any, bool)
func (w *World) RegisterSystem(s System)
func (w *World) Update(ctx context.Context) error

// ComponentStore[T]: wrapper genérico con type safety
type ComponentStore[T any] struct {
    world *World
    id    ComponentID
}

func NewComponentStore[T any](w *World) ComponentStore[T]
func (cs ComponentStore[T]) Set(e Entity, val T)
func (cs ComponentStore[T]) Get(e Entity) T

// System: interfaz que opera sobre el World
type System interface {
    Name() string
    Update(ctx context.Context, w *World) error
}
```

**Ejemplo de API pública:**

```go
type Position struct { X, Y float64 }
type Velocity struct { DX, DY float64 }

world := ecs.NewWorld()
posStore := ecs.NewComponentStore[Position](world)
velStore := ecs.NewComponentStore[Velocity](world)

e := world.NewEntity()
posStore.Set(e, Position{X: 10, Y: 20})
velStore.Set(e, Velocity{DX: 1, DY: 0})

world.RegisterSystem(&MovementSystem{})
world.Update(ctx)  // ejecuta MovementSystem.Update()

// MovementSystem implementa ecs.System
func (s *MovementSystem) Update(ctx context.Context, w *ecs.World) error {
    posStore := ecs.NewComponentStore[Position](w)
    velStore := ecs.NewComponentStore[Velocity](w)
    // iteración manual sobre entidades (sin Query — out of scope)
    return nil
}
```

**Estrategia de tests:** Table-driven tests con `t.Run()`. Test para creación de entity (non-zero ID), Set/Get de componente, Get de componente ausente (zero value), System que muta estado. `go test -race` para verificar concurrent safety.

### data-loader

**Tipos clave:**

```go
// Registry: almacén tipado de datos cargados
type Registry struct {
    data map[string]any
    mu   sync.RWMutex
}

func NewRegistry() *Registry
func (r *Registry) Register(key string, val any)
func (r *Registry) Get(key string) any    // retorna nil si no existe
func (r *Registry) GetAs[T](key string) (T, error)  // type-assert segura

// Loader: recorre directorio y parsea YAML
type Loader struct {
    dir      string
    registry *Registry
}

func NewLoader(dir string, reg *Registry) *Loader
func (l *Loader) Load() error

// Validator: hook opcional post-parseo
type Validator[T any] func(*T) error
type LoadOption[T any] struct {
    Key       string
    Target    *T          // puntero donde deserializar
    Validate  Validator[T] // opcional
}
```

**YAML schema de ejemplo (`data/biomes.yaml`):**

```yaml
biomes:
  - name: grasslands
    description: "Vast plains with tall grass and scattered trees."
    fertility: 0.8
    movement_cost: 1.0
  - name: desert
    description: "Arid expanse of sand and rock."
    fertility: 0.1
    movement_cost: 1.5
  - name: forest
    description: "Dense woodland with rich soil."
    fertility: 0.9
    movement_cost: 2.0
```

**Estrategia de tests:** Directorios temporales con `t.TempDir()`, escribir YAML en ellos. Test: archivo válido → datos correctos; archivo malformado → error; directorio vacío → éxito sin datos; directorio faltante → error; validador rechaza datos inválidos.

### tui-welcome

**Modelo Bubbletea:**

```go
type Model struct {
    quitting bool
    width    int
    height   int
}

func NewModel() Model
func (m Model) Init() tea.Cmd                       { return nil }
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m Model) View() string
```

**Manejo de teclas:**

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width, m.height = msg.Width, msg.Height
        return m, nil
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            m.quitting = true
            return m, tea.Quit
        }
    }
    return m, nil
}
```

**Estilos lipgloss:**

```go
var (
    titleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("#7C3AED")).
        Margin(1, 0, 0, 0)

    versionStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#6B7280"))

    instructionsStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#9CA3AF")).
        Margin(1, 0, 0, 0)
)

func (m Model) View() string {
    title := titleStyle.Render("Evociv-RL")
    version := versionStyle.Render("v0.1.0")
    instructions := instructionsStyle.Render(
        "Press 'q' to quit.",
    )
    return lipgloss.JoinVertical(
        lipgloss.Center,
        title, "", version, "", instructions,
    )
}
```

**Ciclo del programa (main.go):**

```go
func main() {
    model := ui.NewModel()
    p := tea.NewProgram(model, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        log.Fatal(err)
    }
}
```

**Estrategia de tests:**
- Unit: Update('q') → tea.Quit; Update('a') → no-op; View() no panic.
- Golden test: llamar View(), comparar output con `testdata/welcome.golden`. Usar `flag.Update()` para regenerar con `-update`.
- Integration: `teatest.NewTestModel(t, m)`, enviar 'q', verificar que termina.
- Normalización: `*.golden text eol=lf` en `.gitattributes`.

### store-sqlite

**Store interface:**

```go
type Store interface {
    Open() error
    Close() error
    Ping() error
}
```

**SQLiteStore implementación:**

```go
type SQLiteStore struct {
    db  *sql.DB
    dsn string
}

func NewSQLiteStore(path string) *SQLiteStore {
    return &SQLiteStore{dsn: path}
}

func (s *SQLiteStore) Open() error {
    db, err := sql.Open("sqlite", s.dsn)
    if err != nil {
        return fmt.Errorf("sqlite open: %w", err)
    }
    // modernc.org/sqlite solo soporta 1 conexión de escritura
    db.SetMaxOpenConns(1)
    s.db = db
    return s.Ping()
}

func (s *SQLiteStore) Ping() error {
    return s.db.Ping()
}

func (s *SQLiteStore) Close() error {
    return s.db.Close()
}
```

**Connection pool:** 1 conexión máxima (embedded SQLite no se beneficia de múltiples writers). `SetMaxOpenConns(1)`.

**Estrategia de tests:**
- `t.TempDir()` → `filepath.Join(dir, "test.db")`.
- Happy path: Open() ok, Ping() ok, Close() ok.
- Double Open: llamar Open() dos veces → error (o no-op documentado).
- Invalid path: Open() con path en directorio no existente → error.
- Verificar compile-time: `var _ Store = (*SQLiteStore)(nil)`.

---

## 3. Estructura de Archivos Final

```
evociv-rl/
├── .gitattributes              [MODIFY] añadir *.golden text eol=lf
├── go.mod                      [EXISTE] ya con module github.com/marco/evociv-rl
├── go.sum                      [NEW]
├── cmd/
│   └── evociv/
│       └── main.go             [NEW] entry point (inicializa, lanza TUI)
├── internal/
│   ├── ecs/
│   │   ├── entity.go           [NEW] type Entity uint64
│   │   ├── component.go        [NEW] ComponentID, ComponentStore[T]
│   │   ├── world.go            [NEW] World struct, NewEntity, AddComponent, etc.
│   │   ├── system.go           [NEW] System interface, SystemManager
│   │   └── ecs_test.go         [NEW] tests: entity, component, world, system
│   ├── data/
│   │   ├── loader.go           [NEW] Loader: recorre dir, parsea YAML
│   │   ├── registry.go         [NEW] Registry: mapa tipado
│   │   └── data_test.go        [NEW] tests: carga, errores, validación
│   ├── store/
│   │   ├── store.go            [NEW] Store interface
│   │   ├── sqlite.go           [NEW] SQLiteStore impl
│   │   └── store_test.go       [NEW] tests: open/close/ping
│   └── ui/
│       ├── model.go            [NEW] Model Bubbletea, Init, Update
│       ├── view.go             [NEW] View con lipgloss
│       ├── model_test.go       [NEW] tests TUI + golden
│       └── testdata/
│           └── welcome.golden  [NEW] golden file referencia
├── data/
│   └── biomes.yaml             [NEW] 2-3 biomas de prueba
└── openspec/
    └── changes/
        └── fundacion-del-proyecto/
            ├── proposal.md     [EXISTE]
            ├── specs/          [EXISTE]
            └── design.md       [ESTE ARCHIVO]
```

---

## 4. Decisiones Técnicas (ADR)

### ADR-1: ECS Map-Based vs Arquetipos

| Opción | Ventajas | Desventajas |
|--------|----------|-------------|
| Map-based (`map[ComponentID]map[Entity]any`) | Simple, flexible, sin fragmentación, fácil debug | Iteración O(n) sobre entradas, caché miss |
| Archetype ECS (column stores) | Cache-friendly, iteración O(n) sobre arquetipos | Arquitectura rígida, complejidad alta, overkill prematuro |

**Decisión:** Map-based. <10K entidades esperadas en early game. Migrar a arquetipos solo si profiling lo justifica.

### ADR-2: modernc.org/sqlite vs mattn/go-sqlite3

| Opción | Ventajas | Desventajas |
|--------|----------|-------------|
| modernc.org/sqlite | CGO-free, cross-compile fácil, sin toolchain C | ~30% más lento que CGO |
| mattn/go-sqlite3 | Rendimiento nativo | Requiere GCC/MSVC, build frágil en Windows |

**Decisión:** modernc.org/sqlite. Propuesta ya lo especifica. No CGO elimina todo el dolor de toolchain en Windows.

### ADR-3: ComponentStore[T] Genérico + Almacenamiento Interno Untyped

**Contexto:** La API pública debe ser type-safe, pero la máquina ECS interna necesita almacenar tipos arbitrarios.

**Decisión:** `ComponentStore[T any]` expone `Get/Set` con tipos Go concretos. Internamente, `World` usa `map[ComponentID]map[Entity]any`. El cast `any → T` ocurre en la frontera del wrapper genérico, no en el core.

**Consecuencias:** + Type safety en API, + Simplicidad interna. Costo: un type assertion por Get. Aceptable para el dominio.

### ADR-4: Reflect-based ComponentID vs Enteros Registrados

**Contexto:** Necesitamos un identificador único por tipo de componente para el mapa interno.

| Opción | Ventajas | Desventajas |
|--------|----------|-------------|
| `reflect.TypeOf` como key | Zero boilerplate, no registro manual | Dependencia de reflect (mínima) |
| Enteros registrados manualmente | Performance óptimo | Boilerplate: enum + registro; frágil |

**Decisión:** `fmt.Sprintf("%T", zero)` o `reflect.TypeOf`. La creación de componentes no está en hot path de simulación. Se puede optimizar a enteros si profiling lo requiere.

### ADR-5: Golden Files para Tests TUI

**Contexto:** La salida de Bubbletea/lipgloss contiene secuencias ANSI. Verificar manualmente es frágil.

**Decisión:** Golden file `testdata/welcome.golden` con la salida exacta. Tests comparan View() contra golden. Regenerar con `go test -update`.

**Consecuencias:** + Detección inmediata de cambios visuales. + CI puede fallar si golden cambia sin actualizar. - Hay que actualizar golden cuando el diseño cambia.

### ADR-6: Sin Testify en Fundación

**Contexto:** La propuesta ya especifica dependencias mínimas.

**Decisión:** Usar solo `testing` estándar de Go + `teatest` de Charmbracelet (necesario para integración TUI). `testify/assert` no se incluye — las assertions se hacen con `if` manuales o helpers.

**Consecuencias:** + Una dependencia menos. - Tests ligeramente más verbosos.

### ADR-7: Store Interface Mínima (Open/Close/Ping)

**Contexto:** En fundación solo necesitamos probar que SQLite abre y cierra. Store no se usa en el gameplay aún.

**Decisión:** Interface de 3 métodos: `Open()`, `Close()`, `Ping()`. Suficiente para verificar conectividad. Se extenderá (Query/Exec) cuando se implemente persistencia real.

**Consecuencias:** + Fácil de mockear. + Migración trivial a Postgres/otro backend. - Métodos adicionales vendrán después.

---

## 5. Estrategia de Testing

### Unit Tests por Paquete

| Paquete | Tests | Cobertura esperada |
|---------|-------|--------------------|
| `ecs` | Entity non-zero ID, ComponentStore Set/Get/zero, World.Update ejecuta sistemas, System muta estado, concurrent safety | >90% |
| `data` | Load YAML válido, Load dir faltante, Load YAML malformado, Load dir vacío, validador rechaza, Registry Get/GetAs | >90% |
| `store` | Open/Close/Ping, double Open error, invalid path error, compile-time interface check | >85% |
| `ui` | Update('q') → tea.Quit, Update(otra tecla) → no-op, View() no panic, golden match | >85% |

### Integration Tests

- `ui`: `teatest.NewTestModel` — inicia modelo, envía 'q', verifica terminación limpia.
- `data` + `ecs`: end-to-end loading → registry → world (futuro, cuando se integren).

### Golden File Testing

```go
func TestView_Golden(t *testing.T) {
    m := NewModel()
    got := m.View()
    if *update {
        os.WriteFile("testdata/welcome.golden", []byte(got), 0644)
    }
    want, err := os.ReadFile("testdata/welcome.golden")
    if err != nil {
        t.Fatal(err)
    }
    if got != string(want) {
        t.Errorf("View() mismatch (-want +got):\n%s", difftext(got, string(want)))
    }
}
```

### Mocking Strategy

- **Store:** Interface `Store` → mockeable en futuros tests de integración.
- **Loader:** Acepta `io.FS` (Go 1.26 `testing/fstest`) para inyectar filesystem en tests.
- **ECS:** Ningún mock necesario — las dependencias son tipos concretos del mismo paquete.

### Comandos

```bash
go test ./...          # todos los tests
go test -cover ./...   # cobertura
go test -race ./...    # data races
go vet ./...           # lint estático
go test -v -update     # regenerar golden files
```
