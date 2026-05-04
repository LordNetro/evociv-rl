# tui-welcome Specification

## Purpose

Pantalla de bienvenida del simulador usando Bubbletea con lipgloss. Proporciona la base TUI con una tecla para salir.

## Requirements

### Requirement: Welcome Screen Display

The system MUST render a styled welcome screen when the TUI starts.

#### Scenario: Golden test matches expected output

- GIVEN the TUI model is initialized
- WHEN View() is called
- THEN the output MUST match the golden file in `testdata/`

### Requirement: Quit on 'q'

The system MUST quit when the user presses the 'q' key.

#### Scenario: 'q' key produces tea.Quit

- GIVEN the TUI model is running
- WHEN a tea.KeyMsg for 'q' is sent via Update()
- THEN the returned command MUST be tea.Quit

#### Scenario: Other keys are ignored

- GIVEN the TUI model is running
- WHEN any key other than 'q' is sent
- THEN the model MUST remain active (no tea.Quit)

### Requirement: Styled Output

The welcome screen SHOULD use lipgloss for centered, styled text (title, subtitle, instructions).

#### Scenario: Styling uses lipgloss

- GIVEN the View() output
- THEN it MUST contain ANSI style sequences produced by lipgloss

### Requirement: Integration Test

The system MUST have a teatest-based integration test that starts the model and sends 'q'.

#### Scenario: teatest runs model and quits

- GIVEN a teatest.NewTestModel
- WHEN the model runs and receives 'q'
- THEN the model terminates cleanly without error
