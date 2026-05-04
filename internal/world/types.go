package world

// Coord represents a 2D coordinate in the world grid.
type Coord struct {
	X, Y int
}

// Tile represents a single tile in the world map.
type Tile struct {
	Height      float64
	Humidity    float64
	Temperature float64
	BiomeID     string
}

// WorldMap is a 2D grid of tiles stored as a contiguous slice.
type WorldMap struct {
	Width  int
	Height int
	Tiles  []Tile
}
