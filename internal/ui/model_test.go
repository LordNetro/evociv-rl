package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/simulation/npc"
	"github.com/marco/evociv-rl/internal/simulation/settlement"
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

func TestArrowKeysMoveCursor(t *testing.T) {
	m := NewModel()
	m.width = 20
	m.height = 20
	m.cursorX = 10
	m.cursorY = 10

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	mm := newModel.(Model)
	if mm.cursorX != 11 {
		t.Errorf("cursorX after Right = %d, want 11", mm.cursorX)
	}

	newModel, _ = mm.Update(tea.KeyMsg{Type: tea.KeyDown})
	mm = newModel.(Model)
	if mm.cursorY != 11 {
		t.Errorf("cursorY after Down = %d, want 11", mm.cursorY)
	}

	newModel, _ = mm.Update(tea.KeyMsg{Type: tea.KeyLeft})
	mm = newModel.(Model)
	if mm.cursorX != 10 {
		t.Errorf("cursorX after Left = %d, want 10", mm.cursorX)
	}

	newModel, _ = mm.Update(tea.KeyMsg{Type: tea.KeyUp})
	mm = newModel.(Model)
	if mm.cursorY != 10 {
		t.Errorf("cursorY after Up = %d, want 10", mm.cursorY)
	}
}

func TestInspectorOpenOnNPC(t *testing.T) {
	wm := world.NewWorldMap(5, 5)
	m := NewModel()
	m.SetWorldMap(wm)
	m.screen = "map"
	m.cameraX = 0
	m.cameraY = 0
	m.cursorX = 1
	m.cursorY = 1
	m.npcOverlay = []npc.NPCRenderInfo{
		{Entity: ecs.Entity(1), WorldX: 1, WorldY: 1, Symbol: '@'},
	}

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	mm := newModel.(Model)
	if !mm.inspectorOpen {
		t.Error("expected inspector to open")
	}
	if mm.selectedNPC != ecs.Entity(1) {
		t.Errorf("selectedNPC = %d, want 1", mm.selectedNPC)
	}
}

func TestInspectorNoOpOnEmptyTile(t *testing.T) {
	wm := world.NewWorldMap(5, 5)
	m := NewModel()
	m.SetWorldMap(wm)
	m.screen = "map"
	m.npcOverlay = []npc.NPCRenderInfo{
		{Entity: ecs.Entity(1), WorldX: 2, WorldY: 2},
	}

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	mm := newModel.(Model)
	if mm.inspectorOpen {
		t.Error("expected inspector to stay closed on empty tile")
	}
}

func TestInspectorCloseWithQ(t *testing.T) {
	m := NewModel()
	m.inspectorOpen = true
	m.screen = "map"
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	mm := newModel.(Model)
	if mm.inspectorOpen {
		t.Error("expected inspector to close with q")
	}
	if mm.quitting {
		t.Error("q should not quit when inspector is open")
	}
}

func TestInspectorCloseWithEsc(t *testing.T) {
	m := NewModel()
	m.inspectorOpen = true
	m.screen = "map"
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mm := newModel.(Model)
	if mm.inspectorOpen {
		t.Error("expected inspector to close with esc")
	}
}

func TestCursorMovesWithArrows(t *testing.T) {
	m := NewModel()
	m.width = 50
	m.height = 50
	m.cursorX = 25
	m.cursorY = 25

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	mm := newModel.(Model)
	if mm.cursorX != 26 {
		t.Errorf("cursorX = %d, want 26", mm.cursorX)
	}
}

func TestCursorBounds(t *testing.T) {
	m := NewModel()
	m.width = 10
	m.height = 10
	m.cursorX = 9
	m.cursorY = 9

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	mm := newModel.(Model)
	if mm.cursorX != 9 {
		t.Errorf("cursorX at right edge = %d, want 9", mm.cursorX)
	}

	newModel, _ = mm.Update(tea.KeyMsg{Type: tea.KeyDown})
	mm = newModel.(Model)
	if mm.cursorY != 9 {
		t.Errorf("cursorY at bottom edge = %d, want 9", mm.cursorY)
	}

	m2 := NewModel()
	m2.width = 10
	m2.height = 10
	m2.cursorX = 0
	m2.cursorY = 0

	newModel, _ = m2.Update(tea.KeyMsg{Type: tea.KeyLeft})
	mm = newModel.(Model)
	if mm.cursorX != 0 {
		t.Errorf("cursorX at left edge = %d, want 0", mm.cursorX)
	}

	newModel, _ = mm.Update(tea.KeyMsg{Type: tea.KeyUp})
	mm = newModel.(Model)
	if mm.cursorY != 0 {
		t.Errorf("cursorY at top edge = %d, want 0", mm.cursorY)
	}
}

func TestSetSettlementOverlay(t *testing.T) {
	m := NewModel()
	overlay := []settlement.SettlementRenderInfo{
		{WorldX: 1, WorldY: 1, Symbol: '♦', Color: "#8B7355", Name: "Village"},
	}
	m.SetSettlementOverlay(overlay)
	if len(m.settlementOverlay) != 1 {
		t.Errorf("expected 1 settlement overlay, got %d", len(m.settlementOverlay))
	}
}

func TestInspectorOpenOnSettlement(t *testing.T) {
	wm := world.NewWorldMap(5, 5)
	m := NewModel()
	m.SetWorldMap(wm)
	m.screen = "map"
	m.cameraX = 0
	m.cameraY = 0
	m.cursorX = 2
	m.cursorY = 2
	m.settlementOverlay = []settlement.SettlementRenderInfo{
		{Entity: 42, WorldX: 2, WorldY: 2, Symbol: '♦'},
	}

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	mm := newModel.(Model)
	if !mm.inspectorOpen {
		t.Error("expected inspector to open on settlement tile")
	}
	if mm.selectedSettlement != 42 {
		t.Errorf("selectedSettlement = %d, want 42", mm.selectedSettlement)
	}
}
