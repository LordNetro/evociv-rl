package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/marco/evociv-rl/internal/world"
)

func TestModelInit(t *testing.T) {
	m := NewModel()
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected Init to return a non-nil cmd")
	}
}

func TestModelUpdateQuit(t *testing.T) {
	m := NewModel()
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	mm := newModel.(Model)
	if !mm.quitting {
		t.Error("expected quitting to be true after pressing q")
	}
}

func TestModelUpdateWindowSize(t *testing.T) {
	m := NewModel()
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	mm := newModel.(Model)
	if mm.width != 80 {
		t.Errorf("width = %d, want 80", mm.width)
	}
	if mm.height != 24 {
		t.Errorf("height = %d, want 24", mm.height)
	}
	if !mm.ready {
		t.Error("expected ready to be true after WindowSizeMsg")
	}
}

func TestToggleToMap(t *testing.T) {
	m := NewModel()
	m.screen = "welcome"
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	mm := newModel.(Model)
	if mm.screen != "map" {
		t.Errorf("screen = %q, want map", mm.screen)
	}
}

func TestToggleToWelcome(t *testing.T) {
	m := NewModel()
	m.screen = "map"
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	mm := newModel.(Model)
	if mm.screen != "welcome" {
		t.Errorf("screen = %q, want welcome", mm.screen)
	}
}

func TestCameraMoveWASD(t *testing.T) {
	wm := world.NewWorldMap(10, 10)
	m := NewModel()
	m.SetWorldMap(wm)
	m.screen = "map"

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	mm := newModel.(Model)
	if mm.cameraX != 1 {
		t.Errorf("cameraX after 'd' = %d, want 1", mm.cameraX)
	}

	newModel, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	mm = newModel.(Model)
	if mm.cameraY != 1 {
		t.Errorf("cameraY after 's' = %d, want 1", mm.cameraY)
	}

	newModel, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	mm = newModel.(Model)
	if mm.cameraX != 0 {
		t.Errorf("cameraX after 'a' = %d, want 0", mm.cameraX)
	}

	newModel, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	mm = newModel.(Model)
	if mm.cameraY != 0 {
		t.Errorf("cameraY after 'w' = %d, want 0", mm.cameraY)
	}
}

func TestCameraBounds(t *testing.T) {
	wm := world.NewWorldMap(5, 5)
	m := NewModel()
	m.SetWorldMap(wm)
	m.screen = "map"
	m.cameraX = 4 // rightmost valid column
	m.cameraY = 4 // bottommost valid row

	// Try to move beyond right edge
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	mm := newModel.(Model)
	if mm.cameraX != 4 {
		t.Errorf("cameraX at right edge = %d, want 4", mm.cameraX)
	}

	// Try to move beyond bottom edge
	newModel, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	mm = newModel.(Model)
	if mm.cameraY != 4 {
		t.Errorf("cameraY at bottom edge = %d, want 4", mm.cameraY)
	}

	// Move to top-left and try to go beyond
	m2 := NewModel()
	m2.SetWorldMap(wm)
	m2.screen = "map"
	m2.cameraX = 0
	m2.cameraY = 0

	newModel, _ = m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	mm = newModel.(Model)
	if mm.cameraX != 0 {
		t.Errorf("cameraX at left edge = %d, want 0", mm.cameraX)
	}

	newModel, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	mm = newModel.(Model)
	if mm.cameraY != 0 {
		t.Errorf("cameraY at top edge = %d, want 0", mm.cameraY)
	}
}

func TestQuitStillWorks(t *testing.T) {
	m := NewModel()
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	mm := newModel.(Model)
	if !mm.quitting {
		t.Error("expected quitting to be true after pressing q")
	}
	if cmd == nil {
		t.Error("expected tea.Quit command")
	}
}

func TestArrowKeysWork(t *testing.T) {
	wm := world.NewWorldMap(10, 10)
	m := NewModel()
	m.SetWorldMap(wm)
	m.screen = "map"

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	mm := newModel.(Model)
	if mm.cameraX != 1 {
		t.Errorf("cameraX after Right = %d, want 1", mm.cameraX)
	}

	newModel, _ = mm.Update(tea.KeyMsg{Type: tea.KeyDown})
	mm = newModel.(Model)
	if mm.cameraY != 1 {
		t.Errorf("cameraY after Down = %d, want 1", mm.cameraY)
	}

	newModel, _ = mm.Update(tea.KeyMsg{Type: tea.KeyLeft})
	mm = newModel.(Model)
	if mm.cameraX != 0 {
		t.Errorf("cameraX after Left = %d, want 0", mm.cameraX)
	}

	newModel, _ = mm.Update(tea.KeyMsg{Type: tea.KeyUp})
	mm = newModel.(Model)
	if mm.cameraY != 0 {
		t.Errorf("cameraY after Up = %d, want 0", mm.cameraY)
	}
}
