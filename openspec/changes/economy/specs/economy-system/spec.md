# economy-system Specification

## Purpose

Sistema ECS que produce recursos automáticamente según los workers asignados a edificios productivos y consume food por NPC cada tick. Implementa helpers de ResourceStore (Add/Remove/Has) y lazy-init del componente.

## Requirements

### Requirement: SettlementEconomySystem Execution

SettlementEconomySystem MUST ejecutarse cada tick del World. Debe iterar todas las entidades con componente `Settlement`, obtener su `ResourceStore` (creándolo con lazy-init si no existe), y procesar producción y consumo.

#### Scenario: Produce food from farm with workers

- GIVEN un settlement con ResourceStore vacío, un building "farm" con `max_workers=3`, y 2 NPCs con `Job{Role: "farmer"}` y `HomeReference` a este settlement
- WHEN SettlementEconomySystem.Update() se ejecuta con dt=1.0
- THEN ResourceStore["food"] MUST incrementar en +4.0 (2 workers × 2.0 food/worker)
- AND ResourceStore["food"] MUST ser ≥ 4.0

#### Scenario: Blacksmith produces tools

- GIVEN un settlement con un building "blacksmith" (`max_workers=2`) y 1 NPC con `Job{Role: "smith"}` asignado
- WHEN SettlementEconomySystem.Update() se ejecuta
- THEN ResourceStore["tools"] MUST incrementar en +1.0 (1 worker × 1.0 tools/worker)

#### Scenario: Market produces gold and consumes food

- GIVEN un settlement con un building "market" (`max_workers=2`, `consumes: {food: 0.5}`, `produces: {gold: 1.0}`), 1 NPC `Job{Role: "merchant"}`, y ResourceStore con food ≥ 0.5
- WHEN SettlementEconomySystem.Update() se ejecuta
- THEN ResourceStore["gold"] MUST incrementar en +1.0
- AND ResourceStore["food"] MUST decrementar en -0.5

### Requirement: NPC Food Consumption

Cada NPC con componente `HomeReference` MUST consumir 0.01 food/tick del settlement al que pertenece, deducido del ResourceStore del settlement.

#### Scenario: NPC consumes food from settlement

- GIVEN un settlement con ResourceStore["food"] = 100.0 y 5 NPCs con HomeReference
- WHEN SettlementEconomySystem.Update() se ejecuta con dt=1.0
- THEN ResourceStore["food"] MUST decrementar en -0.05 (5 × 0.01)

#### Scenario: No NPCs, no consumption

- GIVEN un settlement con ResourceStore["food"] = 50.0 y 0 NPCs con HomeReference
- WHEN SettlementEconomySystem.Update() se ejecuta
- THEN ResourceStore["food"] MUST permanecer en 50.0

### Requirement: Worker Assignment

Workers MUST ser NPCs cuyo `Job.Role` coincide con el `role` definido en el BuildingDef productivo. El settlement MUST asignar hasta `max_workers` NPCs por building. Si hay más NPCs elegibles que `max_workers`, solo se asignan hasta el límite.

#### Scenario: Workers capped by max_workers

- GIVEN un settlement con building "farm" (`max_workers=2`, `role: farmer`) y 5 NPCs con `Job{Role: "farmer"}` y HomeReference
- WHEN SettlementEconomySystem.Update() se ejecuta
- THEN solo 2 workers MUST ser asignados a la farm
- AND la producción MUST ser +4.0 food (2 × 2.0), no +10.0

#### Scenario: No workers, no production

- GIVEN un settlement con building "farm" pero 0 NPCs con `Job{Role: "farmer"}`
- WHEN SettlementEconomySystem.Update() se ejecuta
- THEN ResourceStore["food"] MUST permanecer sin cambios por producción

### Requirement: ResourceStore Helpers

ResourceStore MUST exponer helpers: `Add(resource string, amount float64)`, `Remove(resource string, amount float64) bool`, `Has(resource string, amount float64) bool`. Remove MUST retornar false si el recurso es insuficiente (sin modificar el store).

#### Scenario: Add increments resource

- GIVEN un ResourceStore con Resources = {"food": 10.0}
- WHEN Add("food", 5.0) es llamado
- THEN Resources["food"] MUST ser 15.0

#### Scenario: Remove decrements when sufficient

- GIVEN un ResourceStore con Resources = {"food": 10.0}
- WHEN Remove("food", 3.0) es llamado
- THEN Remove MUST retornar true
- AND Resources["food"] MUST ser 7.0

#### Scenario: Remove returns false when insufficient

- GIVEN un ResourceStore con Resources = {"food": 1.0}
- WHEN Remove("food", 5.0) es llamado
- THEN Remove MUST retornar false
- AND Resources["food"] MUST permanecer en 1.0

### Requirement: Lazy-Init ResourceStore

Si un settlement no tiene componente `ResourceStore` al iniciar el tick, SettlementEconomySystem MUST crearlo con `Resources: {"food": 0, "gold": 0, "tools": 0}` antes de procesar.

#### Scenario: Lazy-init on first tick

- GIVEN un settlement sin ResourceStore componente
- WHEN SettlementEconomySystem.Update() se ejecuta por primera vez
- THEN el settlement MUST tener un ResourceStore con Resources["food"] = 0, ["gold"] = 0, ["tools"] = 0

### Requirement: Building Without Production

Buildings sin `produces` (e.g., "house", "tavern") MUST ignorarse durante la fase de producción. No generan recursos ni consumen inputs aunque tengan workers.

#### Scenario: House does not produce

- GIVEN un settlement con building "house" (sin produces) y 1 NPC con `Job{Role: "merchant"}`
- WHEN SettlementEconomySystem.Update() se ejecuta
- THEN ningún recurso MUST ser producido por la house
