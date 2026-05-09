# PR C — Job completion detection and RL reward integration

Resumen:
- Añadido `ActionID` y `Reward` a `df.Job` para mapear jobs a GOAP/QLearning actions and per-job rewards.
- `JobSystem` ahora asigna `AIState.AssignedJob` y `AIState.CurrentAction = Job.ActionID` (o `perform_job` si no existe `ActionID`).
- Añadido `JobCompletionSystem` que detecta llegada al objetivo (plan) y:
  - da una recompensa por defecto (1.0) al `QLearningSystem` Q-table (best-effort);
  - limpia `AIState.AssignedJob`, `CurrentAction` y `Plan`.
- Tests añadidos/actualizados: `jobsystem_test.go`, `integration_test.go`, `completion_test.go`.

Notas:
- La integración actual da una recompensa por defecto (1.0). Podemos mejorarla leyendo `Job.Reward` desde una registración central del job en futuras iteraciones.
- La actualización del Q-table se hace buscando el `QLearningSystem` en `World.Systems()` y llamando a `QTable().Update(...)` con heurísticamente reconstruido `state`/`nextState`.

Siguientes pasos propuestos:
- Persistir job definitions en un registro accesible para `JobCompletionSystem` (para usar `Job.Reward`).
- Mapear recompensas más finas y señales para GOAP effects.
- Visualizar recompensas y Q-values en la TUI para debug.
