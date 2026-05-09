# settlement-interior-statusbar Specification

## Purpose

Persistent bottom status bar in the settlement interior view. Shows tile info by default; replaced by inspector content when an inspector is open. Also displays reward activity when NPCs recently received rewards.

## Requirements

### Requirement: Status Bar Always Visible

The settlement interior view MUST render a status bar as the last line of the display. The status bar MUST be visible at all times during settlement view, regardless of cursor position or inspector state.

#### Scenario: Status bar present in settlement view

- GIVEN the model is in `"settlement"` screen with no inspector open
- WHEN the view renders
- THEN the last line of output MUST contain the status bar content

#### Scenario: Status bar uses distinct background

- GIVEN the status bar is rendered
- THEN it MUST use a distinct background color (e.g., dark gray `#333333`) to visually separate it from the viewport content

### Requirement: Default Status Bar Content

When no inspector is open, the status bar MUST show: the tile's relative position within the viewport (e.g., `(3, 5)`), and which entities are under the cursor (building name or NPC name). If nothing is under the cursor, show only the position.

#### Scenario: Status bar shows cursor position and entity

- GIVEN cursor at viewport offset (3, 5) over a building named "Granja" with no NPC
- WHEN no inspector is open
- THEN the status bar MUST contain `(3, 5) Granja` or equivalent

#### Scenario: Status bar shows NPC name over building

- GIVEN cursor at viewport offset (3, 5) with both NPC "Gorim" and building "Granja"
- WHEN no inspector is open
- THEN the status bar MUST show the NPC name "Gorim" (NPC takes display priority)

#### Scenario: Status bar shows only position on empty tile

- GIVEN cursor at viewport offset (1, 1) with no building or NPC
- WHEN no inspector is open
- THEN the status bar MUST show only the position `(1, 1)`

### Requirement: Status Bar Replaced by Inspector

When a building or NPC inspector is open, the status bar content MUST be replaced entirely by the inspector panel content. The inspector panel occupies the same bottom-of-screen area as the status bar.

#### Scenario: Inspector replaces status bar

- GIVEN the building inspector is open
- WHEN the view renders
- THEN the status bar content MUST NOT be visible; the inspector panel content MUST appear in its place

#### Scenario: Closing inspector restores status bar

- GIVEN an inspector is open, then closed via `'q'`
- WHEN the view renders after closing
- THEN the status bar MUST reappear with its default content

### Requirement: Reward Activity in Status Bar

When no inspector is open and an NPC within the settlement recently received a reward (within 3 ticks), the status bar MAY append a summary (e.g., "Rewards: +0.87 Gorim, +0.32 Lidia") after the position info. This MUST NOT exceed the terminal width — truncate if needed.

#### Scenario: Recent reward shown in status bar

- GIVEN NPC "Gorim" with LastReward = 0.87, RewardTick = current_tick
- WHEN the status bar renders with no inspector open
- THEN the status bar MUST append ", +0.87 Gorim" or equivalent

#### Scenario: Stale reward filtered from status bar

- GIVEN NPC "Gorim" with LastReward = 0.87, RewardTick = current_tick - 5
- WHEN the status bar renders
- THEN the reward info MUST NOT appear (reward is stale)

#### Scenario: Status bar truncation

- GIVEN multiple NPCs with rewards causing the status bar to exceed terminal width
- WHEN the status bar renders
- THEN the status bar MUST be truncated to fit the terminal width without wrapping

### Validation Rules

- PASS: Status bar always visible as last line in settlement view
- PASS: Default content shows cursor position + entity under cursor
- PASS: NPC name takes display priority over building name
- PASS: Inspector replaces status bar when open
- PASS: Status bar returns when inspector closes
- PASS: Recent rewards shown, stale rewards filtered
- PASS: Status bar truncated to terminal width
- FAIL: Status bar missing from settlement view
- FAIL: Inspector and status bar visible simultaneously
