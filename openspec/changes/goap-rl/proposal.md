# Proposal: GOAP + Q-Learning Híbrido

## Intent

NPCs pasan de wander aleatorio a tener necesidades (hambre, fatiga), planificar acciones con GOAP forward-chaining, y optimizar objetivos con Q-Learning ε-greedy. Comportamiento inteligente zero-shot con mejora continua vía RL.

## Scope

### In Scope
- Needs: Hunger, Fatigue (0..1, decaimiento por tick)
- 6 acciones YAML data-driven (harvest, forage, rest, socialize, trade, explore)
- GOAP forward-chaining profundidad 3: evaluar necesidades, seleccionar acción, planificar
- Q-Learning ε-greedy, tabla Q discreta ~72 estados/NPC
- GOAP como fallback cuando Q < threshold
- LOD: full GOAP local, simplificado near (5 ticks), wander distant
- Persistencia tabla Q en SQLite (batch write)
- 3 sistemas ECS: NeedsDecaySystem, GOAPSystem, QLearningSystem
- Tests: planificación, Q-learning, integración, persistencia

### Out of Scope
- DQN, redes neuronales, multi-agent RL
- Economía, inventarios, familia, roles
- Diálogo NPC, narrativa procedural

## Capabilities

### New Capabilities
- `goap-rl`: GOAP planning + Q-Learning híbrido. Necesidades, planificación forward-chaining, Q-table ε-greedy, integración LOD, persistencia SQLite.

### Modified Capabilities
- `npc-systems`: +3 sistemas (NeedsDecay, GOAP, QLearning)
- `npc-components`: +Need component (Hunger, Fatigue), actualizar AIState
- `data-loader`: +carga acciones desde YAML
- `store-sqlite`: +tabla qtable + Save/LoadQTable

## Approach

5 fases secuenciales:
1. **Necesidades**: componente Need + NeedsDecaySystem + tests
2. **Acciones YAML**: data/actions.yaml + LoadActions + tipos ActionDef + tests
3. **GOAP planificador**: forward-chaining profundidad 3, greedy, integración AIState + tests
4. **Q-Learning**: tabla Q 72 estados, ε-greedy, Δneed reward, fallback GOAP + tests
5. **Integración**: sistemas ECS, LOD escalonado, SQLite persistencia + tests

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/simulation/npc/components.go` | Modified | +Need |
| `internal/simulation/npc/types.go` | Modified | +ActionDef, QState, NeedType |
| `internal/simulation/npc/systems.go` | Modified | +3 systems |
| `internal/simulation/npc/data.go` | New | LoadActions() |
| `internal/simulation/goap/` | New | GOAP planner |
| `internal/simulation/rl/` | New | Q-table |
| `data/actions.yaml` | New | 6 actions |
| `internal/store/sqlite.go` | Modified | +qtable |
| `internal/store/store.go` | Modified | +Q-table interface |
| `cmd/evociv/main.go` | Modified | register systems |
| `openspec/specs/npc-systems/spec.md` | Modified | +GOAP/RL reqs |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Perf 100 NPCs planificando | Medium | LOD: full local, 5-tick near, wander distant |
| Q-table no converge | Medium | Fallback GOAP cuando Q < threshold |
| Acciones YAML desbalanceadas | Low | Validator + tests |
| SQLite contención writes | Low | Batch write por N ticks |

## Rollback Plan

`git revert` branch `feat/goap-rl`. `DROP TABLE IF EXISTS qtable` si hay migración. No afecta datos de mundo existentes.

## Success Criteria

- [ ] NPCs con hambre priorizan harvest/forage
- [ ] NPCs cansados priorizan rest
- [ ] GOAP planifica secuencias profundidad ≤ 3
- [ ] Q-table converge tras ~100 iteraciones
- [ ] GOAP es fallback cuando Q < threshold
- [ ] LOD distant no ejecuta GOAP
- [ ] Tabla Q persiste en SQLite
- [ ] Todos los tests verdes
