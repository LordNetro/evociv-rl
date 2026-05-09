# store-sqlite Specification

## Purpose

Proporcionar una interfaz Store abstracta y una implementación concreta sobre SQLite (pure-Go, sin CGO). Sirve como base de persistencia para el simulador.

## Requirements

### Requirement: Store Interface

The system MUST define a Store interface with at least Open and Close methods.

#### Scenario: Interface is satisfiable

- GIVEN the Store interface
- WHEN a struct implements both Open() and Close()
- THEN it MUST satisfy the interface at compile time

### Requirement: SQLite Implementation

The system MUST provide a concrete SQLite-backed implementation of Store using `modernc.org/sqlite`.

#### Scenario: Open and Close succeed

- GIVEN a SQLiteStore with a valid file path
- WHEN Open() is called followed by Close()
- THEN no errors MUST be returned

#### Scenario: Open with invalid path returns error

- GIVEN a SQLiteStore with an invalid path (e.g., locked directory)
- WHEN Open() is called
- THEN an error MUST be returned

### Requirement: Test Coverage

The system MUST have unit tests covering happy path and error cases for both interface and implementation.

#### Scenario: Happy path test

- GIVEN a SQLiteStore with a temp file path (t.TempDir())
- WHEN Open and Close are called sequentially
- THEN both MUST succeed

#### Scenario: Double open returns error

- GIVEN an already-opened SQLiteStore
- WHEN Open is called again
- THEN an error MUST be returned (or the operation MUST be a no-op)

### Requirement: Future Extensibility (Edge)

The Store interface SHOULD remain minimal (Open, Close) to allow alternative backends (e.g., in-memory, Postgres) in future changes without breaking consumers.

### Requirement: World Persistence

Store MUST add SaveWorld(seed int64, w, h int, npcSeedOffset int64) error and LoadLatestWorld() (seed int64, w, h int, npcSeedOffset int64, error). SQLite stores in worlds table. The `npc_seed_offset` column MUST be persisted so that NPC placement is deterministically reproducible after save/load cycles.
(Previously: SaveWorld took (seed, w, h) only; LoadLatestWorld returned (seed, w, h) only. No NPC offset column.)

#### Scenario: Save and retrieve with offset
- GIVEN opened store
- WHEN SaveWorld(42, 64, 64, 999) then LoadLatestWorld()
- THEN seed=42, width=64, height=64, npcSeedOffset=999

#### Scenario: Empty store
- GIVEN fresh store, no worlds
- WHEN LoadLatestWorld()
- THEN MUST return error

#### Scenario: Deterministic regeneration after save/load
- GIVEN a world saved with seed=42 and npcSeedOffset=999
- WHEN the world is loaded and NPC spawning uses those values
- THEN the NPC set MUST be identical to the original spawn

### Requirement: Q-Table Persistence

The Store interface MUST add `SaveQTable(npcID int, qTable map[string]map[string]float64) error` and `LoadQTable(npcID int) (map[string]map[string]float64, error)`. The SQLite implementation MUST store Q-values in a `qtable` table.

The `qtable` schema MUST be:

```sql
CREATE TABLE IF NOT EXISTS qtable (
    npc_id    INTEGER,
    state_key TEXT,
    action_id TEXT,
    q_value   REAL,
    PRIMARY KEY (npc_id, state_key, action_id)
);
```

#### Scenario: Save and load Q-table

- GIVEN an opened store and a Q-table with 3 state-action pairs for NPC 1
- WHEN SaveQTable(1, data) then LoadQTable(1) are called
- THEN the loaded data MUST match the saved data exactly for all pairs

#### Scenario: Load empty table returns empty map

- GIVEN an opened store with no rows in qtable
- WHEN LoadQTable(1) is called
- THEN an empty map MUST be returned (no error)

#### Scenario: Persistence across open/close cycle

- GIVEN saved Q-values for NPC 1
- WHEN the store is closed and reopened, then LoadQTable(1) is called
- THEN the Q-values MUST be recoverable

### Requirement: Save on Exit, Load on Start

The system MUST save the Q-table for each NPC on simulation exit (Close) and load it on simulation start (Open).

#### Scenario: Q-table saved on close

- GIVEN an active simulation with modified Q-values
- WHEN Close() is called
- THEN SaveQTable MUST be invoked for each NPC with Q-values
