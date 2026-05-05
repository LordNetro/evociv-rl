# famine-system Specification

## Purpose

Sistema ECS que detecta déficit de food en settlements y causa migración de NPCs. Cuando un settlement tiene food < 0, los NPCs pierden su HomeReference y se vuelven nómadas.

## Requirements

### Requirement: FamineSystem Execution

FamineSystem MUST ejecutarse cada tick del World después de SettlementEconomySystem. Debe iterar settlements con ResourceStore y verificar si food < 0.

#### Scenario: Food deficit detected

- GIVEN un settlement con ResourceStore["food"] = -5.0
- WHEN FamineSystem.Update() se ejecuta
- THEN el sistema MUST detectar el déficit de food

#### Scenario: Positive food, no action

- GIVEN un settlement con ResourceStore["food"] = 50.0 y 5 NPCs con HomeReference
- WHEN FamineSystem.Update() se ejecuta
- THEN ningún NPC MUST perder su HomeReference

### Requirement: NPC Migration on Food Deficit

Si un settlement tiene food < 0, FamineSystem MUST remover el componente `HomeReference` de un NPC por tick hasta que food >= 0. Los NPCs sin `HomeReference` se convierten en nómadas.

#### Scenario: NPC loses HomeReference during famine

- GIVEN un settlement con ResourceStore["food"] = -2.0 y 3 NPCs con HomeReference a este settlement
- WHEN FamineSystem.Update() se ejecuta
- THEN exactamente 1 NPC MUST perder su HomeReference (HomeReference eliminado)
- AND los otros 2 NPCs MUST mantener su HomeReference

#### Scenario: Multiple ticks remove multiple NPCs

- GIVEN un settlement con food = -10.0 y 5 NPCs con HomeReference
- WHEN FamineSystem.Update() se ejecuta por 3 ticks consecutivos
- THEN 3 NPCs MUST haber perdido su HomeReference (1 por tick)
- AND 2 NPCs MUST mantener su HomeReference

#### Scenario: All NPCs migrate when deficit persists

- GIVEN un settlement con food = -5.0 y 2 NPCs con HomeReference
- WHEN FamineSystem.Update() se ejecuta por 3 ticks
- THEN después del tick 2, 0 NPCs MUST tener HomeReference a ese settlement
- AND el tick 3 MUST ser no-op (no hay más NPCs que migrar)

### Requirement: Nomad Behavior

NPCs sin `HomeReference` (nómadas) MUST comportarse según el sistema WanderSystem existente con movimiento wander biome-weighted. No reciben producción ni protección del settlement.

#### Scenario: Nomad wanders like homeless NPC

- GIVEN un NPC sin componente HomeReference en biome plains
- WHEN WanderSystem.Update() se ejecuta
- THEN el NPC MUST moverse a un tile adyacente según las reglas de wander existentes

### Requirement: Recovery After Famine

Si el food del settlement vuelve a ser >= 0 (por producción posterior), FamineSystem MUST dejar de remover HomeReference. Los NPCs que ya migraron NO recuperan automáticamente su HomeReference.

#### Scenario: Famine stops when food recovers

- GIVEN un settlement con food = -1.0, luego de producción food sube a 2.0
- WHEN FamineSystem.Update() se ejecuta después de la recuperación
- THEN ningún NPC MUST perder su HomeReference en ese tick
