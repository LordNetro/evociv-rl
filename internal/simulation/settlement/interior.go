package settlement

// CellType represents the type of cell in an interior grid.
// Based on Dwarf Fortress categories.
type CellType int

const (
	// CellFloor represents walkable floor tiles.
	CellFloor CellType = iota
	// CellWall represents non-walkable wall tiles.
	CellWall
	// CellDoor represents door tiles (entry/exit points).
	CellDoor
)

// DoorPosition represents a door's position in both interior grid coordinates
// and world coordinates (for exterior pathfinding).
type DoorPosition struct {
	// GridX, GridY are coordinates within the interior grid.
	GridX, GridY int
	// WorldX, WorldY are the absolute world coordinates.
	WorldX, WorldY int
}

// BuildingInterior represents the runtime interior state of a building.
// This is separate from Building (spawn config) to maintain backward compatibility.
type BuildingInterior struct {
	// Grid is the 2D interior layout, indexed by [y][x].
	Grid [][]CellType
	// Width is the number of columns in the grid.
	Width int
	// Height is the number of rows in the grid.
	Height int
	// Doors holds the entry/exit positions for this building.
	Doors []DoorPosition
	// WorkersInside is the current count of workers inside (0 to MaxWorkers).
	WorkersInside int
	// MaxWorkers is the maximum capacity of workers for this building.
	MaxWorkers int
	// BuildingSeed is used for deterministic interior generation.
	BuildingSeed int64
}

// BuildingInteriorID is declared in components.go