# Data Specification - Tile Character Mappings

## Purpose

Define ASCII character mappings and layer ordering rules for the tilemap renderer.

## Requirements

### Requirement: Terrain Layer Characters

The system MUST render terrain tiles using the following ASCII characters:

| Biome ID | Character | Color | Description |
|----------|-----------|-------|-------------|
| ocean | `~` | #1E90FF | Water |
| plains | `.` | #90EE90 | Grassland |
| forest | `T` | #228B22 | Trees |
| desert | `d` | #EDC9AF | Sand |
| tundra | `*` | #E0FFFF | Snow/Ice |
| jungle | `J` | #006400 | Dense forest |
| unknown | `?` | #888888 | Undefined |

#### Scenario: Ocean renders as blue tilde

- GIVEN a tile with biomeID "ocean" at layer 0
- WHEN the tilemap renders
- THEN the character MUST be `~` with color #1E90FF

#### Scenario: Forest renders as green T

- GIVEN a tile with biomeID "forest" at layer 0
- WHEN the tilemap renders
- THEN the character MUST be `T` with color #228B22

### Requirement: Building Layer Characters

Buildings MUST render based on footprint position:

| Position | Character | Color | Description |
|----------|-----------|-------|-------------|
| corner | `+` | #8B4513 | Building corners |
| edge | `#` | #A0522D | Building walls |
| interior | `.` | #90EE90 | Walkable floor |

#### Scenario: Multi-tile building renders with corners and edges

- GIVEN a building with footprint (x:5, y:5, w:3, h:2)
- WHEN the tilemap renders the building layer
- THEN corners at (5,5), (7,5), (5,6), (7,6) MUST use `+`
- AND edges at (6,5), (6,6) MUST use `#`
- AND interior at (6,5) within building MUST use `.`

### Requirement: Item Layer Characters

Items MUST render using these ASCII characters:

| Item Type | Character | Color | Description |
|-----------|-----------|-------|-------------|
| generic | `*` | #FFD700 | Generic item |
| tool | `t` | #C0C0C0 | Tool/equipment |
| food | `%` | #FF8C00 | Food item |

#### Scenario: Item renders over terrain

- GIVEN a tile with terrain "plains" and item "food"
- WHEN the tilemap renders layers 0-2
- THEN layer 1 (terrain) shows `.`
- AND layer 2 (item) shows `%` at the same position

### Requirement: Creature Layer Characters

Creatures MUST render using these ASCII characters:

| Creature Type | Character | Color | Description |
|---------------|-----------|-------|-------------|
| NPC/human | `@` | #FF6347 | Player or worker |
| fish | `f` | #00CED1 | Aquatic creature |
| wolf | `w` | #808080 | Predator |

#### Scenario: NPC renders over building

- GIVEN a tile with building and NPC at the same world position
- WHEN the tilemap renders layers 0-3
- THEN layer 3 (creature) MUST show `@` (NPC takes priority)

### Requirement: UI Layer Characters

UI elements MUST use these ASCII characters:

| Element | Character | Color | Description |
|---------|-----------|-------|-------------|
| cursor | `+` | #FFD700 | Selection cursor |
| border | `│` `─` `┌` `┐` `└` `┘` | #7D56F4 | Panel borders |

### Requirement: Layer Ordering Rules

The system MUST render layers in ascending order (0→4):

1. Layer 0 (Terrain): Base biome character
2. Layer 1 (Buildings): Overwrites terrain at building positions
3. Layer 2 (Items): Overwrites terrain/building
4. Layer 3 (Creatures): Overwrites all previous layers
5. Layer 4 (UI): Overwrites all previous layers

#### Scenario: Layer priority - creature over item over building

- GIVEN a tile with building, item, and creature
- WHEN all 5 layers are rendered
- THEN the visible character MUST be from layer 3 (creature)

### Requirement: Fog of War Display Rules

Unexplored tiles MUST display with fog overlay:

| State | Character | Color | Description |
|-------|-----------|-------|-------------|
| explored | (base terrain) | (base color) | Normal render |
| unseen | ` ` | #333333 | Never visited |
| fog | `:` | #555555 | Visited but not current |

#### Scenario: Unexplored area shows fog

- GIVEN a world position that has never been explored
- WHEN the tilemap renders that tile
- THEN the character MUST be ` ` (space) with color #333333