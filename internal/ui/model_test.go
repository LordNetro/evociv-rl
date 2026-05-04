package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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
