# settlement-tui Specification

## Purpose

Renderizado de asentamientos en el mapa TUI con símbolos especiales, nombres visibles, orden de overlay e inspector de datos.

## Requirements

### Requirement: Settlement Symbols

The map view MUST render settlements using their defined symbols: `♦` for village, `▲` for town, `●` for city. The symbol color MUST match the settlement type's color definition from settlements.yaml.

#### Scenario: Village renders as green ♦

- GIVEN a map view with biome tiles rendered and a village at world position (10, 15)
- WHEN the map view renders
- THEN a green `♦` character MUST appear at the corresponding screen position

#### Scenario: City renders as red ●

- GIVEN a city at world position (5, 8)
- WHEN the map view renders
- THEN a red `●` character MUST appear at the corresponding screen position

### Requirement: Settlement Name Visibility

The settlement name MUST be displayed adjacent to its symbol (e.g., "♦ Aldea del Norte") when the player's cursor is within the settlement's radius. The name MUST be visible as a label near the settlement symbol, truncated to 16 characters if longer.

#### Scenario: Settlement name visible on cursor proximity

- GIVEN a settlement "Aldea del Norte" with radius 6 centered at (10, 10)
- WHEN the player cursor is at (12, 10) (within radius)
- THEN the label "♦ Aldea del Norte" MUST appear near the symbol

#### Scenario: Settlement name truncated

- GIVEN a settlement named "La Gran Ciudad del Puerto del Sur" (33 chars)
- WHEN its name is displayed
- THEN it MUST be truncated to 16 characters: "La Gran Ciudad d..."

### Requirement: Overlay Rendering Order

When multiple entities occupy the same screen position, the rendering order MUST be: NPC symbol > Settlement symbol > Biome tile character. A settlement MUST NOT occlude an NPC, and a biome tile MUST NOT occlude a settlement.

#### Scenario: NPC appears over settlement symbol

- GIVEN a settlement and an NPC at the same world position
- WHEN the map view renders
- THEN the NPC symbol (`@`) MUST be visible, and the settlement symbol MUST NOT appear at that exact cell

#### Scenario: Settlement appears over biome tile

- GIVEN a settlement tile with no NPC present
- WHEN the map view renders
- THEN the settlement symbol MUST be visible instead of the biome tile character

### Requirement: Settlement Inspector Panel

Pressing 'e' over a settlement MUST open an inspector panel showing: settlement Name, Type, Population, Level, and a list of building names within the settlement. If both an NPC and a settlement share the cursor position, the NPC inspector MUST take priority.

#### Scenario: Inspector shows settlement details

- GIVEN a settlement "Aldea Verde" of type "village" with Population 12, Level 1, and buildings [house, farm]
- WHEN the player cursor is on the settlement and 'e' is pressed
- THEN an inspector panel MUST show Name "Aldea Verde", Type "village", Population 12, Level 1, and buildings "house, farm"

#### Scenario: NPC inspector takes priority over settlement

- GIVEN cursor over a tile with both an NPC and a settlement
- WHEN 'e' is pressed
- THEN the NPC inspector SHALL be shown, not the settlement inspector

#### Scenario: No entity under cursor shows nothing

- GIVEN cursor over an empty biome tile
- WHEN 'e' is pressed
- THEN no inspector SHALL appear (no-op or empty-state message)
