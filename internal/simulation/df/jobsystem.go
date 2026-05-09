package df

import (
    "fmt"

    "github.com/marco/evociv-rl/internal/ecs"
    "github.com/marco/evociv-rl/internal/simulation/npc"
)

// JobSystem assigns jobs from JobQueue to NPCs matching the role.
type JobSystem struct{}

// NewJobSystem creates a new JobSystem.
func NewJobSystem() *JobSystem { return &JobSystem{} }

func (s *JobSystem) Name() string { return "DFJobSystem" }

func (s *JobSystem) Update(w *ecs.World, dt float64) error {
    jqStore, ok := w.GetStore(JobQueueID).(*ecs.ComponentStore[JobQueue])
    if !ok || jqStore == nil {
        return nil
    }
    // NPC stores
    npcJobStore, _ := w.GetStore(npc.JobID).(*ecs.ComponentStore[npc.Job])
    aiStore, _ := w.GetStore(npc.AIStateID).(*ecs.ComponentStore[npc.AIState])

    for be, jq := range jqStore.All() {
        if len(jq.Jobs) == 0 {
            continue
        }
        // building inventory store, if present
        invStore, _ := w.GetStore(InventoryID).(*ecs.ComponentStore[Inventory])
        // try to assign first job to a matching npc
        for ji, job := range jq.Jobs {
            assigned := false
            // check consumes: ensure building has required items
            if len(job.Consumes) > 0 && invStore != nil {
                if inv, ok := invStore.Get(be); ok {
                    okToAssign := true
                    for item, req := range job.Consumes {
                        if inv.Items[item] < req {
                            okToAssign = false
                            break
                        }
                    }
                    if !okToAssign {
                        // skip this job for now
                        continue
                    }
                    // reserve: decrement items immediately
                    for item, req := range job.Consumes {
                        inv.Items[item] -= req
                    }
                    invStore.Set(be, inv)
                } else {
                    // no inventory, cannot assign consuming job
                    continue
                }
            }
            for e, nj := range npcJobStore.All() {
                if nj.Role != job.Role {
                    continue
                }
                // check NPC AIState is idle
                if aiStore != nil {
                    ai, _ := aiStore.Get(e)
                    if ai.CurrentAction != "" {
                        continue
                    }
                    // assign job: set AssignedJob and set CurrentAction to the job's ActionID
                    ai.AssignedJob = job.ID
                    if job.ActionID != "" {
                        ai.CurrentAction = job.ActionID
                    } else {
                        ai.CurrentAction = "perform_job"
                    }
                    // set target plan if present
                    ai.Plan = []string{fmt.Sprintf("%d,%d", job.TargetX, job.TargetY)}
                    aiStore.Set(e, ai)
                    // remove job from queue
                    jq.Jobs = append(jq.Jobs[:ji], jq.Jobs[ji+1:]...)
                    jqStore.Set(be, jq)
                    assigned = true
                    break
                }
            }
            if assigned {
                break
            }
        }
    }
    return nil
}
