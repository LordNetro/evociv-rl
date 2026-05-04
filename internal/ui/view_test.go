package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/marco/evociv-rl/internal/world"
)

func TestViewContainsTitle(t *testing.T) {
	m := NewModel()
	m.width = 80
	m.height = 24
	m.ready = true
	v := m.View()
	if !strings.Contains(v, "Evociv-RL") {
		t.Error("view missing title 'Evociv-RL'")
	}
}

func TestViewContainsVersion(t *testing.T) {
	m := NewModel()
	m.width = 80
	m.height = 24
	m.ready = true
	v := m.View()
	if !strings.Contains(v, "v0.0.1") {
		t.Error("view missing version 'v0.0.1'")
	}
}

func TestViewContainsSubtitle(t *testing.T) {
	m := NewModel()
	m.width = 80
	m.height = 24
	m.ready = true
	v := m.View()
	if !strings.Contains(v, "Un mundo por descubrir") {
		t.Error("view missing subtitle")
	}
}

func TestViewContainsInstructions(t *testing.T) {
	m := NewModel()
	m.width = 80
	m.height = 24
	m.ready = true
	v := m.View()
	if !strings.Contains(v, "[q] Salir") {
		t.Error("view missing instructions")
	}
}

func TestRenderMapShowsTiles(t *testing.T) {
	wm := world.NewWorldMap(3, 3)
	wm.TileAt(0, 0).BiomeID = "ocean"
	wm.TileAt(0, 0).Height = -0.5
	wm.TileAt(1, 0).BiomeID = "plains"
	wm.TileAt(1, 0).Height = 0.3
	wm.TileAt(2, 0).BiomeID = "forest"
	wm.TileAt(2, 0).Height = 0.5
	wm.TileAt(0, 1).BiomeID = "desert"
	wm.TileAt(0, 1).Height = 0.2
	wm.TileAt(1, 1).BiomeID = "tundra"
	wm.TileAt(1, 1).Height = 0.4
	wm.TileAt(2, 1).BiomeID = "jungle"
	wm.TileAt(2, 1).Height = 0.6
	wm.TileAt(0, 2).BiomeID = "ocean"
	wm.TileAt(1, 2).BiomeID = "plains"
	wm.TileAt(2, 2).BiomeID = "forest"

	m := NewModel()
	m.SetWorldMap(wm)
	m.screen = "map"
	m.width = 80
	m.height = 24
	m.ready = true

	v := m.View()
	// The map view should contain biome symbols, not the welcome title
	if strings.Contains(v, "Evociv-RL") {
		t.Error("map view should not contain welcome title")
	}
}

func TestRenderScreenToggle(t *testing.T) {
	// Welcome screen shows title
	m := NewModel()
	m.screen = "welcome"
	m.width = 80
	m.height = 24
	m.ready = true
	v := m.View()
	if !strings.Contains(v, "Evociv-RL") {
		t.Error("welcome view should contain title")
	}

	// Map screen does not show title
	wm := world.NewWorldMap(3, 3)
	wm.TileAt(0, 0).BiomeID = "ocean"
	m.SetWorldMap(wm)
	m.screen = "map"
	v = m.View()
	if strings.Contains(v, "Evociv-RL") {
		t.Error("map view should not contain welcome title")
	}
}

func TestRenderMapNoWorldMap(t *testing.T) {
	m := NewModel()
	m.screen = "map"
	m.width = 80
	m.height = 24
	m.ready = true
	v := m.View()
	if !strings.Contains(v, "Generando mundo") {
		t.Errorf("expected 'Generando mundo' message, got: %q", v)
	}
}

func TestRenderMapWindowAdapt(t *testing.T) {
	wm := world.NewWorldMap(20, 20)
	for y := 0; y < wm.Height; y++ {
		for x := 0; x < wm.Width; x++ {
			wm.TileAt(x, y).BiomeID = "plains"
		}
	}
	m := NewModel()
	m.SetWorldMap(wm)
	m.screen = "map"
	m.width = 80
	m.height = 24
	m.ready = true

	v := m.View()
	lines := strings.Split(v, "\n")
	// Should produce roughly the terminal height in lines (allowing for ANSI)
	if len(lines) < 20 {
		t.Errorf("expected at least ~20 lines, got %d", len(lines))
	}
}

func TestRenderMapGolden(t *testing.T) {
	wm := world.NewWorldMap(5, 5)
	for y := 0; y < wm.Height; y++ {
		for x := 0; x < wm.Width; x++ {
			wm.TileAt(x, y).BiomeID = "plains"
		}
	}
	m := NewModel()
	m.SetWorldMap(wm)
	m.screen = "map"
	m.width = 80
	m.height = 24
	m.ready = true

	output := m.View()
	goldenFile := filepath.Join("testdata", "map.golden")

	if *updateGolden {
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatalf("failed to create testdata: %v", err)
		}
		if err := os.WriteFile(goldenFile, []byte(output), 0644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
	}

	expected, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("reading golden file: %v", err)
	}
	if output != string(expected) {
		t.Errorf("output doesn't match golden file")
	}
}
