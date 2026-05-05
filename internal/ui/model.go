package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/simulation/npc"
	"github.com/marco/evociv-rl/internal/world"
)

// Model is the Bubbletea model for the Evociv-RL TUI.
type Model struct {
	ready         bool
	width         int
	height        int
	quitting      bool
	screen        string // "welcome" | "map"
	cameraX       int
	cameraY       int
	cursorX       int
	cursorY       int
	worldMap      *world.WorldMap
	ecsWorld      *ecs.World
	npcOverlay    []npc.NPCRenderInfo
	inspectorOpen bool
	selectedNPC   ecs.Entity
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

// SetNPCOverlay injects the current NPC render overlay.
func (m *Model) SetNPCOverlay(overlay []npc.NPCRenderInfo) {
	m.npcOverlay = overlay
}

// SetECSWorld injects the ECS world for inspector lookups.
func (m *Model) SetECSWorld(w *ecs.World) {
	m.ecsWorld = w
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
			if m.inspectorOpen {
				m.inspectorOpen = false
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit
		case "esc":
			if m.inspectorOpen {
				m.inspectorOpen = false
				return m, nil
			}
		case "m":
			if !m.inspectorOpen {
				if m.screen == "welcome" {
					m.screen = "map"
				} else {
					m.screen = "welcome"
				}
			}
		case "e":
			if m.screen == "map" && !m.inspectorOpen {
				m.tryOpenInspector()
			}
		case "w", "up":
			if m.screen == "map" && m.worldMap != nil {
				if m.inspectorOpen {
					if m.cursorY > 0 {
						m.cursorY--
					}
				} else {
					if m.cameraY > 0 {
						m.cameraY--
					}
				}
			}
		case "s", "down":
			if m.screen == "map" && m.worldMap != nil {
				if m.inspectorOpen {
					if m.cursorY < m.worldMap.Height-1 {
						m.cursorY++
					}
				} else {
					if m.cameraY < m.worldMap.Height-1 {
						m.cameraY++
					}
				}
			}
		case "a", "left":
			if m.screen == "map" && m.worldMap != nil {
				if m.inspectorOpen {
					if m.cursorX > 0 {
						m.cursorX--
					}
				} else {
					if m.cameraX > 0 {
						m.cameraX--
					}
				}
			}
		case "d", "right":
			if m.screen == "map" && m.worldMap != nil {
				if m.inspectorOpen {
					if m.cursorX < m.worldMap.Width-1 {
						m.cursorX++
					}
				} else {
					if m.cameraX < m.worldMap.Width-1 {
						m.cameraX++
					}
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

func (m *Model) tryOpenInspector() {
	wx := m.cursorX + m.cameraX
	wy := m.cursorY + m.cameraY
	for _, info := range m.npcOverlay {
		if info.WorldX == wx && info.WorldY == wy {
			m.selectedNPC = info.Entity
			m.inspectorOpen = true
			return
		}
	}
}

// View renders the model (implemented in view.go).
func (m Model) View() string {
	return renderView(m)
}
