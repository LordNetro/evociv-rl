package tilemap

import (
	"testing"
)

func TestLayer_Order(t *testing.T) {
	// Layer enum should have correct order for rendering priority
	if LayerTerrain != 0 {
		t.Errorf("LayerTerrain should be 0, got %d", LayerTerrain)
	}
	if LayerBuilding != 1 {
		t.Errorf("LayerBuilding should be 1, got %d", LayerBuilding)
	}
	if LayerItem != 2 {
		t.Errorf("LayerItem should be 2, got %d", LayerItem)
	}
	if LayerCreature != 3 {
		t.Errorf("LayerCreature should be 3, got %d", LayerCreature)
	}
	if LayerUI != 4 {
		t.Errorf("LayerUI should be 4, got %d", LayerUI)
	}
}

func TestCellType_Order(t *testing.T) {
	// CellType enum should have correct order for interior cell types
	if CellFloor != 0 {
		t.Errorf("CellFloor should be 0, got %d", CellFloor)
	}
	if CellWall != 1 {
		t.Errorf("CellWall should be 1, got %d", CellWall)
	}
	if CellDoor != 2 {
		t.Errorf("CellDoor should be 2, got %d", CellDoor)
	}
	if CellCorridor != 3 {
		t.Errorf("CellCorridor should be 3, got %d", CellCorridor)
	}
}

func TestTile_DefaultValues(t *testing.T) {
	// NewTile returns Tile with all zero bytes (default)
	tile := Tile{}

	if tile.Terrain != 0 {
		t.Errorf("Terrain should be 0 (zero byte), got %d", tile.Terrain)
	}
	if tile.Building != 0 {
		t.Errorf("Building should be 0 (zero byte), got %d", tile.Building)
	}
	if tile.Item != 0 {
		t.Errorf("Item should be 0 (zero byte), got %d", tile.Item)
	}
	if tile.Creature != 0 {
		t.Errorf("Creature should be 0 (zero byte), got %d", tile.Creature)
	}
	if tile.Fog != 0 {
		t.Errorf("Fog should be 0 (zero byte), got %d", tile.Fog)
	}
}

func TestTile_Assignment(t *testing.T) {
	// Tile struct supports 5-layer cell assignment
	tile := Tile{
		Terrain:  '.',
		Building: '+',
		Item:     '*',
		Creature: '@',
		Fog:      ' ',
	}

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
	if tile.Fog != ' ' {
		t.Errorf("Fog should be ' ', got %c", tile.Fog)
	}
}

func TestTile_Comparison(t *testing.T) {
	// Tiles can be compared for equality
	tile1 := Tile{Terrain: '.', Building: '+'}
	tile2 := Tile{Terrain: '.', Building: '+'}
	tile3 := Tile{Terrain: '#', Building: '+'}

	if tile1 != tile2 {
		t.Errorf("tile1 and tile2 should be equal")
	}
	if tile1 == tile3 {
		t.Errorf("tile1 and tile3 should not be equal")
	}
}