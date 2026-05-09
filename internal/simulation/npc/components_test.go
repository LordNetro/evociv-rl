package npc

import (
	"math/rand"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/marco/evociv-rl/internal/ecs"
)

func TestNewPersonalityDeterminism(t *testing.T) {
	rng1 := rand.New(rand.NewSource(42))
	p1 := NewPersonality(rng1)

	rng2 := rand.New(rand.NewSource(42))
	p2 := NewPersonality(rng2)

	if p1 != p2 {
		t.Errorf("same seed produced different personalities: %+v vs %+v", p1, p2)
	}
}

func TestNewPersonalityRange(t *testing.T) {
	rng := rand.New(rand.NewSource(123))
	for i := 0; i < 100; i++ {
		p := NewPersonality(rng)
		for name, v := range map[string]float64{
			"Openness": p.Openness, "Conscientiousness": p.Conscientiousness,
			"Extraversion": p.Extraversion, "Agreeableness": p.Agreeableness, "Neuroticism": p.Neuroticism,
		} {
			if v < 0 || v > 1 {
				t.Errorf("trait %s value %f out of range [0,1] at iteration %d", name, v, i)
			}
		}
	}
}

func TestNewPersonalityDiversity(t *testing.T) {
	rng := rand.New(rand.NewSource(999))
	p1 := NewPersonality(rng)
	p2 := NewPersonality(rng)

	if p1 == p2 {
		t.Error("two personalities from same RNG should differ")
	}
}

func TestComponentStores(t *testing.T) {
	w := ecs.NewWorld()
	RegisterStores(w)

	e := w.NewEntity()
	h := Health{Current: 50, Max: 100}
	ecs.AddComponent(w, e, h)

	got, ok := ecs.GetComponent[Health](w, e)
	if !ok {
		t.Fatal("expected Health component to exist")
	}
	if got != h {
		t.Errorf("got %+v, want %+v", got, h)
	}

	_, ok = ecs.GetComponent[Health](w, ecs.Entity(999))
	if ok {
		t.Error("expected missing component to return false")
	}
}

func TestAllComponentTypes(t *testing.T) {
	w := ecs.NewWorld()
	RegisterStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, Health{Current: 10, Max: 20})
	ecs.AddComponent(w, e, Personality{Openness: 0.5})
	ecs.AddComponent(w, e, Job{Role: "farmer"})
	ecs.AddComponent(w, e, AIState{Goals: []string{"wander"}})
	ecs.AddComponent(w, e, Appearance{Symbol: '@', Color: lipgloss.Color("#FFFFFF")})
	ecs.AddComponent(w, e, LOD{Level: LODLocal})
	ecs.AddComponent(w, e, Needs{Hunger: 0.3, Fatigue: 0.7})

	if h, ok := ecs.GetComponent[Health](w, e); !ok || h.Max != 20 {
		t.Error("Health not retrieved")
	}
	if p, ok := ecs.GetComponent[Personality](w, e); !ok || p.Openness != 0.5 {
		t.Error("Personality not retrieved")
	}
	if j, ok := ecs.GetComponent[Job](w, e); !ok || j.Role != "farmer" {
		t.Error("Job not retrieved")
	}
	if a, ok := ecs.GetComponent[AIState](w, e); !ok || len(a.Goals) != 1 {
		t.Error("AIState not retrieved")
	}
	if ap, ok := ecs.GetComponent[Appearance](w, e); !ok || ap.Symbol != '@' {
		t.Error("Appearance not retrieved")
	}
	if l, ok := ecs.GetComponent[LOD](w, e); !ok || l.Level != LODLocal {
		t.Error("LOD not retrieved")
	}
	if n, ok := ecs.GetComponent[Needs](w, e); !ok || n.Hunger != 0.3 || n.Fatigue != 0.7 {
		t.Errorf("Needs not retrieved: got %+v", n)
	}
}

func TestNeedsZeroInit(t *testing.T) {
	w := ecs.NewWorld()
	RegisterStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, Needs{})

	n, ok := ecs.GetComponent[Needs](w, e)
	if !ok {
		t.Fatal("expected Needs component to exist")
	}
	if n.Hunger != 0.0 || n.Fatigue != 0.0 {
		t.Errorf("expected zero-init Needs, got Hunger=%.2f Fatigue=%.2f", n.Hunger, n.Fatigue)
	}
}

func TestNeedsClampedAtMax(t *testing.T) {
	w := ecs.NewWorld()
	RegisterStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, Needs{Hunger: 0.99, Fatigue: 0.99})

	// Needs component stores raw values; clamping is done by the system
	// But we verify the component can hold values near 1.0
	n, ok := ecs.GetComponent[Needs](w, e)
	if !ok {
		t.Fatal("expected Needs component to exist")
	}
	if n.Hunger < 0 || n.Hunger > 1.0 {
		t.Errorf("Hunger %.2f out of range [0,1]", n.Hunger)
	}
	if n.Fatigue < 0 || n.Fatigue > 1.0 {
		t.Errorf("Fatigue %.2f out of range [0,1]", n.Fatigue)
	}
}

func TestAIStateZeroDefaults(t *testing.T) {
	w := ecs.NewWorld()
	RegisterStores(w)

	e := w.NewEntity()
	ecs.AddComponent(w, e, AIState{})

	ai, ok := ecs.GetComponent[AIState](w, e)
	if !ok {
		t.Fatal("expected AIState component to exist")
	}
	if ai.LastReward != 0.0 {
		t.Errorf("expected LastReward=0.0, got %f", ai.LastReward)
	}
	if ai.RewardTick != 0 {
		t.Errorf("expected RewardTick=0, got %d", ai.RewardTick)
	}
}
