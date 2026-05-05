# npc-tui Specification

## Description

Extensión del TUI de mapa para mostrar NPCs como overlay '@' y un panel inspector modal que se abre con la tecla 'e'.

## Requirements

### Requirement: NPC Overlay on Map

The map view MUST display NPC symbols over biome tiles. The NPC symbol (default '@') MUST be rendered at the NPC's world position, offset by camera, on top of the underlying biome tile character.

#### Scenario: '@' appears on map tile

- GIVEN a map view with biome tiles rendered
- WHEN an NPC exists at world position (10, 15) and camera offset is (0, 0)
- THEN the character '@' MUST be visible at screen position (10, 15)

#### Scenario: Camera offset moves NPC on screen

- GIVEN an NPC at world position (15, 20) and camera at (10, 10)
- WHEN the map view renders
- THEN the '@' MUST appear at screen position (5, 10)

### Requirement: Inspector Panel on 'e'

Pressing the 'e' key MUST open an inspector panel showing the NPC under the player's cursor. The panel MUST display: Name, Health (Current/Max), Job, Personality (O/C/E/A/N), and current Biome.

#### Scenario: Inspector shows NPC details

- GIVEN the player cursor is over a tile containing an NPC named "Gorim" with Health 80/100 and Job "farmer"
- WHEN the 'e' key is pressed
- THEN an inspector panel MUST appear showing Name "Gorim", Health "80/100", Job "farmer", all five Personality traits rounded to 2 decimals, and the tile's biome name

#### Scenario: No NPC under cursor shows nothing

- GIVEN the player cursor is over an empty tile
- WHEN the 'e' key is pressed
- THEN no inspector panel SHALL appear (MUST be a no-op or show an empty-state message)

### Requirement: Close Inspector Panel

The inspector panel MUST close when the player presses 'q' or 'esc'. Closing MUST return to the normal map view.

#### Scenario: Close with 'q'

- GIVEN the inspector panel is open
- WHEN 'q' is pressed
- THEN the panel MUST close and the map view MUST be restored

#### Scenario: Close with 'esc'

- GIVEN the inspector panel is open
- WHEN 'esc' is pressed
- THEN the panel MUST close and the map view MUST be restored
