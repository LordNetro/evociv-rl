package main

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"

	"github.com/marco/evociv-rl/internal/data"
	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/simulation/npc"
	"github.com/marco/evociv-rl/internal/store"
	"github.com/marco/evociv-rl/internal/ui"
	"github.com/marco/evociv-rl/internal/world/gen"
)

func main() {
	if err := run(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func run() error {
	log.SetLevel(log.InfoLevel)
	log.Info("Evociv-RL starting", "version", "v0.0.1")

	// Load data
	fsys := os.DirFS(".")
	loader := data.NewLoader(fsys)
	registry := data.NewRegistry()
	if err := loader.LoadAll("data", registry); err != nil {
		return fmt.Errorf("load data: %w", err)
	}
	log.Info("Data loaded", "types", registry.Types())

	// Load generation config
	genConfig, err := gen.LoadGenConfig(path.Join("data", "gen-config.yaml"), fsys)
	if err != nil {
		log.Warn("Failed to load gen-config, using defaults", "error", err)
		genConfig = gen.GenConfig{
			Seed:       42,
			Width:      64,
			Height:     64,
			Octaves:    6,
			Lacunarity: 2.0,
			Gain:       0.5,
			Scale:      100.0,
		}
	}
	if err := genConfig.Validate(); err != nil {
		log.Warn("Invalid gen-config, using defaults", "error", err)
		genConfig.Width = 64
		genConfig.Height = 64
	}

	// Load biomes
	biomes, err := gen.LoadBiomes(registry)
	if err != nil {
		log.Warn("Failed to load biomes", "error", err)
		biomes = []gen.BiomeDef{
			{ID: "unknown", MinHeight: -1.0, MaxHeight: 1.0, MinHumidity: 0.0, MaxHumidity: 1.0, MinTemperature: 0.0, MaxTemperature: 1.0, Symbol: "?", Color: "#888888"},
		}
	}

	// Generate world
	worldMap, err := gen.Generate(genConfig.Width, genConfig.Height, genConfig, biomes)
	if err != nil {
		log.Warn("World generation failed, continuing without map", "error", err)
		worldMap = nil
	} else {
		log.Info("World generated", "width", genConfig.Width, "height", genConfig.Height, "seed", genConfig.Seed)
	}

	// Load NPC definitions
	raceDefs, err := npc.LoadNpcRaces(registry)
	if err != nil {
		log.Warn("Failed to load NPC races", "error", err)
	}
	roleDefs, err := npc.LoadNpcRoles(registry)
	if err != nil {
		log.Warn("Failed to load NPC roles", "error", err)
	}

	// Initialize ECS world
	ecsWorld := ecs.NewWorld()
	npc.RegisterStores(ecsWorld)

	// Initialize store
	s := store.NewSQLiteStore()
	if err := s.Open("evociv.db"); err != nil {
		log.Warn("Store not available", "error", err)
	} else {
		defer s.Close()
		if err := s.Health(); err != nil {
			log.Warn("Store health check failed", "error", err)
		}
		if worldMap != nil {
			if err := s.SaveWorld(genConfig.Seed, genConfig.Width, genConfig.Height, 999); err != nil {
				log.Warn("Failed to save world", "error", err)
			}
		}
	}

	// Spawn NPCs and run systems once before TUI starts
	var renderSys *npc.NPCRenderSystem
	if worldMap != nil && len(raceDefs) > 0 && len(roleDefs) > 0 {
		renderSys = npc.NewNPCRenderSystem()
		ecsWorld.AddSystem(npc.NewNPCSpawnSystem(worldMap, npc.SpawnConfig{}, genConfig.Seed+999, raceDefs, roleDefs))
		ecsWorld.AddSystem(npc.NewWanderSystem(worldMap, roleDefs, rand.New(rand.NewSource(time.Now().UnixNano()))))
		ecsWorld.AddSystem(npc.NewLODSystem(func() (int, int) { return 0, 0 }))
		ecsWorld.AddSystem(renderSys)

		if err := ecsWorld.Update(0); err != nil {
			log.Warn("ECS update failed", "error", err)
		}

		// Re-run LOD with camera at center for better initial visibility
		cx, cy := genConfig.Width/2, genConfig.Height/2
		lodSys := npc.NewLODSystem(func() (int, int) { return cx, cy })
		if err := lodSys.Update(ecsWorld, 0); err != nil {
			log.Warn("LOD update failed", "error", err)
		}
		if err := renderSys.Update(ecsWorld, 0); err != nil {
			log.Warn("Render update failed", "error", err)
		}
	}

	// Start TUI
	model := ui.NewModel()
	if worldMap != nil {
		model.SetWorldMap(worldMap)
	}
	model.SetECSWorld(ecsWorld)

	// Gather render infos for overlay
	if renderSys != nil {
		model.SetNPCOverlay(renderSys.RenderInfos())
	}

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}

	return nil
}
