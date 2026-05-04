# Exploration: Fundación del Proyecto — Evociv-RL

## Current State

El proyecto está en estado completamente vacío. Existe:
- `go.mod` con módulo `github.com/marco/evociv-rl` y Go 1.26.2
- `openspec/config.yaml` con SDD configurado y Strict TDD activado
- `README.md` minimal
- `.atl/skill-registry.md` con el registro de skills

No hay código Go, no hay tests, no hay directorios de paquete. Es un proyecto desde cero.

## Affected Areas

- **Raíz del proyecto**: estructura completa de directorios, go.mod con dependencias
- **`cmd/evociv/`**: entry point principal (main.go)
- **`internal/ecs/`**: núcleo del Entity Component System
- **`internal/world/`**: generación y simulación del mundo
- **`internal/data/`**: data-driven engine (carga YAML, registry, validación)
- **`internal/store/`**: persistencia SQLite
- **`internal/ui/`**: TUI con Bubbletea
- **`data/`**: archivos YAML de contenido
- **`openspec/`**: specs, design, tasks para el cambio

---

## 1. Estructura de Directorios Propuesta

```
evociv-rl/
├── cmd/
│   └── evociv/
│       └── main.go              # Entry point: inicializa DI, lanza UI
├── internal/
│   ├── ecs/                     # Entity Component System
│   │   ├── entity.go            # Entity ID type, EntityRef
│   │   ├── world.go             # ECS World: registro de componentes y sistemas
│   │   ├── component.go         # Component interface + ComponentStore
│   │   ├── system.go            # System interface
│   │   ├── world_test.go
│   │   └── examples_test.go     # Tests de ejemplo/tutorial del ECS
│   ├── world/                   # Generación y simulación del mundo
│   │   ├── gen/                 # Generación procedural
│   │   │   ├── continent.go
│   │   │   ├── biome.go
│   │   │   └── river.go
│   │   ├── sim/                 # Simulación (LOD)
│   │   │   ├── lod.go
│   │   │   └── tick.go
│   │   └── types.go             # Tipos compartidos del mundo
│   ├── data/                    # Data-Driven Engine
│   │   ├── loader.go            # Carga y parseo de YAML
│   │   ├── registry.go          # Registry de tipos cargados
│   │   ├── validator.go         # Validación de datos cargados
│   │   ├── loader_test.go
│   │   └── testdata/            # YAML de prueba
│   ├── simulation/              # GOAP + RL
│   │   ├── goap/                # Goal-Oriented Action Planning
│   │   │   ├── planner.go
│   │   │   ├── action.go
│   │   │   └── goal.go
│   │   ├── rl/                  # Q-Learning
│   │   │   ├── qtable.go
│   │   │   └── learner.go
│   │   ├── hybrid.go            # Orquestador GOAP+RL
│   │   └── economy/             # Simulación económica
│   ├── llm/                     # Bridge con Ollama
│   │   ├── client.go            # Cliente HTTP para API de Ollama
│   │   ├── prompt.go            # Templates de prompts
│   │   └── client_test.go
│   ├── store/                   # Persistencia SQLite
│   │   ├── store.go             # Interfaz Store
│   │   ├── sqlite.go            # Implementación con modernc.org/sqlite
│   │   └── store_test.go
│   └── ui/                      # Bubbletea TUI
│       ├── model.go             # Modelo principal de la TUI
│       ├── update.go            # Update handler principal
│       ├── view.go              # Vista principal
│       ├── components/          # Componentes TUI reutilizables
│       │   ├── worldview/       # Vista del mundo (ASCII/tiles)
│       │   └── inspector/       # Inspector de entidades
│       ├── model_test.go
│       ├── update_test.go
│       └── testdata/
├── data/                        # YAML data-driven
│   ├── biomes.yaml
│   ├── items.yaml
│   ├── recipes.yaml
│   ├── actions.yaml
│   └── traits.yaml
├── go.mod
├── go.sum
├── openspec/
│   ├── config.yaml
│   ├── specs/                   # Especificaciones activas
│   └── changes/                 # Cambios en progreso
└── README.md
```

---

## 2. Dependencias Go con Versiones

### Core Runtime

| Paquete | Versión | Propósito | CGO-Free |
|---------|---------|-----------|----------|
| `github.com/charmbracelet/bubbletea` | `v1.3.10` | Framework TUI (Model-View-Update) | ✅ |
| `github.com/charmbracelet/lipgloss` | `v1.1.0` | Estilos y colores para la TUI | ✅ |
| `github.com/charmbracelet/bubbles` | `v1.0.0` | Componentes TUI (table, list, viewport, textinput, spinner, help, progress, etc.) | ✅ |
| `modernc.org/sqlite` | `v1.50.0` | SQLite puro Go (sin CGO vía modernc.org/libc) | ✅ |
| `gopkg.in/yaml.v3` | `v3.0.1` | Parseo de archivos YAML data-driven | ✅ |
| `github.com/charmbracelet/log` | `v1.0.0` | Logging estructurado con soporte de niveles y colores | ✅ |

### Testing

| Paquete | Versión | Propósito |
|---------|---------|-----------|
| `github.com/charmbracelet/x/exp/teatest` | `latest` (pseudo-version) | Testing de modelos Bubbletea (NewTestModel, Send, FinalModel, WaitFinished) |
| `github.com/charmbracelet/x/exp/golden` | (incluido en x) | Golden file testing para output TUI |
| `github.com/stretchr/testify` | `v1.11.1` | Aserciones (assert.Equal, require.NoError) — **opcional**, podemos usar solo testing estándar |

### Nota sobre testify
Recomendación: **no incluir testify en la fundación**. Go 1.26.2 tiene `testing` package completo, y el proyecto sigue Strict TDD. Podemos añadirlo más adelante si la necesidad de aserciones más ricas lo justifica. La fundación debe ser lo más lightweight posible.

### Dependencias indirectas clave
- `modernc.org/libc` v1.72.0 — reimplementación pura Go de libc (permite SQLite sin CGO)
- `modernc.org/memory` v1.11.0 — gestión de memoria para libc puro Go
- `github.com/charmbracelet/x/ansi`, `x/term`, `x/cellbuf` — subpackages de x/ para Bubbletea
- `github.com/muesli/termenv` v0.16.0 — renderizado de terminal
- `github.com/mattn/go-runewidth` — ancho de caracteres Unicode

### Por qué NO necesitamos
- **`golang.org/x/exp`**: Go 1.26.2 tiene `slices`, `maps`, `cmp` en stdlib
- **`golang.org/x/sys`**: vendrá como dependencia indirecta de modernc.org/sqlite, no la añadimos directo
- **`github.com/mattn/go-sqlite3`**: requiere CGO. `modernc.org/sqlite` es la alternativa CGO-free
- **Frameworks de test**: solo `testing` estándar + `teatest` para TUI

---

## 3. Arquitectura ECS Propuesta

### Enfoque: Map-Based Simple ECS

Para la fundación, proponemos un ECS **simple basado en mapas** (no arquetipos). Razones:

1. **El proyecto necesita existir primero**. La optimización con arquetipos (tipo Unity DOTS) se puede hacer después.
2. **Go 1.26.2 no es C++ o Rust** — no tenemos cero-cost abstractions para arquetipos sin reflection.
3. **Simplicidad > rendimiento prematuro**. Con miles de entidades (no cientos de miles), mapas son más que suficientes.
4. **Evolución clara**: migrar a arquetipos más adelante manteniendo la misma interfaz `World`.

### Diseño

```go
// internal/ecs/entity.go
type Entity uint64

// internal/ecs/component.go
type Component interface {
    ComponentID() ComponentID
}

type ComponentID uint64

// ComponentStore almacena componentes de un tipo T
type ComponentStore[T any] struct {
    components map[Entity]T
}

func (s *ComponentStore[T]) Get(e Entity) (T, bool) { ... }
func (s *ComponentStore[T]) Set(e Entity, c T) { ... }
func (s *ComponentStore[T]) Delete(e Entity) { ... }

// internal/ecs/world.go
type World struct {
    entities    []Entity        // entidades activas
    nextID      Entity
    stores      map[ComponentID]any  // type-erased stores
}

func (w *World) NewEntity() Entity { ... }
func (w *World) RemoveEntity(e Entity) { ... }
func (w *World) Store(id ComponentID) (store any, ok bool) { ... }
```

### Componentes Base para la Fundación

```go
// Componentes base — package internal/ecs/components/
type Position struct {
    X, Y float64
    Z    int  // altura / nivel
}

type Name struct {
    Name string
}

type Tags struct {
    Tags []string
}
```

Mirando al futuro cercano (próximos cambios), estos componentes crecerán a:
```go
// (No implementar en fundación — solo diseño)
type Health struct { Current, Max int }
type Inventory struct { Items []ItemStack }
type AIComponent struct { 
    Goals   []Goal
    QTable  QTable
    CurrentAction string
}
type Renderable struct { 
    Char rune
    Style lipgloss.Style
}
type LOD struct {
    Level     LODLevel      // Distant, Near, Local
    LastSeen  time.Duration
}
```

### Sistemas

```go
// internal/ecs/system.go
type System interface {
    Update(w *World, dt float64) error
    Name() string
}

// Ejemplo: PositionSystem
type PositionSystem struct{}
func (s *PositionSystem) Name() string { return "position" }
func (s *PositionSystem) Update(w *World, dt float64) error {
    store := w.Store(PositionComponentID)
    // iteramos, actualizamos...
    return nil
}
```

### Integración GOAP + ECS

El GOAP planner será un **System** que opera sobre entidades con componentes `AIComponent` y `Position`:

```go
type GOAPSystem struct {
    planner *goap.Planner
}

func (s *GOAPSystem) Update(w *World, dt float64) error {
    store := w.Store(AIComponentID)
    // Para cada entidad con AI, ejecuta planificación GOAP
    // Si no hay plan, delega a RL para exploración
    return nil
}
```

### Registro de Componentes

Los ComponentID se definen como constantes para type safety en tiempo de compilación:

```go
const (
    PositionComponentID ComponentID = iota + 1
    NameComponentID
    TagsComponentID
    HealthComponentID
    InventoryComponentID
    AIComponentID
    RenderableComponentID
    LODComponentID
)
```

Cada tipo de componente se registra una vez al iniciar el World:

```go
world := ecs.NewWorld()
world.RegisterComponent(PositionComponentID, ecs.NewComponentStore[Position]())
world.RegisterComponent(NameComponentID, ecs.NewComponentStore[Name]())
```

---

## 4. Alcance del Primer Entregable (Fundación)

### Propuesta: Demo Mínima Funcional (no solo estructura)

El primer cambio **debe demostrar que toda la pila funciona**, no solo crear carpetas vacías. Esto es crítico para validar la arquitectura antes de escalar.

### MVP de la Fundación

```
✅ = incluido en fundación
🔲 = próximo cambio
```

#### Infraestructura
- ✅ Estructura de directorios completa
- ✅ go.mod con todas las dependencias resueltas
- ✅ `cmd/evociv/main.go` — entry point que compila y produce binario
- ✅ `Makefile` o `Taskfile.yaml` con comandos: build, test, run

#### ECS
- ✅ `Entity` type, `World` con registro de componentes
- ✅ `ComponentStore[T]` genérico con Get/Set/Delete
- ✅ `System` interface y ciclo de update
- ✅ Al menos un componente (`Position`) y un sistema (`PositionSystem`)
- ✅ Tests: crear entity, asignar componente, leer componente, ejecutar sistema

#### Data-Driven
- ✅ `data/loader.go` — carga YAML desde directorio
- ✅ `data/registry.go` — registry básico (tipo -> []datos)
- ✅ Un archivo YAML de prueba (`data/biomes.yaml` con 2-3 biomas)
- ✅ Tests: cargar YAML, verificar contenido, error en YAML malformado

#### TUI
- ✅ Modelo Bubbletea básico con "Pantalla de bienvenida"
- ✅ View que muestra "Evociv-RL v0.0.1 — Press q to quit"
- ✅ Manejo de tecla 'q' para salir
- ✅ Lipgloss styling básico (título en negrita, color)
- ✅ Tests: model.Update con tecla 'q' → exit, model.View() golden test

#### Persistencia
- ✅ `store.Store` interface con métodos Open/Close
- ✅ Implementación SQLite no-op (open, exec CREATE TABLE, close)
- ✅ Tests: open/close sin errores

#### Build & Verify
- ✅ `go build ./cmd/evociv` produce binario sin errores
- ✅ `go test ./...` pasa todos los tests (Strict TDD)
- ✅ `go vet ./...` sin warnings

### Lo que NO incluye la fundación
- ❌ Generación procedural de mundo (próximo cambio)
- ❌ GOAP o RL real (próximo cambio)
- ❌ Diálogos con LLM (próximo cambio)
- ❌ Múltiples pantallas TUI (próximo cambio)
- ❌ Persistencia real de datos (próximo cambio)
- ❌ Component Viewer/Inspector (próximo cambio)

---

## 5. Riesgos Identificados

### Riesgo 1: Bubbletea 1.x API breaking changes
- **Impacto**: Medio. Bubbletea pasó de v0.x a v1.x con cambios en la API. `tea.NewProgram` ahora usa opciones funcionales. `tea.KeyMsg` cambió. `teatest` se movió a `x/exp/teatest`.
- **Mitigación**: Verificar la API exacta de v1.3.10 durante la implementación. Los tests de la fundación validarán que todo funciona.
- **Concreto**: `tea.NewProgram(m, tea.WithInput(in), tea.WithOutput(out), tea.WithoutSignals())`

### Riesgo 2: modernc.org/sqlite compilación en Windows
- **Impacto**: Bajo-Medio. `modernc.org/sqlite` es puro Go, pero `modernc.org/libc` necesita implementaciones platform-specific. Windows amd64 está soportado (hay tags en las release).
- **Mitigación**: Incluir `modernc.org/libc` como dependencia explícita si es necesario. Verificar que `go build` funciona en Windows.
- **Verificación**: El test de store en la fundación validará que sqlite compila y abre/cierra sin errores.

### Riesgo 3: Path de Go en Windows
- **Impacto**: Alto (ya identificado). `go` no está en PATH del sistema.
- **Mitigación**: Documentar en README. Añadir script `scripts/setup.ps1` que configure PATH. O configurar PATH del sistema permanentemente.
- **Solución permanente**: Añadir `C:\Program Files\Go\bin` al PATH del sistema (vs. sesión por sesión).

### Riesgo 4: go.sum y versiones de dependencias transitivas
- **Impacto**: Bajo. Go modules maneja versionado transitivo automáticamente.
- **Mitigación**: Ejecutar `go mod tidy` al final para limpiar. Commitear `go.sum`.

### Riesgo 5: testdata/ golden files en Windows (line endings)
- **Impacto**: Bajo. Los golden files pueden tener diferencias CRLF vs LF.
- **Mitigación**: Usar `.gitattributes` con `* text=auto` y `*.golden text eol=lf`. Los tests TUI deben normalizar output antes de comparar.

---

## 6. Recomendaciones para la Propuesta

### Orden de Implementación Sugerido

1. **Fase 0: Infraestructura** (estructura, go.mod, go.sum, main.go compilable)
2. **Fase 1: ECS** (core entity/component/system + tests)
3. **Fase 2: Data-Driven** (loader YAML + registry + tests + YAML de prueba)
4. **Fase 3: TUI** (modelo Bubbletea + view + teclas básicas + tests)
5. **Fase 4: Store** (interface + SQLite stub + tests)
6. **Fase 5: Integración** (build, vet, todos los tests pasan)

### Decisiones de Diseño Clave

1. **ECS simple (map-based) > arquetipos** — rendimiento prematuro es la raíz de toda la maldad
2. **Interfaces > concreciones** — `Store` interface permite cambiar SQLite por mock en tests
3. **Data/ separado de internal/** — los YAML no son código, no necesitan compilarse
4. **Strict TDD desde el día 1** — test antes de implementar, siempre
5. **Sin testify en fundación** — menos dependencias, testing estándar es suficiente
6. **teatest para TUI integration tests** — mandatory para Bubbletea

### Ready for Proposal

**Sí.** Esta exploración tiene toda la información necesaria para proceder con la propuesta (sdd-propose), especificaciones (sdd-spec) y diseño técnico (sdd-design).

---

## Anexo: Dependencias Específicas para go.mod

```
require (
    github.com/charmbracelet/bubbletea v1.3.10
    github.com/charmbracelet/lipgloss v1.1.0
    github.com/charmbracelet/bubbles v1.0.0
    github.com/charmbracelet/log v1.0.0
    github.com/charmbracelet/x v0.1.0         // para exp/teatest y exp/golden
    modernc.org/sqlite v1.50.0
    gopkg.in/yaml.v3 v3.0.1
)
```

Nota: `github.com/charmbracelet/x v0.1.0` se añade como dependencia para poder usar `x/exp/teatest` y `x/exp/golden` en tests. No se usa en runtime.

## Anexo: Ejemplo de main.go para Fundación

```go
package main

import (
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

var titleStyle = lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("#00ff00")).
    Padding(1, 2)

type model struct {
    ready bool
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        }
    case tea.WindowSizeMsg:
        m.ready = true
    }
    return m, nil
}

func (m model) View() string {
    if !m.ready {
        return "Initializing..."
    }
    return titleStyle.Render("Evociv-RL v0.0.1") +
        "\n\nPress 'q' to quit."
}

func main() {
    p := tea.NewProgram(model{}, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}
```
