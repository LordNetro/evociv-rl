# settlement-growth Specification

## Purpose

Sistema ECS que verifica thresholds de recursos cada tick y sube de nivel los settlements cuando acumulan recursos suficientes. Los thresholds son data-driven desde YAML.

## Requirements

### Requirement: SettlementGrowthSystem Execution

SettlementGrowthSystem MUST ejecutarse cada tick del World. Debe iterar settlements con componente `Settlement` y `ResourceStore`, y verificar si los recursos acumulados cumplen el threshold para el siguiente nivel.

#### Scenario: Settlement levels up when thresholds met

- GIVEN un settlement con Level=1, Radius=3, ResourceStore con food=100, tools=10, gold=5
- WHEN SettlementGrowthSystem.Update() se ejecuta
- THEN Settlement.Level MUST incrementar a 2
- AND Settlement.Radius MUST incrementar al nuevo radio del threshold
- AND ResourceStore MUST tener food=0, tools=0, gold=0 (recursos deducidos)

#### Scenario: Settlement stays at same level without enough resources

- GIVEN un settlement con Level=1 y ResourceStore con food=50, tools=0, gold=0
- WHEN SettlementGrowthSystem.Update() se ejecuta
- THEN Settlement.Level MUST permanecer en 1
- AND ResourceStore MUST mantener sus valores sin cambios

### Requirement: Level Thresholds

Los thresholds por nivel MUST cargarse desde `data/growth.yaml`. El sistema MUST definir los siguientes thresholds para MVP:

| Level | Food | Tools | Gold | New Radius |
|-------|------|-------|------|-----------|
| 1→2   | 100  | 10    | 5    | +1 (min 4) |
| 2→3   | 500  | 50    | 25   | +2 (min 6) |

#### Scenario: Level 2 to 3 requires more resources

- GIVEN un settlement con Level=2, ResourceStore con food=500, tools=50, gold=25
- WHEN SettlementGrowthSystem.Update() se ejecuta
- THEN Settlement.Level MUST ser 3
- AND Radius MUST incrementar según threshold (min 6)

#### Scenario: Partial resources for next level

- GIVEN un settlement con Level=1, ResourceStore con food=100, tools=10, gold=2 (gold insuficiente)
- WHEN SettlementGrowthSystem.Update() se ejecuta
- THEN Level MUST permanecer en 1

### Requirement: Resource Deduction on Level-Up

Al subir de nivel, el sistema MUST deducir del ResourceStore los recursos consumidos por el threshold. Los recursos deducidos MUST ser exactamente food, tools y gold especificados en el threshold.

#### Scenario: Resources are deducted on level up

- GIVEN un settlement con ResourceStore["food"]=150, ["tools"]=20, ["gold"]=10, Level=1
- WHEN SettlementGrowthSystem.Update() ejecuta level-up a Level 2 (threshold: 100 food, 10 tools, 5 gold)
- THEN ResourceStore["food"] MUST ser 50 (150 - 100)
- AND ResourceStore["tools"] MUST ser 10 (20 - 10)
- AND ResourceStore["gold"] MUST ser 5 (10 - 5)

### Requirement: Max Level Cap

El sistema MUST tener un nivel máximo (Level 3 para MVP). Un settlement en Level 3 MUST saltarse la verificación de thresholds y no intentar subir más.

#### Scenario: Max level settlement does not check thresholds

- GIVEN un settlement con Level=3 (máximo MVP)
- WHEN SettlementGrowthSystem.Update() se ejecuta
- THEN el sistema MUST saltar este settlement sin verificar thresholds
- AND ResourceStore MUST permanecer sin cambios

### Requirement: Missing Threshold Returns Error

Si no existe threshold definido para el siguiente nivel (e.g., Level 3 no tiene Level 4 definido), el sistema MUST tratar el settlement como si hubiera alcanzado el nivel máximo.

#### Scenario: No threshold for next level

- GIVEN un settlement con Level=3 y solo thresholds hasta Level 3 definidos en YAML
- WHEN SettlementGrowthSystem.Update() se ejecuta
- THEN el sistema MUST saltar este settlement sin error
