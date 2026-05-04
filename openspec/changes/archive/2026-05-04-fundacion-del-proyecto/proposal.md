# Proposal: Fundación del proyecto

## Intent

Establecer la base funcional del simulador Evociv-RL: ECS operativo con tests, carga YAML data-driven, TUI de bienvenida con golden test, y stub de persistencia SQLite. Todo compila, tests pasan, `go vet` limpio.

## Scope

### In Scope
- ECS map-based: `Entity`, `ComponentStore[T]`, `World`, `System` interface + tests
- Data-Driven: loader YAML, registry + YAML de prueba (2-3 biomas) + tests
- TUI: pantalla de bienvenida Bubbletea con lipgloss, tecla 'q' para salir + golden test
- Store: `Store` interface + implementación SQLite stub (open/close) + tests
- Infraestructura: estructura de directorios, go.mod con dependencias, `main.go` compilable
- Build pipeline: `go build`, `go test ./...`, `go vet ./...`

### Out of Scope
- Generación procedural de mundo (próximo cambio)
- GOAP/RL (próximo cambio)
- LLM bridge con Ollama (próximo cambio)
- Múltiples pantallas TUI (próximo cambio)
- Persistencia real de datos (próximo cambio)

## Capabilities

### New Capabilities
- `ecs-core`: Entity, World, ComponentStore[T], System interface — núcleo del ECS
- `data-loader`: Loader YAML + registry + validator — carga data-driven
- `tui-welcome`: Pantalla de bienvenida Bubbletea con salida por tecla — TUI base
- `store-sqlite`: Store interface + implementación SQLite — persistencia base

### Modified Capabilities
None — primer cambio, no hay specs existentes.

## Approach

6 fases secuenciales, cada una con tests antes de implementar (Strict TDD):

1. **Fase 0 — Infraestructura**: directorios, go.mod, dependencias, `cmd/evociv/main.go` mínimo que compila
2. **Fase 1 — ECS**: entity.go, world.go, component.go, system.go + tests (crear entity, asignar componente, ejecutar sistema)
3. **Fase 2 — Data-Driven**: loader.go (carga YAML desde directorio), registry.go + biomes.yaml (2-3 biomas) + tests
4. **Fase 3 — TUI**: model.go, update.go, view.go con pantalla de bienvenida + golden test + test de tecla 'q'
5. **Fase 4 — Store**: store.go (interface), sqlite.go (SQLite stub) + tests
6. **Fase 5 — Integración**: `go build`, `go test ./...`, `go vet ./...`, verificar que todo pasa

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `cmd/evociv/main.go` | New | Entry point del proyecto |
| `internal/ecs/` | New | Paquete ECS (entity, world, component, system) |
| `internal/data/` | New | Paquete data-driven (loader, registry) |
| `internal/store/` | New | Paquete de persistencia (interface + SQLite) |
| `internal/ui/` | New | Paquete TUI (model, update, view) |
| `data/biomes.yaml` | New | YAML de prueba con 2-3 biomas |
| `go.mod` / `go.sum` | Modified | Dependencias añadidas |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Bubbletea 1.x API breaking vs docs | Medium | Tests TUI validan uso real, no asumir API de v0.x |
| modernc.org/sqlite compilación Windows | Low | Dependencia pura Go, test store valida compilación |
| Go no está en PATH en Windows | High | Script `scripts/setup.ps1` + documentación en README |
| Golden files CRLF vs LF | Low | `.gitattributes` con `*.golden text eol=lf` + normalización en tests |

## Rollback Plan

Al ser todo archivos nuevos (no se modifica código existente), el rollback es trivial:
- `git revert <commit>` del cambio completo
- `go mod tidy` para limpiar go.sum si quedan dependencias huérfanas
- No hay migraciones de datos ni cambios en producción

## Dependencies

| Paquete | Versión | Propósito |
|---------|---------|-----------|
| `github.com/charmbracelet/bubbletea` | v1.3.10 | Framework TUI |
| `github.com/charmbracelet/lipgloss` | v1.1.0 | Estilos TUI |
| `github.com/charmbracelet/bubbles` | v1.0.0 | Componentes TUI |
| `github.com/charmbracelet/log` | v1.0.0 | Logging estructurado |
| `modernc.org/sqlite` | v1.50.0 | SQLite sin CGO |
| `gopkg.in/yaml.v3` | v3.0.1 | Parseo YAML |
| `github.com/charmbracelet/x` | — | `teatest` + `golden` (solo tests) |

## Success Criteria

- [ ] `go build ./cmd/evociv` produce binario sin errores
- [ ] `go test ./...` pasa todos los tests (mínimo 4 paquetes: ecs, data, ui, store)
- [ ] `go vet ./...` sin warnings
- [ ] ECS: crear entity, asignar/leer componente, ejecutar sistema en test
- [ ] Data: cargar YAML, verificar contenido, error en YAML malformado
- [ ] TUI: golden test de pantalla de bienvenida, tecla 'q' produce tea.Quit
- [ ] Store: open/close sin errores
