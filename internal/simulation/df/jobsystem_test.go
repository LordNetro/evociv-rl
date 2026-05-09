package df

import (
    "testing"

    "github.com/marco/evociv-rl/internal/ecs"
    "github.com/marco/evociv-rl/internal/simulation/npc"
)

func TestJobAssignment(t *testing.T) {
    w := ecs.NewWorld()
    // register necessary stores
    npc.RegisterStores(w)
    RegisterStores(w)

    // create a building with a job queue
    be := w.NewEntity()
    jq := JobQueue{
        Jobs: []Job{{ID: "job1", Role: "farmer", TargetX: 5, TargetY: 5, ActionID: "work_inside", Reward: 1.5}},
    }
    ecs.AddComponent(w, be, jq)

    // create an NPC with matching role
    n := w.NewEntity()
    ecs.AddComponent(w, n, ecs.Position{X: 0, Y: 0})
    ecs.AddComponent(w, n, ecs.Name{Name: "worker"})
    ecs.AddComponent(w, n, npc.Job{Role: "farmer"})
    ecs.AddComponent(w, n, npc.AIState{})

    js := NewJobSystem()
    w.AddSystem(js)

    if err := w.Update(0.1); err != nil {
        t.Fatalf("update failed: %v", err)
    }

    ai, ok := ecs.GetComponent[npc.AIState](w, n)
    if !ok {
        t.Fatal("expected AIState on npc")
    }
    if ai.AssignedJob == "" {
        t.Fatalf("expected NPC to have AssignedJob set")
    }
    if ai.CurrentAction != "work_inside" {
        t.Fatalf("expected CurrentAction=work_inside, got %s", ai.CurrentAction)
    }

    // ensure registry can store job definitions
    RegisterJob(Job{ID: "job1", Role: "farmer", ActionID: "work_inside", Reward: 1.5})

    // ensure job queue is empty
    jq2, ok := ecs.GetComponent[JobQueue](w, be)
    if !ok {
        t.Fatal("expected JobQueue on building")
    }
    if len(jq2.Jobs) != 0 {
        t.Fatalf("expected job queue empty after assignment, got %d", len(jq2.Jobs))
    }
}
