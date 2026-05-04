package gen

import (
	"testing"
	"testing/fstest"

	"github.com/marco/evociv-rl/internal/data"
	"github.com/marco/evociv-rl/internal/world"
)

func TestBiomeMatchPlains(t *testing.T) {
	biomes := []BiomeDef{
		{ID: "ocean", MinHeight: -1.0, MaxHeight: 0.1, MinHumidity: 0.0, MaxHumidity: 1.0, MinTemperature: 0.0, MaxTemperature: 1.0},
		{ID: "plains", MinHeight: 0.1, MaxHeight: 0.6, MinHumidity: 0.2, MaxHumidity: 0.7, MinTemperature: 0.3, MaxTemperature: 0.8},
	}
	tile := world.Tile{Height: 0.3, Humidity: 0.4, Temperature: 0.5}
	got := MatchBiome(tile, biomes)
	if got != "plains" {
		t.Errorf("MatchBiome = %q, want plains", got)
	}
}

func TestBiomeMatchUnknown(t *testing.T) {
	biomes := []BiomeDef{
		{ID: "ocean", MinHeight: -1.0, MaxHeight: 0.1, MinHumidity: 0.0, MaxHumidity: 1.0, MinTemperature: 0.0, MaxTemperature: 1.0},
	}
	tile := world.Tile{Height: 100.0, Humidity: 0.5, Temperature: 0.5}
	got := MatchBiome(tile, biomes)
	if got != "unknown" {
		t.Errorf("MatchBiome = %q, want unknown", got)
	}
}

func TestBiomeMatchOcean(t *testing.T) {
	biomes := []BiomeDef{
		{ID: "ocean", MinHeight: -1.0, MaxHeight: 0.1, MinHumidity: 0.0, MaxHumidity: 1.0, MinTemperature: 0.0, MaxTemperature: 1.0},
		{ID: "plains", MinHeight: 0.1, MaxHeight: 0.6, MinHumidity: 0.2, MaxHumidity: 0.7, MinTemperature: 0.3, MaxTemperature: 0.8},
	}
	tile := world.Tile{Height: -0.5, Humidity: 0.5, Temperature: 0.5}
	got := MatchBiome(tile, biomes)
	if got != "ocean" {
		t.Errorf("MatchBiome = %q, want ocean", got)
	}
}

func TestBiomeMatchFirstInOrder(t *testing.T) {
	// Both could match, first in order should win
	biomes := []BiomeDef{
		{ID: "first", MinHeight: 0.0, MaxHeight: 1.0, MinHumidity: 0.0, MaxHumidity: 1.0, MinTemperature: 0.0, MaxTemperature: 1.0},
		{ID: "second", MinHeight: 0.0, MaxHeight: 1.0, MinHumidity: 0.0, MaxHumidity: 1.0, MinTemperature: 0.0, MaxTemperature: 1.0},
	}
	tile := world.Tile{Height: 0.5, Humidity: 0.5, Temperature: 0.5}
	got := MatchBiome(tile, biomes)
	if got != "first" {
		t.Errorf("MatchBiome = %q, want first", got)
	}
}

func TestLoadBiomesFromRegistry(t *testing.T) {
	fsys := fstest.MapFS{
		"biomes.yaml": &fstest.MapFile{
			Data: []byte(`kind: biomes
data:
  - id: plains
    name: Llanuras
    symbol: "."
    color: "#90EE90"
    minHeight: 0.1
    maxHeight: 0.6
    minHumidity: 0.2
    maxHumidity: 0.7
    minTemperature: 0.3
    maxTemperature: 0.8
`),
		},
	}
	loader := data.NewLoader(fsys)
	registry := data.NewRegistry()
	if err := loader.LoadAll(".", registry); err != nil {
		t.Fatalf("load data: %v", err)
	}

	biomes, err := LoadBiomes(registry)
	if err != nil {
		t.Fatalf("LoadBiomes error: %v", err)
	}
	if len(biomes) != 1 {
		t.Fatalf("expected 1 biome, got %d", len(biomes))
	}
	b := biomes[0]
	if b.ID != "plains" {
		t.Errorf("id = %q, want plains", b.ID)
	}
	if b.Symbol != "." {
		t.Errorf("symbol = %q, want '.'", b.Symbol)
	}
	if b.MinHeight != 0.1 {
		t.Errorf("minHeight = %v, want 0.1", b.MinHeight)
	}
	if b.MaxHeight != 0.6 {
		t.Errorf("maxHeight = %v, want 0.6", b.MaxHeight)
	}
	if b.MinHumidity != 0.2 {
		t.Errorf("minHumidity = %v, want 0.2", b.MinHumidity)
	}
	if b.MaxHumidity != 0.7 {
		t.Errorf("maxHumidity = %v, want 0.7", b.MaxHumidity)
	}
	if b.MinTemperature != 0.3 {
		t.Errorf("minTemperature = %v, want 0.3", b.MinTemperature)
	}
	if b.MaxTemperature != 0.8 {
		t.Errorf("maxTemperature = %v, want 0.8", b.MaxTemperature)
	}
}

func TestGenerateFillsAllTiles(t *testing.T) {
	config := GenConfig{Seed: 42, Width: 3, Height: 3, Octaves: 4, Lacunarity: 2.0, Gain: 0.5, Scale: 10.0}
	biomes := []BiomeDef{
		{ID: "test", MinHeight: -1.0, MaxHeight: 1.0, MinHumidity: 0.0, MaxHumidity: 1.0, MinTemperature: 0.0, MaxTemperature: 1.0},
	}
	wm, err := Generate(config.Width, config.Height, config, biomes)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if len(wm.Tiles) != 9 {
		t.Errorf("expected 9 tiles, got %d", len(wm.Tiles))
	}
	for i, tile := range wm.Tiles {
		if tile.BiomeID == "" {
			t.Errorf("tile %d has empty BiomeID", i)
		}
	}
}

func TestGenerateSeedsDiffer(t *testing.T) {
	config1 := GenConfig{Seed: 42, Width: 8, Height: 8, Octaves: 4, Lacunarity: 2.0, Gain: 0.5, Scale: 10.0}
	config2 := GenConfig{Seed: 99, Width: 8, Height: 8, Octaves: 4, Lacunarity: 2.0, Gain: 0.5, Scale: 10.0}
	biomes := []BiomeDef{
		{ID: "test", MinHeight: -1.0, MaxHeight: 1.0, MinHumidity: 0.0, MaxHumidity: 1.0, MinTemperature: 0.0, MaxTemperature: 1.0},
	}
	wm1, err := Generate(config1.Width, config1.Height, config1, biomes)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	wm2, err := Generate(config2.Width, config2.Height, config2, biomes)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	allSame := true
	for i := range wm1.Tiles {
		if wm1.Tiles[i].Height != wm2.Tiles[i].Height {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("different seeds produced identical height maps")
	}
}

func TestGenerateInvalidParams(t *testing.T) {
	config := GenConfig{Seed: 42, Width: 0, Height: 8, Octaves: 4, Lacunarity: 2.0, Gain: 0.5, Scale: 10.0}
	biomes := []BiomeDef{
		{ID: "test", MinHeight: -1.0, MaxHeight: 1.0, MinHumidity: 0.0, MaxHumidity: 1.0, MinTemperature: 0.0, MaxTemperature: 1.0},
	}
	_, err := Generate(config.Width, config.Height, config, biomes)
	if err == nil {
		t.Error("expected error for width=0")
	}
}

func TestGenerateReproducible(t *testing.T) {
	config := GenConfig{Seed: 42, Width: 5, Height: 5, Octaves: 4, Lacunarity: 2.0, Gain: 0.5, Scale: 10.0}
	biomes := []BiomeDef{
		{ID: "test", MinHeight: -1.0, MaxHeight: 1.0, MinHumidity: 0.0, MaxHumidity: 1.0, MinTemperature: 0.0, MaxTemperature: 1.0},
	}
	wm1, err := Generate(config.Width, config.Height, config, biomes)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	wm2, err := Generate(config.Width, config.Height, config, biomes)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	for i := range wm1.Tiles {
		if wm1.Tiles[i].Height != wm2.Tiles[i].Height {
			t.Errorf("tile %d height differs: %v vs %v", i, wm1.Tiles[i].Height, wm2.Tiles[i].Height)
		}
		if wm1.Tiles[i].Humidity != wm2.Tiles[i].Humidity {
			t.Errorf("tile %d humidity differs: %v vs %v", i, wm1.Tiles[i].Humidity, wm2.Tiles[i].Humidity)
		}
		if wm1.Tiles[i].Temperature != wm2.Tiles[i].Temperature {
			t.Errorf("tile %d temperature differs: %v vs %v", i, wm1.Tiles[i].Temperature, wm2.Tiles[i].Temperature)
		}
		if wm1.Tiles[i].BiomeID != wm2.Tiles[i].BiomeID {
			t.Errorf("tile %d biome differs: %q vs %q", i, wm1.Tiles[i].BiomeID, wm2.Tiles[i].BiomeID)
		}
	}
}

func TestGenerateTemperatureModulated(t *testing.T) {
	// Temperature should vary with height, not be independent
	config := GenConfig{Seed: 42, Width: 8, Height: 8, Octaves: 4, Lacunarity: 2.0, Gain: 0.5, Scale: 10.0}
	biomes := []BiomeDef{
		{ID: "test", MinHeight: -1.0, MaxHeight: 1.0, MinHumidity: 0.0, MaxHumidity: 1.0, MinTemperature: 0.0, MaxTemperature: 1.0},
	}
	wm, err := Generate(config.Width, config.Height, config, biomes)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	// Find tiles with different heights and check that temperatures differ accordingly
	var foundDiff bool
	for i := 1; i < len(wm.Tiles); i++ {
		if wm.Tiles[i].Height != wm.Tiles[0].Height && wm.Tiles[i].Temperature != wm.Tiles[0].Temperature {
			foundDiff = true
			break
		}
	}
	if !foundDiff {
		t.Error("expected temperature to vary with height differences")
	}
}
