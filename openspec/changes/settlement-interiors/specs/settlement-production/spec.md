# Delta for settlement-production

## MODIFIED Requirements

### Requirement: Dynamic Production Based on Workers Inside

The settlement production system MUST calculate building output based on workers currently inside the building. The formula SHALL be `base_output × workers_inside × efficiency`.

(Previously: Buildings had fixed base_output independent of worker presence)

#### Scenario: Empty building produces nothing

- GIVEN a settlement with a farm building that has base_output 2.0
- WHEN there are 0 workers inside that farm
- THEN the farm's production MUST be 0.0

#### Scenario: Worker inside produces output

- GIVEN a settlement with a workshop building that has base_output 3.0 and efficiency 1.0
- WHEN there is 1 worker inside that workshop
- THEN the workshop's production MUST be 3.0

#### Scenario: Multiple workers multiply output

- GIVEN a settlement with a farm building that has base_output 2.0 and efficiency 1.0
- WHEN there are 3 workers inside that farm
- THEN the farm's production MUST be 6.0

#### Scenario: Production scaled by efficiency

- GIVEN a settlement with a workshop that has base_output 4.0 and efficiency 0.75
- WHEN there are 2 workers inside
- THEN the workshop's production MUST be 6.0 (4.0 × 2 × 0.75)

### Requirement: Worker Inside Count Tracking

The production system MUST track how many workers are inside each building. This count MUST be updated when workers enter or exit via GOAP actions.

(Previously: No worker tracking in buildings)

#### Scenario: Production uses worker count from BuildingInterior

- GIVEN a building entity with BuildingInterior component showing 2 workers
- WHEN the production system calculates output
- THEN it SHALL use workers_inside = 2 from the BuildingInterior component

#### Scenario: Production defaults to zero if no interior component

- GIVEN a building without a BuildingInterior component
- WHEN the production system calculates output
- THEN workers_inside SHALL default to 0

### Requirement: Per-Building Production Calculation

Production MUST be calculated individually for each building, not aggregated at the settlement level.

(Previously: Unchanged — production was already per-building)

#### Scenario: Multiple buildings with different outputs

- GIVEN a settlement with a farm (base_output 2.0, 1 worker) and a workshop (base_output 4.0, 2 workers, efficiency 1.0)
- WHEN production is calculated
- THEN farm output MUST be 2.0 AND workshop output MUST be 8.0

## Unchanged Requirements

### Requirement: Production Resource Mapping

Each building MUST define which resources it produces.

- GIVEN a building with Produces: {"food": 2.0, "gold": 1.0}
- WHEN production is calculated
- THEN it MUST produce 2.0 food units per worker per tick
- AND it MUST produce 1.0 gold units per worker per tick

### Requirement: Settlement Consumption

Settlements still consume resources based on population.

- GIVEN a settlement with population 10 and consumption rate 0.01 per NPC
- WHEN consumption is calculated
- THEN total consumption MUST be 0.1 per tick

## New Scenarios Added

### Scenario: Production reaches zero when last worker exits

- GIVEN a farm with 1 worker inside producing 2.0 food
- WHEN that worker exits the building
- THEN food production MUST become 0.0

### Scenario: Production capacity capped at max_workers

- GIVEN a building with max_workers 3 and 5 workers attempting to enter
- WHEN production is calculated
- THEN it SHALL only count 3 workers (the maximum)