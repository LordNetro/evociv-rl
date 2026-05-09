# settlement-components Specification

## Purpose

Componentes ECS que modelan asentamientos, edificios y la relación NPC↔settlement.

## Requirements

### Requirement: Settlement Component

The system MUST define a `Settlement` struct with fields: `Name` (string), `Type` (string — "village"|"town"|"city"), `Radius` (int), `Population` (int), `Level` (int). The component MUST be registered in a `ComponentStore[Settlement]`.

#### Scenario: Create entity with Settlement component

- GIVEN an ECS World with Settlement store registered
- WHEN an entity is created and assigned a Settlement{Name: "Aldea Verde", Type: "village", Radius: 6, Population: 0, Level: 1}
- THEN the component MUST be retrievable by entity ID with all fields matching

### Requirement: Building Component

The system MUST define a `Building` struct with fields: `ID` (string), `Name` (string), `Level` (int). The component MUST be registered in a `ComponentStore[Building]`.

#### Scenario: Create entity with Building component

- GIVEN an ECS World with Building store registered
- WHEN an entity is created and assigned a Building{ID: "farm_01", Name: "Farm", Level: 1}
- THEN the component MUST be retrievable by entity ID with all fields matching

### Requirement: HomeReference Component

The system MUST define a `HomeReference` struct with field: `SettlementEntity` (ecs.Entity). This component MUST be registered in a `ComponentStore[HomeReference]`. It MAY be assigned to NPC entities to indicate their home settlement.

#### Scenario: Assign HomeReference to NPC

- GIVEN an ECS World with HomeReference store registered, an NPC entity (ID 42), and a Settlement entity (ID 7)
- WHEN a HomeReference{SettlementEntity: 7} is assigned to entity 42
- THEN querying HomeReference for entity 42 MUST return SettlementEntity = 7

#### Scenario: NPC without HomeReference returns zero

- GIVEN an NPC entity without HomeReference component
- WHEN querying HomeReference
- THEN the zero value MUST be returned (SettlementEntity = 0)

### Requirement: ResourceStore Component (SHOULD)

The system SHOULD define a `ResourceStore` struct with field: `Resources` (map[string]float64). This component is intended for future economic simulation and MAY remain unused in initial implementation.

#### Scenario: Create entity with ResourceStore (optional)

- GIVEN an ECS World with ResourceStore store registered
- WHEN a settlement entity is assigned ResourceStore{Resources: {"wood": 100.0, "food": 50.0}}
- THEN the Resources map MUST be retrievable by entity ID
