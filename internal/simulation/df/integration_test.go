package df

import (
    "testing"

    "github.com/marco/evociv-rl/internal/ecs"
    "github.com/marco/evociv-rl/internal/simulation/npc"
)

func TestAssignmentIntegration(t *testing.T) {
    w := ecs.NewWorld()
    npc.RegisterStores(w)
    RegisterStores(w)

    // create npc with AIState set to assigned:job1
    n := w.NewEntity()
    ecs.AddComponent(w, n, ecs.Position{X: 0, Y: 0})
    ecs.AddComponent(w, n, ecs.Name{Name: "worker"})
    ecs.AddComponent(w, n, npc.Job{Role: "farmer"})
    ecs.AddComponent(w, n, npc.AIState{CurrentAction: "assigned:job1"})

    sys := NewDFAssignmentIntegrationSystem()
    w.AddSystem(sys)

    if err := w.Update(0.1); err != nil {
        t.Fatalf("update error: %v", err)
    }

    ai, ok := ecs.GetComponent[npc.AIState](w, n)
    if !ok {
        t.Fatalf("expected AIState")
    }
    if ai.CurrentAction != "perform_job" {
        t.Fatalf("expected CurrentAction=perform_job, got %s", ai.CurrentAction)
    }
    if ai.AssignedJob != "job1" {
        t.Fatalf("expected AssignedJob=job1, got %s", ai.AssignedJob)
    }
}
