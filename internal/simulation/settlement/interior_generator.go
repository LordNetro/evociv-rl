package settlement

import (
	"math/rand"
)

// InteriorGenerator implements seeded room placement for building interiors.
type InteriorGenerator struct{}

// Generate creates a deterministic interior layout based on the seed.
// The layout includes rooms, corridors, walls, and doors.
func (g *InteriorGenerator) Generate(seed int64, buildingID string, width, height int) BuildingInterior {
	// Create a deterministic RNG from seed and buildingID
	r := rand.New(rand.NewSource(xorHash(seed, hashBuildingID(buildingID))))

	// Initialize grid with walls
	grid := makeGrid(width, height)

	// Fill with floor tiles inside a border of walls
	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			grid[y][x] = CellFloor
		}
	}

	// Place rooms deterministically
	rooms := placeRooms(r, width, height)

	// Carve rooms into the grid (set as floor)
	for _, room := range rooms {
		carveRoom(grid, room)
	}

	// Connect rooms with corridors
	connectRooms(grid, rooms, r)

	// Add walls around floor tiles
	addWalls(grid, width, height)

	// Place doors on room edges
	doors := placeDoors(grid, rooms, width, height)

	return BuildingInterior{
		Grid:          grid,
		Width:         width,
		Height:        height,
		Doors:         doors,
		MaxWorkers:    2, // default, can be adjusted from BuildingDef
		WorkersInside: 0,
		BuildingSeed:  seed,
	}
}

// hashBuildingID creates a deterministic hash from buildingID for RNG seeding.
func hashBuildingID(id string) int64 {
	var h int64 = 5381
	for _, c := range id {
		h = ((h << 5) + h) + int64(c)
	}
	return h
}

// xorHash combines two int64 values into a single deterministic value.
func xorHash(a, b int64) int64 {
	return a ^ b
}

// room represents a rectangular room in the interior.
type room struct {
	x, y, w, h int
}

// makeGrid creates a 2D grid initialized with walls.
func makeGrid(width, height int) [][]CellType {
	grid := make([][]CellType, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]CellType, width)
		for x := 0; x < width; x++ {
			grid[y][x] = CellWall
		}
	}
	return grid
}

// placeRooms generates deterministic room placement based on RNG.
func placeRooms(r *rand.Rand, width, height int) []room {
	var rooms []room

	// Calculate available space (excluding border)
	availW := width - 2
	availH := height - 2

	// Handle very small interiors
	if availW <= 2 || availH <= 2 {
		// Create a single room that fills the available space
		roomW := max(1, availW)
		roomH := max(1, availH)
		roomX := 1
		roomY := 1
		return []room{{x: roomX, y: roomY, w: roomW, h: roomH}}
	}

	// Determine number of rooms based on interior size
	// Small: 1 room, medium: 1-2, large: 2-3
	numRooms := 1
	if availW >= 8 && availH >= 8 {
		numRooms = 2 + r.Intn(2) // 2-3 rooms
	} else if availW >= 6 || availH >= 6 {
		numRooms = 2
	}

	// Place rooms with some randomness but deterministically
	maxAttempts := 20
	for len(rooms) < numRooms && maxAttempts > 0 {
		// Random room size (min 2x2)
		roomW := 2 + r.Intn(min(availW-2, 4)) // 2 to min(5, availW-2)
		roomH := 2 + r.Intn(min(availH-2, 4)) // 2 to min(5, availH-2)

		// Ensure room fits
		if roomW > availW {
			roomW = availW
		}
		if roomH > availH {
			roomH = availH
		}

		// Random position ensuring room fits and leaves border
		maxX := availW - roomW - 1
		maxY := availH - roomH - 1
		if maxX < 1 {
			maxX = 1
		}
		if maxY < 1 {
			maxY = 1
		}
		roomX := 1 + r.Intn(maxX)
		roomY := 1 + r.Intn(maxY)

		newRoom := room{x: roomX, y: roomY, w: roomW, h: roomH}

		// Check for overlap with existing rooms
		overlaps := false
		for _, existing := range rooms {
			if roomsOverlap(newRoom, existing, 1) {
				overlaps = true
				break
			}
		}

		if !overlaps {
			rooms = append(rooms, newRoom)
		}
		maxAttempts--
	}

	// If no rooms were placed, create a single central room
	if len(rooms) == 0 {
		roomW := availW - 2
		roomH := availH - 2
		if roomW < 2 {
			roomW = 2
		}
		if roomH < 2 {
			roomH = 2
		}
		roomX := 1 + (availW-roomW)/2
		roomY := 1 + (availH-roomH)/2
		rooms = append(rooms, room{x: roomX, y: roomY, w: roomW, h: roomH})
	}

	return rooms
}

// roomsOverlap checks if two rooms overlap (with optional padding).
func roomsOverlap(a, b room, padding int) bool {
	return !(a.x+a.w+padding <= b.x || b.x+b.w+padding <= a.x ||
		a.y+a.h+padding <= b.y || b.y+b.h+padding <= a.y)
}

// carveRoom sets the cells inside a room to floor.
func carveRoom(grid [][]CellType, r room) {
	for y := r.y; y < r.y+r.h && y < len(grid); y++ {
		for x := r.x; x < r.x+r.w && x < len(grid[y]); x++ {
			grid[y][x] = CellFloor
		}
	}
}

// connectRooms creates corridors between rooms.
func connectRooms(grid [][]CellType, rooms []room, r *rand.Rand) {
	if len(rooms) < 2 {
		return
	}

	// Connect each room to the next using L-shaped corridors
	for i := 0; i < len(rooms)-1; i++ {
		curr := rooms[i]
		next := rooms[i+1]

		// Center of each room
		currCenterX := curr.x + curr.w/2
		currCenterY := curr.y + curr.h/2
		nextCenterX := next.x + next.w/2
		nextCenterY := next.y + next.h/2

		// Randomly choose horizontal-first or vertical-first
		if r.Intn(2) == 0 {
			// Horizontal then vertical
			carveHorizontalCorridor(grid, currCenterX, nextCenterX, currCenterY)
			carveVerticalCorridor(grid, currCenterY, nextCenterY, nextCenterX)
		} else {
			// Vertical then horizontal
			carveVerticalCorridor(grid, currCenterY, nextCenterY, currCenterX)
			carveHorizontalCorridor(grid, currCenterX, nextCenterX, nextCenterY)
		}
	}
}

// carveHorizontalCorridor carves a horizontal corridor at row y.
func carveHorizontalCorridor(grid [][]CellType, x1, x2, y int) {
	if y < 0 || y >= len(grid) {
		return
	}
	startX := min(x1, x2)
	endX := max(x1, x2)
	for x := startX; x <= endX && x < len(grid[y]); x++ {
		if x >= 0 {
			grid[y][x] = CellFloor
		}
	}
}

// carveVerticalCorridor carves a vertical corridor at column x.
func carveVerticalCorridor(grid [][]CellType, y1, y2, x int) {
	startY := min(y1, y2)
	endY := max(y1, y2)
	for y := startY; y <= endY && y < len(grid); y++ {
		if x >= 0 && x < len(grid[y]) {
			grid[y][x] = CellFloor
		}
	}
}

// addWalls adds wall tiles around floor tiles.
func addWalls(grid [][]CellType, width, height int) {
	// Add walls around the border (already set to wall, but ensure corners are walls)
	for x := 0; x < width; x++ {
		if grid[0][x] == CellFloor {
			grid[0][x] = CellWall
		}
		if grid[height-1][x] == CellFloor {
			grid[height-1][x] = CellWall
		}
	}
	for y := 0; y < height; y++ {
		if grid[y][0] == CellFloor {
			grid[y][0] = CellWall
		}
		if grid[y][width-1] == CellFloor {
			grid[y][width-1] = CellWall
		}
	}
}

// placeDoors identifies door positions on room edges.
func placeDoors(grid [][]CellType, rooms []room, width, height int) []DoorPosition {
	var doors []DoorPosition

	for _, r := range rooms {
		// Find possible door positions on room edges
		// Check top and bottom edges
		for x := r.x; x < r.x+r.w; x++ {
			// Top edge
			if r.y > 0 && grid[r.y-1][x] == CellFloor {
				doors = append(doors, DoorPosition{
					GridX: x, GridY: r.y - 1,
					WorldX: x, WorldY: r.y - 1,
				})
			}
			// Bottom edge
			if r.y+r.h < height && grid[r.y+r.h][x] == CellFloor {
				doors = append(doors, DoorPosition{
					GridX: x, GridY: r.y + r.h,
					WorldX: x, WorldY: r.y + r.h,
				})
			}
		}

		// Check left and right edges
		for y := r.y; y < r.y+r.h; y++ {
			// Left edge
			if r.x > 0 && grid[y][r.x-1] == CellFloor {
				doors = append(doors, DoorPosition{
					GridX: r.x - 1, GridY: y,
					WorldX: r.x - 1, WorldY: y,
				})
			}
			// Right edge
			if r.x+r.w < width && grid[y][r.x+r.w] == CellFloor {
				doors = append(doors, DoorPosition{
					GridX: r.x + r.w, GridY: y,
					WorldX: r.x + r.w, WorldY: y,
				})
			}
		}
	}

	// Deduplicate doors (same position)
	seen := make(map[string]bool)
	uniqueDoors := make([]DoorPosition, 0, len(doors))
	for _, d := range doors {
		key := coordKey(d.GridX, d.GridY)
		if !seen[key] {
			seen[key] = true
			uniqueDoors = append(uniqueDoors, d)
		}
	}

	// Limit to reasonable number of doors (max 4)
	if len(uniqueDoors) > 4 {
		uniqueDoors = uniqueDoors[:4]
	}

	return uniqueDoors
}

// coordKey creates a unique string key for coordinates.
func coordKey(x, y int) string {
	return string(rune(x+'A')) + string(rune(y+'A')) // Simple mapping
}

// Ensure InteriorGenerator implements InteriorGeneratorInterface
var _ InteriorGeneratorInterface = (*InteriorGenerator)(nil)

// InteriorGeneratorInterface defines the contract for interior generation.
type InteriorGeneratorInterface interface {
	Generate(seed int64, buildingID string, width, height int) BuildingInterior
}

// DefaultInteriorGenerator is the global generator instance used by SpawnSystem.
var DefaultInteriorGenerator InteriorGeneratorInterface = &InteriorGenerator{}