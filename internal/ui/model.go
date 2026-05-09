package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marco/evociv-rl/internal/world"
	"github.com/marco/evociv-rl/internal/ui/tilemap"
)

// Model is the Bubbletea model for the Evociv-RL TUI.
type Model struct {
	ready    bool
	width    int
	height   int
	quitting bool
	screen   string // "welcome" | "map"
	cameraX  int
	cameraY  int
	worldMap *world.WorldMap
	// Tilemap renderer fields (optional - nil when feature flag disabled)
	tilemapView *tilemap.TilemapView
}

// NewModel creates a new TUI model.
func NewModel() Model {
	return Model{
		screen: "welcome",
	}
}

// SetWorldMap injects a world map into the model.
func (m *Model) SetWorldMap(wm *world.WorldMap) {
	m.worldMap = wm
}

// Init initializes the model and returns the initial command.
func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen)
}

// Update handles incoming messages and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			m.quitting = true
			return m, tea.Quit
		case "m":
			if m.screen == "welcome" {
				m.screen = "map"
			} else {
				m.screen = "welcome"
			}
		case "w", "up":
			if m.screen == "map" && m.worldMap != nil {
				if m.cameraY > 0 {
					m.cameraY--
				}
			}
		case "s", "down":
			if m.screen == "map" && m.worldMap != nil {
				if m.cameraY < m.worldMap.Height-1 {
					m.cameraY++
				}
			}
		case "a", "left":
			if m.screen == "map" && m.worldMap != nil {
				if m.cameraX > 0 {
					m.cameraX--
				}
			}
		case "d", "right":
			if m.screen == "map" && m.worldMap != nil {
				if m.cameraX < m.worldMap.Width-1 {
					m.cameraX++
				}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
	}
	return m, nil
}

// View renders the model (implemented in view.go).
func (m Model) View() string {
	return renderView(m)
}
