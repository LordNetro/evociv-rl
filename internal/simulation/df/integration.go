package df

import (
    "strings"

    "github.com/marco/evociv-rl/internal/ecs"
    "github.com/marco/evociv-rl/internal/simulation/npc"
)

// DFAssignmentIntegrationSystem maps JobSystem assignments into NPC AIState actions
// so that existing GOAP/Movement systems can act on them. It is intentionally
// minimal: it records the assigned job id into AIState.AssignedJob and normalizes
// CurrentAction to a `perform_job` semantic.
type DFAssignmentIntegrationSystem struct{}

func NewDFAssignmentIntegrationSystem() *DFAssignmentIntegrationSystem { return &DFAssignmentIntegrationSystem{} }

func (s *DFAssignmentIntegrationSystem) Name() string { return "DFAssignmentIntegrationSystem" }

func (s *DFAssignmentIntegrationSystem) Update(w *ecs.World, dt float64) error {
    aiStore, _ := w.GetStore(npc.AIStateID).(*ecs.ComponentStore[npc.AIState])
    if aiStore == nil {
        return nil
    }

    for e, ai := range aiStore.All() {
        if strings.HasPrefix(ai.CurrentAction, "assigned:") {
            // extract job id
            parts := strings.SplitN(ai.CurrentAction, ":", 2)
            if len(parts) == 2 {
                ai.AssignedJob = parts[1]
                ai.CurrentAction = "perform_job"
                // keep ai.Plan if JobSystem already set it; otherwise handlers will set movement
                aiStore.Set(e, ai)
            }
        }
    }

    return nil
}
