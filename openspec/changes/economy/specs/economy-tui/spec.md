# economy-tui Specification

## Purpose

Extensión de la interfaz TUI para mostrar información económica: status bar con recursos del settlement bajo el cursor, e inspector de settlement expandido con datos de recursos y progreso de nivel.

## Requirements

### Requirement: Status Bar with Resources

Cuando el cursor está sobre un settlement, la status bar MUST mostrar: símbolo, nombre, población, y recursos food/gold/tools del ResourceStore del settlement. Formato: `{símbolo} {nombre} | Pop:{n} | Food:{f} Gold:{g} Tools:{t}`.

#### Scenario: Status bar shows resources on cursor hover

- GIVEN un settlement con símbolo "♦", nombre "Aldea", Population=5, ResourceStore con food=45.0, gold=12.0, tools=3.0
- WHEN el cursor se posiciona sobre el settlement en el mapa
- THEN la status bar MUST mostrar: "♦ Aldea | Pop:5 | Food:45.0 Gold:12.0 Tools:3.0"

#### Scenario: Status bar shows only name and pop when no ResourceStore

- GIVEN un settlement sin ResourceStore componente (aún no inicializado)
- WHEN el cursor se posiciona sobre él
- THEN la status bar MUST mostrar solo símbolo, nombre y población, omitiendo recursos

#### Scenario: No status bar for empty tiles

- GIVEN el cursor sobre un tile sin settlement ni NPC
- WHEN se renderiza el mapa
- THEN no MUST aparecer status bar extra

### Requirement: Inspector with Economic Data

El inspector de settlement (tecla 'e') MUST mostrar recursos adicionales: food, gold, tools actuales, y progreso hacia el siguiente nivel.

#### Scenario: Inspector shows resources and level progress

- GIVEN un settlement con Level=1, ResourceStore food=45.0, gold=12.0, tools=3.0, y threshold para Level 2 requiere food=100, gold=5, tools=10
- WHEN el jugador presiona 'e' sobre el settlement
- THEN el inspector MUST mostrar:
  - Resources: Food 45.0/100, Gold 12.0/5, Tools 3.0/10
  - Next Level: 2

#### Scenario: Inspector at max level shows "MAX"

- GIVEN un settlement con Level=3 (nivel máximo MVP)
- WHEN el jugador abre el inspector
- THEN el inspector MUST indicar "Level: 3 (MAX)" sin progreso

#### Scenario: Inspector shows famine warning

- GIVEN un settlement con ResourceStore["food"] < 0
- WHEN el jugador abre el inspector
- THEN el inspector MUST mostrar una advertencia de hambruna: "⚠ Famine"
