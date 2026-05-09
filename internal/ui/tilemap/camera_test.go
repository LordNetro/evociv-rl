package tilemap

import (
	"testing"
)

func TestNewCamera(t *testing.T) {
	// NewCamera creates camera with specified position and viewport size
	c := NewCamera(10, 20, 0, 80, 24)

	if c.X != 10 {
		t.Errorf("X should be 10, got %d", c.X)
	}
	if c.Y != 20 {
		t.Errorf("Y should be 20, got %d", c.Y)
	}
	if c.Z != 0 {
		t.Errorf("Z should be 0, got %d", c.Z)
	}
	if c.Width != 80 {
		t.Errorf("Width should be 80, got %d", c.Width)
	}
	if c.Height != 24 {
		t.Errorf("Height should be 24, got %d", c.Height)
	}
}

func TestCamera_Viewport_Dimensions(t *testing.T) {
	// Viewport returns correct dimensions matching camera Width/Height
	m := NewTilemap(100, 100)
	c := NewCamera(0, 0, 0, 40, 20)

	viewport := c.Viewport(m)
	if len(viewport) != 20 {
		t.Errorf("Viewport height should be 20, got %d", len(viewport))
	}
	if len(viewport) > 0 && len(viewport[0]) != 40 {
		t.Errorf("Viewport width should be 40, got %d", len(viewport[0]))
	}
}

func TestCamera_Viewport_Clamping(t *testing.T) {
	// Viewport clamps to map bounds when camera extends beyond map
	m := NewTilemap(50, 30)
	c := NewCamera(40, 20, 0, 40, 20) // would extend past bounds

	viewport := c.Viewport(m)

	// Should clamp to available space
	if len(viewport) != 10 { // 30 - 20 = 10 rows available
		t.Errorf("Viewport height should clamp to 10, got %d", len(viewport))
	}
	if len(viewport) > 0 && len(viewport[0]) != 10 { // 50 - 40 = 10 cols available
		t.Errorf("Viewport width should clamp to 10, got %d", len(viewport[0]))
	}
}

func TestCamera_Viewport_EmptyMap(t *testing.T) {
	// Viewport returns empty slice for zero-size map
	m := NewTilemap(0, 0)
	c := NewCamera(0, 0, 0, 10, 10)

	viewport := c.Viewport(m)
	if len(viewport) != 0 {
		t.Errorf("Viewport should be empty for zero-size map, got %d rows", len(viewport))
	}
}

func TestCamera_SetZLevel_Valid(t *testing.T) {
	// SetZLevel succeeds for Z=0 (surface)
	c := NewCamera(0, 0, 0, 10, 10)
	err := c.SetZLevel(0)
	if err != nil {
		t.Errorf("SetZLevel(0) should succeed, got error: %v", err)
	}
	if c.Z != 0 {
		t.Errorf("Z should be 0, got %d", c.Z)
	}
}

func TestCamera_SetZLevel_Interior(t *testing.T) {
	// SetZLevel succeeds for Z=1 (interior)
	c := NewCamera(0, 0, 0, 10, 10)
	err := c.SetZLevel(1)
	if err != nil {
		t.Errorf("SetZLevel(1) should succeed, got error: %v", err)
	}
	if c.Z != 1 {
		t.Errorf("Z should be 1, got %d", c.Z)
	}
}

func TestCamera_SetZLevel_Invalid(t *testing.T) {
	// SetZLevel returns error for invalid Z (e.g., Z=2)
	c := NewCamera(0, 0, 0, 10, 10)
	err := c.SetZLevel(2)
	if err == nil {
		t.Error("SetZLevel(2) should return error")
	}
	// Z should remain unchanged
	if c.Z != 0 {
		t.Errorf("Z should remain 0 after invalid SetZLevel, got %d", c.Z)
	}
}

func TestCamera_SetZLevel_Negative(t *testing.T) {
	// SetZLevel returns error for negative Z
	c := NewCamera(0, 0, 0, 10, 10)
	err := c.SetZLevel(-1)
	if err == nil {
		t.Error("SetZLevel(-1) should return error")
	}
}

func TestCamera_CenterOn(t *testing.T) {
	// CenterOn moves camera so (x, y) is centered in viewport
	c := NewCamera(0, 0, 0, 40, 20)

	c.CenterOn(100, 50)

	// (100, 50) should be centered: camera position = target - viewport/2
	if c.X != 80 { // 100 - 20 = 80
		t.Errorf("X should be 80, got %d", c.X)
	}
	if c.Y != 40 { // 50 - 10 = 40
		t.Errorf("Y should be 40, got %d", c.Y)
	}
}

func TestCamera_CenterOn_EdgeCase(t *testing.T) {
	// CenterOn with position near origin
	c := NewCamera(0, 0, 0, 10, 10)

	c.CenterOn(0, 0)

	// Camera should clamp to valid position
	if c.X != 0 {
		t.Errorf("X should be 0, got %d", c.X)
	}
	if c.Y != 0 {
		t.Errorf("Y should be 0, got %d", c.Y)
	}
}

func TestCamera_Move(t *testing.T) {
	// Move offsets camera position
	c := NewCamera(10, 10, 0, 20, 10)

	c.Move(5, -3)

	if c.X != 15 {
		t.Errorf("X should be 15, got %d", c.X)
	}
	if c.Y != 7 {
		t.Errorf("Y should be 7, got %d", c.Y)
	}
}

func TestCamera_Move_Negative(t *testing.T) {
	// Move can accept negative offsets
	c := NewCamera(10, 10, 0, 20, 10)

	c.Move(-10, -10)

	if c.X != 0 {
		t.Errorf("X should be 0, got %d", c.X)
	}
	if c.Y != 0 {
		t.Errorf("Y should be 0, got %d", c.Y)
	}
}

func TestCamera_Bounds(t *testing.T) {
	// Bounds returns current viewport bounds
	c := NewCamera(10, 20, 0, 30, 15)

	minX, minY, maxX, maxY := c.Bounds()

	if minX != 10 {
		t.Errorf("minX should be 10, got %d", minX)
	}
	if minY != 20 {
		t.Errorf("minY should be 20, got %d", minY)
	}
	if maxX != 39 { // 10 + 30 - 1
		t.Errorf("maxX should be 39, got %d", maxX)
	}
	if maxY != 34 { // 20 + 15 - 1
		t.Errorf("maxY should be 34, got %d", maxY)
	}
}

func TestCamera_Viewport_TilesMatchMap(t *testing.T) {
	// Viewport returns tiles that match the underlying map data
	m := NewTilemap(100, 100)
	m.SetTile(25, 15, 0, LayerTerrain, '.')

	c := NewCamera(20, 10, 0, 10, 10)
	viewport := c.Viewport(m)

	// Tile at viewport position (5, 5) should correspond to map position (25, 15)
	if viewport[5][5].Terrain != '.' {
		t.Errorf("Viewport tile at (5,5) should have Terrain '.', got %c", viewport[5][5].Terrain)
	}
}

func TestCamera_ZLevel_Viewport(t *testing.T) {
	// Viewport respects camera's Z level
	m := NewTilemap(50, 50)
	m.SetTile(10, 10, 0, LayerTerrain, '.')
	m.SetTile(10, 10, 1, LayerTerrain, '#')

	c := NewCamera(5, 5, 0, 10, 10)
	viewport := c.Viewport(m)

	// Should show Z=0 terrain ('.')
	if viewport[5][5].Terrain != '.' {
		t.Errorf("Z=0 viewport should show '.', got %c", viewport[5][5].Terrain)
	}

	// Switch to Z=1
	c.SetZLevel(1)
	viewport = c.Viewport(m)

	// Should show Z=1 terrain ('#')
	if viewport[5][5].Terrain != '#' {
		t.Errorf("Z=1 viewport should show '#', got %c", viewport[5][5].Terrain)
	}
}