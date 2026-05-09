package tilemap

import (
	"errors"
	"testing"
)

// TestInteriorRenderer_RenderInterior_GridSize tests that RenderInterior returns correct grid size
func TestInteriorRenderer_RenderInterior_GridSize(t *testing.T) {
	// Arrange
	state := &MockInteriorWorldState{
		Buildings: []MockBuildingInterior{
			{
				BuildingID: 1,
				Width:      5,
				Height:     4,
				Cells:      createMockInteriorGrid(5, 4),
			},
		},
		NPCsInBldg: map[uint64][]NPCInfo{},
	}

	renderer := NewInteriorRenderer(state.toProvider())

	// Act
	grid, err := renderer.RenderInterior(1, 0)

	// Assert
	if err != nil {
		t.Fatalf("RenderInterior failed: %v", err)
	}

	if len(grid) != 4 { // height
		t.Errorf("Expected grid height 4, got %d", len(grid))
	}
	if len(grid) > 0 && len(grid[0]) != 5 { // width
		t.Errorf("Expected grid width 5, got %d", len(grid[0]))
	}
}

// TestInteriorRenderer_RenderInterior_NPCPlacement tests NPCs are placed at correct grid positions
func TestInteriorRenderer_RenderInterior_NPCPlacement(t *testing.T) {
	// Arrange
	state := &MockInteriorWorldState{
		Buildings: []MockBuildingInterior{
			{
				BuildingID: 1,
				Width:      5,
				Height:     4,
				Cells:      createMockInteriorGrid(5, 4),
			},
		},
		NPCsInBldg: map[uint64][]NPCInfo{
			1: {
				{ID: 1, X: 2, Y: 1, Name: "Dwarf1"},
				{ID: 2, X: 3, Y: 2, Name: "Dwarf2"},
			},
		},
	}

	renderer := NewInteriorRenderer(state.toProvider())

	// Act
	grid, err := renderer.RenderInterior(1, 0)

	// Assert
	if err != nil {
		t.Fatalf("RenderInterior failed: %v", err)
	}

	// NPC at (2,1) should show '@'
	if grid[1][2] != CellType('@') {
		t.Errorf("Expected NPC '@' at (2,1), got %v", grid[1][2])
	}
	// NPC at (3,2) should show '@'
	if grid[2][3] != CellType('@') {
		t.Errorf("Expected NPC '@' at (3,2), got %v", grid[2][3])
	}
}

// TestInteriorRenderer_RenderInterior_InvalidBuildingID tests that invalid building ID returns error
func TestInteriorRenderer_RenderInterior_InvalidBuildingID(t *testing.T) {
	// Arrange
	state := &MockInteriorWorldState{
		Buildings:  []MockBuildingInterior{},
		NPCsInBldg: map[uint64][]NPCInfo{},
	}

	renderer := NewInteriorRenderer(state.toProvider())

	// Act
	_, err := renderer.RenderInterior(999, 0)

	// Assert
	if err == nil {
		t.Error("Expected error for invalid building ID, got nil")
	}
}

// TestInteriorRenderer_RenderInterior_InteriorCells tests interior cells are correctly rendered
func TestInteriorRenderer_RenderInterior_InteriorCells(t *testing.T) {
	// Arrange
	cells := [][]CellType{
		{CellWall, CellWall, CellWall, CellWall, CellWall},
		{CellWall, CellFloor, CellDoor, CellFloor, CellWall},
		{CellWall, CellFloor, CellCorridor, CellFloor, CellWall},
		{CellWall, CellWall, CellWall, CellWall, CellWall},
	}

	state := &MockInteriorWorldState{
		Buildings: []MockBuildingInterior{
			{
				BuildingID: 1,
				Width:      5,
				Height:     4,
				Cells:      cells,
			},
		},
		NPCsInBldg: map[uint64][]NPCInfo{},
	}

	renderer := NewInteriorRenderer(state.toProvider())

	// Act
	grid, err := renderer.RenderInterior(1, 0)

	// Assert
	if err != nil {
		t.Fatalf("RenderInterior failed: %v", err)
	}

	// Verify walls
	if grid[0][0] != CellWall {
		t.Errorf("Expected wall at (0,0), got %v", grid[0][0])
	}
	// Verify floor
	if grid[1][1] != CellFloor {
		t.Errorf("Expected floor at (1,1), got %v", grid[1][1])
	}
	// Verify door
	if grid[1][2] != CellDoor {
		t.Errorf("Expected door at (1,2), got %v", grid[1][2])
	}
	// Verify corridor
	if grid[2][2] != CellCorridor {
		t.Errorf("Expected corridor at (2,2), got %v", grid[2][2])
	}
}

// TestInteriorRenderer_GetNPCsInBuilding tests NPC query for specific building
func TestInteriorRenderer_GetNPCsInBuilding(t *testing.T) {
	// Arrange
	state := &MockInteriorWorldState{
		Buildings: []MockBuildingInterior{
			{
				BuildingID: 1,
				Width:      3,
				Height:     3,
				Cells:      createMockInteriorGrid(3, 3),
			},
		},
		NPCsInBldg: map[uint64][]NPCInfo{
			1: {
				{ID: 1, X: 1, Y: 1, Name: "Worker1"},
			},
			2: {
				{ID: 2, X: 1, Y: 1, Name: "Worker2"},
			},
		},
	}

	renderer := NewInteriorRenderer(state.toProvider())

	// Act
	npcs := renderer.GetNPCsInBuilding(1)

	// Assert
	if len(npcs) != 1 {
		t.Errorf("Expected 1 NPC in building 1, got %d", len(npcs))
	}

	// Building 2 should have 1 NPC
	npcs2 := renderer.GetNPCsInBuilding(2)
	if len(npcs2) != 1 {
		t.Errorf("Expected 1 NPC in building 2, got %d", len(npcs2))
	}

	// Building 999 should have 0 NPCs
	npcs3 := renderer.GetNPCsInBuilding(999)
	if len(npcs3) != 0 {
		t.Errorf("Expected 0 NPCs in building 999, got %d", len(npcs3))
	}
}

// TestInteriorRenderer_ErrorForNegativeBuilding tests error for negative building ID
func TestInteriorRenderer_ErrorForNegativeBuilding(t *testing.T) {
	// Arrange
	state := &MockInteriorWorldState{
		Buildings:  []MockBuildingInterior{},
		NPCsInBldg: map[uint64][]NPCInfo{},
	}

	renderer := NewInteriorRenderer(state.toProvider())

	// Act
	_, err := renderer.RenderInterior(0, 0)

	// Assert
	if err == nil {
		t.Error("Expected error for building ID 0, got nil")
	}
	if !errors.Is(err, ErrInvalidBuildingID) {
		t.Errorf("Expected ErrInvalidBuildingID, got %v", err)
	}
}

// createMockInteriorGrid creates a simple interior grid with floor everywhere
func createMockInteriorGrid(width, height int) [][]CellType {
	grid := make([][]CellType, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]CellType, width)
		for x := 0; x < width; x++ {
			// Border is wall, interior is floor
			if x == 0 || x == width-1 || y == 0 || y == height-1 {
				grid[y][x] = CellWall
			} else {
				grid[y][x] = CellFloor
			}
		}
	}
	return grid
}