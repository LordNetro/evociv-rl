package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/simulation/npc"
	"github.com/marco/evociv-rl/internal/simulation/settlement"
	"github.com/marco/evociv-rl/internal/world"
)

func TestViewContainsTitle(t *testing.T) {
	m := NewModel()
	m.width = 80
	m.height = 24
	m.ready = true
	v := m.View()
	if !strings.Contains(v, "Evociv-RL") {
		t.Error("view missing title 'Evociv-RL'")
	}
}

func TestViewContainsVersion(t *testing.T) {
	m := NewModel()
	m.width = 80
	m.height = 24
	m.ready = true
	v := m.View()
	if !strings.Contains(v, "v0.0.1") {
		t.Error("view missing version 'v0.0.1'")
	}
}

func TestViewContainsSubtitle(t *testing.T) {
	m := NewModel()
	m.width = 80
	m.height = 24
	m.ready = true
	v := m.View()
	if !strings.Contains(v, "Un mundo por descubrir") {
		t.Error("view missing subtitle")
	}
}

func TestViewContainsInstructions(t *testing.T) {
	m := NewModel()
	m.width = 80
	m.height = 24
	m.ready = true
	v := m.View()
	if !strings.Contains(v, "[q] Salir") {
		t.Error("view missing instructions")
	}
}

func TestRenderMapShowsTiles(t *testing.T) {
	wm := world.NewWorldMap(3, 3)
	wm.TileAt(0, 0).BiomeID = "ocean"
	wm.TileAt(0, 0).Height = -0.5
	wm.TileAt(1, 0).BiomeID = "plains"
	wm.TileAt(1, 0).Height = 0.3
	wm.TileAt(2, 0).BiomeID = "forest"
	wm.TileAt(2, 0).Height = 0.5
	wm.TileAt(0, 1).BiomeID = "desert"
	wm.TileAt(0, 1).Height = 0.2
	wm.TileAt(1, 1).BiomeID = "tundra"
	wm.TileAt(1, 1).Height = 0.4
	wm.TileAt(2, 1).BiomeID = "jungle"
	wm.TileAt(2, 1).Height = 0.6
	wm.TileAt(0, 2).BiomeID = "ocean"
	wm.TileAt(1, 2).BiomeID = "plains"
	wm.TileAt(2, 2).BiomeID = "forest"

	m := NewModel()
	m.SetWorldMap(wm)
	m.screen = "map"
	m.width = 80
	m.height = 24
	m.ready = true

	v := m.View()
	// The map view should contain biome symbols, not the welcome title
	if strings.Contains(v, "Evociv-RL") {
		t.Error("map view should not contain welcome title")
	}
}

func TestRenderScreenToggle(t *testing.T) {
	// Welcome screen shows title
	m := NewModel()
	m.screen = "welcome"
	m.width = 80
	m.height = 24
	m.ready = true
	v := m.View()
	if !strings.Contains(v, "Evociv-RL") {
		t.Error("welcome view should contain title")
	}

	// Map screen does not show title
	wm := world.NewWorldMap(3, 3)
	wm.TileAt(0, 0).BiomeID = "ocean"
	m.SetWorldMap(wm)
	m.screen = "map"
	v = m.View()
	if strings.Contains(v, "Evociv-RL") {
		t.Error("map view should not contain welcome title")
	}
}

func TestRenderMapNoWorldMap(t *testing.T) {
	m := NewModel()
	m.screen = "map"
	m.width = 80
	m.height = 24
	m.ready = true
	v := m.View()
	if !strings.Contains(v, "Generando mundo") {
		t.Errorf("expected 'Generando mundo' message, got: %q", v)
	}
}

func TestRenderMapWindowAdapt(t *testing.T) {
	wm := world.NewWorldMap(20, 20)
	for y := 0; y < wm.Height; y++ {
		for x := 0; x < wm.Width; x++ {
			wm.TileAt(x, y).BiomeID = "plains"
		}
	}
	m := NewModel()
	m.SetWorldMap(wm)
	m.screen = "map"
	m.width = 80
	m.height = 24
	m.ready = true

	v := m.View()
	lines := strings.Split(v, "\n")
	// Should produce roughly the terminal height in lines (allowing for ANSI)
	if len(lines) < 20 {
		t.Errorf("expected at least ~20 lines, got %d", len(lines))
	}
}

func TestRenderMapGolden(t *testing.T) {
	wm := world.NewWorldMap(5, 5)
	for y := 0; y < wm.Height; y++ {
		for x := 0; x < wm.Width; x++ {
			wm.TileAt(x, y).BiomeID = "plains"
		}
	}
	m := NewModel()
	m.SetWorldMap(wm)
	m.screen = "map"
	m.width = 80
	m.height = 24
	m.ready = true

	output := m.View()
	goldenFile := filepath.Join("testdata", "map.golden")

	if *updateGolden {
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatalf("failed to create testdata: %v", err)
		}
		if err := os.WriteFile(goldenFile, []byte(output), 0644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
	}

	expected, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("reading golden file: %v", err)
	}
	if output != string(expected) {
		t.Errorf("output doesn't match golden file")
	}
}

func TestRenderOverlayWithNPC(t *testing.T) {
	wm := world.NewWorldMap(3, 3)
	for y := 0; y < wm.Height; y++ {
		for x := 0; x < wm.Width; x++ {
			wm.TileAt(x, y).BiomeID = "plains"
		}
	}
	m := NewModel()
	m.SetWorldMap(wm)
	m.screen = "map"
	m.width = 80
	m.height = 24
	m.ready = true
	m.npcOverlay = []npc.NPCRenderInfo{
		{WorldX: 1, WorldY: 1, Symbol: '@', Color: lipgloss.Color("#FF0000")},
	}

	overlay := renderOverlay(m, 1, 1)
	if overlay == "" {
		t.Error("expected non-empty overlay for tile with NPC")
	}
	if !strings.Contains(overlay, "@") {
		t.Error("expected overlay to contain '@'")
	}

	overlayEmpty := renderOverlay(m, 0, 0)
	if overlayEmpty != "" {
		t.Errorf("expected empty overlay for tile without NPC, got %q", overlayEmpty)
	}
}

func TestRenderMapWithOverlay(t *testing.T) {
	wm := world.NewWorldMap(3, 3)
	for y := 0; y < wm.Height; y++ {
		for x := 0; x < wm.Width; x++ {
			wm.TileAt(x, y).BiomeID = "plains"
		}
	}
	m := NewModel()
	m.SetWorldMap(wm)
	m.screen = "map"
	m.width = 80
	m.height = 24
	m.ready = true
	m.npcOverlay = []npc.NPCRenderInfo{
		{WorldX: 1, WorldY: 1, Symbol: '@', Color: lipgloss.Color("#FF0000")},
	}

	v := m.View()
	// The view should contain the '@' symbol somewhere
	if !strings.Contains(v, "@") {
		t.Error("map view with overlay should contain '@'")
	}
}

func TestRenderOverlayCameraOffset(t *testing.T) {
	wm := world.NewWorldMap(10, 10)
	for y := 0; y < wm.Height; y++ {
		for x := 0; x < wm.Width; x++ {
			wm.TileAt(x, y).BiomeID = "plains"
		}
	}
	m := NewModel()
	m.SetWorldMap(wm)
	m.screen = "map"
	m.cameraX = 5
	m.cameraY = 5
	m.width = 80
	m.height = 24
	m.ready = true
	m.npcOverlay = []npc.NPCRenderInfo{
		{WorldX: 5, WorldY: 5, Symbol: '@', Color: lipgloss.Color("#FF0000")},
	}

	// With camera at (5,5), the NPC at world (5,5) should appear at screen (0,0)
	overlay := renderOverlay(m, 5, 5)
	if overlay == "" {
		t.Error("expected overlay at world (5,5) with NPC")
	}
}

func TestRenderInspectorShowsData(t *testing.T) {
	w := ecs.NewWorld()
	npc.RegisterStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, ecs.Name{Name: "Aldric Torres"})
	ecs.AddComponent(w, e, npc.Health{Current: 80, Max: 100})
	ecs.AddComponent(w, e, npc.Job{Role: "farmer"})
	ecs.AddComponent(w, e, npc.Personality{Openness: 0.55, Conscientiousness: 0.60, Extraversion: 0.45, Agreeableness: 0.70, Neuroticism: 0.30})

	m := NewModel()
	m.inspectorOpen = true
	m.selectedNPC = e
	m.ecsWorld = w

	panel := renderInspector(m)
	if panel == "" {
		t.Fatal("expected non-empty inspector panel")
	}
	if !strings.Contains(panel, "Aldric Torres") {
		t.Error("inspector missing NPC name")
	}
	if !strings.Contains(panel, "farmer") {
		t.Error("inspector missing job")
	}
	if !strings.Contains(panel, "80/100") {
		t.Error("inspector missing health")
	}
	if !strings.Contains(panel, "O: 0.55") {
		t.Error("inspector missing personality openness")
	}
}

func TestRenderInspectorClosed(t *testing.T) {
	m := NewModel()
	m.inspectorOpen = false
	panel := renderInspector(m)
	if panel != "" {
		t.Errorf("expected empty panel when closed, got %q", panel)
	}
}

func TestRenderOverlaySettlementPriority(t *testing.T) {
	wm := world.NewWorldMap(3, 3)
	for y := 0; y < wm.Height; y++ {
		for x := 0; x < wm.Width; x++ {
			wm.TileAt(x, y).BiomeID = "plains"
		}
	}
	m := NewModel()
	m.SetWorldMap(wm)
	m.screen = "map"
	m.width = 80
	m.height = 24
	m.ready = true
	// NPC at (1,1)
	m.npcOverlay = []npc.NPCRenderInfo{
		{WorldX: 1, WorldY: 1, Symbol: '@', Color: lipgloss.Color("#FF0000")},
	}
	// Settlement at (1,1) — NPC should win
	m.settlementOverlay = []settlement.SettlementRenderInfo{
		{WorldX: 1, WorldY: 1, Symbol: '♦', Color: "#8B7355", Name: "Village"},
	}

	overlay := renderOverlay(m, 1, 1)
	if !strings.Contains(overlay, "@") {
		t.Error("expected NPC overlay to take priority over settlement")
	}

	// Settlement-only tile
	overlaySettle := renderOverlay(m, 0, 0)
	if overlaySettle != "" {
		// no settlement at (0,0), should be empty
	}
}

func TestRenderOverlaySettlementOnly(t *testing.T) {
	wm := world.NewWorldMap(3, 3)
	for y := 0; y < wm.Height; y++ {
		for x := 0; x < wm.Width; x++ {
			wm.TileAt(x, y).BiomeID = "plains"
		}
	}
	m := NewModel()
	m.SetWorldMap(wm)
	m.screen = "map"
	m.width = 80
	m.height = 24
	m.ready = true
	m.settlementOverlay = []settlement.SettlementRenderInfo{
		{WorldX: 2, WorldY: 2, Symbol: '▲', Color: "#B8860B", Name: "Town"},
	}

	overlay := renderOverlay(m, 2, 2)
	if overlay == "" {
		t.Error("expected non-empty overlay for tile with settlement")
	}
	if !strings.Contains(overlay, "▲") {
		t.Error("expected overlay to contain settlement symbol")
	}
}

func TestRenderInspectorSettlementData(t *testing.T) {
	w := ecs.NewWorld()
	npc.RegisterStores(w)
	ecs.RegisterComponentStore[ecs.Position](w, ecs.NewComponentID("position"), ecs.NewComponentStore[ecs.Position]())
	ecs.RegisterComponentStore[ecs.Name](w, ecs.NewComponentID("name"), ecs.NewComponentStore[ecs.Name]())
	settlement.RegisterSettlementStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, ecs.Name{Name: "Norte Aldea del Valle"})
	ecs.AddComponent(w, e, settlement.Settlement{
		Name: "Norte Aldea del Valle", Type: "village", Radius: 3,
		Population: 2, Level: 1, Buildings: []string{"house", "farm"},
	})

	m := NewModel()
	m.inspectorOpen = true
	m.selectedSettlement = int(e)
	m.ecsWorld = w

	panel := renderInspector(m)
	if panel == "" {
		t.Fatal("expected non-empty inspector panel")
	}
	if !strings.Contains(panel, "Norte Aldea del Valle") {
		t.Error("inspector missing settlement name")
	}
	if !strings.Contains(panel, "village") {
		t.Error("inspector missing settlement type")
	}
	if !strings.Contains(panel, "Radius: 3") {
		t.Error("inspector missing radius")
	}
	if !strings.Contains(panel, "Population: 2") {
		t.Error("inspector missing population")
	}
	if !strings.Contains(panel, "house") {
		t.Error("inspector missing buildings")
	}
}
