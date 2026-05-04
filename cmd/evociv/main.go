package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"

	"github.com/marco/evociv-rl/internal/data"
	"github.com/marco/evociv-rl/internal/ecs"
	"github.com/marco/evociv-rl/internal/store"
	"github.com/marco/evociv-rl/internal/ui"
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
	loader := data.NewLoader(os.DirFS("."))
	registry := data.NewRegistry()
	if err := loader.LoadAll("data", registry); err != nil {
		return fmt.Errorf("load data: %w", err)
	}
	log.Info("Data loaded", "types", registry.Types())

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
	}

	// Start TUI
	p := tea.NewProgram(ui.NewModel())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}

	return nil
}
