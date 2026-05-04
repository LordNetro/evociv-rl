package world

import (
	"testing"
)

func TestNewWorldMap(t *testing.T) {
	wm := NewWorldMap(10, 20)
	if wm == nil {
		t.Fatal("expected non-nil WorldMap")
	}
	if wm.Width != 10 {
		t.Errorf("width = %d, want 10", wm.Width)
	}
	if wm.Height != 20 {
		t.Errorf("height = %d, want 20", wm.Height)
	}
	if wm.Tiles == nil {
		t.Error("expected Tiles to be non-nil")
	}
	if len(wm.Tiles) != 10*20 {
		t.Errorf("tiles length = %d, want 200", len(wm.Tiles))
	}
}

func TestTileAt(t *testing.T) {
	wm := NewWorldMap(10, 10)

	// Default zero values
	tile := wm.TileAt(3, 7)
	if tile == nil {
		t.Fatal("expected non-nil tile")
	}
	if tile.Height != 0 {
		t.Errorf("default height = %v, want 0", tile.Height)
	}

	// Write and read back
	tile.Height = 5.5
	tile.BiomeID = "plains"
	readBack := wm.TileAt(3, 7)
	if readBack.Height != 5.5 {
		t.Errorf("height after write = %v, want 5.5", readBack.Height)
	}
	if readBack.BiomeID != "plains" {
		t.Errorf("biome after write = %q, want plains", readBack.BiomeID)
	}

	// Different tile should be unaffected
	other := wm.TileAt(0, 0)
	if other.Height != 0 {
		t.Errorf("unaffected tile height = %v, want 0", other.Height)
	}
}

func TestTileAtIndex(t *testing.T) {
	wm := NewWorldMap(10, 10)
	// TileAt(3,7) should index at y*width+x = 7*10+3 = 73
	tile := wm.TileAt(3, 7)
	if tile == nil {
		t.Fatal("expected non-nil tile")
	}
	tile.Height = 99.0

	// Verify via direct slice access
	if wm.Tiles[73].Height != 99.0 {
		t.Errorf(" Tiles[73].Height = %v, want 99.0", wm.Tiles[73].Height)
	}
}

func TestTileAtNil(t *testing.T) {
	wm := NewWorldMap(5, 5)
	cases := []struct {
		x, y int
	}{
		{-1, 0},
		{0, -1},
		{5, 0},
		{0, 5},
		{100, 100},
	}
	for _, c := range cases {
		tile := wm.TileAt(c.x, c.y)
		if tile != nil {
			t.Errorf("TileAt(%d,%d) = %v, want nil", c.x, c.y, tile)
		}
	}
}

func TestInBounds(t *testing.T) {
	wm := NewWorldMap(10, 20)

	// Corners and edges should be in bounds
	inBoundsCases := []struct{ x, y int }{
		{0, 0},
		{9, 0},
		{0, 19},
		{9, 19},
		{5, 10},
	}
	for _, c := range inBoundsCases {
		if !wm.InBounds(c.x, c.y) {
			t.Errorf("InBounds(%d,%d) = false, want true", c.x, c.y)
		}
	}

	// Out of bounds
	outOfBoundsCases := []struct{ x, y int }{
		{-1, 0},
		{0, -1},
		{10, 0},
		{0, 20},
		{-1, -1},
		{100, 100},
	}
	for _, c := range outOfBoundsCases {
		if wm.InBounds(c.x, c.y) {
			t.Errorf("InBounds(%d,%d) = true, want false", c.x, c.y)
		}
	}
}
