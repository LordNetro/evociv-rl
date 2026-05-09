# Rendering Specification - Layer Pipeline

## Purpose

Define the layer rendering pipeline with viewport culling, multi-tile buildings, and UI overlays.

## Requirements

### Requirement: Render Layers in Order 0→4

The renderer MUST process and render layers in ascending order:

1. Layer 0 (Terrain) - base biome
2. Layer 1 (Buildings) - building footprint
3. Layer 2 (Items) - item characters
4. Layer 3 (Creatures) - NPC/animals
5. Layer 4 (UI) - cursor/overlays

- GIVEN a Tilemap with all layers populated at position (10, 10)
- WHEN the renderer produces output for that position
- THEN the final character MUST be from the highest non-empty layer

#### Scenario: Empty layers fall through

- GIVEN a tile where LayerBuilding=0, LayerItem=0, LayerCreature=0
- WHEN rendering layer 1
- THEN Layer 0 (Terrain) character MUST be visible

### Requirement: Camera Viewport Culling

The renderer MUST only process tiles visible in the camera viewport:

- GIVEN Camera at X=50, Y=30, Width=80, Height=24
- WHEN rendering proceeds
- THEN only tiles in world range X:[50,129], Y:[30,53] MUST be processed

#### Scenario: Culling skips out-of-view tiles

- GIVEN Camera at (50, 30) with Width=80
- WHEN tile at world position (10, 10) is checked
- THEN it MUST be skipped (outside viewport X range)

### Requirement: Multi-Tile Building Rendering

Buildings MUST render with correct footprint: corners, edges, interior:

- GIVEN building footprint at (10,10) with Width=3, Height=2
- WHEN layer 1 renders
- THEN corner positions (10,10), (12,10), (10,11), (12,11) MUST show '+'
- AND edge positions (11,10), (11,11) MUST show '#'
- AND interior position (11,10) MUST show '.'

#### Scenario: Adjacent buildings don't merge

- GIVEN building A at (10,10) with size 2×2 and building B at (13,10) with size 2×2
- WHEN both render
- THEN there MUST be at least 1 tile of separation between them

### Requirement: Creature Rendering

Creatures MUST render at their world positions:

- GIVEN NPC at world position (15, 20) with symbol '@'
- WHEN layer 3 renders
- THEN the character '@' MUST appear at viewport position corresponding to (15, 20)

#### Scenario: Multiple creatures same position shows one

- GIVEN two NPCs at the same world position (15, 20)
- WHEN layer 3 renders
- THEN only one '@' character MUST appear (first in render order)

### Requirement: Creature Type Mapping

Different creature types MUST render with distinct characters:

| Type | Symbol | Color |
|------|--------|-------|
| NPC/human | @ | #FF6347 |
| fish | f | #00CED1 |
| wolf | w | #808080 |

- GIVEN a fish creature at (10, 10)
- WHEN layer 3 renders
- THEN character 'f' with color #00CED1 MUST appear

### Requirement: UI Layer Rendering

UI elements MUST overlay all other layers:

- GIVEN cursor at viewport position (5, 10)
- WHEN layer 4 renders
- THEN character '+' with gold background MUST appear at that position

#### Scenario: UI cursor overlays creature

- GIVEN creature '@' at world position and cursor at same viewport position
- WHEN all 5 layers render
- THEN the cursor character MUST be visible (layer 4)

### Requirement: Fog of War Integration

Unexplored areas MUST show fog overlay:

- GIVEN tile at world position (50, 50) that is unexplored
- WHEN layers render
- THEN layer 4 (Fog) MUST show either ' ' (unseen) or ':' (fog)

#### Scenario: Explored area shows base layers

- GIVEN tile at world position (50, 50) that was explored
- WHEN layers render
- THEN fog MUST NOT obscure the terrain/building/item/creature layers