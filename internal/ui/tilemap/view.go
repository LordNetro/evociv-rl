package tilemap

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TilemapView is the BubbleTea model for the tilemap renderer
type TilemapView struct {
	Tilemap           *Tilemap
	Camera            *Camera
	Builder           *TileBuilder
	InteriorRenderer  *InteriorRenderer
	CurrentZ          int            // 0 = surface, 1 = interior
	SelectedBuilding  uint64         // 0 = none
	InteriorGrid      [][]CellType   // Grid for current interior view

	// BubbleTea fields
	width  int
	height int
}

// NewTilemapView creates a new TilemapView with the given tilemap and camera
func NewTilemapView(tilemap *Tilemap, camera *Camera) TilemapView {
	return TilemapView{
		Tilemap:          tilemap,
		Camera:           camera,
		Builder:          nil,
		InteriorRenderer: nil,
		CurrentZ:         0,
		SelectedBuilding: 0,
		InteriorGrid:     nil,
		width:            0,
		height:           0,
	}
}

// Init initializes the model and returns the initial command
func (m TilemapView) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages and updates the model state
func (m TilemapView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Update camera dimensions to match viewport
		if m.Camera != nil {
			m.Camera.Width = msg.Width
			m.Camera.Height = msg.Height
		}

	default:
		// Ignore other messages
	}

	return m, nil
}

// handleKey handles keyboard input
func (m TilemapView) handleKey(msg tea.KeyMsg) (TilemapView, tea.Cmd) {
	// Handle escape key specially (also by type)
	if msg.Type == tea.KeyEscape {
		// In interior: return to surface
		if m.CurrentZ == 1 {
			m.CurrentZ = 0
			m.SelectedBuilding = 0
			m.InteriorGrid = nil
			// Reset camera to Z=0
			if m.Camera != nil {
				_ = m.Camera.SetZLevel(0)
			}
		}
		return m, nil
	}

	switch msg.String() {
	case "q", "Q", "ctrl+c":
		// Quit the program
		return m, tea.Quit

	case "z", "Z":
		// Toggle Z level
		m.toggleZLevel()

	case "w", "up":
		// Move camera up
		if m.Camera != nil && m.CurrentZ == 0 {
			m.Camera.Y--
			if m.Camera.Y < 0 {
				m.Camera.Y = 0
			}
		}

	case "s", "down":
		// Move camera down
		if m.Camera != nil && m.CurrentZ == 0 {
			m.Camera.Y++
			if m.Camera.Height > 0 && m.Tilemap != nil {
				maxY := m.Tilemap.Height() - m.Camera.Height
				if maxY < 0 {
					maxY = 0
				}
				if m.Camera.Y > maxY {
					m.Camera.Y = maxY
				}
			}
		}

	case "a", "left":
		// Move camera left
		if m.Camera != nil && m.CurrentZ == 0 {
			m.Camera.X--
			if m.Camera.X < 0 {
				m.Camera.X = 0
			}
		}

	case "d", "right":
		// Move camera right
		if m.Camera != nil && m.CurrentZ == 0 {
			m.Camera.X++
			if m.Camera.Width > 0 && m.Tilemap != nil {
				maxX := m.Tilemap.Width() - m.Camera.Width
				if maxX < 0 {
					maxX = 0
				}
				if m.Camera.X > maxX {
					m.Camera.X = maxX
				}
			}
		}
	}

	return m, nil
}

// toggleZLevel toggles between Z=0 (surface) and Z=1 (interior)
func (m *TilemapView) toggleZLevel() {
	if m.CurrentZ == 0 {
		// Try to switch to Z=1 (interior)
		// Ensure Z=1 level exists
		if m.Tilemap != nil {
			m.Tilemap.SetZLevel(1)
		}
		if m.Camera != nil {
			_ = m.Camera.SetZLevel(1)
		}
		m.CurrentZ = 1
	} else {
		// Switch back to Z=0 (surface)
		if m.Camera != nil {
			_ = m.Camera.SetZLevel(0)
		}
		m.CurrentZ = 0
	}
}

// View renders the model
func (m TilemapView) View() string {
	style := DefaultStyleConfig()

	if m.Tilemap == nil || m.Camera == nil {
		return "No tilemap data"
	}

	var result string
	if m.CurrentZ == 0 {
		// Render surface (Z=0)
		result = RenderViewport(m.Tilemap, m.Camera, style)
	} else {
		// Render interior (Z=1)
		if m.InteriorGrid != nil {
			result = RenderInterior(m.Tilemap, m.Camera, m.InteriorGrid, style)
		} else {
			// No interior grid set - show empty interior
			result = RenderViewport(m.Tilemap, m.Camera, style)
		}
	}

	// Wrap in lipgloss style with dimensions
	viewStyle := lipgloss.NewStyle().
		Width(m.Camera.Width).
		Height(m.Camera.Height)

	return viewStyle.Render(result)
}

// SetInteriorGrid sets the interior grid for Z=1 rendering
func (m *TilemapView) SetInteriorGrid(grid [][]CellType) {
	m.InteriorGrid = grid
}

// EnterBuilding enters the interior of a building
func (m *TilemapView) EnterBuilding(buildingID uint64) {
	if buildingID == 0 {
		return
	}

	m.SelectedBuilding = buildingID
	m.CurrentZ = 1

	if m.Camera != nil {
		_ = m.Camera.SetZLevel(1)
	}
}

// ExitBuilding exits the building interior and returns to surface
func (m *TilemapView) ExitBuilding() {
	m.SelectedBuilding = 0
	m.CurrentZ = 0
	m.InteriorGrid = nil

	if m.Camera != nil {
		_ = m.Camera.SetZLevel(0)
	}
}