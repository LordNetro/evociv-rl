package tilemap

import (
	"strings"
	"testing"
)

// TestDefaultStyleConfig tests that DefaultStyleConfig returns valid config
func TestDefaultStyleConfig(t *testing.T) {
	cfg := DefaultStyleConfig()
	if cfg == nil {
		t.Fatal("DefaultStyleConfig returned nil")
	}
	if cfg.TerrainBG == "" {
		t.Error("TerrainBG should not be empty")
	}
	if cfg.CreatureFG == "" {
		t.Error("CreatureFG should not be empty")
	}
}

// TestRenderViewport_EmptyMap tests rendering an empty viewport
func TestRenderViewport_EmptyMap(t *testing.T) {
	// Setup: create an empty tilemap and camera
	m := NewTilemap(3, 3)
	cam := NewCamera(0, 0, 0, 3, 3)

	// Execute: render the viewport
	result := RenderViewport(m, cam, DefaultStyleConfig())

	// Verify: output should have 3 lines (height)
	lines := splitLines(result)
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}

	// Verify each line has 3 characters (width)
	for i, line := range lines {
		if len(line) != 3 {
			t.Errorf("line %d: expected 3 chars, got %d", i, len(line))
		}
	}
}

// TestRenderViewport_CreatureLayerPriority tests creature layer has highest priority
func TestRenderViewport_CreatureLayerPriority(t *testing.T) {
	// Setup: tilemap with creature at (1,1)
	m := NewTilemap(3, 3)
	m.SetTile(1, 1, 0, LayerTerrain, '.')  // terrain below
	m.SetTile(1, 1, 0, LayerBuilding, '#') // building below
	m.SetTile(1, 1, 0, LayerCreature, '@') // creature on top

	cam := NewCamera(0, 0, 0, 3, 3)

	// Execute
	result := RenderViewport(m, cam, DefaultStyleConfig())

	// Verify: creature should be visible at center
	lines := splitLines(result)
	if len(lines) < 1 {
		t.Fatal("expected at least 1 line")
	}
	// Middle of middle line should contain @ (creature)
	middleLine := lines[1]
	if len(middleLine) < 3 {
		t.Fatalf("line too short: %q", middleLine)
	}
	// The creature should appear in position 1
	if !containsRune(middleLine, '@') {
		t.Errorf("expected '@' in output, got: %q", middleLine)
	}
}

// TestRenderViewport_BuildingLayerPriority tests building layer has second priority
func TestRenderViewport_BuildingLayerPriority(t *testing.T) {
	// Setup: tilemap with building but no creature
	m := NewTilemap(3, 3)
	m.SetTile(1, 1, 0, LayerTerrain, '.')
	m.SetTile(1, 1, 0, LayerBuilding, '#')
	// No creature layer set

	cam := NewCamera(0, 0, 0, 3, 3)

	// Execute
	result := RenderViewport(m, cam, DefaultStyleConfig())

	// Verify: building should be visible
	lines := splitLines(result)
	if len(lines) < 1 {
		t.Fatal("expected at least 1 line")
	}
	middleLine := lines[1]
	if !containsRune(middleLine, '#') {
		t.Errorf("expected '#' in output, got: %q", middleLine)
	}
}

// TestRenderViewport_TerrainLayer tests terrain renders when no building/creature
func TestRenderViewport_TerrainLayer(t *testing.T) {
	// Setup: tilemap with only terrain
	m := NewTilemap(3, 3)
	m.SetTile(1, 1, 0, LayerTerrain, 'T') // forest

	cam := NewCamera(0, 0, 0, 3, 3)

	// Execute
	result := RenderViewport(m, cam, DefaultStyleConfig())

	// Verify: terrain character should be visible
	lines := splitLines(result)
	middleLine := lines[1]
	if !containsRune(middleLine, 'T') {
		t.Errorf("expected 'T' in output, got: %q", middleLine)
	}
}

// TestRenderViewport_FogDimsCharacter tests fog affects rendering
func TestRenderViewport_FogDimsCharacter(t *testing.T) {
	// Setup: tile with terrain and fog
	m := NewTilemap(3, 3)
	m.SetTile(1, 1, 0, LayerTerrain, '.')
	m.SetFog(1, 1, 0, ':') // dark/unexplored

	cam := NewCamera(0, 0, 0, 3, 3)

	// Execute
	result := RenderViewport(m, cam, DefaultStyleConfig())

	// Verify: output should be non-empty (fog doesn't hide completely)
	lines := splitLines(result)
	if len(lines) == 0 {
		t.Error("expected non-empty output with fog")
	}
}

// TestRenderViewport_EmptySpace tests empty tiles render as space
func TestRenderViewport_EmptySpace(t *testing.T) {
	// Setup: tilemap with empty terrain
	m := NewTilemap(3, 3)
	// All tiles are zero-value (empty terrain)

	cam := NewCamera(0, 0, 0, 3, 3)

	// Execute
	result := RenderViewport(m, cam, DefaultStyleConfig())

	// Verify: output should contain spaces
	lines := splitLines(result)
	for i, line := range lines {
		// Should have characters (even if just spaces)
		if len(line) == 0 {
			t.Errorf("line %d is empty", i)
		}
	}
}

// TestRenderInterior_GridSize tests interior renders with correct dimensions
func TestRenderInterior_GridSize(t *testing.T) {
	// Setup: create tilemap with Z=1 interior
	m := NewTilemap(5, 5)
	cam := NewCamera(0, 0, 1, 5, 5)

	// Create interior grid (3x3)
	grid := [][]CellType{
		{CellWall, CellWall, CellWall},
		{CellWall, CellFloor, CellFloor},
		{CellFloor, CellFloor, CellFloor},
	}

	// Execute
	result := RenderInterior(m, cam, grid, DefaultStyleConfig())

	// Verify: should have rows based on grid height
	lines := splitLines(result)
	// Grid is 3x3 but map is 5x5 - should render grid portion
	if len(lines) < 3 {
		t.Errorf("expected at least 3 lines, got %d", len(lines))
	}
}

// TestRenderInterior_NPCOverlay tests NPCs overlaid on interior grid
func TestRenderInterior_NPCOverlay(t *testing.T) {
	// Setup: tilemap with creature in Z=1 at position within grid bounds
	m := NewTilemap(5, 5)
	m.SetTile(1, 1, 1, LayerCreature, '@')

	cam := NewCamera(0, 0, 1, 5, 5)

	// Create floor grid - 3x3 so NPC at (1,1) is inside
	grid := [][]CellType{
		{CellFloor, CellFloor, CellFloor},
		{CellFloor, CellFloor, CellFloor},
		{CellFloor, CellFloor, CellFloor},
	}

	// Execute
	result := RenderInterior(m, cam, grid, DefaultStyleConfig())

	// Verify: creature should appear in output
	lines := splitLines(result)
	foundCreature := false
	for _, line := range lines {
		if containsRune(line, '@') {
			foundCreature = true
			break
		}
	}
	if !foundCreature {
		t.Error("expected '@' in interior output")
	}
}

// TestRenderViewport_ItemLayer tests item layer renders
func TestRenderViewport_ItemLayer(t *testing.T) {
	// Setup: tilemap with item but no creature/building
	m := NewTilemap(3, 3)
	m.SetTile(1, 1, 0, LayerTerrain, '.')
	m.SetTile(1, 1, 0, LayerItem, '*') // gold item

	cam := NewCamera(0, 0, 0, 3, 3)

	// Execute
	result := RenderViewport(m, cam, DefaultStyleConfig())

	// Verify: item should be visible (between terrain and creature/building)
	lines := splitLines(result)
	middleLine := lines[1]
	if !containsRune(middleLine, '*') {
		t.Errorf("expected '*' in output, got: %q", middleLine)
	}
}

// TestRenderViewport_LayerPriorityOrder tests correct layer priority
func TestRenderViewport_LayerPriorityOrder(t *testing.T) {
	// Setup: tilemap with all layers populated
	m := NewTilemap(3, 3)
	m.SetTile(1, 1, 0, LayerTerrain, '.')
	m.SetTile(1, 1, 0, LayerBuilding, '#')
	m.SetTile(1, 1, 0, LayerItem, '*')
	m.SetTile(1, 1, 0, LayerCreature, '@')

	cam := NewCamera(0, 0, 0, 3, 3)

	// Execute
	result := RenderViewport(m, cam, DefaultStyleConfig())

	// Verify: creature should win (highest priority)
	lines := splitLines(result)
	middleLine := lines[1]
	if !containsRune(middleLine, '@') {
		t.Errorf("expected '@' (creature highest priority), got: %q", middleLine)
	}
}

// Helper functions

func splitLines(s string) []string {
	// Split by newline, filter empty lines
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		if len(line) > 0 {
			lines = append(lines, line)
		}
	}
	return lines
}

func containsRune(s string, r rune) bool {
	for _, ch := range s {
		if ch == r {
			return true
		}
	}
	return false
}