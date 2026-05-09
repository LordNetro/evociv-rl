package tilemap

import "fmt"

// Mock types for testing (also used by mock helper functions in this file)

// MockTile simulates world.Tile
type MockTile struct {
	BiomeID string
}

// MockWorldMap simulates world.WorldMap
type MockWorldMap struct {
	Width  int
	Height int
	Tiles  []MockTile
}

// MockNPC simulates an NPC entity with position
type MockNPC struct {
	ID       uint64
	X, Y     int
	Name     string
}

// MockWorldState simulates simulation.WorldState
type MockWorldState struct {
	NPCs []MockNPC
}

// MockBuilding simulates a building with footprint
type MockBuilding struct {
	ID       uint64
	X, Y     int // Top-left corner
	Width    int
	Height   int
	Name     string
}

// MockSettlement provides buildings
type MockSettlement struct {
	Buildings []MockBuilding
}

// WorldProvider interface abstracts world data access for TileBuilder.
// This allows mock implementations for testing.
type WorldProvider interface {
	GetWorldMap() (width int, height int, tiles func(x, y int) (biomeID string))
	GetNPCs() []NPCInfo
	GetBuildings() []BuildingInfo
	GetExploredTiles() map[string]bool
}

// NPCInfo represents NPC data from the simulation
type NPCInfo struct {
	ID   uint64
	X, Y int
	Name string
}

// BuildingInfo represents building data from settlements
type BuildingInfo struct {
	ID     uint64
	X, Y   int // top-left corner
	Width  int
	Height int
	Name   string
}

// TileBuilder populates a Tilemap from world data
type TileBuilder struct {
	world WorldProvider
}

// NewTileBuilder creates a new TileBuilder with the given world provider
func NewTileBuilder(world WorldProvider) *TileBuilder {
	return &TileBuilder{world: world}
}

// Build populates the tilemap with data from the world provider
func (b *TileBuilder) Build(tilemap *Tilemap) error {
	// Populate terrain layer from world map
	width, height, tilesFn := b.world.GetWorldMap()

	// Ensure tilemap has correct dimensions
	if tilemap.Width() != width || tilemap.Height() != height {
		// Rebuild tilemap with correct dimensions
		*tilemap = *NewTilemap(width, height)
	}

	// Layer 0: Terrain - map biome IDs to ASCII characters
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			biomeID := tilesFn(x, y)
			char := biomeToChar(biomeID)
			tilemap.SetTile(x, y, 0, LayerTerrain, char)
		}
	}

	// Layer 3: Creature - place NPCs
	for _, npc := range b.world.GetNPCs() {
		if npc.X >= 0 && npc.X < width && npc.Y >= 0 && npc.Y < height {
			tilemap.SetTile(npc.X, npc.Y, 0, LayerCreature, '@')
		}
	}

	// Layer 1: Building - render building footprints
	for _, bldg := range b.world.GetBuildings() {
		b.renderBuildingFootprint(tilemap, bldg)
	}

	// Layer 4: Fog - mark explored vs unexplored tiles
	explored := b.world.GetExploredTiles()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			key := coordKey(x, y)
			if explored[key] {
				tilemap.SetFog(x, y, 0, '.') // explored
			} else {
				tilemap.SetFog(x, y, 0, ':') // unexplored/dark
			}
		}
	}

	return nil
}

// renderBuildingFootprint renders a building with corners as '+', edges as '#', interior as '.'
func (b *TileBuilder) renderBuildingFootprint(tilemap *Tilemap, bldg BuildingInfo) {
	x, y := bldg.X, bldg.Y
	w, h := bldg.Width, bldg.Height

	// Use BuildingFootprint helper for corner/edge detection
	fp := NewBuildingFootprint(x, y, w, h, '#', 0)

	// Render the building footprint
	for by := 0; by < h; by++ {
		for bx := 0; bx < w; bx++ {
			// Calculate world coordinates
			tx := x + bx
			ty := y + by

			// Skip if out of bounds
			if tx < 0 || tx >= tilemap.Width() || ty < 0 || ty >= tilemap.Height() {
				continue
			}

			var char rune
			if fp.IsCorner(bx, by) {
				char = '+' // corners
			} else if fp.IsEdge(bx, by) {
				char = '#' // edges
			} else {
				char = '.' // interior/floor
			}

			tilemap.SetTile(tx, ty, 0, LayerBuilding, char)
		}
	}
}

// biomeToChar maps biome ID to ASCII character
func biomeToChar(biomeID string) rune {
	switch biomeID {
	case "ocean":
		return '~'
	case "plains":
		return '.'
	case "forest":
		return 'T'
	case "mountain":
		return '#'
	case "hill":
		return '^'
	case "desert":
		return 's'
	case "tundra":
		return ','
	case "taiga":
		return 't'
	case "jungle":
		return 'J'
	case "swamp":
		return '~'
	default:
		return '.' // default to plains
	}
}

// coordKey creates a string key for coordinate pair
func coordKey(x, y int) string {
	return fmt.Sprintf("%d,%d", x, y)
}

// Mock implementation for testing

// mockWorldWrapper provides mock world data for testing
type mockWorldWrapper struct {
	worldMap   MockWorldMap
	state      *MockWorldState
	settlement *MockSettlement
	explored   map[string]bool
}

func (m *mockWorldWrapper) GetWorldMap() (int, int, func(x, y int) string) {
	width := m.worldMap.Width
	height := m.worldMap.Height
	tiles := m.worldMap.Tiles
	return width, height, func(x, y int) string {
		idx := y*width + x
		if idx >= 0 && idx < len(tiles) {
			return tiles[idx].BiomeID
		}
		return ""
	}
}

func (m *mockWorldWrapper) GetNPCs() []NPCInfo {
	npcs := make([]NPCInfo, len(m.state.NPCs))
	for i, n := range m.state.NPCs {
		npcs[i] = NPCInfo{
			ID:   n.ID,
			X:    n.X,
			Y:    n.Y,
			Name: n.Name,
		}
	}
	return npcs
}

func (m *mockWorldWrapper) GetBuildings() []BuildingInfo {
	if m.settlement == nil {
		return nil
	}
	buildings := make([]BuildingInfo, len(m.settlement.Buildings))
	for i, b := range m.settlement.Buildings {
		buildings[i] = BuildingInfo{
			ID:     b.ID,
			X:      b.X,
			Y:      b.Y,
			Width:  b.Width,
			Height: b.Height,
			Name:   b.Name,
		}
	}
	return buildings
}

func (m *mockWorldWrapper) GetExploredTiles() map[string]bool {
	if m.explored == nil {
		// Default: all tiles explored
		exp := make(map[string]bool)
		for y := 0; y < m.worldMap.Height; y++ {
			for x := 0; x < m.worldMap.Width; x++ {
				exp[coordKey(x, y)] = true
			}
		}
		return exp
	}
	return m.explored
}