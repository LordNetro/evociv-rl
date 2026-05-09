# settlement-interior-inspectors Specification

## Purpose

Dual inspector system within the settlement interior view. Building inspector shows construction and economic details with assigned NPCs. NPC inspector shows entity status and Q-learning statistics. NPC inspector takes priority when cursor covers both.

## Requirements

### Requirement: Building Inspector

Pressing `'e'` while cursor is over a building in settlement view MUST open a building inspector panel showing: Name, Level, Role, Workers (current/max), Produces (list of resources with rates), Consumes (list of resources with rates), and a list of assigned NPCs with their last reward value.

#### Scenario: Building inspector shows full details

- GIVEN cursor over a farm building with Level 1, Role "farmer", MaxWorkers 3, producing {food: 2.0}
- WHEN `'e'` is pressed in settlement view
- THEN the inspector panel MUST show: Name "Granja", Level 1, Role "farmer", Workers "0/3", Produces "food: 2.0"

#### Scenario: Building inspector lists assigned NPCs

- GIVEN a farm with 2 assigned NPCs (NPC1 with LastReward 0.5, NPC2 with LastReward 0.3)
- WHEN the building inspector is open
- THEN the panel MUST list both NPCs by name, each with their last reward value

### Requirement: NPC Inspector

Pressing `'e'` while cursor is over an NPC in settlement view MUST open an NPC inspector panel showing: Name, Role, Health (current/max), Home settlement name, Workplace building name (matched by role), and Q-learning stats: Last Reward, Current Policy (ε-greedy), Epsilon (current ε value), and Current State.

#### Scenario: NPC inspector shows full details

- GIVEN a farmer NPC "Gorim" with Health 80/100, HomeReference to settlement "Aldea Verde", AIState{LastReward: 0.87, CurrentAction: "harvest"}, and QTable epsilon = 0.12
- WHEN `'e'` is pressed over this NPC in settlement view
- THEN the inspector MUST show: Name "Gorim", Role "farmer", Health "80/100", Home "Aldea Verde", Workplace building name, Last Reward "0.87", Policy "ε-greedy", Epsilon "0.12", State derived from needs+biome

#### Scenario: NPC inspector when AIState is zero

- GIVEN an NPC with no AIState component (LastReward = 0, no current action)
- WHEN the NPC inspector opens
- THEN Q-learning stats SHALL show "N/A" or zero values without errors

### Requirement: Priority Rule — NPC Over Building

When the cursor covers both a building and an NPC at the same viewport position, pressing `'e'` MUST open the NPC inspector. The building inspector MUST be accessible only when no NPC shares the cursor tile.

#### Scenario: NPC inspector takes priority

- GIVEN cursor over a tile with a farmer NPC and a farm building
- WHEN `'e'` is pressed
- THEN the NPC inspector SHALL open, NOT the building inspector

#### Scenario: Building inspector opens when no NPC

- GIVEN cursor over a building with no NPC on the same tile
- WHEN `'e'` is pressed
- THEN the building inspector SHALL open

### Requirement: Worker Assignment Display

NPCs SHALL be matched to buildings by role: an NPC with `Job.Role` equal to a building's `BuildingDef.Role` is considered assigned to that building. The building inspector MUST compute and display `workers_count / max_workers` using this role-matching.

#### Scenario: Role matching assigns workers

- GIVEN a farm with Role "farmer" and 2 NPCs in the settlement with Job{Role: "farmer"}
- WHEN the building inspector renders
- THEN it MUST show Workers "2/3"

#### Scenario: No workers matched

- GIVEN a blacksmith building with Role "smith" but no NPCs have Job{Role: "smith"}
- WHEN the building inspector renders
- THEN it MUST show Workers "0/2"

### Requirement: Close Inspector in Settlement View

Pressing `'q'` or `'esc'` while an inspector is open in settlement view MUST close the inspector panel and return to the settlement view cursor navigation (NOT exit to map).

#### Scenario: Close inspector stays in settlement

- GIVEN the building inspector is open in settlement view
- WHEN `'q'` is pressed
- THEN the inspector MUST close but the screen MUST remain `"settlement"`

### Validation Rules

- PASS: Building inspector shows all required fields
- PASS: NPC inspector shows all required fields including Q-learning stats
- PASS: NPC inspector takes priority over building inspector
- PASS: Worker assignment computed by role matching
- PASS: Close inspector stays in settlement view
- FAIL: Inspector opens on empty tile (no-op)
- FAIL: NPC with missing AIState doesn't crash inspector
