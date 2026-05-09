# Tasks: NPC System

## Phase 1: Foundation

- [x] 1.1 **Types + ECS Components** — `internal/simulation/npc/types.go` (RaceDef, RoleDef, TraitDef, NamePool, RoleWeight, SpawnConfig, NPCRenderInfo), `internal/simulation/npc/components.go` (Health, Personality, Job, AIState, Appearance, LOD + NewPersonality()). Register 6 component stores in world.go. Extend RemoveEntity. **Tests (TDD)**: Personality determinism (same seed→same 5 traits), values in [0,1], traits differ across entities, zero-value retrieval for missing components.
- [x] 1.2 **Data YAML** — `data/npcs.yaml` (human/dwarf/elf races with traits, roles, name_pool), `data/npc-roles.yaml` (farmer/hunter/merchant/artisan/miner/smith with symbol, color, biomes). Add LoadNpcRaces/LoadNpcRoles typed loaders. **Tests (TDD)**: YAML load via data.Loader, registry access, race-role compatibility, required fields.

## Phase 2: Core Simulation

- [x] 2.1 **Spawner** — `internal/simulation/npc/spawner.go` with Spawn(). Biome-weighted placement (plains/forest=1.0, tundra/desert=0.2, ocean/jungle=0.0). Seed offset +999 via `rand.New(rand.NewSource(seed+999))`. Clamp count to [50,100] on 256×256. **Tests (TDD)**: 50–100 NPCs, determinism (identical output per seed), zero NPCs in ocean/jungle, plains > tundra statistically, race-role rejections.

- [x] 2.2 **Systems** — `internal/simulation/npc/systems.go`: NPCSpawnSystem (spawn once, no-op later), WanderSystem (LOD≥1, 8-dir, biome-compat, world-bounded), LODSystem (Chebyshev ≤5→local, ≤15→near, >15→distant), NPCRenderSystem ([]NPCRenderInfo for LOD≥1). **Tests (TDD)**: 4 systems execute per tick, LOD transitions with player move, wander within bounds, stay when surrounded by ocean, skip LOD0 on render.

## Phase 3: UI + Persistence + Integration

- [x] 3.1 **Store Persistence** — `internal/store/store.go`: extend interface with npcSeedOffset. `internal/store/sqlite.go`: migrate adds `npc_seed_offset INTEGER DEFAULT 999`, ALTER TABLE for existing DBs. SaveWorld/LoadLatestWorld handle new param. **Tests (TDD)**: save/load with offset(999), error on empty store, idempotent migration.

- [x] 3.2 **TUI Overlay** — `internal/ui/model.go`: add ecsWorld, cursorX/Y, npcOverlay fields. `internal/ui/view.go`: add renderOverlay() returning NPC symbol+color over biome tile, modify renderMap to call per tile. **Tests (TDD)**: '@' appears at correct world offset, camera offset moves NPC, golden test for overlay map.

- [x] 3.3 **TUI Inspector** — `internal/ui/model.go`: add inspectorOpen, selectedNPC. Model.Update for 'e' (open at cursor), 'q'/'esc' (close), arrows (move cursor only when open). `internal/ui/view.go`: add renderInspector() panel (Name, Health C/M, Job, Personality O/C/E/A/N to 2dp, Biome). **Tests (TDD)**: inspector shows NPC data, empty tile no-op, close via q/esc, cursor movement bounds.

- [x] 3.4 **Main Integration** — `cmd/evociv/main.go`: create ecsWorld (not discarded), register 6 component stores, load npc YAML via loader, pass to spawner, register 4 NPC systems (Spawn→Wander→LOD→Render), wire ecsWorld+npcOverlay into ui.Model. **Tests**: build compiles, run() no panic.

## Summary

| Phase | Tasks | Focus |
|-------|-------|-------|
| Phase 1 | 2 | Types, ECS components, YAML data definitions |
| Phase 2 | 2 | Spawner biome-weighted, 4 ECS systems |
| Phase 3 | 4 | Persistence, TUI overlay, inspector, main wiring |
| **Total** | **8** | |

**Order**: 1.1 → 1.2 → 2.1 → 2.2 → 3.1 → 3.2 → 3.3 → 3.4. Tasks within a phase depend on the previous (components needed by spawner, spawner needed by systems, systems output needed by TUI overlay/inspector). 3.1 (persistence) is semi-independent — can run in parallel with 3.2/3.3 if needed.
