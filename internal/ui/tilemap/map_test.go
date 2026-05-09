package tilemap

import (
	"testing"
)

func TestNewTilemap_Dimensions(t *testing.T) {
	// NewTilemap creates a map with specified dimensions
	m := NewTilemap(100, 50)

	if m.Width() != 100 {
		t.Errorf("Width should be 100, got %d", m.Width())
	}
	if m.Height() != 50 {
		t.Errorf("Height should be 50, got %d", m.Height())
	}
}

func TestNewTilemap_InitialZLevels(t *testing.T) {
	// NewTilemap creates map with Z=0 level by default
	m := NewTilemap(10, 10)

	levels := m.ZLevels()
	if len(levels) != 1 {
		t.Errorf("Should have 1 Z-level, got %d", len(levels))
	}
	if _, ok := levels[0]; !ok {
		t.Errorf("Z-level 0 should exist")
	}
}

func TestTilemap_TileAt_InBounds(t *testing.T) {
	m := NewTilemap(10, 10)
	m.SetTile(5, 5, 0, LayerTerrain, '.')

	tile := m.TileAt(5, 5, 0)
	if tile == nil {
		t.Fatal("TileAt should not return nil for in-bounds access")
	}
	if tile.Terrain != '.' {
		t.Errorf("Terrain should be '.', got %c", tile.Terrain)
	}
}

func TestTilemap_TileAt_OutOfBounds(t *testing.T) {
	// TileAt returns nil for out of bounds coordinates
	m := NewTilemap(10, 10)

	// Negative coordinates
	if m.TileAt(-1, 5, 0) != nil {
		t.Error("TileAt should return nil for negative x")
	}
	if m.TileAt(5, -1, 0) != nil {
		t.Error("TileAt should return nil for negative y")
	}

	// Beyond width/height
	if m.TileAt(10, 5, 0) != nil {
		t.Error("TileAt should return nil for x >= width")
	}
	if m.TileAt(5, 10, 0) != nil {
		t.Error("TileAt should return nil for y >= height")
	}
}

func TestTilemap_TileAt_InvalidZLevel(t *testing.T) {
	// TileAt returns nil for non-existent Z level
	m := NewTilemap(10, 10)

	if m.TileAt(5, 5, 1) != nil {
		t.Error("TileAt should return nil for non-existent Z level")
	}
}

func TestTilemap_SetTile_LayerAssignment(t *testing.T) {
	// SetTile assigns character to specific layer
	m := NewTilemap(10, 10)

	// Set different layers
	m.SetTile(3, 3, 0, LayerTerrain, '.')
	m.SetTile(3, 3, 0, LayerBuilding, '+')
	m.SetTile(3, 3, 0, LayerItem, '*')
	m.SetTile(3, 3, 0, LayerCreature, '@')

	tile := m.TileAt(3, 3, 0)
	if tile.Terrain != '.' {
		t.Errorf("Terrain should be '.', got %c", tile.Terrain)
	}
	if tile.Building != '+' {
		t.Errorf("Building should be '+', got %c", tile.Building)
	}
	if tile.Item != '*' {
		t.Errorf("Item should be '*', got %c", tile.Item)
	}
	if tile.Creature != '@' {
		t.Errorf("Creature should be '@', got %c", tile.Creature)
	}
}

func TestTilemap_SetTile_CreatesZLevel(t *testing.T) {
	// SetTile creates Z level if it doesn't exist
	m := NewTilemap(10, 10)

	m.SetTile(5, 5, 1, LayerTerrain, '#')

	levels := m.ZLevels()
	if len(levels) != 2 {
		t.Errorf("Should have 2 Z-levels, got %d", len(levels))
	}
	if _, ok := levels[1]; !ok {
		t.Errorf("Z-level 1 should exist after SetTile")
	}
}

func TestTilemap_SetZLevel_ExistingLevel(t *testing.T) {
	// SetZLevel returns existing level if already created
	m := NewTilemap(10, 10)

	// Create Z=1 level first
	m.SetTile(0, 0, 1, LayerTerrain, '.')

	// Get it back
	level := m.SetZLevel(1)
	if level == nil {
		t.Fatal("SetZLevel should return existing level")
	}
	if len(level) != 10 {
		t.Errorf("Level height should be 10, got %d", len(level))
	}
	if len(level[0]) != 10 {
		t.Errorf("Level width should be 10, got %d", len(level[0]))
	}
}

func TestTilemap_ZLevels_AllLevels(t *testing.T) {
	// ZLevels returns all created levels
	m := NewTilemap(10, 10)

	m.SetTile(0, 0, 0, LayerTerrain, '.')
	m.SetTile(0, 0, 1, LayerTerrain, '#')

	levels := m.ZLevels()
	if len(levels) != 2 {
		t.Errorf("Should have 2 Z-levels, got %d", len(levels))
	}
}

func TestTilemap_LayerPriority(t *testing.T) {
	// Tiles support 5-layer system with correct priority order
	m := NewTilemap(5, 5)

	// Set all layers at same position
	m.SetTile(2, 2, 0, LayerTerrain, '.')   // 0 - lowest priority
	m.SetTile(2, 2, 0, LayerBuilding, '#')  // 1
	m.SetTile(2, 2, 0, LayerItem, '*')       // 2
	m.SetTile(2, 2, 0, LayerCreature, '@')  // 3
	m.SetTile(2, 2, 0, LayerUI, 'X')       // 4 - highest priority

	tile := m.TileAt(2, 2, 0)
	if tile.Terrain != '.' {
		t.Errorf("Terrain should be '.', got %c", tile.Terrain)
	}
	if tile.Building != '#' {
		t.Errorf("Building should be '#', got %c", tile.Building)
	}
	if tile.Item != '*' {
		t.Errorf("Item should be '*', got %c", tile.Item)
	}
	if tile.Creature != '@' {
		t.Errorf("Creature should be '@', got %c", tile.Creature)
	}
	if tile.Fog != 0 {
		t.Errorf("Fog should be 0, got %c", tile.Fog)
	}
}