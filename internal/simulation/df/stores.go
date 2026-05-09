package df

import "github.com/marco/evociv-rl/internal/ecs"

// Component IDs for DF simulation package.
var (
    JobQueueID = ecs.NewComponentID("df_jobqueue")
    InventoryID = ecs.NewComponentID("df_inventory")
)

// RegisterStores registers DF component stores into the world.
func RegisterStores(w *ecs.World) {
    ecs.RegisterComponentStore[JobQueue](w, JobQueueID, ecs.NewComponentStore[JobQueue]())
    ecs.RegisterComponentStore[Inventory](w, InventoryID, ecs.NewComponentStore[Inventory]())
}
