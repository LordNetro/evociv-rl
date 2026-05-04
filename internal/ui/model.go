package ui

import tea "github.com/charmbracelet/bubbletea"

// Model is the Bubbletea model for the Evociv-RL TUI.
type Model struct {
	ready    bool
	width    int
	height   int
	quitting bool
}

// NewModel creates a new TUI model.
func NewModel() Model {
	return Model{}
}

// Init initializes the model and returns the initial command.
func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen)
}

// Update handles incoming messages and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
	}
	return m, nil
}

// View renders the model (implemented in view.go).
func (m Model) View() string {
	return renderView(m)
}
