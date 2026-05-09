# PR A — Job component & JobSystem (DF refactor initial slice)

Summary:
- Added DF package component store for `JobQueue`.
- Implemented a minimal `JobSystem` that assigns jobs from a building's `JobQueue` to idle NPCs whose `npc.Job.Role` matches the `Job.Role`.
- Added `Job`, `JobQueue`, `Inventory`, and `Item` types in `internal/simulation/df/components.go`.
- Added unit test `internal/simulation/df/jobsystem_test.go` validating assignment and queue removal.

Files changed/added:
- internal/simulation/df/components.go (new)
- internal/simulation/df/stores.go (new)
- internal/simulation/df/jobsystem.go (new)
- internal/simulation/df/jobsystem_test.go (new)
- openspec/changes/df-pr-a.md (this file)

Notes:
- The `JobSystem` is intentionally minimal; it acts as an adapter between existing `npc` components and new DF job queues.
- Integration with GOAP/QLearning will come in next slices: e.g., mapping `assigned:<id>` actions into GOAP actions or QLearning rewards.
- Tests ensure the change is non-invasive and compiles with the rest of the suite.
