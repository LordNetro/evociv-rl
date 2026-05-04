# tui-map Specification

## Purpose

Mapa navegable TUI: grid coloreado, cámara wasd, toggle 'm'.

## Requirements

### Requirement: Rendering

MUST render grid colored by biome. MUST adapt to WindowSizeMsg.

#### Scenario: Colored tiles
- GIVEN WorldMap and terminal 80x24
- WHEN map screen active
- THEN view displays colored characters

### Requirement: Camera

MUST move camera with wasd. Arrows MAY also work.

#### Scenario: wasd moves
- GIVEN camera at (0,0)
- WHEN 'd' pressed
- THEN offset shifts right by 1

### Requirement: Screen Toggle

MUST toggle welcome↔map on 'm'.

#### Scenario: To map
- GIVEN welcome active
- WHEN 'm' pressed
- THEN screen switches to map

#### Scenario: To welcome
- GIVEN map active
- WHEN 'm' pressed
- THEN screen switches to welcome

### Requirement: Bounds

MUST NOT scroll beyond bounds. MUST NOT panic.

#### Scenario: Edge stop
- GIVEN camera at rightmost column
- WHEN 'd' pressed again
- THEN camera stays, no panic
