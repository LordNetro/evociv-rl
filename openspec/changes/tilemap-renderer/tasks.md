# Tasks: tilemap-renderer

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~800-1000 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | Phase 1-3: Core types & Camera → Phase 4-5: Builder & Interior → Phase 6: Integration |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Core types (Tile, Layer, CellType) + Tilemap + Camera | PR 1 | Base foundation; ~300 lines |
| 2 | TileBuilder + InteriorRenderer + Fog | PR 2 | World integration; ~250 lines |
| 3 | Renderer pipeline + migration to model.go/view.go | PR 3 | Final integration; ~250 lines |

## Phase 1: Core Types (Foundation)

- [x] 1.1 **PHASE**: Tile struct with 5 layer fields
  - File: `internal/ui/tilemap/tile.go`
  - Description: Create Tile struct with Terrain, Building, Item, Creature, Fog rune fields

- [x] 1.2 **PHASE**: Layer enum constants
  - File: `internal/ui/tilemap/tile.go`
  - Description: Define Layer type with iota: Terrain=0, Building=1, Item=2, Creature=3, UI=4

- [x] 1.3 **PHASE**: CellType enum for building footprints
  - File: `internal/ui/tilemap/tile.go`
  - Description: Define CellType: Floor, Wall, Door, Corridor, Corner

## Phase 2: Tilemap Type

- [x] 2.1 **PHASE**: Tilemap struct and NewTilemap constructor
  - File: `internal/ui/tilemap/map.go`
  - Description: Create Tilemap with Width, Height, ZLevels=2, levels map[int][][]Tile

- [x] 2.2 **PHASE**: TileAt method
  - File: `internal/ui/tilemap/map.go`
  - Description: Safe tile access, returns nil for out-of-bounds

- [x] 2.3 **PHASE**: SetTile method with layer parameter
  - File: `internal/ui/tilemap/map.go`
  - Description: Set character at tile position for specific layer

- [x] 2.4 **PHASE**: Z-level support (SetZLevel)
  - File: `internal/ui/tilemap/map.go`
  - Description: Create or access Z-level slices

## Phase 3: Camera

- [x] 3.1 **PHASE**: Camera struct
  - File: `internal/ui/tilemap/camera.go`
  - Description: Camera with X, Y, Z, Width, Height; default Z=0

- [x] 3.2 **PHASE**: Viewport method with bounds clamping
  - File: `internal/ui/tilemap/camera.go`
  - Description: Returns [][]*Tile for visible area

- [x] 3.3 **PHASE**: Camera.SetZLevel with validation
  - File: `internal/ui/tilemap/camera.go`
  - Description: Error for Z<0 or Z>1

- [x] 3.4 **PHASE**: CenterOn method with clamping
  - File: `internal/ui/tilemap/camera.go`
  - Description: Camera moves to center on coordinates

## Phase 4: Tilemap Builder (World Integration)

- [x] 4.1 **PHASE**: TileBuilder.BuildFromWorld
  - File: `internal/ui/tilemap/builder.go`
  - Description: Populate terrain layer from WorldMap biome data

- [x] 4.2 **PHASE**: NPC rendering in Creature layer
  - File: `internal/ui/tilemap/builder.go`
  - Description: Map NPC entities to '@' symbol

- [x] 4.3 **PHASE**: Multi-tile building footprint rendering
  - File: `internal/ui/tilemap/builder.go`
  - Description: Render corners as '+', edges as '#', interior as '.'

- [x] 4.4 **PHASE**: Fog of war layer
  - File: `internal/ui/tilemap/builder.go`
  - Description: Mark explored vs visible tiles

## Phase 5: Interior Renderer (Z=1)

- [x] 5.1 **PHASE**: InteriorRenderer with building ID
  - File: `internal/ui/tilemap/interior.go`
  - Description: Render BuildingInterior grid at Z=1

- [x] 5.2 **PHASE**: NPC placement in interior
  - File: `internal/ui/tilemap/interior.go`
  - Description: Position NPCs from AIState into Z=1 Creature layer

- [x] 5.3 **PHASE**: Workers count overlay
  - File: `internal/ui/tilemap/interior.go`
  - Description: Show workers count in UI layer

## Phase 6: Renderer & Migration

- [x] 6.1 **PHASE**: RenderLayer with lipgloss styles
  - File: `internal/ui/tilemap/renderer.go`
  - Description: Render single layer to string with color

- [x] 6.2 **PHASE**: RenderAll (multi-layer composition)
  - File: `internal/ui/tilemap/renderer.go`
  - Description: Compose layers 0→4 with priority

- [x] 6.3 **PHASE**: Integrate into BubbleTea View
  - File: `internal/ui/model.go`
  - Description: Add Tilemap field, update View() to use renderer

- [x] 6.4 **PHASE**: Migrate view.go to tilemap
  - File: `internal/ui/view.go`
  - Description: Delegate to tilemap package

- [x] 6.5 **PHASE**: Remove old biomeStyles code
  - File: `internal/ui/view.go`
  - Description: Clean up replaced rendering functions