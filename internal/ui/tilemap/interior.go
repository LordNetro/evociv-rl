package tilemap

import (
	"errors"
)

// ErrInvalidBuildingID is returned when an invalid building ID is requested
var ErrInvalidBuildingID = errors.New("invalid building ID")

// Mock types for interior testing

// MockBuildingInterior simulates a building interior grid
type MockBuildingInterior struct {
	BuildingID uint64
	Width      int
	Height     int
	Cells      [][]CellType
}

// MockInteriorWorldState simulates world state for interior testing
type MockInteriorWorldState struct {
	Buildings   []MockBuildingInterior
	NPCsInBldg map[uint64][]NPCInfo
}

// InteriorProvider provides building interior data
type InteriorProvider interface {
	GetBuildingInterior(buildingID uint64) (MockBuildingInterior, bool)
	GetNPCsInBuilding(buildingID uint64) []NPCInfo
}

// InteriorRenderer renders Z=1 — building interior view
type InteriorRenderer struct {
	state InteriorProvider
}

// NewInteriorRenderer creates a new InteriorRenderer with the given state
func NewInteriorRenderer(state InteriorProvider) *InteriorRenderer {
	return &InteriorRenderer{state: state}
}

// RenderInterior renders the interior of a building and returns a 2D grid of CellType.
// The grid includes the interior layout plus any NPCs currently in the building.
// Returns ErrInvalidBuildingID if the building does not exist.
func (r *InteriorRenderer) RenderInterior(buildingID uint64, zLevel int) ([][]CellType, error) {
	// Validate building ID
	if buildingID == 0 {
		return nil, ErrInvalidBuildingID
	}

	// Get building interior data
	interior, ok := r.state.GetBuildingInterior(buildingID)
	if !ok {
		return nil, ErrInvalidBuildingID
	}

	// Create the grid from interior data
	grid := make([][]CellType, interior.Height)
	for y := 0; y < interior.Height; y++ {
		grid[y] = make([]CellType, interior.Width)
		copy(grid[y], interior.Cells[y])
	}

	// Place NPCs in the grid
	npcs := r.state.GetNPCsInBuilding(buildingID)
	for _, npc := range npcs {
		if npc.X >= 0 && npc.X < interior.Width && npc.Y >= 0 && npc.Y < interior.Height {
			// NPC is represented as CellType('@') to distinguish from floor/wall
			grid[npc.Y][npc.X] = CellType('@')
		}
	}

	return grid, nil
}

// GetNPCsInBuilding returns all NPCs currently in a specific building.
// Queries the ECS for NPCs where AIState.CurrentBuilding == buildingID.
func (r *InteriorRenderer) GetNPCsInBuilding(buildingID uint64) []NPCInfo {
	if buildingID == 0 {
		return nil
	}
	return r.state.GetNPCsInBuilding(buildingID)
}

// mockInteriorProvider provides mock interior data for testing
type mockInteriorProvider struct {
	buildings map[uint64]MockBuildingInterior
	npcs      map[uint64][]NPCInfo
}

func (m *mockInteriorProvider) GetBuildingInterior(buildingID uint64) (MockBuildingInterior, bool) {
	b, ok := m.buildings[buildingID]
	return b, ok
}

func (m *mockInteriorProvider) GetNPCsInBuilding(buildingID uint64) []NPCInfo {
	return m.npcs[buildingID]
}

// Convert MockInteriorWorldState to InteriorProvider interface
func (s *MockInteriorWorldState) toProvider() InteriorProvider {
	buildings := make(map[uint64]MockBuildingInterior)
	for _, b := range s.Buildings {
		buildings[b.BuildingID] = b
	}
	return &mockInteriorProvider{
		buildings: buildings,
		npcs:      s.NPCsInBldg,
	}
}