package tilemap

// Layer enum — 5 layers for rendering priority
type Layer int

const (
	LayerTerrain Layer = iota // 0
	LayerBuilding             // 1
	LayerItem                 // 2
	LayerCreature             // 3
	LayerUI                   // 4
)

// CellType enum — interior cell types for Z=1
type CellType uint8

const (
	CellFloor CellType = iota // 0
	CellWall                   // 1
	CellDoor                   // 2
	CellCorridor               // 3
)

// Tile struct — 5-layer cell
type Tile struct {
	Terrain  rune // Layer 0: '~' ocean, '.' plains, 'T' forest, '#' mountain
	Building rune // Layer 1: '+' door, 'o' window, '#' wall, 0=none
	Item     rune // Layer 2: '*' gold, ':' food, etc.
	Creature rune // Layer 3: '@' NPC, 'f' fish, 'w' wolf
	Fog      rune // Layer 4: ' ' visible, '.' explored, ':' dark
}

// NewTile returns a Tile with all zero bytes (default).
// This is the same as the zero value, but provides a constructor for clarity.
func NewTile() Tile {
	return Tile{}
}

// BuildingFootprint defines a rectangular building area for multi-tile rendering.
// Used by the builder to render building edges and corners.
type BuildingFootprint struct {
	X      int  // World X position (top-left corner)
	Y      int  // World Y position (top-left corner)
	Width  int  // Tile width
	Height int  // Tile height
	Symbol rune // Character for building display
	Color  int  // Color code for building
}

// NewBuildingFootprint creates a BuildingFootprint with the given dimensions and symbol.
func NewBuildingFootprint(x, y, w, h int, symbol rune, color int) BuildingFootprint {
	return BuildingFootprint{
		X:      x,
		Y:      y,
		Width:  w,
		Height: h,
		Symbol: symbol,
		Color:  color,
	}
}

// Contains checks if a world coordinate is within this building's footprint.
func (bf BuildingFootprint) Contains(wx, wy int) bool {
	return wx >= bf.X && wx < bf.X+bf.Width && wy >= bf.Y && wy < bf.Y+bf.Height
}

// IsCorner returns true if the given local coordinates are a corner of the footprint.
func (bf BuildingFootprint) IsCorner(lx, ly int) bool {
	return (lx == 0 || lx == bf.Width-1) && (ly == 0 || ly == bf.Height-1)
}

// IsEdge returns true if the given local coordinates are on an edge (but not corner).
func (bf BuildingFootprint) IsEdge(lx, ly int) bool {
	return (lx == 0 || lx == bf.Width-1 || ly == 0 || ly == bf.Height-1) && !bf.IsCorner(lx, ly)
}