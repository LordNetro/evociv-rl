# data-loader Specification

## Purpose

Cargar archivos YAML desde `data/` y registrarlos en un Registry tipado para consumo data-driven del simulador.

## Requirements

### Requirement: YAML Loading

The system MUST load all `.yaml` files from a configured directory into a typed Registry.

#### Scenario: Load valid YAML file

- GIVEN a `data/biomes.yaml` file with 2-3 biome definitions
- WHEN Load() is called
- THEN the biomes MUST be available in the Registry with correct field values

#### Scenario: Registry returns loaded data by type

- GIVEN a Registry with biomes loaded
- WHEN Get("biomes") is called
- THEN the biome data MUST be returned as the expected Go type

### Requirement: Error Handling

The system MUST report errors for missing directories and malformed YAML.

#### Scenario: Missing directory returns error

- GIVEN a path to a non-existent directory
- WHEN Load() is called
- THEN an error MUST be returned describing the missing directory

#### Scenario: Malformed YAML returns error

- GIVEN a YAML file with invalid syntax
- WHEN Load() is called
- THEN a parse error MUST be returned

#### Scenario: Empty directory loads successfully with no data

- GIVEN an empty directory
- WHEN Load() is called
- THEN no error is returned and the Registry is empty

### Requirement: Optional Validation Hook

The system MAY accept a validation function per data type. If provided, it MUST be called after YAML parse and before registration.

#### Scenario: Validator rejects invalid data

- GIVEN a validator that checks biome names are non-empty
- WHEN a biome with an empty name is loaded
- THEN a validation error MUST be returned
