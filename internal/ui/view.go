package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
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

	return strings.Join(lines, "\n")
}
