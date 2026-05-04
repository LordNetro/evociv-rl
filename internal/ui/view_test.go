package ui

import (
	"strings"
	"testing"
)

func TestViewContainsTitle(t *testing.T) {
	m := NewModel()
	m.width = 80
	m.height = 24
	m.ready = true
	v := m.View()
	if !strings.Contains(v, "Evociv-RL") {
		t.Error("view missing title 'Evociv-RL'")
	}
}

func TestViewContainsVersion(t *testing.T) {
	m := NewModel()
	m.width = 80
	m.height = 24
	m.ready = true
	v := m.View()
	if !strings.Contains(v, "v0.0.1") {
		t.Error("view missing version 'v0.0.1'")
	}
}

func TestViewContainsSubtitle(t *testing.T) {
	m := NewModel()
	m.width = 80
	m.height = 24
	m.ready = true
	v := m.View()
	if !strings.Contains(v, "Un mundo por descubrir") {
		t.Error("view missing subtitle")
	}
}

func TestViewContainsInstructions(t *testing.T) {
	m := NewModel()
	m.width = 80
	m.height = 24
	m.ready = true
	v := m.View()
	if !strings.Contains(v, "[q] Salir") {
		t.Error("view missing instructions")
	}
}
