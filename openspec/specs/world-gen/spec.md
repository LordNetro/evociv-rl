# world-gen Specification

## Purpose

Generación procedural: Perlin 2D + fBm, WorldMap, pipeline 4 fases.

## Requirements

### Requirement: Perlin + fBm

MUST produce consistent float64 per seed. SHOULD differ across seeds. fBm MUST blend octaves.

#### Scenario: Same seed identical
- GIVEN two instances seed=42
- WHEN Noise2D(x,y) on both
- THEN values MUST match

#### Scenario: Different seeds differ
- GIVEN seeds 42 and 99
- WHEN Noise2D(0,0)
- THEN values SHOULD differ

#### Scenario: fBm blends octaves
- GIVEN 4 octaves
- WHEN fBm2D(x,y)
- THEN output blends all

### Requirement: WorldMap

MUST provide TileAt(x,y) and InBounds(x,y). Index: y*width+x. MUST validate bounds.

#### Scenario: TileAt correct
- GIVEN 10x10 WorldMap
- WHEN TileAt(3,7)
- THEN returns index 73

#### Scenario: InBounds rejects
- GIVEN 10x10 grid
- WHEN InBounds(-1,5) or InBounds(10,10)
- THEN MUST return false

### Requirement: 4-Phase Pipeline

MUST run height→humidity→temperature→biome in order.

#### Scenario: Tile gets biome
- GIVEN WorldMap + valid config
- WHEN Generate() completes
- THEN each tile has BiomeID

#### Scenario: Seeds differ worlds
- GIVEN configs differing only in seed
- WHEN Generate() each
- THEN height maps MUST differ

#### Scenario: Invalid params
- GIVEN width=0 config
- WHEN Generate()
- THEN MUST return error
