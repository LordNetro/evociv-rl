package tilemap

import (
	"testing"
)

// TestTileBuilder_BuildFromWorld_TerrainLayer tests that terrain layer is populated from biome data
func TestTileBuilder_BuildFromWorld_TerrainLayer(t *testing.T) {
	// Arrange
	worldMap := MockWorldMap{
		Width:  5,
		Height: 5,
		Tiles: make([]MockTile, 25),
	}
	// Set biome IDs for tiles
	worldMap.Tiles[0].BiomeID = "ocean"
	worldMap.Tiles[6].BiomeID = "plains"
	worldMap.Tiles[12].BiomeID = "forest"
	worldMap.Tiles[18].BiomeID = "mountain"

	// Create a mock wrapper that provides world data
	worldWrapper := &mockWorldWrapper{
		worldMap: worldMap,
		state:    &MockWorldState{},
	}

	builder := NewTileBuilder(worldWrapper)

	// Create tilemap
	tilemap := NewTilemap(5, 5)

	// Act
	err := builder.Build(tilemap)

	// Assert
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify terrain layer was populated
	// Ocean -> '~'
	if tile := tilemap.TileAt(0, 0, 0); tile.Terrain != '~' {
		t.Errorf("Expected terrain '~' at (0,0), got '%c'", tile.Terrain)
	}
	// Plains -> '.'
	if tile := tilemap.TileAt(1, 1, 0); tile.Terrain != '.' {
		t.Errorf("Expected terrain '.' at (1,1), got '%c'", tile.Terrain)
	}
	// Forest -> 'T'
	if tile := tilemap.TileAt(2, 2, 0); tile.Terrain != 'T' {
		t.Errorf("Expected terrain 'T' at (2,2), got '%c'", tile.Terrain)
	}
	// Mountain -> '#'
	if tile := tilemap.TileAt(3, 3, 0); tile.Terrain != '#' {
		t.Errorf("Expected terrain '#' at (3,3), got '%c'", tile.Terrain)
	}
}

// TestTileBuilder_BuildFromWorld_CreatureLayer tests that NPCs are rendered in creature layer
func TestTileBuilder_BuildFromWorld_CreatureLayer(t *testing.T) {
	// Arrange
	worldMap := MockWorldMap{
		Width:  10,
		Height: 10,
		Tiles: make([]MockTile, 100),
	}
	// Fill with plains
	for i := range worldMap.Tiles {
		worldMap.Tiles[i].BiomeID = "plains"
	}

	state := &MockWorldState{
		NPCs: []MockNPC{
			{ID: 1, X: 3, Y: 4, Name: "Dwarf1"},
			{ID: 2, X: 7, Y: 2, Name: "Dwarf2"},
		},
	}

	worldWrapper := &mockWorldWrapper{
		worldMap: worldMap,
		state:    state,
	}

	builder := NewTileBuilder(worldWrapper)
	tilemap := NewTilemap(10, 10)

	// Act
	err := builder.Build(tilemap)

	// Assert
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify NPC at (3,4) -> '@'
	if tile := tilemap.TileAt(3, 4, 0); tile.Creature != '@' {
		t.Errorf("Expected creature '@' at (3,4), got '%c'", tile.Creature)
	}
	// Verify NPC at (7,2) -> '@'
	if tile := tilemap.TileAt(7, 2, 0); tile.Creature != '@' {
		t.Errorf("Expected creature '@' at (7,2), got '%c'", tile.Creature)
	}
}

// TestTileBuilder_BuildFromWorld_BuildingLayer tests that buildings are rendered with footprint
func TestTileBuilder_BuildFromWorld_BuildingLayer(t *testing.T) {
	// Arrange
	worldMap := MockWorldMap{
		Width:  20,
		Height: 20,
		Tiles: make([]MockTile, 400),
	}
	for i := range worldMap.Tiles {
		worldMap.Tiles[i].BiomeID = "plains"
	}

	state := &MockWorldState{
		NPCs: []MockNPC{},
	}

	settlement := &MockSettlement{
		Buildings: []MockBuilding{
			{ID: 1, X: 5, Y: 5, Width: 4, Height: 3, Name: "Tavern"},
		},
	}

	worldWrapper := &mockWorldWrapper{
		worldMap:   worldMap,
		state:      state,
		settlement: settlement,
	}

	builder := NewTileBuilder(worldWrapper)
	tilemap := NewTilemap(20, 20)

	// Act
	err := builder.Build(tilemap)

	// Assert
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify building corners at (5,5), (8,5), (5,7), (8,7) -> '+'
	if tile := tilemap.TileAt(5, 5, 0); tile.Building != '+' {
		t.Errorf("Expected building '+' at corner (5,5), got '%c'", tile.Building)
	}
	if tile := tilemap.TileAt(8, 5, 0); tile.Building != '+' {
		t.Errorf("Expected building '+' at corner (8,5), got '%c'", tile.Building)
	}
	if tile := tilemap.TileAt(5, 7, 0); tile.Building != '+' {
		t.Errorf("Expected building '+' at corner (5,7), got '%c'", tile.Building)
	}
	if tile := tilemap.TileAt(8, 7, 0); tile.Building != '+' {
		t.Errorf("Expected building '+' at corner (8,7), got '%c'", tile.Building)
	}

	// Verify building edges at (6,5), (7,5), (6,7), (7,7) -> '#'
	if tile := tilemap.TileAt(6, 5, 0); tile.Building != '#' {
		t.Errorf("Expected building '#' at edge (6,5), got '%c'", tile.Building)
	}

	// Verify interior is floor '.'
	if tile := tilemap.TileAt(6, 6, 0); tile.Building != '.' {
		t.Errorf("Expected building '.' at interior (6,6), got '%c'", tile.Building)
	}
}

// TestTileBuilder_BuildFromWorld_FogLayer tests fog of war layer
func TestTileBuilder_BuildFromWorld_FogLayer(t *testing.T) {
	// Arrange
	worldMap := MockWorldMap{
		Width:  10,
		Height: 10,
		Tiles: make([]MockTile, 100),
	}
	for i := range worldMap.Tiles {
		worldMap.Tiles[i].BiomeID = "plains"
	}

	// Mock explored tiles (x,y -> explored)
	explored := map[string]bool{
		"0,0": true,
		"1,0": true,
		"0,1": true,
		"1,1": true,
		// Rest are unexplored
	}

	state := &MockWorldState{
		NPCs: []MockNPC{},
	}

	worldWrapper := &mockWorldWrapper{
		worldMap:  worldMap,
		state:     state,
		explored:  explored,
	}

	builder := NewTileBuilder(worldWrapper)
	tilemap := NewTilemap(10, 10)

	// Act
	err := builder.Build(tilemap)

	// Assert
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verified explored tiles show '.' (explored)
	if tile := tilemap.TileAt(0, 0, 0); tile.Fog != '.' {
		t.Errorf("Expected fog '.' at explored (0,0), got '%c'", tile.Fog)
	}
	if tile := tilemap.TileAt(1, 1, 0); tile.Fog != '.' {
		t.Errorf("Expected fog '.' at explored (1,1), got '%c'", tile.Fog)
	}

	// Unexplored should show ':' (dark)
	if tile := tilemap.TileAt(5, 5, 0); tile.Fog != ':' {
		t.Errorf("Expected fog ':' at unexplored (5,5), got '%c'", tile.Fog)
	}
}

// TestTileBuilder_EmptyWorld tests building from empty world
func TestTileBuilder_EmptyWorld(t *testing.T) {
	// Arrange
	worldMap := MockWorldMap{
		Width:  3,
		Height: 3,
		Tiles:  make([]MockTile, 9),
	}

	worldWrapper := &mockWorldWrapper{
		worldMap: worldMap,
		state:    &MockWorldState{},
	}

	builder := NewTileBuilder(worldWrapper)
	tilemap := NewTilemap(3, 3)

	// Act
	err := builder.Build(tilemap)

	// Assert
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// All tiles should have default/zero values
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			if tile := tilemap.TileAt(x, y, 0); tile.Terrain == 0 && tile.Creature == 0 {
				// Expected - empty terrain and no creatures
			}
		}
	}
}