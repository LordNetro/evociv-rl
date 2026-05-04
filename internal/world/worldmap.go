package world

// NewWorldMap creates a new WorldMap with the given dimensions.
// Tiles are stored in a contiguous slice of length width*height.
func NewWorldMap(w, h int) *WorldMap {
	return &WorldMap{
		Width:  w,
		Height: h,
		Tiles:  make([]Tile, w*h),
	}
}

// TileAt returns a pointer to the tile at (x, y), or nil if out of bounds.
func (m *WorldMap) TileAt(x, y int) *Tile {
	if !m.InBounds(x, y) {
		return nil
	}
	return &m.Tiles[y*m.Width+x]
}

// InBounds returns true if the coordinate is within the world map bounds.
func (m *WorldMap) InBounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < m.Width && y < m.Height
}
