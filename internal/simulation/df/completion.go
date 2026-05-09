package df

import (
    "fmt"

    "github.com/marco/evociv-rl/internal/ecs"
    "github.com/marco/evociv-rl/internal/simulation/npc"
)

// JobCompletionSystem detects when NPCs complete assigned jobs (reach the plan target)
// and rewards the QLearningSystem accordingly, then clears the assignment.
type JobCompletionSystem struct{}

func NewJobCompletionSystem() *JobCompletionSystem { return &JobCompletionSystem{} }

func (s *JobCompletionSystem) Name() string { return "DFJobCompletionSystem" }

func (s *JobCompletionSystem) Update(w *ecs.World, dt float64) error {
    posStore := w.GetStore(ecs.NewComponentID("position")).(*ecs.ComponentStore[ecs.Position])
    aiStore, _ := w.GetStore(npc.AIStateID).(*ecs.ComponentStore[npc.AIState])
    if posStore == nil || aiStore == nil {
        return nil
    }

    // find QLearningSystem if present
    var ql *npc.QLearningSystem
    for _, sys := range w.Systems() {
        if ssys, ok := sys.(*npc.QLearningSystem); ok {
            ql = ssys
            break
        }
    }

    for e, ai := range aiStore.All() {
        if ai.AssignedJob == "" {
            continue
        }
        // if no plan, skip
        if len(ai.Plan) == 0 {
            continue
        }
        // parse target from plan
        var tx, ty int
        fmt.Sscanf(ai.Plan[0], "%d,%d", &tx, &ty)
        pos, ok := posStore.Get(e)
        if !ok {
            continue
        }
        // consider completion when within Chebyshev distance <=1
        if chebyshev(int(pos.X), int(pos.Y), tx, ty) > 1 {
            continue
        }

        // Completed: reward QLearningSystem using job definition if available
        reward := 1.0
        if ai.AssignedJob != "" {
            if jd, ok := GetJob(ai.AssignedJob); ok {
                reward = jd.Reward
                // apply produces to target entity (building) or to NPC's inventory
                if len(jd.Produces) > 0 {
                    invStore, _ := w.GetStore(InventoryID).(*ecs.ComponentStore[Inventory])
                    if invStore != nil {
                        // prefer job TargetEntity if present
                        if jd.TargetEntity != 0 {
                            te := ecs.Entity(jd.TargetEntity)
                            inv, _ := invStore.Get(te)
                            if inv.Items == nil {
                                inv.Items = map[string]int{}
                            }
                            for it, qty := range jd.Produces {
                                inv.Items[it] = inv.Items[it] + qty
                            }
                            invStore.Set(te, inv)
                        } else {
                            // fallback: give produced items to NPC entity
                            inv, _ := invStore.Get(e)
                            if inv.Items == nil {
                                inv.Items = map[string]int{}
                            }
                            for it, qty := range jd.Produces {
                                inv.Items[it] = inv.Items[it] + qty
                            }
                            invStore.Set(e, inv)
                        }
                    }
                }
            }
        }

        if ql != nil {
            // compute state/action similar to QLearningSystem
            needsStore, _ := w.GetStore(npc.NeedsID).(*ecs.ComponentStore[npc.Needs])
            var state, nextState string
            biome := ""
            if ql.QTable() != nil {
                // try to reconstruct state using similar heuristics; best-effort
                if needsStore != nil {
                    if needs, ok := needsStore.Get(e); ok {
                        needType := "none"
                        if needs.Hunger > needs.Fatigue && needs.Hunger > 0.3 {
                            needType = "hunger"
                        } else if needs.Fatigue > needs.Hunger && needs.Fatigue > 0.3 {
                            needType = "fatigue"
                        }
                        state = fmt.Sprintf("%s|%s|day", needType, biome)
                        nextState = state
                    }
                }
            }
            actionID := ai.CurrentAction
            if state != "" && actionID != "" {
                ql.QTable().Update(state, actionID, reward, nextState, 0.1, 0.9)
                ql.QTable().DecayEpsilon()
            }
        }

        // set last reward on AIState and clear assignment
        ai.LastReward = reward
        ai.RewardTick = 0
        ai.AssignedJob = ""
        ai.CurrentAction = ""
        ai.Plan = nil
        aiStore.Set(e, ai)
    }

    return nil
}

func chebyshev(x1, y1, x2, y2 int) int {
    dx := x1 - x2
    if dx < 0 {
        dx = -dx
    }
    dy := y1 - y2
    if dy < 0 {
        dy = -dy
    }
    if dx > dy {
        return dx
    }
    return dy
}
