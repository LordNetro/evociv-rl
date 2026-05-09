package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/simulation/npc"
	"github.com/marco/evociv-rl/internal/simulation/df"
	"github.com/marco/evociv-rl/internal/simulation/settlement"
	"github.com/marco/evociv-rl/internal/world"
)

// tickMsg is sent periodically to advance the simulation.
type tickMsg struct{}

// SettlementViewState tracks the active settlement interior view.
type SettlementViewState struct {
	SettlementEntity  ecs.Entity
	SettlementCenterX int
	SettlementCenterY int
	CursorX, CursorY  int
	ViewportRadius    int
}

// RewardPopup is a floating text showing an NPC's recent reward.
type RewardPopup struct {
	WorldX, WorldY int
	Text           string
	TicksLeft      int
}

// Model is the Bubbletea model for the Evociv-RL TUI.
type Model struct {
	ready               bool
	width               int
	height              int
	quitting            bool
	screen              string // "welcome" | "map" | "settlement"
	previousScreen      string
	cameraX             int
	cameraY             int
	cursorX             int
	cursorY             int
	worldMap            *world.WorldMap
	ecsWorld            *ecs.World
	npcOverlay          []npc.NPCRenderInfo
	settlementOverlay   []settlement.SettlementRenderInfo
	settlementBuildings []settlement.BuildingRenderInfo
	settlementNPCs      []npc.NPCRenderInfo
	rewardPopups        []RewardPopup
	settlementViewState *SettlementViewState
	inspectorOpen       bool
	selectedNPC         ecs.Entity
	selectedSettlement  int
	selectedBuilding    int
	simTick             int
	renderTick          int
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

// SetSettlementBuildings injects the current building render overlay for settlement view.
func (m *Model) SetSettlementBuildings(overlay []settlement.BuildingRenderInfo) {
	m.settlementBuildings = overlay
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
				m.closeSettlementView()
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
				m.closeSettlementView()
				return m, nil
			}
		case "m":
			if !m.inspectorOpen && m.screen != "settlement" {
				if m.screen == "welcome" {
					m.screen = "map"
				} else {
					m.screen = "welcome"
				}
			}
		case "e":
			if m.screen == "map" && !m.inspectorOpen {
				m.tryOpenInspectorOrSettlement()
			} else if m.screen == "settlement" && !m.inspectorOpen {
				m.tryOpenSettlementInspector()
			}
		case "up":
			if m.screen == "settlement" && m.settlementViewState != nil {
				if m.settlementViewState.CursorY > 0 {
					m.settlementViewState.CursorY--
				}
			} else if m.cursorY > 0 {
				m.cursorY--
			}
		case "down":
			if m.screen == "settlement" && m.settlementViewState != nil {
				maxY := m.settlementViewState.ViewportRadius*2
				if m.settlementViewState.CursorY < maxY {
					m.settlementViewState.CursorY++
				}
			} else if m.cursorY < m.height-1 {
				m.cursorY++
			}
		case "left":
			if m.screen == "settlement" && m.settlementViewState != nil {
				if m.settlementViewState.CursorX > 0 {
					m.settlementViewState.CursorX--
				}
			} else if m.cursorX > 0 {
				m.cursorX--
			}
		case "right":
			if m.screen == "settlement" && m.settlementViewState != nil {
				maxX := m.settlementViewState.ViewportRadius*2
				if m.settlementViewState.CursorX < maxX {
					m.settlementViewState.CursorX++
				}
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

		// Enqueue a job onto selected building's JobQueue
		if msg.String() == "j" {
			if m.screen == "settlement" && m.inspectorOpen && m.selectedBuilding != 0 && m.ecsWorld != nil {
				// select job whose role matches building role
				var chosen df.Job
				var found bool
				var role string
				for _, bi := range m.settlementBuildings {
					if int(bi.Entity) == m.selectedBuilding {
						role = bi.Role
						break
					}
				}
				for _, j := range df.AllJobs() {
					if role != "" && j.Role == role {
						chosen = j
						found = true
						break
					}
				}
				if !found {
					jobs := df.AllJobs()
					if len(jobs) > 0 {
						chosen = jobs[0]
						found = true
					}
				}
				if found {
					// attach/append to JobQueue component of building
					be := ecs.Entity(m.selectedBuilding)
					jqStore, ok := m.ecsWorld.GetStore(df.JobQueueID).(*ecs.ComponentStore[df.JobQueue])
					if ok {
						q, qok := jqStore.Get(ecs.Entity(be))
						if !qok {
							q = df.JobQueue{}
						}
						// if job has no target entity, set to this building
						chosen.TargetEntity = int64(be)
						q.Jobs = append(q.Jobs, df.Job(chosen))
						jqStore.Set(ecs.Entity(be), q)
					}
				}
			}

			// Add wood to selected building inventory
			if msg.String() == "k" {
				if m.screen == "settlement" && m.inspectorOpen && m.selectedBuilding != 0 && m.ecsWorld != nil {
					be := ecs.Entity(m.selectedBuilding)
					invStore, ok := m.ecsWorld.GetStore(df.InventoryID).(*ecs.ComponentStore[df.Inventory])
					if ok {
						inv, iok := invStore.Get(be)
						if !iok {
							inv = df.Inventory{OwnerEntity: int64(be), Items: map[string]int{}, Cap: 10}
						}
						inv.Items["wood"] = inv.Items["wood"] + 1
						invStore.Set(be, inv)
					}
				}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.cursorX = msg.Width / 2
		m.cursorY = msg.Height / 2
	case tickMsg:
		if m.ecsWorld != nil {
			if m.screen == "map" {
				m.simTick++
				dt := 0.2 // 200ms per tick
				_ = m.ecsWorld.Update(dt)
				m.refreshOverlay()
			} else if m.screen == "settlement" {
				m.simTick++
				dt := 0.2
				// Boost LOD for settlement NPCs before world tick so WanderSystem moves them
				m.boostSettlementNPCLOD()
				_ = m.ecsWorld.Update(dt)
				m.refreshSettlementOverlay()
				m.processRewardPopups()
			}
		}
		return m, m.simTickCmd()
	}
	return m, nil
}

func (m *Model) refreshOverlay() {
	// Find NPCRenderSystem and get render infos
	for _, sys := range m.ecsWorld.Systems() {
		if rs, ok := sys.(*npc.NPCRenderSystem); ok {
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

func (m *Model) refreshSettlementOverlay() {
	if m.settlementViewState == nil || m.ecsWorld == nil {
		return
	}
	settlementEntity := m.settlementViewState.SettlementEntity
	// Find NPCRenderSystem and filter by settlement
	for _, sys := range m.ecsWorld.Systems() {
		if rs, ok := sys.(*npc.NPCRenderSystem); ok {
			_ = rs.Update(m.ecsWorld, 0)
			m.settlementNPCs = rs.RenderInfosForSettlement(m.ecsWorld, settlementEntity)
		}
	}
	// Find BuildingRenderSystem and filter by settlement
	for _, sys := range m.ecsWorld.Systems() {
		if rs, ok := sys.(*settlement.BuildingRenderSystem); ok {
			_ = rs.Update(m.ecsWorld, 0)
			m.settlementBuildings = rs.RenderInfosForSettlement(m.ecsWorld, settlementEntity)
			break
		}
	}
}

// boostSettlementNPCLOD forces LOD to Local for all NPCs belonging to the active settlement
// and registers them with LODSystem so it won't downgrade them.
func (m *Model) boostSettlementNPCLOD() {
	if m.settlementViewState == nil || m.ecsWorld == nil {
		return
	}
	settlementEntity := m.settlementViewState.SettlementEntity
	homeRefID := ecs.NewComponentID("home_reference")
	homeStore, ok := m.ecsWorld.GetStore(homeRefID).(*ecs.ComponentStore[settlement.HomeReference])
	if !ok {
		return
	}
	lodStore, ok := m.ecsWorld.GetStore(npc.LODID).(*ecs.ComponentStore[npc.LOD])
	if !ok {
		return
	}
	// Find LODSystem once and cache reference
	var lodSys *npc.LODSystem
	for _, sys := range m.ecsWorld.Systems() {
		if ls, ok := sys.(*npc.LODSystem); ok {
			lodSys = ls
			break
		}
	}
	for e, hr := range homeStore.All() {
		if hr.SettlementEntity == settlementEntity {
			if l, ok := lodStore.Get(e); ok && l.Level < npc.LODLocal {
				lodStore.Set(e, npc.LOD{Level: npc.LODLocal})
			}
			// Register with LODSystem so it won't downgrade this NPC
			if lodSys != nil {
				lodSys.SetSettlementBoost(e)
			}
		}
	}
}

func (m *Model) tryOpenInspectorOrSettlement() {
	if m.worldMap == nil {
		return
	}
	wx := m.cursorX + m.cameraX
	wy := m.cursorY + m.cameraY
	if !m.worldMap.InBounds(wx, wy) {
		return
	}
	// NPC inspector takes priority
	for _, info := range m.npcOverlay {
		if info.WorldX == wx && info.WorldY == wy {
			m.selectedNPC = info.Entity
			m.inspectorOpen = true
			m.selectedSettlement = 0
			return
		}
	}
	// Settlement entry — use actual radius from Settlement component
	for _, info := range m.settlementOverlay {
		if info.WorldX == wx && info.WorldY == wy {
			radius := 3 // default fallback
			if m.ecsWorld != nil {
				if setComp, ok := ecs.GetComponent[settlement.Settlement](m.ecsWorld, ecs.Entity(info.Entity)); ok {
					radius = setComp.Radius
				}
			}
			m.openSettlementView(ecs.Entity(info.Entity), info.WorldX, info.WorldY, radius)
			return
		}
	}
}

func (m *Model) tryOpenSettlementInspector() {
	if m.settlementViewState == nil {
		return
	}
	vx := m.settlementViewState.CursorX
	vy := m.settlementViewState.CursorY
	radius := m.settlementViewState.ViewportRadius
	wx := vx + m.settlementViewState.SettlementCenterX - radius
	wy := vy + m.settlementViewState.SettlementCenterY - radius

	// NPC inspector takes priority
	for _, info := range m.settlementNPCs {
		if info.WorldX == wx && info.WorldY == wy {
			m.selectedNPC = info.Entity
			m.inspectorOpen = true
			m.selectedSettlement = 0
			return
		}
	}
	// Building inspector
	for _, info := range m.settlementBuildings {
		if info.WorldX == wx && info.WorldY == wy {
			m.selectedBuilding = int(info.Entity)
			m.inspectorOpen = true
			m.selectedNPC = 0
			return
		}
	}
}

func (m *Model) openSettlementView(entity ecs.Entity, cx, cy, radius int) {
	m.previousScreen = m.screen
	m.screen = "settlement"
	m.settlementViewState = &SettlementViewState{
		SettlementEntity:  entity,
		SettlementCenterX: cx,
		SettlementCenterY: cy,
		CursorX:           radius,
		CursorY:           radius,
		ViewportRadius:    radius,
	}
	m.inspectorOpen = false
	m.selectedNPC = 0
	m.selectedSettlement = 0
	m.rewardPopups = nil
	m.refreshSettlementOverlay()
}

func (m *Model) closeSettlementView() {
	m.screen = m.previousScreen
	if m.screen == "" {
		m.screen = "map"
	}
	// Clear LODSystem settlement boost so NPCs return to normal LOD
	if m.ecsWorld != nil {
		for _, sys := range m.ecsWorld.Systems() {
			if ls, ok := sys.(*npc.LODSystem); ok {
				ls.ClearSettlementBoost()
				break
			}
		}
	}
	m.settlementViewState = nil
	m.settlementBuildings = nil
	m.settlementNPCs = nil
	m.rewardPopups = nil
	m.inspectorOpen = false
	m.selectedNPC = 0
	m.selectedSettlement = 0
	m.selectedBuilding = 0
}

func (m *Model) processRewardPopups() {
	// Decrement existing popups
	var active []RewardPopup
	for _, p := range m.rewardPopups {
		p.TicksLeft--
		if p.TicksLeft > 0 {
			active = append(active, p)
		}
	}
	m.rewardPopups = active

	// Create new popups from recent rewards, avoiding duplicates at same position
	existingPositions := make(map[string]bool)
	for _, p := range m.rewardPopups {
		key := fmt.Sprintf("%d,%d", p.WorldX, p.WorldY)
		existingPositions[key] = true
	}
	var newPopups []RewardPopup
	for _, npc := range m.settlementNPCs {
		if npc.LastReward == 0 {
			continue
		}
		if npc.LastReward > -0.1 && npc.LastReward < 0.1 {
			continue
		}
		if m.simTick-npc.RewardTick >= 5 {
			continue
		}
		key := fmt.Sprintf("%d,%d", npc.WorldX, npc.WorldY)
		if existingPositions[key] {
			continue
		}
		existingPositions[key] = true
		newPopups = append(newPopups, RewardPopup{
			WorldX:    npc.WorldX,
			WorldY:    npc.WorldY,
			Text:      fmt.Sprintf("+%.2f", npc.LastReward),
			TicksLeft: 5,
		})
	}
	// Limit to max 3
	if len(newPopups) > 3 {
		newPopups = newPopups[:3]
	}
	m.rewardPopups = append(m.rewardPopups, newPopups...)
}

// View renders the model (implemented in view.go).
func (m Model) View() string {
	return renderView(m)
}
