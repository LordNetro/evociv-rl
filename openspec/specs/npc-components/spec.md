# npc-components Specification

## Purpose

Componentes ECS que modelan un NPC como agente autónomo: identidad, personalidad, ocupación, estado cognitivo, apariencia y nivel de detalle.

## Requirements

### Requirement: Six Component Types

The system MUST define exactly six ECS component types for NPCs: Health, Personality, Job, AIState, Appearance, and LOD.

| Component | Fields | Purpose |
|-----------|--------|---------|
| Health | Current, Max float64 | Vitalidad del NPC |
| Personality | O, C, E, A, N float64 | Big 5 (Openness, Conscientiousness, Extraversion, Agreeableness, Neuroticism) |
| Job | Role string | Ocupación (farmer, hunter, trader, etc.) |
| AIState | Goals []string, Plan []string | Estado cognitivo GOAP-ready |
| Appearance | Symbol rune, Color color.Attribute | Visual en TUI |
| LOD | Level int | Nivel de detalle (0=off, 1=far, 2=near) |

#### Scenario: Create NPC entity with all components

- GIVEN an ECS World
- WHEN an entity is created and all six component types are assigned
- THEN each component MUST be retrievable by its type for that entity

#### Scenario: Missing components return zero values

- GIVEN an entity with only a Health component
- WHEN querying the Personality component for that entity
- THEN the zero value of Personality MUST be returned

### Requirement: Personality Distribution

Personality values MUST be drawn from a normal distribution clamped to [0, 1], seeded deterministically per entity. Each trait (O, C, E, A, N) MUST be independently sampled.

#### Scenario: Deterministic personality by seed

- GIVEN a fixed global seed and entity ID
- WHEN generating Personality for the same entity twice
- THEN both values MUST be identical across all five traits

#### Scenario: Independence across traits

- GIVEN a generated Personality for entity A
- WHEN comparing with entity B (different ID, same seed)
- THEN the five-trait vectors MUST differ

### Requirement: Appearance Assignment

Appearance MUST assign a symbol and color based on the NPC's race and role. Races define the base symbol; roles MAY override the color.

#### Scenario: Appearance varies by race and role

- GIVEN two NPCs with different race/role combinations
- WHEN their Appearance components are generated
- THEN they MUST differ in symbol or color

#### Scenario: Same race and role produce same appearance

- GIVEN two NPCs with identical race and role
- WHEN their Appearance components are generated
- THEN symbol and color MUST be identical
