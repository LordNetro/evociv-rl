# Delta for store-sqlite

## ADDED Requirements

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
