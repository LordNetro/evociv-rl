# Tasks: GOAP + Q-Learning Hybrid

## Phase 1: Foundation

- [x] 1.1 **Needs Component** — `components.go`: Needs{Hunger, Fatigue float64}, NeedsID, RegisterStores update. `types.go`: NeedsValues. **Tests**: range [0,1], zero-init, clamp at max.
- [x] 1.2 **ActionDef Types + YAML** — `types.go`: ActionDef, ActionRequires, ActionEffects, ActionReward, NeedsValues. `data.go`: LoadActions(registry) for kind `npc-actions`. `data/actions.yaml`: 6 acciones (harvest, forage, rest, socialize, trade, explore). **Tests**: load 6 actions, biome restriction, missing field error, invalid YAML error.

## Phase 2: Core Logic

- [x] 2.1 **GOAP Planner** — `internal/simulation/goap/planner.go`: Plan(Needs, []ActionDef, biome) → ActionDef, forward-chaining depth≤3. **Tests**: high hunger→harvest/forage, high fatigue→rest, low needs→explore/socialize, biome target in plan, replan on need change, LOD scaling levels.

- [x] 2.2 **Q-Learning Engine** — `internal/simulation/rl/qtable.go`: QTable{values map[string]map[string]float64}, eGreedy(state, actions, rng), Update(state, action, reward, nextState, alpha, gamma). **Tests**: zero-init Q, ε-greedy ~50% explore at ε=0.5, ε decays 0.5→0.05 over 1000 iters, positive reward reinforces, GOAP fallback when all Q<threshold.

## Phase 3: ECS Systems

- [x] 3.1 **NeedsDecaySystem** — `systems.go`: Hung+=0.01×lodMul×dt, Fatig+=0.005×lodMul×dt, clamp [0,1], LOD0=0.5×. **Tests**: 10 ticks→Hunger 0.10, Fatigue 0.05, clamp, LOD0 half rate.

- [x] 3.2 **GOAPSystem** — `systems.go`: planifica NPCs LOD≥1, LOD2=full plan depth≤3, LOD1=1-step, LOD0=skip. **Tests**: full plan at LOD2, 1-step at LOD1, no plan at LOD0, need-selects correct action.

- [x] 3.3 **QLearningSystem** — `systems.go`: ε-greedy select, execute, reward, update Q. GOAP fallback when Q<0.1. **Tests**: explore vs exploit, reward includes completion bonus, GOAP fallback at zero Q.

## Phase 4: Persistence

- [x] 4.1 **SQLite Q-Table** — `store.go`: SaveQTable(npcID, qtable), LoadQTable(npcID). `sqlite.go`: CREATE TABLE qtable(npc_id, state_key, action_id, q_value, PK), migrate, batch Save/Load. **Tests**: save→load match, empty→empty map, open/close recover.

## Phase 5: Integration

- [x] 5.1 **Main Wiring** — `cmd/evociv/main.go`: register 3 systems (NeedsDecay→GOAP→QLearning) tras WanderSystem, load actions via data.Loader. **Smoke**: build compiles, run() no panic, LOD0 skips GOAP/QL.

## Summary

| Phase | Tasks | Focus |
|-------|-------|-------|
| Phase 1 | 2 | Needs component + ActionDef types + YAML load |
| Phase 2 | 2 | GOAP planner + Q-Learning engine |
| Phase 3 | 3 | NeedsDecaySystem + GOAPSystem + QLearningSystem |
| Phase 4 | 1 | SQLite qtable persistence |
| Phase 5 | 1 | main.go wiring + smoke test |
| **Total** | **9** | |

**Order**: 1.1 → 1.2 → 2.1 → 2.2 → 3.1 → 3.2 → 3.3 → 4.1 → 5.1. Strict dependencies: Planner necesita ActionDef (1.2), QLSystem necesita QTable (2.2), GOAPSystem necesita Planner (2.1) y Needs (1.1), Integración necesita todos.
