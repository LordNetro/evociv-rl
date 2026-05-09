# settlement-interior-data Specification

## Purpose

Data-driven definitions for settlement interior rendering. Extends BuildingDef and buildings YAML with interior symbol and color fields to render buildings inside the settlement view.

## Requirements

### Requirement: BuildingDef Interior Fields

The `BuildingDef` struct MUST include two new fields: `InteriorSymbol` (string, a single Unicode rune) and `Color` (string, hex color). The `Building` component MUST include the same fields for runtime rendering.

| Building ID | Interior Symbol | Color    |
|-------------|----------------|----------|
| house       | ⌂              | #8B4513  |
| farm        | ╬              | #DEB887  |
| market      | §              | #FFD700  |
| tavern      | ♨              | #FF6347  |
| temple      | ϟ              | #E6E6FA  |
| blacksmith  | ⚒              | #A0522D  |

#### Scenario: BuildingDef loads InteriorSymbol and Color from YAML

- GIVEN a `data/buildings.yaml` entry `{id: farm, interior_symbol: "╬", color: "#DEB887"}`
- WHEN `LoadBuildingTypes()` is called
- THEN the resulting `BuildingDef` MUST have `InteriorSymbol = "╬"` and `Color = "#DEB887"`

#### Scenario: Building component carries InteriorSymbol and Color

- GIVEN a Building entity spawned from a BuildingDef with InteriorSymbol "⌂" and Color "#8B4513"
- WHEN the Building component is retrieved
- THEN its InteriorSymbol MUST be "⌂" and Color MUST be "#8B4513"

#### Scenario: Missing interior fields default to building symbol

- GIVEN a building type without `interior_symbol` or `color` in YAML
- WHEN the BuildingDef is loaded
- THEN the system MUST fall back to the building's existing `symbol` and `color` fields from the original definition

### Requirement: YAML Extension for buildings.yaml

The `data/buildings.yaml` file MUST accept `interior_symbol` (string) and `color` (string) as optional fields on each building entry. The loader MUST NOT reject entries missing these fields.

#### Scenario: Existing buildings.yaml loads without errors

- GIVEN the current `data/buildings.yaml` with the existing 6 building types (no interior_symbol or color added)
- WHEN `LoadBuildingTypes()` is called
- THEN no error MUST be returned

#### Scenario: Interior fields are optional

- GIVEN a buildings.yaml entry `{id: house, name: Casa}` without `interior_symbol` or `color`
- WHEN the YAML is parsed
- THEN the interior_symbol and color fields MUST be empty strings (zero values)

### Requirement: Loader Parses Interior Fields

The `LoadBuildingTypes()` function in `internal/simulation/settlement/data.go` MUST parse `interior_symbol` (string) and `color` (string) from each building entry in the registry. Building entities spawned by `SettlementSpawnSystem` MUST copy these fields from the BuildingDef.

#### Scenario: Loader reads interior_symbol

- GIVEN a building entry `{id: tavern, interior_symbol: "♨", color: "#FF6347"}` in the registry
- WHEN `LoadBuildingTypes()` parses it
- THEN the resulting BuildingDef MUST contain `InteriorSymbol = "♨"` and `Color = "#FF6347"`

#### Scenario: Spawned building copies interior fields

- GIVEN a tavern with `InteriorSymbol = "♨"` and `Color = "#FF6347"`
- WHEN `SettlementSpawnSystem` spawns a tavern Building entity
- THEN the entity's Building component MUST have `InteriorSymbol = "♨"` and `Color = "#FF6347"`

### Validation Rules

- PASS: BuildingDef with valid interior_symbol and color loads correctly
- PASS: BuildingDef without interior_symbol or color loads with zero values (backward compatible)
- PASS: Spawned Building entity copies interior fields from BuildingDef
- FAIL: BuildingDef with multi-rune interior_symbol returns an error or is truncated to first rune
- FAIL: BuildingDef with invalid hex color string — advisory warning, no error (default to fallback)
