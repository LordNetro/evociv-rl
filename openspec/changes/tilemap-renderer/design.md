# Design: tilemap-renderer

## Technical Approach

Replace the current single-layer biome renderer in `view.go` with a layered ASCII tilemap system supporting multi-tile buildings (rectangular footprints) and 2 Z-levels (surface + building interiors). The implementation follows the phased migration strategy from the proposal: create `internal/ui/tilemap/` as a parallel package, then swap in the TUI model.

## Architecture Decisions

### Decision: Layer Enum and Priority

**Choice**: Explicit Layer enum with 5 values (0=Terrain through 4=UI)
**Alternatives considered**: Boolean flags per layer, bitmask layering
**Rationale**: Matches current overlay priority (NPC > Building > Biome) and provides explicit ordering for multi-tile buildings. The enum enables type-safe layer indexing and matches the existing renderOverlay precedence.

### Decision: Tilemap Storage Structure

**Choice**: `map[int][][]Tile` for Z-levels (key is Z level, value is 2D tile grid)
**Alternatives considered**: Single flat slice with stride, `[][][]Tile` (Z-first), separate Tilemap per level
**Rationale**: Enables sparse Z-level storage (only populate levels when needed) and matches camera.Z transition pattern in existing settlement view. The current code already uses overlay maps—tilemap stores equivalent data in-layer.

### Decision: Building Footprint Rendering

**Choice**: Render corners as '+', horizontal/vertical edges as '#', interior as '.' (floor)
**Alternatives considered**: Uniform building character, ASCII box-drawing characters
**Rationale**: Matches existing `biomeStyles` visual vocabulary. The corner/edge differentiation provides clear building boundary visualization at a glance.

### Decision: Z-Level Transition Method

**Choice**: `Camera.SetZLevel(z int) error` with error for Z<0 or Z>1
**Alternatives considered**: No error (clamp to valid range), panic
**Rationale**: Error return makes bug detection explicit during migration and matches Go idioms for invalid state transitions.

## Data Flow

```
WorldMap + WorldState
        │
        ├─► TileBuilder.BuildFromWorld()
        │         │
        │         ▼
        │   Tilemap (layers populated)
        │         │
        │         ▼
        │   Camera.Viewport()
        │         │
        │         ▼
        │   Renderer.RenderLayer() × 5
        │         │
        │         ▼
        │   Buffer (string grid with lipgloss.Style)
        │         │
        ▼         ▼
  renderMap() ←─► renderSettlementView()
        │              │
        └──────────────┘
                 │
                 ▼
           BubbleTea View
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/ui/tilemap/tile.go` | Create | Tile struct with 5 layer fields, Layer enum, CellType |
| `internal/ui/tilemap/map.go` | Create | Tilemap type with Z-level storage, TileAt/SetTile methods |
| `internal/ui/tilemap/camera.go` | Create | Camera with X,Y,Z, viewport, SetZLevel, bounds clamping |
| `internal/ui/tilemap/renderer.go` | Create | RenderLayer pipeline, lipgloss styling per layer |
| `internal/ui/tilemap/builder.go` | Create | TileBuilder, populate Tilemap from WorldMap + WorldState |
| `internal/ui/tilemap/interior.go` | Create | InteriorRenderer for Z=1 building interiors |
| `internal/ui/model.go` | Modify | Add Tilemap field, camera state for Z-levels |
| `internal/ui/view.go` | Modify | Deprecate or delegate to tilemap package |
| `internal/ui/tilemap/style.go` | Create | lipgloss.Style mappings (biomeStyles reimplemented) |

## Interfaces / Contracts

```go
// tile.go
type Layer int

const (
    LayerTerrain Layer = iota
    LayerBuilding
    LayerItem
    LayerCreature
    LayerUI
)

type Tile struct {
    Terrain   rune
    Building rune
    Item     rune
    Creature rune
    Fog      rune
}

// map.go
type Tilemap struct {
    Width   int
    Height  int
    ZLevels int          // always 2
    levels  map[int][][]Tile
}

func (t *Tilemap) TileAt(x, y, z int) *Tile
func (t *Tilemap) SetTile(x, y, z int, layer Layer, char rune)

// camera.go
type Camera struct {
    X, Y, Z int
    Width, Height int
}

func (c *Camera) Viewport(t *Tilemap) [][]*Tile
func (c *Camera) SetZLevel(z int) error
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Tile struct, Layer enum, Tilemap methods | Tilemap tests with world bounds |
| Unit | Camera viewport, bounds clamping | Camera with fixed dimensions |
| Unit | Building footprint corner/edge logic | BuildingRender tests |
| Integration | TileBuilder from WorldMap | Mock WorldMap with tile data |
| E2E | End-to-end render pipeline | Visual diff against view.go output |

## Migration / Rollout

Phase 1: Create `internal/ui/tilemap/` package alongside view.go ✓  
Phase 2: Implement `Tile`, `Tilemap`, `Camera` with unit tests ✓  
Phase 3: Add `RenderToBuffer()` for BubbleTea integration ✓  
Phase 4: Swap in model.go — wrap TilemapRenderer behind existing View interface ✓  
Phase 5: Remove old `biomeStyles` rendering code from view.go ✓  

Rollback: Revert model.go to call original `renderMap()` / `renderSettlementView()` — tilemap package stays as unused code.

## Open Questions

- [ ] Should interior tiles use BuildingInterior ECS data or generate from Building footprint?
- [ ] Do we need Z=1 population for building interiors, or is grid-based rendering sufficient?
- [ ] Should Tilemap store fog state separately or per-tile in the Fog layer?