# Indoor Pathfinding Specification

## Purpose

Provide pathfinding capabilities within building interiors using A* algorithm on the 2D interior grid. Workers use these paths to navigate from doors to work locations inside buildings.

## Requirements

### Requirement: A* Pathfinding on Interior Grid

The IndoorPathfinder MUST compute paths using the A* algorithm on the interior grid. Only floor, corridor, and door cells are traversable; wall cells are not.

#### Scenario: Find path from door to room center

- GIVEN a building interior with a door at (1, 1) and a room center at (3, 2)
- WHEN FindPath(doorPos, targetPos) is called
- THEN a valid path MUST be returned that moves only through traversable cells

#### Scenario: Path avoids walls

- GIVEN a building interior with obstacles
- WHEN a path is computed
- THEN no cell in the path SHALL be a wall

#### Scenario: No path exists for isolated target

- GIVEN a building interior where target is completely surrounded by walls
- WHEN FindPath is called
- THEN nil MUST be returned with an error indicating no path

### Requirement: Door-to-Door Navigation

Workers MUST be able to navigate from any door to any other door within the same building.

#### Scenario: Path between two doors

- GIVEN a building with door A at (0, 2) and door B at (4, 2)
- WHEN path from door A to door B is requested
- THEN a valid path MUST be returned

#### Scenario: Path respects door cells

- GIVEN a path between doors
- WHEN the path is inspected
- THEN door cells MAY be included in the path (doors are traversable)

### Requirement: Entry Point Calculation

The system MUST compute valid entry positions at building edges where workers transition from world navigation to interior navigation.

#### Scenario: Entry adjacent to door

- GIVEN a building at world position (10, 10) with a door at interior position (0, 2)
- WHEN entry position is calculated
- THEN the result MUST be adjacent to the building's world position

#### Scenario: Exit position calculation

- GIVEN a worker inside a building at interior position (2, 2)
- WHEN exit to world is requested
- THEN the world position MUST correspond to a door cell

### Requirement: Path Caching

Path computation MAY be cached to improve performance when the same path is requested multiple times.

#### Scenario: Cached path returned

- GIVEN the same path requested twice
- WHEN the second request is made before the building interior changes
- THEN the cached path SHOULD be returned

#### Scenario: Cache invalidated on layout change

- GIVEN a cached path exists
- WHEN the building interior is modified
- THEN the cache MUST be invalidated

### Requirement: Empty Building Path

The pathfinder MUST handle buildings with empty interiors gracefully.

#### Scenario: Path in empty interior

- GIVEN a building with a single room and doors
- WHEN a path is requested within that room
- THEN a valid path MUST be returned if traversable cells exist

### Requirement: Single Cell Path

A path from a position to itself MUST return a valid single-element path.

#### Scenario: Zero-distance path

- GIVEN any valid interior position
- WHEN FindPath is called with same start and end
- THEN the returned path MUST contain exactly one cell (the start position)

## Out of Scope

- Pathfinding between buildings
- Outdoor navigation
- Dynamic obstacle avoidance during path execution
- Path smoothing or optimization beyond A*