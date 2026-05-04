package main

import (
	"os"
	"os/exec"
	"testing"
	"testing/fstest"
	"time"

	"github.com/marco/evociv-rl/internal/data"
	"github.com/marco/evociv-rl/internal/world/gen"
)

func TestRunNoPanic(t *testing.T) {
	done := make(chan struct{})
	var runErr error
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("run() panicked: %v", r)
			}
		}()
		_ = os.Chdir("../..")
		runErr = run()
	}()
	select {
	case <-done:
		_ = runErr // acceptable to error when no TTY
	case <-time.After(2 * time.Second):
		t.Error("run() timed out")
	}
}

func TestMainBuild(t *testing.T) {
	// Try building from repo root first, then from cmd/evociv dir
	cmd := exec.Command("go", "build", "./cmd/evociv")
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback: maybe running from cmd/evociv directory
		cmd = exec.Command("go", "build", ".")
		out, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("build failed: %v\n%s", err, out)
		}
	}
}

func TestMainInitRuns(t *testing.T) {
	fsys := fstest.MapFS{
		"data/gen-config.yaml": &fstest.MapFile{
			Data: []byte(`kind: gen-config
data:
  seed: 42
  width: 16
  height: 16
  octaves: 4
  lacunarity: 2.0
  gain: 0.5
  scale: 10.0
`),
		},
		"data/biomes.yaml": &fstest.MapFile{
			Data: []byte(`kind: biomes
data:
  - id: test
    name: Test
    symbol: "."
    color: "#FFFFFF"
    minHeight: -1.0
    maxHeight: 1.0
    minHumidity: 0.0
    maxHumidity: 1.0
    minTemperature: 0.0
    maxTemperature: 1.0
`),
		},
	}

	loader := data.NewLoader(fsys)
	registry := data.NewRegistry()
	if err := loader.LoadAll("data", registry); err != nil {
		t.Fatalf("load data: %v", err)
	}

	cfg, err := gen.LoadGenConfig("data/gen-config.yaml", fsys)
	if err != nil {
		t.Fatalf("load gen config: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate config: %v", err)
	}

	biomes, err := gen.LoadBiomes(registry)
	if err != nil {
		t.Fatalf("load biomes: %v", err)
	}
	if len(biomes) == 0 {
		t.Fatal("expected at least one biome")
	}
}

func TestGenerateWithConfig(t *testing.T) {
	fsys := fstest.MapFS{
		"gen-config.yaml": &fstest.MapFile{
			Data: []byte(`kind: gen-config
data:
  seed: 42
  width: 8
  height: 8
  octaves: 4
  lacunarity: 2.0
  gain: 0.5
  scale: 10.0
`),
		},
	}

	cfg, err := gen.LoadGenConfig("gen-config.yaml", fsys)
	if err != nil {
		t.Fatalf("load gen config: %v", err)
	}

	biomes := []gen.BiomeDef{
		{ID: "test", MinHeight: -1.0, MaxHeight: 1.0, MinHumidity: 0.0, MaxHumidity: 1.0, MinTemperature: 0.0, MaxTemperature: 1.0},
	}

	wm, err := gen.Generate(cfg.Width, cfg.Height, cfg, biomes)
	if err != nil {
		t.Fatalf("generate world: %v", err)
	}
	if len(wm.Tiles) != 64 {
		t.Errorf("expected 64 tiles, got %d", len(wm.Tiles))
	}
	for i, tile := range wm.Tiles {
		if tile.BiomeID == "" {
			t.Errorf("tile %d has empty BiomeID", i)
		}
	}
}
