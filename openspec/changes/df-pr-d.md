# PR D — Job Registry and reward wiring

Resumen:
- Añadido `internal/simulation/df/registry.go`: un registro global de definiciones de jobs (`RegisterJob`, `GetJob`).
- `JobCompletionSystem` ahora lee `Job.Reward` desde el registro en vez de usar un valor por defecto.
- Tests actualizados para registrar la definición del job y verificar que la finalización del job limpia la asignación.

Motivación:
- Persistir metadatos de jobs (recompensas, action mapping, payloads) fuera de colas facilita la edición desde UI y mantiene colas ligeras.

Siguientes pasos:
- UI: permitir crear/editar jobs (enqueue) y mostrar job registry.
- Persistencia: cargar job definitions desde `data/` YAML a `jobRegistry` en bootstrap.

Files:
- internal/simulation/df/registry.go (new)
- internal/simulation/df/completion.go (updated to use registry)
- tests updated accordingly

