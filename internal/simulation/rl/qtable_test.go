package rl

import (
	"math/rand"
	"testing"
)

func TestQTableZeroInit(t *testing.T) {
	qt := NewQTable()
	val := qt.Get("state1", "action1")
	if val != 0.0 {
		t.Errorf("expected zero-init Q, got %f", val)
	}
}

func TestEGreedyExplore(t *testing.T) {
	qt := NewQTable()
	rng := rand.New(rand.NewSource(42))

	actions := []string{"a", "b", "c"}
	qt.Set("s", "a", 10.0)
	qt.Set("s", "b", 1.0)
	qt.Set("s", "c", 1.0)

	exploreCount := 0
	const trials = 1000
	for i := 0; i < trials; i++ {
		choice := qt.EGreedy("s", actions, 0.5, rng)
		if choice != "a" {
			exploreCount++
		}
	}

	// With epsilon=0.5, approximately 50% should explore (not pick greedy "a")
	ratio := float64(exploreCount) / trials
	if ratio < 0.35 || ratio > 0.65 {
		t.Errorf("expected ~50%% exploration at epsilon=0.5, got %.2f%%", ratio*100)
	}
}

func TestEGreedyExploit(t *testing.T) {
	qt := NewQTable()
	rng := rand.New(rand.NewSource(42))

	actions := []string{"a", "b", "c"}
	qt.Set("s", "a", 10.0)
	qt.Set("s", "b", 1.0)
	qt.Set("s", "c", 1.0)

	// With epsilon=0, always exploit
	choice := qt.EGreedy("s", actions, 0.0, rng)
	if choice != "a" {
		t.Errorf("expected greedy choice 'a', got %s", choice)
	}
}

func TestEpsilonDecay(t *testing.T) {
	qt := NewQTable()
	qt.SetEpsilon(0.5)
	qt.SetEpsilonDecay(0.997) // decays to ~0.025 over 1000 iterations

	for i := 0; i < 1000; i++ {
		qt.DecayEpsilon()
	}

	if qt.Epsilon() > 0.05 {
		t.Errorf("expected epsilon <= 0.05 after 1000 decays, got %f", qt.Epsilon())
	}
}

func TestUpdateReinforces(t *testing.T) {
	qt := NewQTable()
	qt.Set("s1", "a1", 0.0)

	// Q(s,a) += alpha * (reward + gamma * maxQ(s') - Q(s,a))
	qt.Update("s1", "a1", 1.0, "s2", 0.1, 0.9)

	val := qt.Get("s1", "a1")
	if val <= 0.0 {
		t.Errorf("expected positive Q after reward, got %f", val)
	}
}

func TestShouldFallbackWhenAllQZero(t *testing.T) {
	qt := NewQTable()
	actions := []string{"a", "b"}
	if !qt.ShouldFallback("s", actions, 0.1) {
		t.Error("expected fallback when all Q are zero")
	}
}

func TestShouldNotFallbackWhenQHigh(t *testing.T) {
	qt := NewQTable()
	qt.Set("s", "a", 0.5)
	actions := []string{"a", "b"}
	if qt.ShouldFallback("s", actions, 0.1) {
		t.Error("expected no fallback when at least one Q is high")
	}
}
