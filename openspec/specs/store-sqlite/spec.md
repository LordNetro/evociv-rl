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
