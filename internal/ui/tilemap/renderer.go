package tilemap

import "github.com/charmbracelet/lipgloss"

// StyleConfig holds the lipgloss style configuration for rendering
type StyleConfig struct {
	TerrainBG   lipgloss.Color
	BuildingBG  lipgloss.Color
	CreatureFG  lipgloss.Color
	FogExplored lipgloss.Color
	FogDark     lipgloss.Color
	DefaultStyle lipgloss.Style
}

// DefaultStyleConfig returns sensible defaults for DF-style rendering
func DefaultStyleConfig() *StyleConfig {
	return &StyleConfig{
		TerrainBG:   lipgloss.Color("212"),
		BuildingBG:  lipgloss.Color("250"),
		CreatureFG:  lipgloss.Color("226"), // yellow for NPCs
		FogExplored: lipgloss.Color("244"),
		FogDark:     lipgloss.Color("235"),
		DefaultStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
	}
}

// RenderViewport renders the tilemap for the camera viewport
// For each tile in camera viewport:
//   - Check layers in priority order: Creature > Building > Item > Terrain
//   - If Creature != 0 → show creature (lipgloss yellow for '@')
//   - Else if Building != 0 → show building (lipgloss white for '#')
//   - Else if Item != 0 → show item
//   - Else if Terrain != 0 → show terrain
//   - Else → show ' ' (space)
//   - Apply fog: if Fog == ':' → dimmed char
//   - Return string with newlines between rows
func RenderViewport(t *Tilemap, cam *Camera, style *StyleConfig) string {
	if t == nil || cam == nil || style == nil {
		return ""
	}

	viewport := cam.Viewport(t)
	if viewport == nil {
		return ""
	}

	var lines []string
	for y := 0; y < len(viewport); y++ {
		var line string
		for x := 0; x < len(viewport[y]); x++ {
			tile := viewport[y][x]
			char := renderTile(tile, style)
			line += char
		}
		lines = append(lines, line)
	}

	return joinLines(lines)
}

// renderTile returns the character to render for a single tile based on layer priority
func renderTile(tile *Tile, style *StyleConfig) string {
	if tile == nil {
		return " "
	}

	// Check layers in priority order (highest to lowest)
	// Priority: Creature > Building > Item > Terrain
if tile.Creature != 0 {
		return style.DefaultStyle.Copy().
			Foreground(style.CreatureFG).
			Render(string(tile.Creature))
	}
	if tile.Building != 0 {
		return style.DefaultStyle.Copy().
			Background(style.BuildingBG).
			Render(string(tile.Building))
	}
	if tile.Item != 0 {
		return style.DefaultStyle.Copy().
			Render(string(tile.Item))
	}
	if tile.Terrain != 0 {
		return style.DefaultStyle.Copy().
			Background(style.TerrainBG).
			Render(string(tile.Terrain))
	}

	// Apply fog if present
	if tile.Fog != 0 {
		switch tile.Fog {
		case ':': // dark/unexplored
			return style.DefaultStyle.Copy().
				Foreground(style.FogDark).
				Render(" ")
		case '.': // explored
			return style.DefaultStyle.Copy().
				Foreground(style.FogExplored).
				Render(" ")
		}
	}

	// Empty tile
	return " "
}

// RenderInterior renders Z=1 interior grid with NPC overlay
func RenderInterior(t *Tilemap, cam *Camera, grid [][]CellType, style *StyleConfig) string {
	if t == nil || cam == nil || style == nil || grid == nil {
		return ""
	}

	var lines []string
	for y := 0; y < len(grid); y++ {
		var line string
		for x := 0; x < len(grid[y]); x++ {
			cell := grid[y][x]

			// Check if there's a creature at this position in Z=1
			tile := t.TileAt(x, y, 1)
			char := renderInteriorCell(cell, tile, style)
			line += char
		}
		lines = append(lines, line)
	}

	return joinLines(lines)
}

// renderInteriorCell returns the character to render for an interior cell
func renderInteriorCell(cell CellType, tile *Tile, style *StyleConfig) string {
	// Check if NPC is present (Creature layer in Z=1)
	if tile != nil && tile.Creature != 0 {
		return style.DefaultStyle.Copy().
			Foreground(style.CreatureFG).
			Render(string(tile.Creature))
	}

	// Render based on cell type
	switch cell {
	case CellWall:
		return style.DefaultStyle.Copy().
			Background(style.BuildingBG).
			Render("#")
	case CellDoor:
		return style.DefaultStyle.Copy().
			Render("+")
	case CellCorridor:
		return style.DefaultStyle.Copy().
			Render(".")
	case CellFloor:
		return style.DefaultStyle.Copy().
			Render(".")
	default:
		return " "
	}
}

// joinLines joins strings with newlines
func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}