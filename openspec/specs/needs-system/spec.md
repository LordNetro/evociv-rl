# needs-system Specification

## Purpose

Sistema de necesidades básicas para NPCs: Hunger y Fatigue como valores continuos [0, 1] que decaen por tick, afectando la planificación GOAP.

## Requirements

### Requirement: Need Component

The system MUST define a Need component with two float64 fields: Hunger and Fatigue, each clamped to the range [0.0, 1.0].

| Field | Range | Direction | Decay per Tick |
|-------|-------|-----------|----------------|
| Hunger | [0, 1] | Increases | +0.01 |
| Fatigue | [0, 1] | Increases | +0.005 |

#### Scenario: Hunger increases over ticks

- GIVEN an NPC with Hunger=0.0
- WHEN 10 simulation ticks pass
- THEN Hunger MUST equal 0.10 (± float precision)

#### Scenario: Fatigue increases over ticks

- GIVEN an NPC with Fatigue=0.0
- WHEN 10 simulation ticks pass
- THEN Fatigue MUST equal 0.05 (± float precision)

#### Scenario: Values clamp at maximum

- GIVEN an NPC with Hunger=0.99
- WHEN 5 ticks pass (would exceed 1.0)
- THEN Hunger MUST NOT exceed 1.0

### Requirement: LOD Scaling

The system MUST decay needs at ALL LOD levels. At LOD level 0 (distant), the decay rate MUST be multiplied by 0.5×.

| LOD Level | Decay Multiplier |
|-----------|------------------|
| 2 (local) | 1.0× |
| 1 (near) | 1.0× |
| 0 (distant) | 0.5× |

#### Scenario: Distant NPCs decay slower

- GIVEN an NPC with LOD=0 and Hunger=0.0
- WHEN 10 ticks pass
- THEN Hunger MUST equal 0.05 (± float precision, half the normal rate)

#### Scenario: Local NPCs decay at normal rate

- GIVEN an NPC with LOD=2 and Hunger=0.0
- WHEN 10 ticks pass
- THEN Hunger MUST equal 0.10 (± float precision)
