package tilemap

import "errors"

// Camera represents a viewport into the tilemap.
type Camera struct {
	X, Y, Z int
	Width   int
	Height  int
}

// NewCamera creates a new camera with specified position and viewport size.
func NewCamera(x, y, z, w, h int) *Camera {
	return &Camera{
		X:      x,
		Y:      y,
		Z:      z,
		Width:  w,
		Height: h,
	}
}

// Viewport returns a 2D slice of tile pointers representing the visible area.
// Returns tiles clamped to map bounds if camera extends beyond map.
func (c *Camera) Viewport(t *Tilemap) [][]*Tile {
	if t == nil || c.Width <= 0 || c.Height <= 0 {
		return nil
	}

	// Calculate clamped bounds
	startX := c.X
	startY := c.Y
	endX := c.X + c.Width
	endY := c.Y + c.Height

	// Clamp to map bounds
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}
	if endX > t.Width() {
		endX = t.Width()
	}
	if endY > t.Height() {
		endY = t.Height()
	}

	// Handle empty viewport (map too small)
	if startX >= endX || startY >= endY {
		return nil
	}

	// Build viewport slice
	height := endY - startY
	width := endX - startX
	result := make([][]*Tile, height)
	for y := 0; y < height; y++ {
		result[y] = make([]*Tile, width)
		for x := 0; x < width; x++ {
			result[y][x] = t.TileAt(startX+x, startY+y, c.Z)
		}
	}

	return result
}

// SetZLevel sets the camera's Z level.
// Returns error if Z is not 0 or 1.
func (c *Camera) SetZLevel(z int) error {
	if z < 0 || z > 1 {
		return errors.New("invalid Z level: must be 0 (surface) or 1 (interior)")
	}
	c.Z = z
	return nil
}

// CenterOn moves the camera so the given world coordinates are centered in the viewport.
func (c *Camera) CenterOn(x, y int) {
	// Center: camera position = target - viewport/2
	c.X = x - c.Width/2
	c.Y = y - c.Height/2

	// Clamp to non-negative
	if c.X < 0 {
		c.X = 0
	}
	if c.Y < 0 {
		c.Y = 0
	}
}

// Move offsets the camera position by the given delta.
func (c *Camera) Move(dx, dy int) {
	c.X += dx
	c.Y += dy

	// Clamp to non-negative
	if c.X < 0 {
		c.X = 0
	}
	if c.Y < 0 {
		c.Y = 0
	}
}

// Bounds returns the current viewport bounds as (minX, minY, maxX, maxY).
// maxX and maxY are inclusive.
func (c *Camera) Bounds() (minX, minY, maxX, maxY int) {
	return c.X, c.Y, c.X + c.Width - 1, c.Y + c.Height - 1
}