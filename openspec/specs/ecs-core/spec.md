# ecs-core Specification

## Purpose

Núcleo del motor ECS (Entity-Component-System) del simulador. Proporciona las abstracciones fundamentales para modelar entidades, sus componentes, y los sistemas que operan sobre ellos.

## Requirements

### Requirement: Entity Creation

The system MUST allow creation of Entity values with unique identifiers.

#### Scenario: Create entity with unique ID

- GIVEN no entities exist
- WHEN a new Entity is created
- THEN it MUST have a non-zero unique ID

### Requirement: Component Assignment

The system MUST support assigning typed components to entities via ComponentStore[T].

#### Scenario: Get and set typed component

- GIVEN a ComponentStore for a concrete component type
- WHEN a component is stored for a given entity ID
- THEN it MUST be retrievable by its type for that same entity

#### Scenario: Missing component returns zero value

- GIVEN a ComponentStore
- WHEN querying a component type for an entity that has none assigned
- THEN the zero value of that component type MUST be returned

### Requirement: World Management

The World MUST manage entities and execute registered Systems on Update.

#### Scenario: World executes all systems on Update

- GIVEN a World with N registered Systems
- WHEN World.Update() is called
- THEN every registered System MUST execute exactly once

#### Scenario: World adds entities and queries components

- GIVEN a World
- WHEN an entity is created via the World
- THEN the entity ID MUST be non-zero and components can be attached via the World

### Requirement: System Interface

A System MUST implement a fixed interface with an Update method that receives the World.

#### Scenario: System mutates component state

- GIVEN a System that increments a counter component per entity
- WHEN World.Update() is called
- THEN all entities with that component MUST have their value incremented

### Requirement: Concurrent Safety (Edge)

The ComponentStore SHOULD be safe for concurrent reads from multiple Systems during a single Update tick, though writes happen sequentially (single-threaded update loop).

#### Scenario: Read during update does not panic

- GIVEN a World with two Systems reading the same component type
- WHEN World.Update() is called
- THEN both Systems MUST complete without data races
