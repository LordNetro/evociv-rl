# Proposal: tilemap-renderer

## Intent

Enable a layered ASCII tilemap renderer that supports multi-tile buildings (rectangular footprints) and 2 Z-levels (surface world + building interiors), replacing the current single-layer biome renderer in `view.go`. This is essential for the DF-like experience where players explore the world map and enter settlement buildings to see NPCs working inside.

## Scope

### In Scope
- New `internal/ui/tilemap/` package with `Tile`, `Tilemap`, `Camera` types
- `Tile` struct with 5 layers: 0=Terrain, 1=Buildings, 2=Items, 3=Creatures, 4=UI
- `Tilemap` type: 2D grid of Tiles
- `Camera` with X, Y, Z (Z-level), Width, Height
- 5-layer rendering pipeline with ASCII character mapping
- Multi-tile building support (rectangular footprints)
- 2 Z-levels: surface (Z=0, world map) and interior (Z=1, building interiors)
- Migration of current `view.go` renderer to layer-based system
- Settlement interior view integrated into tilemap (Z=1)

### Out of Scope
- Q-table visualization overlays (PR E)
- World generation refactor (PR D)
- Inventory/item rendering details beyond basic ASCII

## Capabilities

### New Capabilities
- `tilemap-renderer`: Multi-layered ASCII tilemap with Z-levels and multi-tile buildings

### Modified Capabilities
- `settlement-tui`: Existing TUI spec will need delta for integrated view

## Approach

Phased migration to minimize TUI breakage:

1. **Parallel package**: Create `internal/ui/tilemap/` alongside existing `view.go`
2. **Core types**: Implement `Tile`, `Tilemap`, `Camera` with 5-layer pipeline
3. **迁移**: Port world map rendering from `renderMap()` to new pipeline
4. **集成**: Merge `renderSettlementView()` as Z=1 interior view
5. **Swap**: Update TUI model to use tilemap package, remove old renderer

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/ui/tilemap/` | New | Package for layered tilemap rendering |
| `internal/ui/view.go` | Modified | Migrate to tilemap package |
| `internal/ui/model.go` | Modified | Update camera/view state for Z-levels |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Breaking TUI during migration | Medium | Run parallel until verified |
| Multi-tile building pathfinding | Low | Use same ECS Position component |
| Z-level camera transitions | Medium | Simple Z-switch on building entry |

## Rollback Plan

If migration breaks the TUI:
1. Revert `internal/ui/model.go` to use original `view.go` functions
2. Keep `internal/ui/tilemap/` as unused package
3. Roll back in single commit

## Dependencies

- PR A (Inventory/Item/Job types): **Already merged**
- PR B (JobSystem, Job completion): **Already merged**
- None blocking this change

## Success Criteria

- [ ] Tests pass with new tilemap renderer
- [ ] World map displays with 5-layer rendering
- [ ] Building interiors viewable at Z=1
- [ ] Multi-tile buildings render with correct footprints
- [ ] Camera transitions between Z=0 and Z=1 work