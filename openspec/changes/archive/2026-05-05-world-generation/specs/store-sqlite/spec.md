# Delta for store-sqlite

## ADDED Requirements

### Requirement: World Persistence

Store MUST add SaveWorld(seed int64, w, h int) error and LoadLatestWorld() (seed int64, w, h int, error). SQLite stores in worlds table.

#### Scenario: Save+retrieve
- GIVEN opened store
- WHEN SaveWorld(42,64,64) then LoadLatestWorld()
- THEN seed=42, width=64, height=64

#### Scenario: Empty store
- GIVEN fresh store, no worlds
- WHEN LoadLatestWorld()
- THEN MUST return error
