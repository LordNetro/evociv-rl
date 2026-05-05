# qlearning-engine Specification

## Purpose

Motor Q-Learning ε-greedy con tabla discreta Q[state]→map[action]→float64. Estado = (need_type, biome_id, time_of_day). GOAP actúa como fallback cuando Q(s,a) < threshold para todas las acciones.

## Requirements

### Requirement: Q-Table Structure

The system MUST maintain a discrete Q-table as a map keyed by state (string encoding of need_type + biome_id + time_of_day), mapping to a sub-map of action_id → q_value float64.

#### Scenario: Q-table returns zero-initialized values

- GIVEN a fresh Q-table with no entries
- WHEN querying any state-action pair
- THEN the Q-value MUST be 0.0

### Requirement: ε-Greedy Policy

The system MUST use ε-greedy action selection: with probability ε, select a random available action (explore); with probability 1-ε, select argmax Q(s,a) (exploit). The exploration rate ε MUST decay from 0.5 to 0.05 over the simulation.

#### Scenario: ε-greedy explores randomly

- GIVEN ε=0.5 and a state with multiple actions
- WHEN SelectAction is called 100 times
- THEN approximately 50% of selections MUST differ from the greedy choice

#### Scenario: ε decays over time

- GIVEN initial ε=0.5
- WHEN the simulation progresses (e.g., 1000 ticks, or N updates)
- THEN ε MUST be ≤ 0.05

### Requirement: Reward Calculation

The reward for an action MUST be computed as: `reward = hunger_reduction + fatigue_reduction + completion_bonus`, where reductions are the difference in need values before and after the action.

#### Scenario: Positive reward reinforces action

- GIVEN an NPC with Hunger=0.8 that executes harvest (hunger_change=-0.3)
- WHEN ComputeReward is called
- THEN the reward MUST be positive AND reflect the hunger reduction plus completion bonus

### Requirement: GOAP Fallback

The system MUST use the GOAP planner as fallback when Q(s,a) < threshold (e.g., 0.01) for ALL available actions in the current state.

#### Scenario: GOAP fallback activates

- GIVEN a state where all action Q-values are 0.0 (below threshold)
- WHEN SelectAction is called
- THEN the system MUST delegate to the GOAP planner for action selection
