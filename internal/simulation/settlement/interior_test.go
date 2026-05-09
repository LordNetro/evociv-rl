package settlement

import (
	"testing"
)

// TestCellTypeValues verifies CellType enum values match spec
func TestCellTypeValues(t *testing.T) {
	t.Run("CellFloor is zero", func(t *testing.T) {
		if CellFloor != 0 {
			t.Errorf("CellFloor should be 0, got %d", CellFloor)
		}
	})
	t.Run("CellWall is first", func(t *testing.T) {
		if CellWall != 1 {
			t.Errorf("CellWall should be 1, got %d", CellWall)
		}
	})
	t.Run("CellDoor is second", func(t *testing.T) {
		if CellDoor != 2 {
			t.Errorf("CellDoor should be 2, got %d", CellDoor)
		}
	})
	t.Run("CellForbidden is third", func(t *testing.T) {
		if CellForbidden != 3 {
			t.Errorf("CellForbidden should be 3, got %d", CellForbidden)
		}
	})
}

// TestBuildingInteriorCreation verifies BuildingInterior struct creation
func TestBuildingInteriorCreation(t *testing.T) {
	t.Run("creates interior with correct dimensions", func(t *testing.T) {
		bi := BuildingInterior{
			Width:         5,
			Height:        4,
			MaxWorkers:    3,
			WorkersInside: 0,
			BuildingSeed:  12345,
		}
		if bi.Width != 5 {
			t.Errorf("Width should be 5, got %d", bi.Width)
		}
		if bi.Height != 4 {
			t.Errorf("Height should be 4, got %d", bi.Height)
		}
		if bi.MaxWorkers != 3 {
			t.Errorf("MaxWorkers should be 3, got %d", bi.MaxWorkers)
		}
		if bi.WorkersInside != 0 {
			t.Errorf("WorkersInside should be 0, got %d", bi.WorkersInside)
		}
		if bi.BuildingSeed != 12345 {
			t.Errorf("BuildingSeed should be 12345, got %d", bi.BuildingSeed)
		}
	})
	t.Run("grid is initially nil", func(t *testing.T) {
		bi := BuildingInterior{}
		if bi.Grid != nil {
			t.Error("Grid should be nil initially")
		}
	})
	t.Run("doors is initially nil", func(t *testing.T) {
		bi := BuildingInterior{}
		if bi.Doors != nil {
			t.Error("Doors should be nil initially")
		}
	})
}

// TestDoorPosition verifies DoorPosition struct
func TestDoorPosition(t *testing.T) {
	t.Run("creates door with grid coordinates", func(t *testing.T) {
		door := DoorPosition{
			GridX:    2,
			GridY:    3,
			WorldX:   100,
			WorldY:   200,
		}
		if door.GridX != 2 {
			t.Errorf("GridX should be 2, got %d", door.GridX)
		}
		if door.GridY != 3 {
			t.Errorf("GridY should be 3, got %d", door.GridY)
		}
		if door.WorldX != 100 {
			t.Errorf("WorldX should be 100, got %d", door.WorldX)
		}
		if door.WorldY != 200 {
			t.Errorf("WorldY should be 200, got %d", door.WorldY)
		}
	})
}

// TestBuildingInteriorIDNotZero verifies BuildingInteriorID is registered
func TestBuildingInteriorIDNotZero(t *testing.T) {
	t.Run("BuildingInteriorID is not zero", func(t *testing.T) {
		if BuildingInteriorID == 0 {
			t.Error("BuildingInteriorID should not be zero")
		}
	})
}