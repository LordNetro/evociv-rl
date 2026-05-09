# Camera Specification - Viewport Behavior

## Purpose

Define camera behavior for viewport rendering, Z-level switching, and world bounds clamping.

## Requirements

### Requirement: Camera Position Tracks World Coordinates

The camera X and Y MUST represent the world position of the viewport's top-left corner:

```go
func (c *Camera) ViewportTopLeft() (x, y int)
```

- GIVEN Camera with X=50, Y=30
- WHEN ViewportTopLeft is called
- THEN it MUST return (50, 30)

#### Scenario: Camera scrolls with movement

- GIVEN Camera at (50, 30) viewing world (100, 80)
- WHEN camera moves to (55, 35)
- THEN the new top-left corner MUST be (55, 35)

### Requirement: Viewport Renders Visible Tiles

The camera MUST render exactly Width×Height tiles from the camera position:

- GIVEN Camera with Width=80, Height=24 and world position (50, 30)
- WHEN tiles are rendered for the viewport
- THEN exactly 80×24 = 1920 tiles MUST be processed

#### Scenario: Viewport shows correct world coordinates

- GIVEN Camera at X=10, Y=5, Width=5, Height=3
- WHEN tile at viewport column 2, row 1 is rendered
- THEN it MUST correspond to world position (12, 6)

### Requirement: Z-Level Switching

The camera Z field MUST support switching between surface (Z=0) and interior (Z=1):

```go
func (c *Camera) SetZLevel(z int) error
```

- GIVEN Camera at Z=0
- WHEN SetZLevel(1) is called
- THEN camera Z MUST be 1

#### Scenario: Surface to interior transition

- GIVEN Camera viewing world map at Z=0
- WHEN player enters a building
- THEN camera Z MUST transition to 1 (interior view)

#### Scenario: Interior to surface transition

- GIVEN Camera viewing building interior at Z=1
- WHEN player exits the building
- THEN camera Z MUST transition to 0 (world map)

#### Scenario: Z-level restricted to 0-1

- GIVEN Camera at Z=0
- WHEN SetZLevel(2) is called
- THEN an error MUST be returned (only Z=0, Z=1 allowed)

### Requirement: Camera Clamping to World Bounds

The camera position MUST be clamped to prevent rendering outside world bounds:

- GIVEN world dimensions Width=200, Height=100 and Camera Width=80, Height=24
- WHEN camera X is set to 150
- THEN camera X MUST be clamped to max(0, 200-80) = 120

#### Scenario: Camera clamped at left boundary

- GIVEN Camera with X = -10
- WHEN rendering occurs
- THEN X MUST be clamped to 0

#### Scenario: Camera clamped at top boundary

- GIVEN Camera with Y = -5
- WHEN rendering occurs
- THEN Y MUST be clamped to 0

#### Scenario: Camera clamped at right boundary

- GIVEN world Width=100, Camera Width=30, Camera X=80
- WHEN rendering occurs
- THEN X MUST be clamped to 70

### Requirement: Camera Viewport Culling

Tiles outside the viewport MUST NOT be rendered:

- GIVEN Camera at (50, 30) with Width=80, Height=24
- WHEN tile at world position (20, 10) is evaluated
- THEN it MUST be skipped (not rendered)

#### Scenario: Tile at viewport edge is rendered

- GIVEN Camera at (50, 30), Width=80, Height=24
- WHEN tile at world position (50, 30) is evaluated
- THEN it MUST be rendered (first tile in viewport)

#### Scenario: Tile just outside viewport not rendered

- GIVEN Camera at (50, 30), Width=80, Height=24
- WHEN tile at world position (49, 30) is evaluated
- THEN it MUST be skipped (column -1 from viewport)