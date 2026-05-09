package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/simulation/npc"
	"github.com/marco/evociv-rl/internal/simulation/df"
	"github.com/marco/evociv-rl/internal/simulation/rl"
	"github.com/marco/evociv-rl/internal/simulation/settlement"
)

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		PaddingTop(2).
		PaddingBottom(1).
		Align(lipgloss.Center)

	subtitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0A0A0")).
		PaddingBottom(1).
		Align(lipgloss.Center)

	versionStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50FA7B")).
		PaddingBottom(2).
		Align(lipgloss.Center)

	instructionsStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2")).
		PaddingTop(1).
		Align(lipgloss.Center)
)

// cursorStyle highlights the cursor position with a gold background.
var cursorStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#FFD700")).
	Foreground(lipgloss.Color("#000000"))

// biomeStyles maps biome IDs to their display symbol and color.
var biomeStyles = map[string]struct {
	symbol string
	color  string
}{
	"ocean":   {symbol: "~", color: "#1E90FF"},
	"plains":  {symbol: ".", color: "#90EE90"},
	"forest":  {symbol: "T", color: "#228B22"},
	"desert":  {symbol: "d", color: "#EDC9AF"},
	"tundra":  {symbol: "*", color: "#E0FFFF"},
	"jungle":  {symbol: "J", color: "#006400"},
	"unknown": {symbol: "?", color: "#888888"},
}

// statusBarStyle is the dark background style for the status bar.
var statusBarStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#333333")).
	Foreground(lipgloss.Color("#FFFFFF"))

// renderView dispatches rendering based on the current screen.
func renderView(m Model) string {
	if !m.ready {
		return "Cargando..."
	}

	switch m.screen {
	case "map":
		return renderMap(m)
	case "settlement":
		return renderSettlementView(m)
	default:
		return renderWelcome(m)
	}
}

func renderWelcome(m Model) string {
	title := titleStyle.Render("Evociv-RL")
	version := versionStyle.Render("v0.0.1 — Fundación")
	subtitle := subtitleStyle.Render("Un mundo por descubrir...")
	instructions := instructionsStyle.Render("[q] Salir  [m] Mapa")

	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		version,
		subtitle,
		instructions,
	)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func renderInspector(m Model) string {
	if !m.inspectorOpen {
		return ""
	}
	if m.ecsWorld == nil {
		return ""
	}

	if m.screen == "settlement" {
		if m.selectedNPC != 0 {
			return renderSettlementNPCInspector(m)
		}
		if m.selectedBuilding != 0 {
			return renderBuildingInspector(m)
		}
		return ""
	}

	if m.selectedNPC != 0 {
		return renderNPCInspector(m)
	}
	if m.selectedSettlement != 0 {
		return renderSettlementInspector(m)
	}
	return ""
}

func renderNPCInspector(m Model) string {
	nameComp, _ := ecs.GetComponent[ecs.Name](m.ecsWorld, m.selectedNPC)
	healthComp, _ := ecs.GetComponent[npc.Health](m.ecsWorld, m.selectedNPC)
	jobComp, _ := ecs.GetComponent[npc.Job](m.ecsWorld, m.selectedNPC)
	persComp, _ := ecs.GetComponent[npc.Personality](m.ecsWorld, m.selectedNPC)
	posComp, posOk := ecs.GetComponent[ecs.Position](m.ecsWorld, m.selectedNPC)

	biomeName := "unknown"
	if m.worldMap != nil && posOk {
		if tile := m.worldMap.TileAt(int(posComp.X), int(posComp.Y)); tile != nil {
			biomeName = tile.BiomeID
		}
	}

	var b strings.Builder
	b.WriteString("=== NPC ===\n")
	b.WriteString("Name: " + nameComp.Name + "\n")
	b.WriteString(fmt.Sprintf("Health: %.0f/%.0f\n", healthComp.Current, healthComp.Max))
	b.WriteString("Job: " + jobComp.Role + "\n")
	b.WriteString("Biome: " + biomeName + "\n")
	b.WriteString(fmt.Sprintf("O: %.2f C: %.2f E: %.2f A: %.2f N: %.2f\n",
		persComp.Openness, persComp.Conscientiousness, persComp.Extraversion, persComp.Agreeableness, persComp.Neuroticism))

	return b.String()
}

func renderSettlementInspector(m Model) string {
	setComp, ok := ecs.GetComponent[settlement.Settlement](m.ecsWorld, ecs.Entity(m.selectedSettlement))
	if !ok {
		return ""
	}

	var b strings.Builder
	b.WriteString("=== Settlement ===\n")
	b.WriteString("Name: " + setComp.Name + "\n")
	b.WriteString("Type: " + setComp.Type + "\n")
	b.WriteString(fmt.Sprintf("Population: %d\n", setComp.Population))
	b.WriteString(fmt.Sprintf("Radius: %d\n", setComp.Radius))
	b.WriteString(fmt.Sprintf("Level: %d\n", setComp.Level))
	b.WriteString("Buildings: " + strings.Join(setComp.Buildings, ", ") + "\n")

	// Resources
	if rs, ok := ecs.GetComponent[settlement.ResourceStore](m.ecsWorld, ecs.Entity(m.selectedSettlement)); ok {
		b.WriteString(fmt.Sprintf("Food: %.0f\n", rs.Resources["food"]))
		b.WriteString(fmt.Sprintf("Gold: %.0f\n", rs.Resources["gold"]))
		b.WriteString(fmt.Sprintf("Tools: %.0f\n", rs.Resources["tools"]))
		if rs.Resources["food"] < 0 {
			b.WriteString("⚠ Famine\n")
		}
	}

	// Level progress
	if setComp.Level >= 3 {
		b.WriteString("Level: 3 (MAX)\n")
	} else {
		b.WriteString(fmt.Sprintf("Next Level: %d\n", setComp.Level+1))
	}

	return b.String()
}

func renderBuildingInspector(m Model) string {
	if m.ecsWorld == nil {
		return ""
	}
	bComp, ok := ecs.GetComponent[settlement.Building](m.ecsWorld, ecs.Entity(m.selectedBuilding))
	if !ok {
		return ""
	}

	var info settlement.BuildingRenderInfo
	for _, bi := range m.settlementBuildings {
		if int(bi.Entity) == m.selectedBuilding {
			info = bi
			break
		}
	}

	// Count workers by role matching
	workerCount := 0
	var assigned []npc.NPCRenderInfo
	for _, n := range m.settlementNPCs {
		if n.JobRole == info.Role {
			workerCount++
			assigned = append(assigned, n)
		}
	}

	var b strings.Builder
	b.WriteString("=== Building ===\n")
	b.WriteString("Name: " + bComp.Name + "\n")
	b.WriteString(fmt.Sprintf("Level: %d\n", bComp.Level))
	b.WriteString("Role: " + info.Role + "\n")
	b.WriteString(fmt.Sprintf("Workers: %d/%d\n", workerCount, info.MaxWorkers))

	// Produces
	if len(info.Produces) > 0 {
		var parts []string
		for res, rate := range info.Produces {
			parts = append(parts, fmt.Sprintf("%s +%.1f/tick/worker", res, rate))
		}
		b.WriteString("Produces: " + strings.Join(parts, ", ") + "\n")
	} else {
		b.WriteString("Produces: (none)\n")
	}

	// Consumes
	if len(info.Consumes) > 0 {
		var parts []string
		for res, rate := range info.Consumes {
			parts = append(parts, fmt.Sprintf("%s -%.1f/tick/worker", res, rate))
		}
		b.WriteString("Consumes: " + strings.Join(parts, ", ") + "\n")
	} else {
		b.WriteString("Consumes: (none)\n")
	}

	if info.Role != "" {
		b.WriteString("--- Assigned NPCs ---\n")
		for _, n := range assigned {
			name := "Unknown"
			if nameComp, ok := ecs.GetComponent[ecs.Name](m.ecsWorld, n.Entity); ok {
				name = nameComp.Name
			}
			b.WriteString(fmt.Sprintf("%s (reward: %.2f)\n", name, n.LastReward))
		}
	}

	// Available jobs
	b.WriteString("--- Available Jobs (press 'j' to enqueue first) ---\n")
	for _, job := range df.AllJobs() {
		b.WriteString(fmt.Sprintf("%s: role=%s action=%s reward=%.2f\n", job.ID, job.Role, job.ActionID, job.Reward))
	}

	// Inventory
	if m.ecsWorld != nil {
		if invStore := m.ecsWorld.GetStore(df.InventoryID); invStore != nil {
			if s, ok := invStore.(*ecs.ComponentStore[df.Inventory]); ok {
				if inv, ok := s.Get(ecs.Entity(m.selectedBuilding)); ok {
					b.WriteString("--- Inventory (press 'k' to add wood) ---\n")
					if len(inv.Items) == 0 {
						b.WriteString("(empty)\n")
					} else {
						for it, c := range inv.Items {
							b.WriteString(fmt.Sprintf("%s x%d\n", it, c))
						}
					}
				}
			}
		}
	}

	return b.String()
}

func renderSettlementNPCInspector(m Model) string {
	if m.ecsWorld == nil {
		return ""
	}
	nameComp, _ := ecs.GetComponent[ecs.Name](m.ecsWorld, m.selectedNPC)
	healthComp, _ := ecs.GetComponent[npc.Health](m.ecsWorld, m.selectedNPC)
	jobComp, _ := ecs.GetComponent[npc.Job](m.ecsWorld, m.selectedNPC)
	aiComp, _ := ecs.GetComponent[npc.AIState](m.ecsWorld, m.selectedNPC)
	homeComp, _ := ecs.GetComponent[settlement.HomeReference](m.ecsWorld, m.selectedNPC)

	homeName := "Unknown"
	if homeComp != (settlement.HomeReference{}) {
		if setComp, ok := ecs.GetComponent[settlement.Settlement](m.ecsWorld, homeComp.SettlementEntity); ok {
			homeName = setComp.Name
		}
	}

	// Find workplace by role matching
	workplaceName := "None"
	if jobComp.Role != "" {
		for _, bi := range m.settlementBuildings {
			if bi.Role == jobComp.Role {
				workplaceName = bi.Name
				break
			}
		}
	}

	var b strings.Builder
	b.WriteString("=== NPC ===\n")
	b.WriteString("Name: " + nameComp.Name + "\n")
	b.WriteString("Role: " + jobComp.Role + "\n")
	b.WriteString(fmt.Sprintf("Health: %.0f/%.0f\n", healthComp.Current, healthComp.Max))
	b.WriteString("Home: " + homeName + "\n")
	b.WriteString("Workplace: " + workplaceName + "\n")
	b.WriteString(fmt.Sprintf("Last Reward: %.2f\n", aiComp.LastReward))
	b.WriteString("Policy: ε-greedy\n")
	b.WriteString(fmt.Sprintf("Epsilon: %.2f\n", getEpsilon(m.ecsWorld)))
	b.WriteString("State: " + aiComp.CurrentAction + "\n")

	return b.String()
}

func getEpsilon(w *ecs.World) float64 {
	for _, sys := range w.Systems() {
		if ql, ok := sys.(*npc.QLearningSystem); ok {
			return ql.QTable().Epsilon()
		}
	}
	return rl.NewQTable().Epsilon()
}

func renderMap(m Model) string {
	if m.worldMap == nil {
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			"Generando mundo...",
		)
	}

	termHeight := m.height
	termWidth := m.width

	var lines []string
	for row := 0; row < termHeight; row++ {
		worldY := m.cameraY + row
		if worldY >= m.worldMap.Height {
			lines = append(lines, strings.Repeat(" ", termWidth))
			continue
		}

		var line strings.Builder
		for col := 0; col < termWidth; col++ {
			worldX := m.cameraX + col
			if worldX >= m.worldMap.Width {
				line.WriteString(" ")
				continue
			}

			isCursor := worldX == m.cameraX+m.cursorX && worldY == m.cameraY+m.cursorY

			styled := renderOverlay(m, worldX, worldY, isCursor)
			if styled != "" {
				line.WriteString(styled)
				continue
			}

			tile := m.worldMap.TileAt(worldX, worldY)
			if tile == nil {
				if isCursor {
					line.WriteString(cursorStyle.Render(" "))
				} else {
					line.WriteString(" ")
				}
				continue
			}

			style, ok := biomeStyles[tile.BiomeID]
			if !ok {
				style = biomeStyles["unknown"]
			}
			if isCursor {
				line.WriteString(cursorStyle.Render(style.symbol))
			} else {
				line.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color(style.color)).
					Render(style.symbol))
			}
		}
		lines = append(lines, line.String())
	}

	result := strings.Join(lines, "\n")

	// Status bar: show settlement info if cursor is over one
	if m.screen == "map" && !m.inspectorOpen {
		wx := m.cameraX + m.cursorX
		wy := m.cameraY + m.cursorY
		for _, info := range m.settlementOverlay {
			if info.WorldX == wx && info.WorldY == wy {
				status := fmt.Sprintf(" %s %s | Pop: %d ", string(info.Symbol), info.Name, info.Population)
				if info.HasResources {
					status = fmt.Sprintf(" %s %s | Pop: %d | Food: %.0f Gold: %.0f Tools: %.0f ",
						string(info.Symbol), info.Name, info.Population, info.Food, info.Gold, info.Tools)
				}
				result += "\n" + lipgloss.NewStyle().
					Background(lipgloss.Color("#333333")).
					Foreground(lipgloss.Color(info.Color)).
					Render(status)
				break
			}
		}
	}

	if m.inspectorOpen {
		result += "\n" + renderInspector(m)
	}
	return result
}

// renderOverlay returns the styled overlay symbol at the given world coordinate.
// Priority: NPC > Settlement > Biome (empty string falls through to biome).
// If isCursor is true, wraps the symbol in cursor styling (gold background).
func renderOverlay(m Model, worldX, worldY int, isCursor bool) string {
	var symbol rune
	var colorStr string

	// 1. NPC overlay (highest priority)
	for _, info := range m.npcOverlay {
		if info.WorldX == worldX && info.WorldY == worldY {
			symbol = info.Symbol
			colorStr = string(info.Color)
			goto render
		}
	}
	// 2. Settlement overlay (middle priority)
	for _, info := range m.settlementOverlay {
		if info.WorldX == worldX && info.WorldY == worldY {
			symbol = info.Symbol
			colorStr = info.Color
			goto render
		}
	}
	// 3. Biome tile (default — handled in renderMap)
	return ""

render:
	if isCursor {
		return cursorStyle.Render(string(symbol))
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(colorStr)).Render(string(symbol))
}

func renderSettlementView(m Model) string {
	if m.settlementViewState == nil {
		return ""
	}

	radius := m.settlementViewState.ViewportRadius
	size := radius*2 + 1
	cx := m.settlementViewState.SettlementCenterX
	cy := m.settlementViewState.SettlementCenterY

	// Build settlement header from ECS
	header := ""
	setType := ""
	if m.ecsWorld != nil {
		if setComp, ok := ecs.GetComponent[settlement.Settlement](m.ecsWorld, m.settlementViewState.SettlementEntity); ok {
			typeName := setComp.Type // "village" | "town" | "city"
			switch typeName {
			case "village":
				setType = "Aldea"
			case "town":
				setType = "Pueblo"
			case "city":
				setType = "Ciudad"
			default:
				setType = typeName
			}

			// Check for resource info
			foodStr, goldStr, toolsStr := "", "", ""
			if rs, ok := ecs.GetComponent[settlement.ResourceStore](m.ecsWorld, m.settlementViewState.SettlementEntity); ok {
				foodStr = fmt.Sprintf(" Food:%.0f", rs.Resources["food"])
				goldStr = fmt.Sprintf(" Gold:%.0f", rs.Resources["gold"])
				toolsStr = fmt.Sprintf(" Tools:%.0f", rs.Resources["tools"])
			}

			header = fmt.Sprintf("%s │ %s │ Pop:%d │ Lvl:%d%s%s%s",
				setComp.Name, setType, setComp.Population, setComp.Level,
				foodStr, goldStr, toolsStr)
		}
	}
	if header == "" {
		header = "Settlement Interior"
	}
	// Truncate header to terminal width
	if len(header) > m.width {
		header = header[:m.width]
	}
	headerLine := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Render(header)

	// Build lookup maps
	buildingMap := make(map[string]settlement.BuildingRenderInfo)
	for _, b := range m.settlementBuildings {
		key := fmt.Sprintf("%d,%d", b.WorldX, b.WorldY)
		buildingMap[key] = b
	}
	npcMap := make(map[string]npc.NPCRenderInfo)
	for _, n := range m.settlementNPCs {
		key := fmt.Sprintf("%d,%d", n.WorldX, n.WorldY)
		npcMap[key] = n
	}
	popupMap := make(map[string]RewardPopup)
	for _, p := range m.rewardPopups {
		key := fmt.Sprintf("%d,%d", p.WorldX, p.WorldY)
		popupMap[key] = p
	}

	// Determine viewport display bounds
	// Reserve: 1 line header + bottom area (status bar or inspector)
	bottomLines := 1 // status bar is always 1 line
	if m.inspectorOpen {
		// Pre-render inspector to count its actual height
		inspectorText := renderInspector(m)
		bottomLines = strings.Count(inspectorText, "\n") + 1
		if inspectorText == "" {
			bottomLines = 0
		}
	}
	reservedLines := 1 + bottomLines // header + bottom
	maxRows := m.height - reservedLines
	if maxRows < 3 {
		maxRows = 3 // never go below 3 rows for the grid
	}

	startRow := 0
	endRow := size
	if size < maxRows {
		startRow = (maxRows - size) / 2
		endRow = startRow + size
	}
	if endRow > maxRows {
		endRow = maxRows
	}

	var lines []string
	for i := 0; i < startRow; i++ {
		lines = append(lines, strings.Repeat(" ", size))
	}

	for vy := 0; vy < size && len(lines) < endRow; vy++ {
		var line strings.Builder
		for vx := 0; vx < size; vx++ {
			wx := vx + cx - radius
			wy := vy + cy - radius

			isCursor := vx == m.settlementViewState.CursorX && vy == m.settlementViewState.CursorY
			key := fmt.Sprintf("%d,%d", wx, wy)

			// Priority: NPC > Building > Popup > Biome
			if n, ok := npcMap[key]; ok {
				if isCursor {
					line.WriteString(cursorStyle.Render(string(n.Symbol)))
				} else {
					line.WriteString(lipgloss.NewStyle().Foreground(n.Color).Render(string(n.Symbol)))
				}
				continue
			}

			if b, ok := buildingMap[key]; ok {
				if isCursor {
					line.WriteString(cursorStyle.Render(string(b.Symbol)))
				} else {
					line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(b.Color)).Render(string(b.Symbol)))
				}
				continue
			}

			if p, ok := popupMap[key]; ok {
				line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render(p.Text))
				continue
			}

			// Biome background
			if m.worldMap != nil && m.worldMap.InBounds(wx, wy) {
				tile := m.worldMap.TileAt(wx, wy)
				if tile != nil {
					style, ok := biomeStyles[tile.BiomeID]
					if !ok {
						style = biomeStyles["unknown"]
					}
					if isCursor {
						line.WriteString(cursorStyle.Render(style.symbol))
					} else {
						line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(style.color)).Render(style.symbol))
					}
					continue
				}
			}

			if isCursor {
				line.WriteString(cursorStyle.Render(" "))
			} else {
				line.WriteString(" ")
			}
		}
		lines = append(lines, line.String())
	}

	for len(lines) < maxRows {
		lines = append(lines, strings.Repeat(" ", size))
	}

	result := headerLine + "\n" + strings.Join(lines, "\n")

	if m.inspectorOpen {
		result += "\n" + renderInspector(m)
	} else {
		result += "\n" + renderSettlementStatusBar(m)
	}

	return result
}

func renderSettlementStatusBar(m Model) string {
	if m.settlementViewState == nil {
		return statusBarStyle.Render(strings.Repeat(" ", m.width))
	}

	radius := m.settlementViewState.ViewportRadius
	vx := m.settlementViewState.CursorX
	vy := m.settlementViewState.CursorY
	wx := vx + m.settlementViewState.SettlementCenterX - radius
	wy := vy + m.settlementViewState.SettlementCenterY - radius

	var parts []string
	parts = append(parts, fmt.Sprintf("(%d,%d)", vx, vy))

	// Find entity under cursor (NPC priority)
	var entityName string
	for _, n := range m.settlementNPCs {
		if n.WorldX == wx && n.WorldY == wy {
			name := "NPC"
			if m.ecsWorld != nil {
				if nameComp, ok := ecs.GetComponent[ecs.Name](m.ecsWorld, n.Entity); ok {
					name = nameComp.Name
				}
			}
			entityName = name
			break
		}
	}
	if entityName == "" {
		for _, b := range m.settlementBuildings {
			if b.WorldX == wx && b.WorldY == wy {
				entityName = b.Name
				// Show worker count for this building
				if b.Role != "" {
					// Use WorkersInside from BuildingInterior (actual workers inside)
					// Fall back to role-based count if WorkersInside is 0
					workerCount := b.WorkersInside
					if workerCount == 0 {
						for _, n := range m.settlementNPCs {
							if n.JobRole == b.Role {
								workerCount++
							}
						}
					}
					parts = append(parts, entityName)
					parts = append(parts, fmt.Sprintf("workers: %d/%d", workerCount, b.MaxWorkers))
				} else {
					parts = append(parts, entityName)
				}
				break
			}
		}
	} else {
		parts = append(parts, entityName)
	}

	// Recent rewards summary (last 3 ticks)
	var rewards []string
	for _, n := range m.settlementNPCs {
		if n.LastReward == 0 {
			continue
		}
		if m.simTick-n.RewardTick > 3 {
			continue
		}
		name := "?"
		if m.ecsWorld != nil {
			if nameComp, ok := ecs.GetComponent[ecs.Name](m.ecsWorld, n.Entity); ok {
				name = nameComp.Name
			}
		}
		rewards = append(rewards, fmt.Sprintf("+%.2f %s", n.LastReward, name))
	}
	if len(rewards) > 0 {
		parts = append(parts, strings.Join(rewards, ", "))
	}

	status := strings.Join(parts, " ")
	if len(status) > m.width {
		status = status[:m.width]
	}
	padding := m.width - len(status)
	if padding < 0 {
		padding = 0
	}
	status += strings.Repeat(" ", padding)

	return statusBarStyle.Render(status)
}
