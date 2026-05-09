package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/simulation/npc"
	"github.com/marco/evociv-rl/internal/simulation/settlement"
	"github.com/marco/evociv-rl/internal/ui/tilemap"
	"github.com/marco/evociv-rl/internal/world"
)

// tickMsg is sent periodically to advance the simulation.
type tickMsg struct{}

// Model is the Bubbletea model for the Evociv-RL TUI.
type Model struct {
	ready    bool
	width    int
	height   int
	quitting bool
	screen   string // "welcome" | "map" | "settlement"
	cameraX  int
	cameraY  int
	cursorX  int // cursor position within viewport
	cursorY  int
	worldMap *world.WorldMap
	ecsWorld *ecs.World
	npcOverlay        []npc.NPCRenderInfo
	settlementOverlay []settlement.SettlementRenderInfo
	inspectorOpen     bool
	selectedNPC       ecs.Entity
	selectedSettlement ecs.Entity
	simTick           int
	renderTick        int
	// Settlement view state
	settlementCameraX int
	settlementCameraY int
	// Track which settlement entity we're viewing in settlement mode
	settlementViewEntity ecs.Entity
	settlementViewInfo settlement.SettlementRenderInfo
	// Tilemap renderer fields (optional - nil when feature flag disabled)
	tilemapView *tilemap.TilemapView
}

// NewModel creates a new TUI model.
func NewModel() Model {
	return Model{
		screen:  "welcome",
		cursorX: 40,
		cursorY: 12,
	}
}

// SetWorldMap injects a world map into the model and centers the camera.
func (m *Model) SetWorldMap(wm *world.WorldMap) {
	m.worldMap = wm
	m.cameraX = wm.Width/2 - 40
	m.cameraY = wm.Height/2 - 12
	if m.cameraX < 0 {
		m.cameraX = 0
	}
	if m.cameraY < 0 {
		m.cameraY = 0
	}
}

// SetNPCOverlay injects the current NPC render overlay.
func (m *Model) SetNPCOverlay(overlay []npc.NPCRenderInfo) {
	m.npcOverlay = overlay
}

// SetSettlementOverlay injects the current settlement render overlay.
func (m *Model) SetSettlementOverlay(overlay []settlement.SettlementRenderInfo) {
	m.settlementOverlay = overlay
}

// SetECSWorld injects the ECS world for inspector lookups.
func (m *Model) SetECSWorld(w *ecs.World) {
	m.ecsWorld = w
}

// Init initializes the model and returns the initial command.
func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, m.simTickCmd())
}

// simTickCmd returns a command that fires a tick after a short delay.
func (m Model) simTickCmd() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
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
			if m.screen == "settlement" {
				m.screen = "map"
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit
		case "esc":
			if m.inspectorOpen {
				m.inspectorOpen = false
				return m, nil
			}
			if m.screen == "settlement" {
				m.screen = "map"
				return m, nil
			}
		case "enter":
			if m.screen == "map" && !m.inspectorOpen {
				// Try to enter settlement at cursor
				m.tryEnterSettlement()
			}
		case "m":
			if !m.inspectorOpen {
				if m.screen == "welcome" {
					m.screen = "map"
				} else if m.screen == "settlement" {
					m.screen = "map"
				} else {
					m.screen = "welcome"
				}
			}
		case "e":
			if (m.screen == "map" || m.screen == "settlement") && !m.inspectorOpen {
				m.tryOpenInspector()
			}
		case "up":
			if m.screen == "settlement" {
				if m.settlementCameraY > 0 {
					m.settlementCameraY--
				}
			} else if m.cursorY > 0 {
				m.cursorY--
			}
		case "down":
			if m.screen == "settlement" {
				m.settlementCameraY++
			} else if m.cursorY < m.height-1 {
				m.cursorY++
			}
		case "left":
			if m.screen == "settlement" {
				if m.settlementCameraX > 0 {
					m.settlementCameraX--
				}
			} else if m.cursorX > 0 {
				m.cursorX--
			}
		case "right":
			if m.screen == "settlement" {
				m.settlementCameraX++
			} else if m.cursorX < m.width-1 {
				m.cursorX++
			}
		case "w":
			if m.screen == "map" && m.worldMap != nil && m.cameraY > 0 {
				m.cameraY--
			}
		case "s":
			if m.screen == "map" && m.worldMap != nil && m.cameraY < m.worldMap.Height-1 {
				m.cameraY++
			}
		case "a":
			if m.screen == "map" && m.worldMap != nil && m.cameraX > 0 {
				m.cameraX--
			}
		case "d":
			if m.screen == "map" && m.worldMap != nil && m.cameraX < m.worldMap.Width-1 {
				m.cameraX++
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.cursorX = msg.Width / 2
		m.cursorY = msg.Height / 2
	case tickMsg:
		if m.ecsWorld != nil && (m.screen == "map" || m.screen == "settlement") {
			m.simTick++
			dt := 0.2 // 200ms per tick
			_ = m.ecsWorld.Update(dt)
			// Re-fetch NPC overlay from the render system
			m.refreshOverlay()
		}
		return m, m.simTickCmd()
	}
	return m, nil
}

func (m *Model) refreshOverlay() {
	// Find NPCRenderSystem and get render infos
	for _, sys := range m.ecsWorld.Systems() {
		if rs, ok := sys.(*npc.NPCRenderSystem); ok {
			// Re-run render system to update positions
			_ = rs.Update(m.ecsWorld, 0)
			m.npcOverlay = rs.RenderInfos()
		}
	}
	// Find SettlementRenderSystem and get render infos
	for _, sys := range m.ecsWorld.Systems() {
		if rs, ok := sys.(*settlement.SettlementRenderSystem); ok {
			_ = rs.Update(m.ecsWorld, 0)
			m.settlementOverlay = rs.RenderInfos()
			break
		}
	}
}

func (m *Model) tryOpenInspector() {
	if m.worldMap == nil {
		return
	}
	wx := m.cursorX + m.cameraX
	wy := m.cursorY + m.cameraY
	if !m.worldMap.InBounds(wx, wy) {
		return
	}
	for _, info := range m.npcOverlay {
		if info.WorldX == wx && info.WorldY == wy {
			m.selectedNPC = info.Entity
			m.inspectorOpen = true
			m.selectedSettlement = 0
			return
		}
	}
	for _, info := range m.settlementOverlay {
		if info.WorldX == wx && info.WorldY == wy {
			m.selectedSettlement = ecs.Entity(info.Entity)
			m.inspectorOpen = true
			m.selectedNPC = 0
			return
		}
	}
}

// tryEnterSettlement enters the settlement interior view if cursor is on a settlement
func (m *Model) tryEnterSettlement() {
	if m.worldMap == nil {
		return
	}
	wx := m.cursorX + m.cameraX
	wy := m.cursorY + m.cameraY
	for _, info := range m.settlementOverlay {
		if info.WorldX == wx && info.WorldY == wy {
			// Found settlement under cursor, enter settlement view
			m.screen = "settlement"
			m.settlementViewEntity = ecs.Entity(info.Entity)
			m.settlementViewInfo = info
			// Center the settlement camera on the settlement position
			m.settlementCameraX = info.WorldX - m.width/2
			m.settlementCameraY = info.WorldY - m.height/2
			if m.settlementCameraX < 0 {
				m.settlementCameraX = 0
			}
			if m.settlementCameraY < 0 {
				m.settlementCameraY = 0
			}
			return
		}
	}
}

// View renders the model (implemented in view.go).
func (m Model) View() string {
	return renderView(m)
}