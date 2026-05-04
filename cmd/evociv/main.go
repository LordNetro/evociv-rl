package main

import (
	"fmt"
	"os"
	"path"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"

	"github.com/marco/evociv-rl/internal/data"
	"github.com/marco/evociv-rl/internal/ecs"
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

	// Initialize ECS world
	_ = ecs.NewWorld()

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
			if err := s.SaveWorld(genConfig.Seed, genConfig.Width, genConfig.Height); err != nil {
				log.Warn("Failed to save world", "error", err)
			}
		}
	}

	// Start TUI
	model := ui.NewModel()
	if worldMap != nil {
		model.SetWorldMap(worldMap)
	}
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}

	return nil
}
