# PR B — Integración de asignaciones con AI (DF refactor slice)

Resumen:
- Añadido `AssignedJob` a `npc.AIState` para almacenar el id del trabajo asignado.
- Implementado `DFAssignmentIntegrationSystem` que detecta `AIState.CurrentAction` con el prefijo `assigned:<id>` y:
  - guarda `AssignedJob = <id>` en `AIState`
  - normaliza `CurrentAction` a `perform_job` para que GOAP/Movement/Exec systems actúen sobre ello
- Añadido test `internal/simulation/df/integration_test.go`.

Rationale:
- Mantener un paso intermedio claro entre el `JobSystem` (gestión y colas) y la lógica de ejecución de acciones existente (GOAP/QLearning). Esto permite mapear asignaciones a acciones GOAP o a recompensas QLearning en slices posteriores.

Siguientes pasos sugeridos:
- PR C: detectar completitud del job (cuando el NPC alcanza la ubicación/realiza la acción) y emitir una recompensa a `QLearningSystem` o desencadenar la acción GOAP final.
- PR D: enlazar `JobQueue` con edificios generados y UI para encolar jobs desde el jugador.

Files:
- internal/simulation/npc/components.go (AIState modification)
- internal/simulation/df/integration.go (new)
- internal/simulation/df/integration_test.go (new)
- openspec/changes/df-pr-b.md (this file)
