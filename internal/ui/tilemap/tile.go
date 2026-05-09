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
	Terrain  byte // '~' ocean, '.' plains, 'T' forest, '#' mountain
	Building byte // '+' door, 'o' window, '#' wall, 0=none
	Item     byte // '*' gold, ':' food, etc.
	Creature byte // '@' NPC, 'f' fish, 'w' wolf
	Fog      byte // ' ' visible, '.' explored, ':' dark
}

// NewTile returns a Tile with all zero bytes (default).
// This is the same as the zero value, but provides a constructor for clarity.
func NewTile() Tile {
	return Tile{}
}