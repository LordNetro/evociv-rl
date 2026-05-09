# Types Specification - Core Tilemap Types

## Purpose

Define the core types for the `internal/ui/tilemap/` package.

## Requirements

### Requirement: Tile Struct

The Tile struct MUST contain 5 layer fields representing visual elements:

```go
type Tile struct {
    Terrain   rune  // Layer 0: biome character
    Building  rune  // Layer 1: building footprint
    Item      rune  // Layer 2: item character
    Creature  rune  // Layer 3: NPC/animal
    Fog       rune  // Layer 4: fog of war state
}
```

#### Scenario: Empty tile has zero values

- GIVEN a new Tile struct
- WHEN all fields are inspected
- THEN all rune fields MUST be zero value (0)

#### Scenario: Tile with all layers populated

- GIVEN a Tile with Terrain='T', Building='+', Item='*', Creature='@', Fog=' '
- WHEN each layer is read
- THEN each field MUST return the set value

### Requirement: Layer Enum

The system MUST define a Layer type for explicit layer ordering:

```go
type Layer int

const (
    LayerTerrain Layer = iota  // 0
    LayerBuilding               // 1
    LayerItem                  // 2
    LayerCreature              // 3
    LayerUI                    // 4
)
```

#### Scenario: Layer enum values are sequential

- GIVEN LayerTerrain, LayerBuilding, LayerItem, LayerCreature, LayerUI
- WHEN their int values are compared
- THEN they MUST be 0, 1, 2, 3, 4 respectively

### Requirement: Tilemap Struct

The Tilemap struct MUST contain:

```go
type Tilemap struct {
    Tiles    [][]Tile  // 2D grid of tiles
    Width    int       // World width in tiles
    Height   int       // World height in tiles
    ZLevels  int       // Number of Z-levels (MUST be 2)
}
```

#### Scenario: Tilemap initialized with dimensions

- GIVEN Tilemap initialization with width 100, height 80, zLevels 2
- WHEN Width, Height, and ZLevels are read
- THEN Width MUST be 100, Height MUST be 80, ZLevels MUST be 2

#### Scenario: Tiles 2D slice has correct dimensions

- GIVEN a Tilemap with width W, height H
- WHEN len(Tiles) and len(Tiles[0]) are checked
- THEN they MUST equal H and W respectively

### Requirement: Camera Struct

The Camera struct MUST contain viewport state:

```go
type Camera struct {
    X         int  // World X position (top-left)
    Y         int  // World Y position (top-left)
    Z         int  // Z-level (0 = surface, 1 = interior)
    Width     int  // Viewport width in tiles
    Height    int  // Viewport height in tiles
}
```

#### Scenario: Camera defaults to Z=0

- GIVEN a new Camera struct
- WHEN Z is read
- THEN Z MUST be 0

#### Scenario: Camera viewport matches terminal size

- GIVEN Camera with Width=80, Height=24
- WHEN rendered
- THEN the viewport MUST contain exactly 80×24 tiles

### Requirement: TileAt Method

The Tilemap MUST provide a method to safely access tiles:

```go
func (t *Tilemap) TileAt(x, y int) *Tile
```

- GIVEN a Tilemap with tiles at valid coordinates
- WHEN TileAt is called with those coordinates
- THEN a pointer to the Tile MUST be returned

#### Scenario: TileAt returns nil for out-of-bounds

- GIVEN a Tilemap with Width=100, Height=80
- WHEN TileAt(150, 50) is called
- THEN nil MUST be returned

#### Scenario: TileAt returns nil for negative coordinates

- GIVEN a Tilemap
- WHEN TileAt(-1, 5) is called
- THEN nil MUST be returned

### Requirement: Building Footprint Type

The system MUST define building footprint for multi-tile rendering:

```go
type BuildingFootprint struct {
    X       int  // World X position
    Y       int  // World Y position
    Width   int  // Tile width
    Height  int  // Tile height
    Symbol  rune // Character for building
    Color   string
}
```

#### Scenario: Building footprint defines rectangular area

- GIVEN a BuildingFootprint at (10,10) with Width=4, Height=3
- WHEN all positions in range [10-13] × [10-12] are checked
- THEN they MUST all belong to this building