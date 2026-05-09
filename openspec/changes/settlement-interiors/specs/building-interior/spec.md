# Building Interior Specification

## Purpose

Enable buildings to contain navigable interior spaces with rooms, corridors, and doors. Workers can enter, navigate inside, and exit buildings. Production scales based on the number of workers inside.

## Requirements

### Requirement: Interior Grid Representation

Each building with the BuildingInterior component MUST have a 2D grid representing walkable floor space. The grid size SHALL be determined by the building's footprint dimensions (width × height).

#### Scenario: Building generates interior grid

- GIVEN a building entity with footprint dimensions (width: 5, height: 4)
- WHEN the BuildingInterior component is initialized
- THEN the interior grid SHALL have 5 columns × 4 rows of cells

#### Scenario: Grid cells have correct types

- GIVEN an interior grid with cells
- WHEN each cell is inspected
- THEN each cell MUST be one of: floor, wall, door, corridor

### Requirement: Room Generation

The interior grid MUST contain one or more rooms. Room placement SHALL be deterministic based on the building seed.

#### Scenario: Farm generates two rooms

- GIVEN a building with type "farm" and seed 123
- WHEN the interior is generated
- THEN there SHALL be at least 2 distinct room areas

#### Scenario: House generates one room

- GIVEN a building with type "house" and seed 456
- WHEN the interior is generated
- THEN there SHALL be exactly 1 room area

### Requirement: Door Positions

Each building MUST expose door positions that serve as entry and exit points. Doors MUST connect interior rooms to the exterior.

#### Scenario: Building has at least one door

- GIVEN a building entity with an interior
- WHEN door positions are queried
- THEN there MUST be at least 1 door position

#### Scenario: Door at building edge

- GIVEN a building interior
- WHEN a door position is examined
- THEN at least one door MUST be on the perimeter of the interior grid

### Requirement: Maximum Worker Capacity

Each building MUST have a maximum worker capacity that limits how many workers can be inside simultaneously.

#### Scenario: Farm allows three workers

- GIVEN a building with type "farm"
- WHEN MaxWorkers is read
- THEN the value MUST be 3

#### Scenario: House allows one worker

- GIVEN a building with type "house"
- WHEN MaxWorkers is read
- THEN the value MUST be 1

### Requirement: Workers Inside Tracking

The BuildingInterior component MUST track which worker entities are currently inside the building.

#### Scenario: Worker enters building

- GIVEN a building with no workers inside
- WHEN a worker enters via EnterBuilding action
- THEN workers_inside list MUST contain that worker entity

#### Scenario: Worker exits building

- GIVEN a building with one worker inside
- WHEN that worker exits via ExitBuilding action
- THEN workers_inside list MUST be empty

#### Scenario: At capacity prevents entry

- GIVEN a building with workers_inside equal to MaxWorkers
- WHEN a worker attempts to enter
- THEN the entry MUST be rejected

### Requirement: Deterministic Interior Generation

Interior layout MUST be deterministic given the same building seed. Different seeds MUST produce different layouts.

#### Scenario: Same seed produces identical layout

- GIVEN building seed S
- WHEN generating two buildings with seed S
- THEN both buildings MUST have identical room positions, door positions, and cell types

#### Scenario: Different seed produces different layout

- GIVEN two buildings with seeds S1 and S2 where S1 ≠ S2
- WHEN their interiors are compared
- THEN at least one aspect of the layout MUST differ

## Out of Scope

- Multi-floor buildings
- Complex furniture placement
- Dynamic building expansion
- Inventory storage within buildings