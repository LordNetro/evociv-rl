package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/simulation/npc"
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

			overlay := renderOverlay(m, worldX, worldY)
			if overlay != "" {
				line.WriteString(overlay)
				continue
			}

			tile := m.worldMap.TileAt(worldX, worldY)
			if tile == nil {
				line.WriteString(" ")
				continue
			}

			style, ok := biomeStyles[tile.BiomeID]
			if !ok {
				style = biomeStyles["unknown"]
			}
			styled := lipgloss.NewStyle().
				Foreground(lipgloss.Color(style.color)).
				Render(style.symbol)
			line.WriteString(styled)
		}
		lines = append(lines, line.String())
	}

	result := strings.Join(lines, "\n")
	if m.inspectorOpen {
		result += "\n" + renderInspector(m)
	}
	return result
}

// renderOverlay returns the styled NPC symbol if an NPC is present at the
// given world coordinate, otherwise an empty string.
func renderOverlay(m Model, worldX, worldY int) string {
	for _, info := range m.npcOverlay {
		if info.WorldX == worldX && info.WorldY == worldY {
			return lipgloss.NewStyle().
				Foreground(info.Color).
				Render(string(info.Symbol))
		}
	}
	return ""
}
