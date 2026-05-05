# Delta for settlement-data: economy-data

## ADDED Requirements

### Requirement: Economic Building Fields

`data/buildings.yaml` MUST extender cada building type con los campos opcionales: `role` (string), `produces` (map[string]float64), `consumes` (map[string]float64), `max_workers` (int, default 0). Edificios sin estos campos (solo `id` y `name`) MUST seguir siendo válidos — los campos ausentes se interpretan como "no produce, no consume, max_workers=0".

#### Scenario: Load buildings.yaml with produces and consumes

- GIVEN a `data/buildings.yaml` con farm (`produces: {food: 2.0}`, `max_workers: 3`, `role: farmer`) y blacksmith (`produces: {tools: 1.0}`, `max_workers: 2`, `role: smith`)
- WHEN `LoadBuildingTypes()` es llamado
- THEN farm MUST tener Produces = {"food": 2.0}, MaxWorkers = 3, Role = "farmer"
- AND blacksmith MUST tener Produces = {"tools": 1.0}, MaxWorkers = 2, Role = "smith"

#### Scenario: Legacy building without economic fields still loads

- GIVEN un building "house" en buildings.yaml sin `produces`, `consumes`, `role`, `max_workers`
- WHEN `LoadBuildingTypes()` es llamado
- THEN house MUST cargar con Produces = nil, Consumes = nil, Role = "", MaxWorkers = 0
- AND no MUST retornar error

### Requirement: Growth Threshold Definitions

`data/growth.yaml` (nuevo archivo) MUST definir thresholds de crecimiento por nivel. Formato: `kind: growth-thresholds`, data como array con campos: `level` (int), `food` (float64), `tools` (float64), `gold` (float64), `new_radius` (int), `new_buildings` ([]string, MAY be empty).

#### Scenario: Load valid growth.yaml

- GIVEN un `data/growth.yaml` con thresholds para Level 2 (food: 100, tools: 10, gold: 5, new_radius: 4) y Level 3 (food: 500, tools: 50, gold: 25, new_radius: 6)
- WHEN `LoadGrowthThresholds()` es llamado
- THEN MUST retornar 2 thresholds accesibles por level
- AND threshold para Level 2 MUST tener food=100, tools=10, gold=5, new_radius=4

#### Scenario: Missing growth.yaml returns empty

- GIVEN no existe `data/growth.yaml`
- WHEN `LoadGrowthThresholds()` es llamado
- THEN MUST retornar slice vacío sin error

### Requirement: BuildingDef Struct Extension

El struct `BuildingDef` en `internal/simulation/settlement/types.go` MUST extenderse con campos: `Role string`, `Produces map[string]float64`, `Consumes map[string]float64`, `MaxWorkers int`. El struct `GrowthThreshold` (nuevo) MUST definirse con campos: `Level int`, `Food float64`, `Tools float64`, `Gold float64`, `NewRadius int`, `NewBuildings []string`.

#### Scenario: BuildingDef struct has new fields

- GIVEN un BuildingDef cargado desde YAML con produces, consumes, role, max_workers
- WHEN se accede a los campos
- THEN Role, Produces, Consumes, MaxWorkers MUST estar disponibles

### Requirement: Validation — Produces Format

El loader MUST validar que `produces` y `consumes` sean mapas con valores float64 ≥ 0. Si un valor es negativo, MUST retornar error.

#### Scenario: Reject negative production rate

- GIVEN buildings.yaml con farm `produces: {food: -1.0}`
- WHEN `LoadBuildingTypes()` es llamado
- THEN MUST retornar error indicando rate negativo

## MODIFIED Requirements

### Requirement: Building Type Definitions

El archivo `data/buildings.yaml` MUST definir al menos: `house`, `farm`, `market`, `tavern`, `temple`, `blacksmith`. Cada edificio MUST tener: `id`, `name`, `symbol` (rune), `color`. Edificios productivos SHOULD tener: `role`, `produces` (map[string]float64), `consumes` (map[string]float64, MAY be empty), `max_workers` (int ≥ 0).
(Previously: buildings had `produces` as []string, no role/consumes/max_workers)

#### Scenario: Load valid buildings.yaml with economic data

- GIVEN a `data/buildings.yaml` with all six building types, where productive ones include role, produces, max_workers
- WHEN `LoadBuildingTypes()` is called
- THEN all buildings MUST be accessible with correct fields
- AND farm MUST have Produces = {"food": 2.0}, MaxWorkers = 3, Role = "farmer"
- AND market MUST have Produces = {"gold": 1.0}, Consumes = {"food": 0.5}, MaxWorkers = 2, Role = "merchant"
