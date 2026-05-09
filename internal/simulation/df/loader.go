package df

import (
    "fmt"

    "github.com/marco/evociv-rl/internal/data"
)

// LoadJobsFromRegistry reads job definitions from the data.Registry under the "jobs" key
// and registers them into the JobRegistry.
func LoadJobsFromRegistry(registry *data.Registry) error {
    raw, ok := data.Get[[]any](registry, "jobs")
    if !ok {
        // nothing to load is fine
        return nil
    }
    for _, item := range raw {
        m, ok := item.(map[string]any)
        if !ok {
            continue
        }
        j := Job{}
        if v, ok := m["id"].(string); ok {
            j.ID = v
        } else {
            return fmt.Errorf("job missing id")
        }
        if v, ok := m["role"].(string); ok {
            j.Role = v
        }
        if v, ok := m["action_id"].(string); ok {
            j.ActionID = v
        }
        if v, ok := m["reward"].(float64); ok {
            j.Reward = v
        } else if v, ok := m["reward"].(int); ok {
            j.Reward = float64(v)
        }
        // optional consumes/produces maps
        if c, ok := m["consumes"].(map[string]any); ok {
            j.Consumes = make(map[string]int)
            for k, vv := range c {
                switch val := vv.(type) {
                case int:
                    j.Consumes[k] = val
                case float64:
                    j.Consumes[k] = int(val)
                }
            }
        }
        if p, ok := m["produces"].(map[string]any); ok {
            j.Produces = make(map[string]int)
            for k, vv := range p {
                switch val := vv.(type) {
                case int:
                    j.Produces[k] = val
                case float64:
                    j.Produces[k] = int(val)
                }
            }
        }
        // optional target coordinates
        if tx, ok := m["target_x"].(int); ok {
            j.TargetX = tx
        }
        if ty, ok := m["target_y"].(int); ok {
            j.TargetY = ty
        }

        RegisterJob(j)
    }
    return nil
}
