package goap

import "math"

// Needs represents basic NPC needs.
type Needs struct {
	Hunger  float64
	Fatigue float64
}

// Action represents a GOAP action for planning.
type Action struct {
	ID            string
	Biomes        []string
	NeedsMin      Needs
	NeedsMax      Needs
	HungerChange  float64
	FatigueChange float64
	RewardBase    float64
}

// Plan selects the best action for the given needs, available actions, and current biome.
// It uses forward-chaining greedy evaluation: the most urgent need determines the action category,
// and the highest-reward compatible action is chosen.
func Plan(needs Needs, actions []Action, biome string) Action {
	if len(actions) == 0 {
		return Action{}
	}

	var best Action
	bestScore := -math.MaxFloat64

	// Determine the most urgent need and its value
	urgentNeed := "hunger"
	urgentVal := needs.Hunger
	if needs.Fatigue > needs.Hunger {
		urgentNeed = "fatigue"
		urgentVal = needs.Fatigue
	}

	lowNeeds := urgentVal <= 0.3

	for _, action := range actions {
		if !actionCompatible(action, biome) {
			continue
		}
		if !needsInRange(needs, action.NeedsMin, action.NeedsMax) {
			continue
		}

		score := action.RewardBase

		if lowNeeds {
			// When needs are low, penalize actions that reduce needs
			// (survival actions are unnecessary; prefer leisure)
			if action.HungerChange < 0 || action.FatigueChange < 0 {
				score -= 100.0 // strongly deprioritize need-reducing actions
			}
		} else {
			// Prioritize actions that address the most urgent need
			switch urgentNeed {
			case "hunger":
				if action.HungerChange < 0 {
					score += 10.0 + math.Abs(action.HungerChange)
				}
			case "fatigue":
				if action.FatigueChange < 0 {
					score += 10.0 + math.Abs(action.FatigueChange)
				}
			}
		}

		if score > bestScore {
			bestScore = score
			best = action
		}
	}

	return best
}

func actionCompatible(action Action, biome string) bool {
	if len(action.Biomes) == 0 {
		return true
	}
	for _, b := range action.Biomes {
		if b == biome {
			return true
		}
	}
	return false
}

func needsInRange(needs, min, max Needs) bool {
	if needs.Hunger < min.Hunger || needs.Hunger > max.Hunger {
		return false
	}
	if needs.Fatigue < min.Fatigue || needs.Fatigue > max.Fatigue {
		return false
	}
	return true
}
