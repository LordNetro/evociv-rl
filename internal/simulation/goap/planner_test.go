package goap

import "testing"

func TestPlanHighHunger(t *testing.T) {
	actions := []Action{
		{ID: "harvest", Biomes: []string{"plains"}, NeedsMin: Needs{Hunger: 0, Fatigue: 0}, NeedsMax: Needs{Hunger: 1, Fatigue: 1}, HungerChange: -0.3, RewardBase: 1.0},
		{ID: "forage", NeedsMin: Needs{Hunger: 0, Fatigue: 0}, NeedsMax: Needs{Hunger: 1, Fatigue: 1}, HungerChange: -0.2, RewardBase: 0.5},
		{ID: "rest", NeedsMin: Needs{Hunger: 0, Fatigue: 0}, NeedsMax: Needs{Hunger: 1, Fatigue: 1}, FatigueChange: -0.4, RewardBase: 1.0},
		{ID: "explore", NeedsMin: Needs{Hunger: 0, Fatigue: 0}, NeedsMax: Needs{Hunger: 1, Fatigue: 1}, RewardBase: 0.2},
	}
	needs := Needs{Hunger: 0.8, Fatigue: 0.1}

	plan := Plan(needs, actions, "plains")
	if plan.ID != "harvest" && plan.ID != "forage" {
		t.Errorf("expected harvest or forage for high hunger, got %s", plan.ID)
	}
}

func TestPlanHighFatigue(t *testing.T) {
	actions := []Action{
		{ID: "harvest", NeedsMin: Needs{Hunger: 0, Fatigue: 0}, NeedsMax: Needs{Hunger: 1, Fatigue: 1}, HungerChange: -0.3, RewardBase: 1.0},
		{ID: "rest", NeedsMin: Needs{Hunger: 0, Fatigue: 0}, NeedsMax: Needs{Hunger: 1, Fatigue: 1}, FatigueChange: -0.4, RewardBase: 1.0},
		{ID: "explore", NeedsMin: Needs{Hunger: 0, Fatigue: 0}, NeedsMax: Needs{Hunger: 1, Fatigue: 1}, RewardBase: 0.2},
	}
	needs := Needs{Hunger: 0.1, Fatigue: 0.8}

	plan := Plan(needs, actions, "plains")
	if plan.ID != "rest" {
		t.Errorf("expected rest for high fatigue, got %s", plan.ID)
	}
}

func TestPlanLowNeeds(t *testing.T) {
	actions := []Action{
		{ID: "harvest", NeedsMin: Needs{Hunger: 0, Fatigue: 0}, NeedsMax: Needs{Hunger: 1, Fatigue: 1}, HungerChange: -0.3, RewardBase: 1.0},
		{ID: "rest", NeedsMin: Needs{Hunger: 0, Fatigue: 0}, NeedsMax: Needs{Hunger: 1, Fatigue: 1}, FatigueChange: -0.4, RewardBase: 1.0},
		{ID: "explore", NeedsMin: Needs{Hunger: 0, Fatigue: 0}, NeedsMax: Needs{Hunger: 1, Fatigue: 1}, RewardBase: 0.2},
		{ID: "socialize", NeedsMin: Needs{Hunger: 0, Fatigue: 0}, NeedsMax: Needs{Hunger: 0.5, Fatigue: 0.5}, RewardBase: 0.3},
	}
	needs := Needs{Hunger: 0.1, Fatigue: 0.1}

	plan := Plan(needs, actions, "plains")
	if plan.ID != "explore" && plan.ID != "socialize" {
		t.Errorf("expected explore or socialize for low needs, got %s", plan.ID)
	}
}

func TestPlanBiomeRestriction(t *testing.T) {
	actions := []Action{
		{ID: "harvest", Biomes: []string{"plains", "forest"}, NeedsMin: Needs{Hunger: 0, Fatigue: 0}, NeedsMax: Needs{Hunger: 1, Fatigue: 1}, HungerChange: -0.3, RewardBase: 1.0},
		{ID: "forage", NeedsMin: Needs{Hunger: 0, Fatigue: 0}, NeedsMax: Needs{Hunger: 1, Fatigue: 1}, HungerChange: -0.2, RewardBase: 0.5},
	}
	needs := Needs{Hunger: 0.8, Fatigue: 0.1}

	// On ocean, harvest should not be available
	plan := Plan(needs, actions, "ocean")
	if plan.ID != "forage" {
		t.Errorf("expected forage on ocean (harvest restricted), got %s", plan.ID)
	}
}

func TestPlanReplanOnNeedChange(t *testing.T) {
	actions := []Action{
		{ID: "harvest", NeedsMin: Needs{Hunger: 0, Fatigue: 0}, NeedsMax: Needs{Hunger: 1, Fatigue: 1}, HungerChange: -0.3, RewardBase: 1.0},
		{ID: "rest", NeedsMin: Needs{Hunger: 0, Fatigue: 0}, NeedsMax: Needs{Hunger: 1, Fatigue: 1}, FatigueChange: -0.4, RewardBase: 1.0},
	}

	// Start with fatigue higher
	needs1 := Needs{Hunger: 0.1, Fatigue: 0.8}
	plan1 := Plan(needs1, actions, "plains")
	if plan1.ID != "rest" {
		t.Fatalf("expected rest initially, got %s", plan1.ID)
	}

	// Hunger rises above fatigue
	needs2 := Needs{Hunger: 0.9, Fatigue: 0.1}
	plan2 := Plan(needs2, actions, "plains")
	if plan2.ID != "harvest" {
		t.Errorf("expected harvest after need change, got %s", plan2.ID)
	}
}

func TestPlanEmptyActions(t *testing.T) {
	needs := Needs{Hunger: 0.5, Fatigue: 0.5}
	plan := Plan(needs, nil, "plains")
	if plan.ID != "" {
		t.Errorf("expected empty plan for no actions, got %s", plan.ID)
	}
}
