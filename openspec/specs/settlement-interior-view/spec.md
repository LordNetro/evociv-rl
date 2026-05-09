# settlement-interior-view Specification

## Purpose

The "settlement" screen — a full-screen interior view of a settlement showing buildings, NPCs, and reward popups within a dynamic viewport. Navigable with arrow keys, inspectable with 'e', closable with 'esc'/'q'.

## Requirements

### Requirement: Settlement Screen Mode

The model MUST support a `"settlement"` screen value. Pressing `'e'` on a settlement tile in `"map"` screen MUST transition to `"settlement"` and set the settlement entity as the active context.

#### Scenario: Enter settlement from map

- GIVEN the model is in `"map"` screen, cursor over a settlement tile with `inspectorOpen = false`
- WHEN `'e'` is pressed
- THEN the model screen MUST change to `"settlement"` and the matching `Settlement` entity MUST be stored

#### Scenario: `'e'` on non-settlement tile does nothing

- GIVEN the model is in `"map"` screen, cursor over a biome tile with no settlement
- WHEN `'e'` is pressed
- THEN the screen MUST remain `"map"` (no-op)

#### Scenario: NPC inspector takes priority over settlement entry

- GIVEN cursor over a tile with both an NPC and a settlement
- WHEN `'e'` is pressed
- THEN the NPC inspector SHALL open, NOT the settlement interior view

### Requirement: Exit Settlement Screen

Pressing `'q'` or `'esc'` in `"settlement"` screen MUST return to `"map"` screen.

#### Scenario: Exit with q

- GIVEN the model is in `"settlement"` screen
- WHEN `'q'` is pressed
- THEN the screen MUST return to `"map"` and the active settlement context MUST be cleared

#### Scenario: Exit with esc

- GIVEN the model is in `"settlement"` screen
- WHEN `'esc'` is pressed
- THEN the screen MUST return to `"map"` and the active settlement context MUST be cleared

### Requirement: Dynamic Viewport

The settlement interior viewport MUST be centered on the settlement's world position. The viewport side length MUST be `2 * settlement.Radius + 1` (e.g., village with Radius 3 → 7×7, town Radius 5 → 11×11, city Radius 8 → 17×17). The viewport height MUST be clamped to available terminal height.

#### Scenario: Village renders 7×7 viewport

- GIVEN a village with Radius = 3
- WHEN the settlement view renders
- THEN the visible grid MUST be 7 tiles wide and 7 tiles tall (or less if terminal is smaller)

#### Scenario: City renders 17×17 viewport

- GIVEN a city with Radius = 8
- WHEN the settlement view renders
- THEN the visible grid MUST be 17 tiles wide and 17 tiles tall (or less if terminal is smaller)

#### Scenario: Terminal smaller than viewport

- GIVEN a village with Radius = 3 (7×7) but terminal height = 5 rows
- WHEN the settlement view renders
- THEN only 5 rows SHALL be visible; content above/below SHALL be clipped

### Requirement: Building Grid Rendering

All Building entities whose `SettlementEntity` matches the active settlement MUST be rendered as their `InteriorSymbol` with their `Color` at positions relative to the settlement center. Buildings MUST be rendered as a background layer before NPCs.

#### Scenario: Building renders at correct relative position

- GIVEN a settlement at world (10, 10) and a farm building at world (8, 9)
- WHEN the settlement viewport renders centered at (10, 10)
- THEN the farm MUST appear at viewport offset row = -1, col = -2 with symbol `╬` and color `#DEB887`

#### Scenario: Building outside viewport clipped

- GIVEN a village (Radius 3) and a building at Chebyshev distance 4 from center
- WHEN the viewport renders
- THEN the building MUST NOT appear (outside the 7×7 grid)

#### Scenario: Multiple buildings render at distinct positions

- GIVEN two buildings at different positions within the settlement radius
- WHEN the viewport renders
- THEN both buildings MUST appear at their respective viewport positions

### Requirement: NPC Rendering in Interior

NPCs with `HomeReference.SettlementEntity` matching the active settlement MUST be rendered on top of buildings using their `Symbol` and `Color`. This overrides the LOD system — all home settlement NPCs SHALL be visible regardless of LOD level.

#### Scenario: NPC appears over building

- GIVEN a settlement with a house (`⌂`) at viewport (3, 3) and an NPC at the same world position
- WHEN the viewport renders
- THEN the NPC symbol MUST be visible at (3, 3), NOT the building symbol

#### Scenario: NPC outside settlement not shown

- GIVEN an NPC whose HomeReference does NOT match the active settlement
- WHEN the viewport renders
- THEN that NPC MUST NOT appear in the viewport

#### Scenario: NPC moves within viewport

- GIVEN an NPC inside the settlement moving from (2, 2) to (3, 3)
- WHEN the viewport renders after the move
- THEN the NPC MUST appear at (3, 3) instead of (2, 2)

### Requirement: Cursor Navigation

Arrow keys (up/down/left/right) MUST move a cursor within the settlement viewport. The cursor MUST be constrained to the viewport bounds.

#### Scenario: Arrow keys move cursor

- GIVEN cursor at viewport position (0, 0) in settlement view
- WHEN `right` is pressed
- THEN cursor MUST move to (1, 0)

#### Scenario: Cursor clamps at viewport edge

- GIVEN cursor at viewport position (6, 3) and viewport width = 7
- WHEN `right` is pressed
- THEN cursor MUST stay at (6, 3)

### Requirement: Reward Popups

When an NPC's `LastReward` is non-zero and the current tick is within 5 ticks of `RewardTick`, a floating text `+{reward}` MUST be rendered near the NPC's position in the viewport. The popup MUST fade (not render) after 5 ticks. Maximum 3 concurrent popups.

#### Scenario: Reward popup appears near NPC

- GIVEN an NPC with LastReward = 0.87, RewardTick = 100, and current tick = 101
- WHEN the viewport renders
- THEN the text `+0.87` MUST appear near the NPC's viewport position

#### Scenario: Reward popup fades after 5 ticks

- GIVEN an NPC with RewardTick = 100 and current tick = 106
- WHEN the viewport renders
- THEN the reward popup MUST NOT appear

#### Scenario: Maximum concurrent popups

- GIVEN 5 NPCs with recent rewards within 5 ticks
- WHEN the viewport renders
- THEN at most 3 popups SHALL be shown (closest 3 by distance to cursor)

#### Scenario: Minimum reward threshold

- GIVEN an NPC with LastReward = 0.05
- WHEN the viewport renders
- THEN the reward popup MUST NOT appear (below threshold 0.1)

### Validation Rules

- PASS: Enter settlement from map via `'e'` on settlement tile
- PASS: Exit settlement via `'q'` or `'esc'`
- PASS: Buildings render at correct viewport-relative positions
- PASS: NPCs from the settlement appear regardless of LOD
- PASS: Cursor navigates within viewport bounds
- PASS: Reward popups appear and fade correctly
- PASS: Max 3 concurrent popups enforced
- FAIL: Entering settlement from non-settlement tile
- FAIL: NPC from other settlement appears in viewport
