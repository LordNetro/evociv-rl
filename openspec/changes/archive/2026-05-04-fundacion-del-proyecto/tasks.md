# Tasks: Fundación del proyecto

## Capability: infraestructura

### [ ] 0.1 Crear estructura de directorios y configurar .gitattributes
  - Crear directorios: `cmd/evociv/`, `internal/ecs/`, `internal/data/`, `internal/store/`, `internal/ui/`, `internal/ui/testdata/`, `data/`
  - Modificar `.gitattributes`: añadir `*.golden text eol=lf` debajo de `* text=auto` para normalizar golden files en CI/CD
  - Tests: N/A — verificar con `ls -R` que los directorios existen y `.gitattributes` contiene la linea golden

### [ ] 0.2 Agregar dependencias externas y generar go.sum
  - Ejecutar `go get` para cada dependencia del proyecto:
    - `github.com/charmbracelet/bubbletea@v1.3.10`
    - `github.com/charmbracelet/lipgloss@v1.1.0`
    - `github.com/charmbracelet/bubbles@v1.0.0`
    - `github.com/charmbracelet/log@v1.0.0`
    - `modernc.org/sqlite@v1.50.0`
    - `gopkg.in/yaml.v3@v3.0.1`
    - `github.com/charmbracelet/x` (teatest + golden — solo tests)
  - Verificar que `go.sum` se genera automáticamente
  - Tests: N/A — `go.sum` debe existir y contener entradas de verificación

### [ ] 0.3 Crear cmd/evociv/main.go mínimo compilable
  - `cmd/evociv/main.go`: `package main`, `import "fmt"`, `func main() { fmt.Println("Evociv-RL v0.1.0") }`
  - Propósito: smoke test de que la estructura de directorios y dependencias compilan
  - Tests: `go build ./cmd/evociv/` produce binario sin errores

## Capability: ecs-core

### [ ] 1.1 Implementar Entity, ComponentID y ComponentStore[T]
  - `internal/ecs/entity.go`: `type Entity uint64` — identificador único numérico, zero value inválido
  - `internal/ecs/component.go`:
    - `ComponentID` = `string` derivado de `fmt.Sprintf("%T", zero)` para identificación única por tipo Go
    - `ComponentStore[T any]` struct con referencia a `*World` y `ComponentID`
    - Métodos: `NewComponentStore[T any](w *World) ComponentStore[T]`, `Set(e Entity, val T)`, `Get(e Entity) T` (retorna zero value si no existe)
    - Internamente `Set` llama a `w.AddComponent(e, id, val)` y `Get` llama a `w.GetComponent(e, id)` con type assertion
  - Tests: Entity con ID != 0; ComponentID único para tipos distintos; ComponentStore.Set + Get; Get devuelve zero value para entidad sin componente
  - Verificar: `go vet ./internal/ecs/` sin warnings

### [ ] 1.2 Implementar World con manejo de entidades y componentes
  - `internal/ecs/world.go`: `World` struct con:
    - `nextID Entity` — contador incremental para IDs únicos
    - `components map[ComponentID]map[Entity]any` — almacenamiento map-based
    - `systems []System` — sistemas registrados
    - `mu sync.RWMutex` — safety concurrente para lecturas
  - Métodos: `NewWorld() *World`, `NewEntity() Entity` (retorna nextID y avanza), `AddComponent(e Entity, id ComponentID, val any)`, `GetComponent(e Entity, id ComponentID) (any, bool)`
  - Tests: NewEntity produce IDs únicos secuenciales; AddComponent + GetComponent retrieves valor; GetComponent devuelve false para entidad sin componente; ids no-zero desde el primer NewEntity
  - Verificar: `go test -run TestEntity ./internal/ecs/` pasa

### [ ] 1.3 Implementar System interface, SystemManager y World.Update
  - `internal/ecs/system.go`:
    - `System` interface: `Name() string`, `Update(ctx context.Context, w *World) error`
    - En `world.go`: `RegisterSystem(s System)` añade a `w.systems`; `Update(ctx context.Context) error` itera todos los sistemas registrados y ejecuta `s.Update(ctx, w)` secuencialmente
    - Si un sistema retorna error, Update detiene y propaga el error (fail-fast)
  - `internal/ecs/ecs_test.go`:
    - Test: World.Update ejecuta exactamente N sistemas registrados
    - Test: System que muta un componente (ej. incrementa contador) cambia estado del World
    - Test: lectura concurrente desde múltiples sistemas — `go test -race` sin panics
    - Test: sistema que retorna error detiene la ejecución de los siguientes
  - Verificar: `go test -race ./internal/ecs/` pasa sin data races

## Capability: data-loader

### [ ] 2.1 Implementar Registry tipado
  - `internal/data/registry.go`:
    - `Registry` struct: `data map[string]any`, `mu sync.RWMutex`
    - `NewRegistry() *Registry`
    - `Register(key string, val any)` — almacena bajo clave
    - `Get(key string) any` — retorna nil si la clave no existe
    - `GetAs[T](key string) (T, error)` — type assertion segura con error descriptivo si el tipo no coincide
  - Tests: Register + Get retorna valor; Get de clave inexistente retorna nil; GetAs[T] con tipo correcto retorna valor; GetAs[T] con tipo incorrecto retorna error; concurrencia en lecturas (RLock)
  - Verificar: `go vet ./internal/data/` sin warnings

### [ ] 2.2 Implementar Loader YAML y datos de prueba
  - `internal/data/loader.go`:
    - `Loader` struct: `dir string`, `registry *Registry`
    - `NewLoader(dir string, reg *Registry) *Loader`
    - `Load() error`: lee directorio con `os.ReadDir`, filtra archivos `*.yaml`/`*.yml`, abre cada uno, decodifica con `yaml.Decode` de `gopkg.in/yaml.v3`, registra en Registry con clave = nombre de archivo sin extensión
    - Manejo de errores: directorio no existe → error; archivo con YAML malformado → error; directorio vacío → éxito (no hay nada que cargar)
  - `data/biomes.yaml`: 3 biomas (grasslands, desert, forest) con campos `name` (string), `description` (string), `fertility` (float64), `movement_cost` (float64)
  - `internal/data/data_test.go`: test con `t.TempDir()` — YAML válido produce datos correctos en Registry; directorio faltante retorna error; YAML malformado retorna error de parseo; directorio vacío éxito sin datos
  - Verificar: `go test -v ./internal/data/` pasa

### [ ] 2.3 Implementar Validator opcional y test de validación
  - En `internal/data/loader.go`: tipo `Validator[T any] func(*T) error`; `LoadOption[T any]` struct con `Key string`, `Target *T`, `Validate Validator[T]` (opcional)
  - Registrar datos parseados con validación: si `Validate` está definida y retorna error, `Load()` propaga el error
  - Extender `data_test.go`: test donde validador chequea que biome name no sea empty — YAML con name vacío retorna validation error
  - Verificar: `go test -v -run TestValidator ./internal/data/` pasa

## Capability: tui-welcome

### [ ] 3.1 Implementar Model Bubbletea con Init y Update
  - `internal/ui/model.go`:
    - `Model` struct: `quitting bool`, `width int`, `height int`
    - `NewModel() Model` — estado inicial (quitting=false)
    - `Init() tea.Cmd` — retorna `nil` (sin comandos iniciales)
    - `Update(msg tea.Msg) (tea.Model, tea.Cmd)`:
      - `tea.WindowSizeMsg`: actualiza width/height
      - `tea.KeyMsg`: `"q"` o `"ctrl+c"` → `m.quitting = true`, retorna `m, tea.Quit`; otras teclas → no-op
      - default: retorna `m, nil` (modelo sin cambios)
  - Tests: Update con `tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}` produce `tea.Quit`; Update con `tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}` no modifica `quitting`; Update con `tea.WindowSizeMsg{Width: 80, Height: 24}` actualiza dimensiones
  - Verificar: `go vet ./internal/ui/` sin warnings

### [ ] 3.2 Implementar View con lipgloss
  - `internal/ui/view.go`:
    - `titleStyle`: `lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED")).Margin(1, 0, 0, 0)`
    - `versionStyle`: `lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))`
    - `instructionsStyle`: `lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Margin(1, 0, 0, 0)`
    - `View() string`: renderiza con `lipgloss.JoinVertical(lipgloss.Center, title, "", version, "", instructions)`
    - Textos: título `"Evociv-RL"`, versión `"v0.1.0"`, instrucciones `"Press 'q' to quit."`
  - Tests: View() retorna string no vacío; output contiene secuencias ANSI (escape codes) de lipgloss; View() no panic con modelo recién creado
  - Verificar: `go test -run TestView ./internal/ui/` pasa

### [ ] 3.3 Crear golden file y suite completa de tests TUI
  - `internal/ui/testdata/welcome.golden`: salida de referencia de `View()` (generar con `go test -update`)
  - `internal/ui/model_test.go`:
    - Golden test con flag `-update` para regenerar archivo: si `*update` es true, escribe `os.WriteFile("testdata/welcome.golden", []byte(got), 0644)`; si no, compara `got` con contenido del golden file
    - Test de integración con `teatest.NewTestModel(t, m)`: inicia modelo con teatest, envía `tea.KeyMsg` para 'q', verifica que el programa termina limpiamente
  - Tests: View() coincide exactamente con golden file (usando `flag.Update` de ADR-5); teatest corre modelo completo, recibe 'q', termina sin error
  - Verificar: `go test -v ./internal/ui/` pasa; si golden no existe, falla con mensaje claro

## Capability: store-sqlite

### [ ] 4.1 Implementar Store interface y SQLiteStore
  - `internal/store/store.go`: `Store` interface con `Open() error`, `Close() error`, `Ping() error`
  - `internal/store/sqlite.go`:
    - `SQLiteStore` struct: `db *sql.DB`, `dsn string`
    - `NewSQLiteStore(path string) *SQLiteStore` — almacena path como DSN
    - `Open() error`: `sql.Open("sqlite", s.dsn)`, `SetMaxOpenConns(1)`, `Ping()`, retorna error wrapped
    - `Ping() error`: `s.db.Ping()` (panic-safe si db es nil)
    - `Close() error`: `s.db.Close()` (panic-safe si db es nil)
  - Compile-time check: `var _ store.Store = (*store.SQLiteStore)(nil)` en un test
  - Tests: Open/Close con `t.TempDir()` — path válido funciona; double Open retorna error (store ya abierto); Ping() después de Open funciona; Open con path en directorio inexistente retorna error
  - Verificar: `go test -v ./internal/store/` pasa (sin CGO, build nativo Windows)

## Capability: integracion

### [ ] 5.1 Integrar main.go completo con todos los paquetes
  - `cmd/evociv/main.go`: reemplazar main mínimo con lógica completa de arranque:
    1. Crear World: `ecs.NewWorld()`
    2. Crear Registry + Loader: `reg := data.NewRegistry()`, `loader := data.NewLoader("data", reg)`, `loader.Load()`
    3. Abrir Store: `sqlStore := store.NewSQLiteStore("evociv.db")`, `sqlStore.Open()`, `defer sqlStore.Close()`
    4. Crear modelo TUI: `model := ui.NewModel()`
    5. Lanzar programa: `p := tea.NewProgram(model, tea.WithAltScreen())`, `p.Run()`
    6. Manejar errores con `log.Fatal` en cada punto de fallo
  - Importaciones necesarias: `ecs`, `data`, `store`, `ui`, `log` (de charmbracelet/log), `tea`
  - Tests: `go build ./cmd/evociv/` produce binario sin errores

### [ ] 5.2 Verificación final: build, tests y vet
  - `go build ./cmd/evociv/` → binario `evociv.exe` (Windows) sin errores de compilación
  - `go test ./...` → todos los tests pasan (ecs, data, ui, store)
  - `go vet ./...` → sin warnings de ningún paquete
  - `go mod tidy` → go.mod y go.sum limpios, sin dependencias huérfanas
  - `go test -race ./...` → sin data races (opcional, verificación extra)
  - Criterios de éxito del proposal: 4 paquetes con tests, binario compila, vet limpio
