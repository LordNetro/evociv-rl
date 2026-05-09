package tilemap

// Tilemap holds a 2D grid of tiles across multiple Z levels.
type Tilemap struct {
	width  int
	height int
	levels map[int][][]Tile // key=Z level (0=surface, 1=interior)
}

// NewTilemap creates a new tilemap with specified width and height.
// Initializes Z-level 0 (surface) by default.
func NewTilemap(width, height int) *Tilemap {
	t := &Tilemap{
		width:  width,
		height: height,
		levels: make(map[int][][]Tile),
	}
	// Create Z=0 level by default
	t.levels[0] = make([][]Tile, height)
	for y := 0; y < height; y++ {
		t.levels[0][y] = make([]Tile, width)
	}
	return t
}

// TileAt returns a pointer to the tile at (x, y, z).
// Returns nil if coordinates are out of bounds or Z level doesn't exist.
func (t *Tilemap) TileAt(x, y, z int) *Tile {
	if x < 0 || x >= t.width || y < 0 || y >= t.height {
		return nil
	}
	level, ok := t.levels[z]
	if !ok {
		return nil
	}
	return &level[y][x]
}

// SetTile sets the character for a specific layer at (x, y, z).
// Creates the Z level if it doesn't exist.
func (t *Tilemap) SetTile(x, y, z int, layer Layer, char byte) {
	// Ensure Z level exists
	level := t.setZLevel(z)
	if x < 0 || x >= t.width || y < 0 || y >= t.height {
		return
	}
	switch layer {
	case LayerTerrain:
		level[y][x].Terrain = char
	case LayerBuilding:
		level[y][x].Building = char
	case LayerItem:
		level[y][x].Item = char
	case LayerCreature:
		level[y][x].Creature = char
	case LayerUI:
		// LayerUI is reserved for UI overlay (no corresponding field in Tile)
		// Fog field is reserved for visibility (' ' visible, '.' explored, ':' dark)
	}
}

// SetZLevel creates or returns the existing Z level slice.
// Returns the 2D slice for the specified Z level.
func (t *Tilemap) SetZLevel(z int) [][]Tile {
	return t.setZLevel(z)
}

// setZLevel is the internal helper that creates the level if needed.
func (t *Tilemap) setZLevel(z int) [][]Tile {
	if level, ok := t.levels[z]; ok {
		return level
	}
	// Create new level
	level := make([][]Tile, t.height)
	for y := 0; y < t.height; y++ {
		level[y] = make([]Tile, t.width)
	}
	t.levels[z] = level
	return level
}

// Width returns the map width.
func (t *Tilemap) Width() int {
	return t.width
}

// Height returns the map height.
func (t *Tilemap) Height() int {
	return t.height
}

// ZLevels returns all created Z levels as a map.
// Key is the Z level index, value is the 2D tile slice.
func (t *Tilemap) ZLevels() map[int][][]Tile {
	return t.levels
}