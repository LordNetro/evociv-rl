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
	"github.com/marco/evociv-rl/internal/simulation/df"
	"github.com/marco/evociv-rl/internal/simulation/economy"
	"github.com/marco/evociv-rl/internal/simulation/npc"
	"github.com/marco/evociv-rl/internal/simulation/settlement"
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
	actionDefs, err := npc.LoadActions(registry)
	if err != nil {
		log.Warn("Failed to load actions", "error", err)
	}

	// Load settlement definitions
	settlementDefs, err := settlement.LoadSettlementTypes(registry)
	if err != nil {
		log.Warn("Failed to load settlement types", "error", err)
	}
	buildingDefs, err := settlement.LoadBuildingTypes(registry)
	if err != nil {
		log.Warn("Failed to load building types", "error", err)
	}
	growthThresholds, err := settlement.LoadGrowthThresholds(registry)
	if err != nil {
		log.Warn("Failed to load growth thresholds", "error", err)
	}

	// Initialize ECS world
	ecsWorld := ecs.NewWorld()

	// Register df stores
	// load DF job defs from data and register stores
	if err := df.LoadJobsFromRegistry(registry); err != nil {
		log.Warn("Failed to load DF job defs", "error", err)
	}
	df.RegisterStores(ecsWorld)
	npc.RegisterStores(ecsWorld)
	settlement.RegisterSettlementStores(ecsWorld)

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
	var npcRenderSys *npc.NPCRenderSystem
	var setRenderSys *settlement.SettlementRenderSystem
	var buildingRenderSys *settlement.BuildingRenderSystem
	var qlSys *npc.QLearningSystem
	if worldMap != nil {
		setRenderSys = settlement.NewSettlementRenderSystem()
		buildingRenderSys = settlement.NewBuildingRenderSystem(buildingDefs)
		ecsWorld.AddSystem(settlement.NewSettlementSpawnSystem(worldMap, genConfig.Seed, settlementDefs, buildingDefs))
		ecsWorld.AddSystem(settlement.NewPopulationSystem())
		ecsWorld.AddSystem(economy.NewSettlementEconomySystem(buildingDefs))
		ecsWorld.AddSystem(economy.NewSettlementGrowthSystem(growthThresholds))
		ecsWorld.AddSystem(economy.NewFamineSystem())
		ecsWorld.AddSystem(setRenderSys)
		ecsWorld.AddSystem(buildingRenderSys)
	}
	if worldMap != nil && len(raceDefs) > 0 && len(roleDefs) > 0 {
		npcRenderSys = npc.NewNPCRenderSystem()
		ecsWorld.AddSystem(npc.NewNPCSpawnSystem(worldMap, npc.SpawnConfig{}, genConfig.Seed+999, raceDefs, roleDefs))
		ecsWorld.AddSystem(npc.NewWanderSystem(worldMap, roleDefs, rand.New(rand.NewSource(time.Now().UnixNano()))))
		playerX, playerY := genConfig.Width/2, genConfig.Height/2
		ecsWorld.AddSystem(npc.NewLODSystem(func() (int, int) { return playerX, playerY }))
		ecsWorld.AddSystem(npc.NewNeedsDecaySystem())
		if len(actionDefs) > 0 {
			ecsWorld.AddSystem(npc.NewGOAPSystem(worldMap, actionDefs))
			qlSys = npc.NewQLearningSystem(worldMap, actionDefs, rand.New(rand.NewSource(time.Now().UnixNano())))
			ecsWorld.AddSystem(qlSys)
		}
		// DF job systems: assignment, integration and completion
		ecsWorld.AddSystem(df.NewJobSystem())
		ecsWorld.AddSystem(df.NewDFAssignmentIntegrationSystem())
		ecsWorld.AddSystem(df.NewJobCompletionSystem())
		ecsWorld.AddSystem(npcRenderSys)
	}

	if err := ecsWorld.Update(0); err != nil {
		log.Warn("ECS update failed", "error", err)
	}

	// Re-run LOD with camera at center for better initial visibility
	if worldMap != nil && len(raceDefs) > 0 && len(roleDefs) > 0 {
		cx, cy := genConfig.Width/2, genConfig.Height/2
		lodSys := npc.NewLODSystem(func() (int, int) { return cx, cy })
		if err := lodSys.Update(ecsWorld, 0); err != nil {
			log.Warn("LOD update failed", "error", err)
		}
		if npcRenderSys != nil {
			if err := npcRenderSys.Update(ecsWorld, 0); err != nil {
				log.Warn("NPC render update failed", "error", err)
			}
		}
	}
	if setRenderSys != nil {
		if err := setRenderSys.Update(ecsWorld, 0); err != nil {
			log.Warn("Settlement render update failed", "error", err)
		}
	}
	if buildingRenderSys != nil {
		if err := buildingRenderSys.Update(ecsWorld, 0); err != nil {
			log.Warn("Building render update failed", "error", err)
		}
	}

	// Start TUI
	model := ui.NewModel()
	if worldMap != nil {
		model.SetWorldMap(worldMap)
	}
	model.SetECSWorld(ecsWorld)

	// Gather render infos for overlay
	if npcRenderSys != nil {
		model.SetNPCOverlay(npcRenderSys.RenderInfos())
	}
	if setRenderSys != nil {
		model.SetSettlementOverlay(setRenderSys.RenderInfos())
	}
	if buildingRenderSys != nil {
		model.SetSettlementBuildings(buildingRenderSys.RenderInfos())
	}

	// Load Q-table from store if available
	if qlSys != nil && s != nil {
		data, err := s.LoadQTable(0)
		if err == nil && len(data) > 0 {
			qlSys.QTable().LoadValues(data)
			log.Info("Q-table loaded", "states", len(data))
		}
	}

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}

	// Save Q-table after TUI closes
	if qlSys != nil && s != nil {
		if err := s.SaveQTable(0, qlSys.QTable().Values()); err != nil {
			log.Warn("Failed to save Q-table", "error", err)
		} else {
			log.Info("Q-table saved")
		}
	}

	return nil
}
