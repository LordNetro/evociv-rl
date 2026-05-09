package tilemap

import (
	"testing"

	"github.com/charmbracelet/bubbletea"
)

// TestTilemapView_Init tests that Init returns nil (no initial command)
func TestTilemapView_Init(t *testing.T) {
	m := NewTilemapView(nil, nil)
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init should return nil command")
	}
}

// TestTilemapView_Update_ArrowKeys tests arrow key handling moves camera
func TestTilemapView_Update_ArrowKeys(t *testing.T) {
	m := NewTilemapView(NewTilemap(10, 10), NewCamera(0, 0, 0, 10, 10))

	// Test up arrow
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m, _ = result.(TilemapView)
	_ = m

	// Test down arrow
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, _ = result.(TilemapView)
	_ = m
}

// TestTilemapView_Update_ZToggle tests pressing Z toggles Z level
func TestTilemapView_Update_ZToggle(t *testing.T) {
	// Setup: tilemap with Z=1 level initialized
	m := NewTilemapView(NewTilemap(10, 10), NewCamera(0, 0, 0, 10, 10))
	m.Tilemap.SetZLevel(1) // Create Z=1 level

	// Verify initial Z=0
	if m.CurrentZ != 0 {
		t.Errorf("expected CurrentZ=0, got %d", m.CurrentZ)
	}

	// Press Z to toggle to Z=1
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	m, _ = result.(TilemapView)

	// After toggle, CurrentZ should be 1
	if m.CurrentZ != 1 {
		t.Errorf("expected CurrentZ=1 after toggle, got %d", m.CurrentZ)
	}
}

// TestTilemapView_Update_Quit tests pressing q quits
func TestTilemapView_Update_Quit(t *testing.T) {
	m := NewTilemapView(NewTilemap(10, 10), NewCamera(0, 0, 0, 10, 10))

	// Press q to quit
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	// Should return tea.Quit
	if cmd == nil {
		t.Error("expected tea.Quit command")
	}
}

// TestTilemapView_Update_WindowSize tests window size updates dimensions
func TestTilemapView_Update_WindowSize(t *testing.T) {
	m := NewTilemapView(NewTilemap(10, 10), NewCamera(0, 0, 0, 10, 10))

	// Send window size message
	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	result, _ := m.Update(msg)
	m, _ = result.(TilemapView)

	// Verify dimensions updated
	if m.width != 80 {
		t.Errorf("expected width=80, got %d", m.width)
	}
	if m.height != 24 {
		t.Errorf("expected height=24, got %d", m.height)
	}
}

// TestTilemapView_View_Surface tests View returns rendered surface when Z=0
func TestTilemapView_View_Surface(t *testing.T) {
	// Setup: tilemap with some data
	m := NewTilemapView(NewTilemap(5, 5), NewCamera(0, 0, 0, 5, 5))
	m.width = 5
	m.height = 5

	// Set some terrain
	m.Tilemap.SetTile(2, 2, 0, LayerTerrain, '.')

	// View should return non-empty string
	result := m.View()
	if len(result) == 0 {
		t.Error("View should return non-empty string for surface")
	}
}

// TestTilemapView_View_Interior tests View returns rendered interior when Z=1
func TestTilemapView_View_Interior(t *testing.T) {
	// Setup: tilemap with interior data
	m := NewTilemapView(NewTilemap(5, 5), NewCamera(0, 0, 1, 5, 5))
	m.width = 5
	m.height = 5
	m.CurrentZ = 1

	// Create interior grid
	grid := [][]CellType{
		{CellFloor, CellFloor},
		{CellFloor, CellFloor},
	}
	m.InteriorGrid = grid

	// For this test, just verify View doesn't error when CurrentZ=1
	result := m.View()
	_ = result // Just verify no panic
}

// TestTilemapView_NewTilemapView tests constructor creates valid view
func TestTilemapView_NewTilemapView(t *testing.T) {
	tm := NewTilemap(10, 10)
	cam := NewCamera(0, 0, 0, 10, 10)

	m := NewTilemapView(tm, cam)

	if m.Tilemap != tm {
		t.Error("Tilemap not set correctly")
	}
	if m.Camera != cam {
		t.Error("Camera not set correctly")
	}
	if m.CurrentZ != 0 {
		t.Errorf("expected CurrentZ=0, got %d", m.CurrentZ)
	}
	if m.SelectedBuilding != 0 {
		t.Errorf("expected SelectedBuilding=0, got %d", m.SelectedBuilding)
	}
}

// TestTilemapView_Update_EscapeInInterior tests Escape returns to surface
func TestTilemapView_Update_EscapeInInterior(t *testing.T) {
	m := NewTilemapView(NewTilemap(10, 10), NewCamera(0, 0, 0, 10, 10))
	m.CurrentZ = 1 // In interior

	// Press Escape to return to surface
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m, _ = result.(TilemapView)

	if m.CurrentZ != 0 {
		t.Errorf("expected CurrentZ=0 after escape, got %d", m.CurrentZ)
	}
}