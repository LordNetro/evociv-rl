# Production Scaling Specification

## Purpose

Scale building production output based on the number of workers currently inside. Production follows the formula: `base_output * workers_inside * efficiency`.

## Requirements

### Requirement: Production Formula

Each building's production output MUST be calculated using the formula: `base_output × workers_inside × efficiency`.

#### Scenario: Zero workers produces zero output

- GIVEN a farm with base_output 2.0 and 0 workers inside
- WHEN production is calculated
- THEN the output MUST be 0.0

#### Scenario: One worker produces base output

- GIVEN a farm with base_output 2.0 and efficiency 1.0, with 1 worker inside
- WHEN production is calculated
- THEN the output MUST be 2.0 (2.0 × 1 × 1.0)

#### Scenario: Two workers doubles output

- GIVEN a farm with base_output 2.0 and efficiency 1.0, with 2 workers inside
- WHEN production is calculated
- THEN the output MUST be 4.0 (2.0 × 2 × 1.0)

### Requirement: Efficiency Factor

The efficiency factor MUST be applied to the production calculation. Efficiency MAY vary based on building type, worker assignment quality, or other factors.

#### Scenario: Efficiency reduces output

- GIVEN a farm with base_output 2.0, efficiency 0.5, with 1 worker inside
- WHEN production is calculated
- THEN the output MUST be 1.0 (2.0 × 1 × 0.5)

#### Scenario: Efficiency above 1.0 increases output

- GIVEN a workshop with base_output 2.0, efficiency 1.5, with 1 worker inside
- WHEN production is calculated
- THEN the output MUST be 3.0 (2.0 × 1 × 1.5)

### Requirement: Building Type Bonus

Different building types MAY have production bonuses. The system SHALL support type-specific multipliers.

#### Scenario: Farm has normal bonus

- GIVEN a farm building with efficiency 1.0 and 1 worker inside
- WHEN production is calculated
- THEN it SHALL use the standard farm multiplier (1.0x)

#### Scenario: Workshop has production bonus

- GIVEN a workshop building with efficiency 1.0 and 1 worker inside
- WHEN production is calculated
- THEN it SHALL apply the workshop bonus (e.g., 1.25x)

### Requirement: Worker Count Accuracy

The worker count MUST accurately reflect workers that have entered the building via the GOAP action system.

#### Scenario: Worker enters updates count

- GIVEN a building with 1 worker inside
- WHEN a second worker enters
- THEN workers_inside count MUST become 2

#### Scenario: Worker exit updates count

- GIVEN a building with 2 workers inside
- WHEN one worker exits
- THEN workers_inside count MUST become 1

### Requirement: Production Per Tick

Production MUST be calculated per game tick. The accumulated production over time represents the building's output rate.

#### Scenario: Production ticks continuously

- GIVEN a building with 1 worker inside and base_output 2.0
- WHEN the game updates twice
- THEN each tick MUST produce 2.0 units (assuming efficiency 1.0)

### Requirement: Per-Resource Calculation

Production MUST be calculated for each resource type the building produces.

#### Scenario: Multiple resource outputs

- GIVEN a building that produces food (base 2.0) and gold (base 0.5), with 1 worker inside and efficiency 1.0
- WHEN production is calculated
- THEN food output MUST be 2.0 AND gold output MUST be 0.5

## Out of Scope

- Resource consumption by buildings
- Inventory overflow handling
- Production storage in building
- Worker skill levels affecting efficiency