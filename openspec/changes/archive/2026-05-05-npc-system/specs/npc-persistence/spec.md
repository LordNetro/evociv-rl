# Delta for store-sqlite

## MODIFIED Requirements

### Requirement: World Persistence

Store MUST add SaveWorld(seed int64, w, h int, npcSeedOffset int64) error and LoadLatestWorld() (seed int64, w, h int, npcSeedOffset int64, error). SQLite stores in worlds table. The `npc_seed_offset` column MUST be persisted so that NPC placement is deterministically reproducible after save/load cycles.
(Previously: SaveWorld took (seed, w, h) only; LoadLatestWorld returned (seed, w, h) only. No NPC offset column.)

#### Scenario: Save and retrieve with offset

- GIVEN an opened store
- WHEN SaveWorld(42, 64, 64, 999) is called, then LoadLatestWorld()
- THEN seed MUST be 42, width 64, height 64, npcSeedOffset 999

#### Scenario: Empty store returns error

- GIVEN a fresh store with no worlds saved
- WHEN LoadLatestWorld() is called
- THEN an error MUST be returned

#### Scenario: Deterministic regeneration after save/load

- GIVEN a world saved with seed=42 and npcSeedOffset=999
- WHEN the world is loaded and NPC spawning uses those values
- THEN the NPC set MUST be identical to the original spawn
