package settlement

import (
	"container/heap"
	"fmt"

	"github.com/marco/evociv-rl/internal/ecs"
)

// pos represents a 2D position for pathfinding.
type pos struct {
	x, y int
}

// node represents an A* search node.
type node struct {
	pos       pos
	g         int // cost from start
	f         int // total cost (g + h)
	parent    *node
	heapIndex int // for heap.Interface
}

// pathResult holds a cached path result.
type pathResult struct {
	path []pos
	err  error
}

// IndoorPathfinder provides A* pathfinding within building interiors.
type IndoorPathfinder struct {
	cache map[string]pathResult
}

// NewIndoorPathfinder creates a new pathfinder with an empty cache.
func NewIndoorPathfinder() *IndoorPathfinder {
	return &IndoorPathfinder{
		cache: make(map[string]pathResult),
	}
}

// FindPath finds a path between two door positions using A* algorithm.
// Returns the path as a slice of positions, or an error if no path exists.
func (pf *IndoorPathfinder) FindPath(grid [][]CellType, from, to DoorPosition) ([]pos, error) {
	if grid == nil {
		return nil, fmt.Errorf("grid is nil")
	}
	if len(grid) == 0 || len(grid[0]) == 0 {
		return nil, fmt.Errorf("grid has zero dimensions")
	}

	// Check if from and to are valid
	if !isWalkable(grid, from.GridX, from.GridY) {
		return nil, fmt.Errorf("start position is not walkable")
	}
	if !isWalkable(grid, to.GridX, to.GridY) {
		return nil, fmt.Errorf("target position is not walkable")
	}

	// Check cache
	cacheKey := pf.cacheKey(grid, from, to)
	if result, ok := pf.cache[cacheKey]; ok {
		return result.path, result.err
	}

	// Run A* search
	path, err := pf.aStar(grid, pos{x: from.GridX, y: from.GridY}, pos{x: to.GridX, y: to.GridY})

	// Cache the result
	pf.cache[cacheKey] = pathResult{path: path, err: err}

	return path, err
}

// aStar performs A* search from start to goal.
func (pf *IndoorPathfinder) aStar(grid [][]CellType, start, goal pos) ([]pos, error) {
	open := &nodeHeap{}
	closed := make(map[pos]*node)
	startNode := &node{
		pos: start,
		g:   0,
		f:   heuristic(start, goal),
	}
	heap.Init(open)
	heap.Push(open, startNode)
	closed[start] = startNode

	for open.Len() > 0 {
		current := heap.Pop(open).(*node)

		// Goal reached
		if current.pos == goal {
			return reconstructPath(current), nil
		}

		// Check neighbors
		for _, neighbor := range getNeighbors(current.pos, grid) {
			g := current.g + 1 // uniform cost

			if existing, ok := closed[neighbor]; ok {
				if g < existing.g {
					// Found a better path
					existing.g = g
					existing.f = g + heuristic(neighbor, goal)
					existing.parent = current
					heap.Push(open, existing)
				}
			} else {
				neighborNode := &node{
					pos:    neighbor,
					g:      g,
					f:      g + heuristic(neighbor, goal),
					parent: current,
				}
				closed[neighbor] = neighborNode
				heap.Push(open, neighborNode)
			}
		}
	}

	return nil, fmt.Errorf("no path found from (%d,%d) to (%d,%d)", start.x, start.y, goal.x, goal.y)
}

// heuristic calculates Manhattan distance between two positions.
func heuristic(a, b pos) int {
	dx := a.x - b.x
	if dx < 0 {
		dx = -dx
	}
	dy := a.y - b.y
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

// getNeighbors returns valid neighboring positions.
func getNeighbors(p pos, grid [][]CellType) []pos {
	height := len(grid)
	width := len(grid[0])

	var neighbors []pos
	// 4-directional movement (no diagonals for indoor pathfinding)
	directions := []pos{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

	for _, dir := range directions {
		nx, ny := p.x+dir.x, p.y+dir.y
		if nx >= 0 && nx < width && ny >= 0 && ny < height && isWalkable(grid, nx, ny) {
			neighbors = append(neighbors, pos{x: nx, y: ny})
		}
	}

	return neighbors
}

// isWalkable checks if a cell is walkable (floor or door).
func isWalkable(grid [][]CellType, x, y int) bool {
	if y < 0 || y >= len(grid) {
		return false
	}
	if x < 0 || x >= len(grid[y]) {
		return false
	}
	cell := grid[y][x]
	return cell == CellFloor || cell == CellDoor
}

// reconstructPath builds the path from goal to start by following parent pointers.
func reconstructPath(n *node) []pos {
	var path []pos
	for current := n; current != nil; current = current.parent {
		path = append([]pos{current.pos}, path...)
	}
	return path
}

// cacheKey generates a unique cache key for a path query.
func (pf *IndoorPathfinder) cacheKey(grid [][]CellType, from, to DoorPosition) string {
	return fmt.Sprintf("%dx%d-%dx%d", from.GridX, from.GridY, to.GridX, to.GridY)
}

// WorldEntryPos converts a world building position and door to world coordinates.
// For interior grid, coordinates are 1:1 with world tiles.
func (pf *IndoorPathfinder) WorldEntryPos(buildingPos ecs.Position, door DoorPosition) ecs.Position {
	return ecs.Position{
		X: buildingPos.X + float64(door.WorldX),
		Y: buildingPos.Y + float64(door.WorldY),
	}
}

// InteriorToWorld converts interior grid coordinates to world coordinates relative to building.
func (pf *IndoorPathfinder) InteriorToWorld(buildingPos ecs.Position, interiorX, interiorY int) ecs.Position {
	return ecs.Position{
		X: buildingPos.X + float64(interiorX),
		Y: buildingPos.Y + float64(interiorY),
	}
}

// WorldToInterior converts world coordinates to interior grid coordinates relative to building.
func (pf *IndoorPathfinder) WorldToInterior(buildingPos ecs.Position, worldX, worldY int) (int, int) {
	return worldX - int(buildingPos.X), worldY - int(buildingPos.Y)
}

// ClearCache removes all cached paths.
func (pf *IndoorPathfinder) ClearCache() {
	pf.cache = make(map[string]pathResult)
}

// nodeHeap implements a min-heap for A* open set.
type nodeHeap []*node

func (h nodeHeap) Len() int { return len(h) }

func (h nodeHeap) Less(i, j int) bool {
	return h[i].f < h[j].f
}

func (h nodeHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].heapIndex = i
	h[j].heapIndex = j
}

func (h *nodeHeap) Push(x any) {
	n := len(*h)
	node := x.(*node)
	node.heapIndex = n
	*h = append(*h, node)
}

func (h *nodeHeap) Pop() any {
	old := *h
	n := len(old)
	node := old[n-1]
	old[n-1] = nil
	*h = old[0 : n-1]
	return node
}