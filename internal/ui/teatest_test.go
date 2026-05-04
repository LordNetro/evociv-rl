package ui

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

var updateGolden = flag.Bool("update-golden", false, "update golden files")

func TestViewGolden(t *testing.T) {
	m := NewModel()
	m.ready = true
	m.width = 80
	m.height = 24

	output := m.View()
	golden := filepath.Join("testdata", "welcome.golden")

	if *updateGolden {
		os.WriteFile(golden, []byte(output), 0644)
	}

	expected, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("reading golden file: %v", err)
	}
	if output != string(expected) {
		t.Errorf("output doesn't match golden file")
	}
}

func TestTeatestIntegration(t *testing.T) {
	m := NewModel()
	tm := teatest.NewTestModel(t, m)

	tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	tm.WaitFinished(t)

	finalModel := tm.FinalModel(t).(Model)
	if !finalModel.quitting {
		t.Error("expected quitting to be true")
	}
}
