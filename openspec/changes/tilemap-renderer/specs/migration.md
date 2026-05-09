# Migration Specification - From view.go to Tilemap

## Purpose

Define the migration strategy from the current view.go renderer to the layered tilemap system.

## Requirements

### Requirement: Parallel Package Creation

The new tilemap package MUST exist alongside view.go during migration:

- GIVEN the project before migration
- WHEN the tilemap package is created
- THEN both `internal/ui/tilemap/` and `internal/ui/view.go` MUST exist

#### Scenario: Parallel rendering coexisting

- GIVEN tilemap package with Tile, Tilemap, Camera types
- WHEN view.go functions are called
- THEN no errors occur (parallel packages work independently)

### Requirement: Biome Styles Migration

The existing `biomeStyles` map MUST be incorporated into the tilemap data layer:

- GIVEN biomeStyles map in view.go with entries: ocean, plains, forest, desert, tundra, jungle, unknown
- WHEN tilemap data.md spec is implemented
- THEN all 7 biome entries MUST exist in the tilemap with same characters and colors

#### Scenario: Biome colors preserved

- GIVEN biomeStyles["plains"] has color #90EE90
- WHEN tilemap renders terrain at a plains tile
- THEN the color MUST be #90EE90

### Requirement: Layer Priority Migration

The current overlay priority (NPC > Building > Popup > Biome) MUST become explicit layer ordering:

| Old Priority | New Layer |
|--------------|-----------|
| NPC | LayerCreature (3) |
| Building | LayerBuilding (1) |
| Popup | LayerUI (4) |
| Biome | LayerTerrain (0) |

- GIVEN current renderOverlay with priority: NPC > Settlement > Biome
- WHEN migrated to tilemap
- THEN Layer 3 (Creature) > Layer 1 (Building) > Layer 0 (Terrain)

#### Scenario: Old priority matches new layers

- GIVEN NPC and settlement at same position
- WHEN tilemap renders
- THEN NPC character MUST appear (higher layer)

### Requirement: renderMap to Tilemap Migration

The `renderMap` function MUST be reimplemented using tilemap:

- GIVEN current view.go renderMap function
- WHEN migration is complete
- THEN `renderMap` in model.go MUST use Tilemap.TileAt() and Camera viewport

#### Scenario: Map renders correctly after migration

- GIVEN world map with tiles at various positions
- WHEN tilemap renders the viewport
- THEN the output MUST match the current renderMap output (visual regression)

### Requirement: Settlement View Integration as Z=1

The `renderSettlementView` function MUST become Z=1 (interior) view:

- GIVEN current renderSettlementView showing settlement interior
- WHEN camera Z=1
- THEN interior tiles MUST render with building interiors at layer 1

#### Scenario: Settlement view uses Z-level

- GIVEN Camera.Z = 1 viewing building interior
- WHEN tilemap renders
- THEN layer 1 shows interior grid (rooms, corridors, doors)

### Requirement: Migration Strategy - Parallel Then Swap

The migration MUST follow this sequence:

1. Create `internal/ui/tilemap/` package (parallel)
2. Implement Tile, Tilemap, Camera types
3. Port world map rendering to tilemap pipeline
4. Integrate settlement view as Z=1
5. Update model.go to use tilemap package
6. Remove old renderer (view.go)

- GIVEN migration phase 1 complete
- WHEN model.go uses tilemap for map view
- THEN view.go functions may be deprecated but still present

#### Scenario: Migration can be rolled back

- GIVEN migration causes TUI breakage
- WHEN model.go reverts to original view.go functions
- THEN the TUI MUST display correctly (rollback works)

### Requirement: Overlay Maps Replacement

Current overlay slices (`m.npcOverlay`, `m.settlementOverlay`, `m.rewardPopups`) MUST be replaced by tilemap layers:

- GIVEN NPC overlay slice in model
- WHEN tilemap is implemented
- THEN NPC positions MUST be stored in Tilemap.Tiles[worldY][worldX].Creature

#### Scenario: Overlay data populates tilemap

- GIVEN npcOverlay slice with entries
- WHEN tilemap is populated
- THEN each NPC position in tilemap MUST have Creature layer set