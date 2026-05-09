package settlement

import (
	"testing"

	"github.com/marco/evociv-rl/internal/ecs"
)

func TestIndoorPathfinderFindPath(t *testing.T) {
	pf := NewIndoorPathfinder()

	t.Run("nil grid returns error", func(t *testing.T) {
		_, err := pf.FindPath(nil, DoorPosition{GridX: 0, GridY: 0}, DoorPosition{GridX: 1, GridY: 1})
		if err == nil {
			t.Error("expected error for nil grid")
		}
	})

	t.Run("empty grid returns error", func(t *testing.T) {
		_, err := pf.FindPath([][]CellType{}, DoorPosition{GridX: 0, GridY: 0}, DoorPosition{GridX: 1, GridY: 1})
		if err == nil {
			t.Error("expected error for empty grid")
		}
	})
}

func TestIndoorPathfinderDirectNeighbors(t *testing.T) {
	pf := NewIndoorPathfinder()

	// Create a simple 5x5 grid with all floors
	grid := createFloorGrid(5, 5)

	t.Run("direct neighbor path", func(t *testing.T) {
		// Adjacent cells - simple path
		path, err := pf.FindPath(grid, DoorPosition{GridX: 2, GridY: 2}, DoorPosition{GridX: 3, GridY: 2})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(path) == 0 {
			t.Error("expected non-empty path")
		}
		if path[0].x != 2 || path[0].y != 2 {
			t.Errorf("path should start at (2,2), got (%d,%d)", path[0].x, path[0].y)
		}
		if path[len(path)-1].x != 3 || path[len(path)-1].y != 2 {
			t.Errorf("path should end at (3,2), got (%d,%d)", path[len(path)-1].x, path[len(path)-1].y)
		}
	})

	t.Run("same start and end returns single point path", func(t *testing.T) {
		path, err := pf.FindPath(grid, DoorPosition{GridX: 2, GridY: 2}, DoorPosition{GridX: 2, GridY: 2})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(path) != 1 {
			t.Errorf("expected path length 1, got %d", len(path))
		}
		if path[0].x != 2 || path[0].y != 2 {
			t.Errorf("path should be at (2,2), got (%d,%d)", path[0].x, path[0].y)
		}
	})

	t.Run("longer path through grid", func(t *testing.T) {
		path, err := pf.FindPath(grid, DoorPosition{GridX: 0, GridY: 0}, DoorPosition{GridX: 4, GridY: 4})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(path) < 2 {
			t.Error("expected path with multiple steps")
		}
		// First and last should be correct
		if path[0].x != 0 || path[0].y != 0 {
			t.Errorf("path should start at (0,0), got (%d,%d)", path[0].x, path[0].y)
		}
		last := path[len(path)-1]
		if last.x != 4 || last.y != 4 {
			t.Errorf("path should end at (4,4), got (%d,%d)", last.x, last.y)
		}
	})
}

func TestIndoorPathfinderObstacles(t *testing.T) {
	pf := NewIndoorPathfinder()

	t.Run("path blocked by wall returns error", func(t *testing.T) {
		// Create a grid with a wall barrier that blocks all paths
		// Grid:
		// F F F F F  <- can pass
		// F F W F F  <- wall in middle
		// F F F F F  <- can pass
		// Path from (0,0) to (4,3) requires going through (2,1) or (2,0)
		// Actually, let's create a grid that truly blocks all paths
		grid := [][]CellType{
			{CellFloor, CellFloor, CellFloor, CellFloor, CellFloor},
			{CellFloor, CellFloor, CellFloor, CellFloor, CellFloor},
			{CellWall, CellWall, CellWall, CellWall, CellWall}, // Full wall row
			{CellFloor, CellFloor, CellFloor, CellFloor, CellFloor},
			{CellFloor, CellFloor, CellFloor, CellFloor, CellFloor},
		}

		// Try to path from top to bottom - should fail due to wall row
		_, err := pf.FindPath(grid, DoorPosition{GridX: 2, GridY: 0}, DoorPosition{GridX: 2, GridY: 4})
		if err == nil {
			t.Error("expected error for blocked path through wall row")
		}
	})

	t.Run("path blocked by isolated target", func(t *testing.T) {
		// Grid with a room isolated by walls
		grid := [][]CellType{
			{CellFloor, CellFloor, CellWall, CellFloor, CellFloor},
			{CellFloor, CellFloor, CellWall, CellFloor, CellFloor},
			{CellFloor, CellFloor, CellFloor, CellFloor, CellFloor},
		}

		// Target isolated in the middle
		_, err := pf.FindPath(grid, DoorPosition{GridX: 0, GridY: 0}, DoorPosition{GridX: 2, GridY: 0})
		if err == nil {
			t.Error("expected error when target is isolated by walls")
		}
	})

	t.Run("wall position is not walkable", func(t *testing.T) {
		// Grid with walls at corners
		grid := [][]CellType{
			{CellWall, CellFloor, CellFloor},
			{CellFloor, CellFloor, CellFloor},
			{CellFloor, CellFloor, CellFloor},
		}

		// Start at wall (0,0) - should return error
		_, err := pf.FindPath(grid, DoorPosition{GridX: 0, GridY: 0}, DoorPosition{GridX: 2, GridY: 2})
		if err == nil {
			t.Error("expected error when start position is on wall")
		}

		// Same grid, target at wall - should return error
		_, err2 := pf.FindPath(grid, DoorPosition{GridX: 1, GridY: 1}, DoorPosition{GridX: 0, GridY: 0})
		if err2 == nil {
			t.Error("expected error when target position is on wall")
		}
	})
}

func TestIndoorPathfinderCache(t *testing.T) {
	pf := NewIndoorPathfinder()
	grid := createFloorGrid(10, 10)

	t.Run("cache hit on repeated query", func(t *testing.T) {
		// First call
		_, err1 := pf.FindPath(grid, DoorPosition{GridX: 0, GridY: 0}, DoorPosition{GridX: 5, GridY: 5})
		if err1 != nil {
			t.Fatalf("expected no error, got %v", err1)
		}

		// Second call with same params should hit cache
		_, err2 := pf.FindPath(grid, DoorPosition{GridX: 0, GridY: 0}, DoorPosition{GridX: 5, GridY: 5})
		if err2 != nil {
			t.Fatalf("expected no error from cache, got %v", err2)
		}

		// Cache should have entry
		if len(pf.cache) == 0 {
			t.Error("cache should have entries")
		}
	})

	t.Run("cache clear works", func(t *testing.T) {
		// Generate some paths
		pf.FindPath(grid, DoorPosition{GridX: 0, GridY: 0}, DoorPosition{GridX: 5, GridY: 5})
		pf.FindPath(grid, DoorPosition{GridX: 1, GridY: 1}, DoorPosition{GridX: 6, GridY: 6})

		if len(pf.cache) < 2 {
			t.Errorf("expected at least 2 cached paths, got %d", len(pf.cache))
		}

		// Clear cache
		pf.ClearCache()

		if len(pf.cache) != 0 {
			t.Error("cache should be empty after clear")
		}
	})
}

func TestIndoorPathfinderCoordinateConversion(t *testing.T) {
	pf := NewIndoorPathfinder()

	t.Run("InteriorToWorld converts correctly", func(t *testing.T) {
		buildingPos := ecs.Position{X: 10, Y: 20}

		world := pf.InteriorToWorld(buildingPos, 3, 5)

		if world.X != 13 {
			t.Errorf("WorldX = %f, want 13", world.X)
		}
		if world.Y != 25 {
			t.Errorf("WorldY = %f, want 25", world.Y)
		}
	})

	t.Run("WorldToInterior converts correctly", func(t *testing.T) {
		buildingPos := ecs.Position{X: 10, Y: 20}

		ix, iy := pf.WorldToInterior(buildingPos, 15, 27)

		if ix != 5 {
			t.Errorf("interiorX = %d, want 5", ix)
		}
		if iy != 7 {
			t.Errorf("interiorY = %d, want 7", iy)
		}
	})

	t.Run("WorldEntryPos returns correct world position", func(t *testing.T) {
		buildingPos := ecs.Position{X: 10, Y: 20}
		door := DoorPosition{WorldX: 3, WorldY: 4}

		entry := pf.WorldEntryPos(buildingPos, door)

		if entry.X != 13 {
			t.Errorf("entry.X = %f, want 13", entry.X)
		}
		if entry.Y != 24 {
			t.Errorf("entry.Y = %f, want 24", entry.Y)
		}
	})
}

func TestHeuristic(t *testing.T) {
	tests := []struct {
		name  string
		a     pos
		b     pos
		expect int
	}{
		{"same pos", pos{0, 0}, pos{0, 0}, 0},
		{"horizontal", pos{0, 0}, pos{3, 0}, 3},
		{"vertical", pos{0, 0}, pos{0, 4}, 4},
		{"diagonal manhattan", pos{1, 1}, pos{4, 5}, 7},
		{"negative coords", pos{-1, -1}, pos{2, 3}, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := heuristic(tt.a, tt.b)
			if got != tt.expect {
				t.Errorf("heuristic(%v, %v) = %d, want %d", tt.a, tt.b, got, tt.expect)
			}
		})
	}
}

func TestIsWalkable(t *testing.T) {
	grid := [][]CellType{
		{CellWall, CellFloor, CellDoor, CellWall},
		{CellFloor, CellFloor, CellFloor, CellFloor},
		{CellWall, CellFloor, CellWall, CellWall},
	}

	tests := []struct {
		name    string
		x, y    int
		walkable bool
	}{
		{"floor is walkable", 1, 0, true},
		{"door is walkable", 2, 0, true},
		{"wall is not walkable", 0, 0, false},
		{"out of bounds returns false", -1, 0, false},
		{"out of bounds y returns false", 0, 10, false},
		{"floor in middle", 1, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isWalkable(grid, tt.x, tt.y)
			if got != tt.walkable {
				t.Errorf("isWalkable(%d,%d) = %v, want %v", tt.x, tt.y, got, tt.walkable)
			}
		})
	}
}

func TestGetNeighbors(t *testing.T) {
	grid := [][]CellType{
		{CellFloor, CellFloor, CellFloor},
		{CellFloor, CellFloor, CellFloor},
		{CellFloor, CellFloor, CellFloor},
	}

	neighbors := getNeighbors(pos{1, 1}, grid)

	// Should have 4 neighbors (no diagonals)
	if len(neighbors) != 4 {
		t.Errorf("expected 4 neighbors, got %d", len(neighbors))
	}
}

func TestGetNeighborsEdge(t *testing.T) {
	grid := [][]CellType{
		{CellFloor, CellFloor, CellFloor},
		{CellFloor, CellFloor, CellFloor},
		{CellFloor, CellFloor, CellFloor},
	}

	// Corner neighbors
	neighbors := getNeighbors(pos{0, 0}, grid)

	// Should only have 2 neighbors (right and down)
	if len(neighbors) != 2 {
		t.Errorf("expected 2 corner neighbors, got %d", len(neighbors))
	}
}

func TestGetNeighborsBlocked(t *testing.T) {
	grid := [][]CellType{
		{CellWall, CellWall, CellWall},
		{CellWall, CellFloor, CellWall},
		{CellWall, CellWall, CellWall},
	}

	neighbors := getNeighbors(pos{1, 1}, grid)

	// All neighbors should be walls, so no walkable neighbors
	if len(neighbors) != 0 {
		t.Errorf("expected 0 neighbors in walled cell, got %d", len(neighbors))
	}
}

// Helper function to create a grid with all floors
func createFloorGrid(width, height int) [][]CellType {
	grid := make([][]CellType, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]CellType, width)
		for x := 0; x < width; x++ {
			grid[y][x] = CellFloor
		}
	}
	return grid
}