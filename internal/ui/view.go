package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/simulation/npc"
	"github.com/marco/evociv-rl/internal/simulation/settlement"
	"github.com/marco/evociv-rl/internal/ui/tilemap"
	"github.com/marco/evociv-rl/internal/world"
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

	// cursorStyle highlights the cursor position with a gold background.
	cursorStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#FFD700")).
		Foreground(lipgloss.Color("#000000"))

	// useTilemapRenderer is the feature flag for tilemap rendering
	// Default: false (use biome renderer). Set RENDER_TILEMAP=true to enable.
	useTilemapRenderer = os.Getenv("RENDER_TILEMAP") == "true"
)

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
	setComp, ok := ecs.GetComponent[settlement.Settlement](m.ecsWorld, m.selectedSettlement)
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

	return b.String()
}

func renderMap(m Model) string {
	// If tilemap view is initialized and feature flag enabled, use tilemap renderer
	if useTilemapRenderer && m.tilemapView != nil {
		return m.tilemapView.View()
	}

	// Fall back to existing biomeStyles rendering
	return renderMapBiome(m)
}

// renderMapBiome is the original biome-based map renderer with overlay support.
func renderMapBiome(m Model) string {
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
				result += "\n" + lipgloss.NewStyle().
					Background(lipgloss.Color("#333333")).
					Foreground(lipgloss.Color(info.Color)).
					Render(fmt.Sprintf(" %s %s | Pop: %d ", string(info.Symbol), info.Name, info.Population))
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

// initTilemapRenderer initializes the tilemap renderer if feature flag is enabled
// Returns nil if feature flag is disabled
func initTilemapRenderer(worldMap *world.WorldMap) *tilemap.TilemapView {
	if !useTilemapRenderer {
		return nil
	}

	if worldMap == nil {
		return nil
	}

	// Create tilemap from world dimensions
	tm := tilemap.NewTilemap(worldMap.Width, worldMap.Height)

	// Create camera
	cam := tilemap.NewCamera(0, 0, 0, 80, 24) // Default viewport size

	// Create and return tilemap view
	return &tilemap.TilemapView{
		Tilemap: tm,
		Camera:  cam,
	}
}

// SetTilemapView sets the tilemap view for the model
func (m *Model) SetTilemapView(tv *tilemap.TilemapView) {
	m.tilemapView = tv
}

// renderSettlementView renders the settlement interior with terrain + buildings + NPCs
func renderSettlementView(m Model) string {
	// Show settlement interior with terrain + buildings + NPCs
	if m.worldMap == nil {
		return "No world map"
	}

	// Use stored settlement info
	setInfo := m.settlementViewInfo

	// Get settlement radius from component
	radius := 5 // default
	if m.settlementViewEntity != 0 && m.ecsWorld != nil {
		if comp, ok := ecs.GetComponent[settlement.Settlement](m.ecsWorld, m.settlementViewEntity); ok {
			radius = comp.Radius
		}
	}
	radiusSq := radius * radius

	// Header with settlement info
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Bold(true).
		Render(fmt.Sprintf("[SETTLEMENT] %s - Radius: %d", setInfo.Name, radius))
	lines := []string{header}

	termHeight := m.height - 4 // Leave room for header and status bar
	termWidth := m.width

	// Render the settlement area with terrain + buildings + NPCs
	for row := 0; row < termHeight; row++ {
		worldY := m.settlementCameraY + row
		var line strings.Builder

		for col := 0; col < termWidth; col++ {
			worldX := m.settlementCameraX + col

			// Calculate distance from settlement center
			dx := worldX - setInfo.WorldX
			dy := worldY - setInfo.WorldY
			distSq := dx*dx + dy*dy

			showChar := ""
			charColor := ""
			bold := false

			// Check for NPC at this position (within settlement radius)
			if distSq <= radiusSq {
				for _, npcInfo := range m.npcOverlay {
					if npcInfo.WorldX == worldX && npcInfo.WorldY == worldY {
						showChar = "@"
						charColor = string(npcInfo.Color)
						bold = true
						break
					}
				}
			}

			// Check for settlement center
			if showChar == "" && distSq == 0 {
				showChar = string(setInfo.Symbol)
				charColor = setInfo.Color
				bold = true
			}

			// Show building edges (walls at the border of settlement radius)
			if showChar == "" && distSq <= radiusSq {
				// On the edge of settlement radius
				if distSq == radiusSq-1 || distSq == radiusSq-2 {
					showChar = "#"
					charColor = "#8B7355"
				} else if distSq < radiusSq-2 && distSq > 0 {
					// Inside settlement, show floor
					showChar = "."
					charColor = "#90EE90"
				}
			}

			// If we have a char to show, render it
			if showChar != "" {
				style := lipgloss.NewStyle().
					Foreground(lipgloss.Color(charColor))
				if bold {
					style = style.Bold(true)
				}
				line.WriteString(style.Render(showChar))
				continue
			}

			// Out of bounds check
			if !m.worldMap.InBounds(worldX, worldY) {
				line.WriteString(" ")
				continue
			}

			// Show terrain
			tile := m.worldMap.TileAt(worldX, worldY)
			if tile != nil {
				style := biomeStyles["unknown"]
				if s, ok := biomeStyles[tile.BiomeID]; ok {
					style = s
				}
				line.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color(style.color)).
					Render(style.symbol))
			} else {
				line.WriteString(" ")
			}
		}
		lines = append(lines, line.String())
	}

	// Footer with controls
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Render(fmt.Sprintf("[m] Map  [esc/q] Back  | %s | Pop: %d | Pos: [%d,%d]",
		setInfo.Name, setInfo.Population, m.settlementCameraX, m.settlementCameraY))
	lines = append(lines, footer)

	return lipgloss.JoinVertical(lipgloss.Top, lines...)
}

// findSettlementAt finds the settlement at the given world coordinates
func (m Model) findSettlementAt(wx, wy int) settlement.SettlementRenderInfo {
	return m.settlementViewInfo
}