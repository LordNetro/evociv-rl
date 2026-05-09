package df

import "sync"

// JobRegistry holds canonical job definitions (used for rewards, metadata).
var (
    registryMu sync.RWMutex
    jobRegistry = make(map[string]Job)
)

// RegisterJob registers or replaces a job definition by ID.
func RegisterJob(j Job) {
    registryMu.Lock()
    defer registryMu.Unlock()
    jobRegistry[j.ID] = j
}

// GetJob returns a job definition if present.
func GetJob(id string) (Job, bool) {
    registryMu.RLock()
    defer registryMu.RUnlock()
    j, ok := jobRegistry[id]
    return j, ok
}

// AllJobs returns a slice of all registered job definitions.
func AllJobs() []Job {
    registryMu.RLock()
    defer registryMu.RUnlock()
    out := make([]Job, 0, len(jobRegistry))
    for _, j := range jobRegistry {
        out = append(out, j)
    }
    return out
}
