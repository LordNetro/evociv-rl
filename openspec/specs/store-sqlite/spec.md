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
