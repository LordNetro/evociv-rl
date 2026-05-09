package df

import (
    "testing"

    "github.com/marco/evociv-rl/internal/ecs"
    "github.com/marco/evociv-rl/internal/simulation/npc"
    "github.com/marco/evociv-rl/internal/world"
)

func TestJobCompletionRewardsQL(t *testing.T) {
    w := ecs.NewWorld()
    npc.RegisterStores(w)
    RegisterStores(w)

    // create npc at target
    n := w.NewEntity()
    ecs.AddComponent(w, n, ecs.Position{X: 5, Y: 5})
    ecs.AddComponent(w, n, ecs.Name{Name: "worker"})
    ecs.AddComponent(w, n, npc.Job{Role: "farmer"})
    ecs.AddComponent(w, n, npc.AIState{AssignedJob: "job1", CurrentAction: "work_inside", Plan: []string{"5,5"}})
    ecs.AddComponent(w, n, npc.Needs{Hunger: 0.6, Fatigue: 0.2})

    // register job definition so completion uses its Reward
    RegisterJob(Job{ID: "job1", Role: "farmer", ActionID: "work_inside", Reward: 2.5})

    // add QLearningSystem so JobCompletionSystem can update its QTable
    wm := (*world.WorldMap)(nil)
    w.AddSystem(npc.NewQLearningSystem(wm, nil, nil))

    sys := NewJobCompletionSystem()
    w.AddSystem(sys)

    if err := w.Update(0.1); err != nil {
        t.Fatalf("update failed: %v", err)
    }

    ai, _ := ecs.GetComponent[npc.AIState](w, n)
    if ai.AssignedJob != "" {
        t.Fatalf("expected AssignedJob cleared after completion")
    }
}
