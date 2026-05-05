package rl

import (
	"math"
	"math/rand"
	"sync"
)

// QTable stores Q-values as state → action → value.
type QTable struct {
	mu      sync.RWMutex
	values  map[string]map[string]float64
	epsilon float64
	decay   float64
}

// NewQTable creates a new empty Q-table with default epsilon=0.5 and decay=0.9995.
func NewQTable() *QTable {
	return &QTable{
		values:  make(map[string]map[string]float64),
		epsilon: 0.5,
		decay:   0.9995,
	}
}

// Get returns the Q-value for a state-action pair (zero if uninitialized).
func (qt *QTable) Get(state, action string) float64 {
	qt.mu.RLock()
	defer qt.mu.RUnlock()
	if actions, ok := qt.values[state]; ok {
		if v, ok := actions[action]; ok {
			return v
		}
	}
	return 0.0
}

// Set sets a Q-value directly (used for initialization and testing).
func (qt *QTable) Set(state, action string, value float64) {
	qt.mu.Lock()
	defer qt.mu.Unlock()
	if qt.values[state] == nil {
		qt.values[state] = make(map[string]float64)
	}
	qt.values[state][action] = value
}

// Epsilon returns the current exploration rate.
func (qt *QTable) Epsilon() float64 {
	qt.mu.RLock()
	defer qt.mu.RUnlock()
	return qt.epsilon
}

// SetEpsilon sets the exploration rate.
func (qt *QTable) SetEpsilon(e float64) {
	qt.mu.Lock()
	defer qt.mu.Unlock()
	qt.epsilon = math.Max(0.0, math.Min(1.0, e))
}

// SetEpsilonDecay sets the per-step decay factor.
func (qt *QTable) SetEpsilonDecay(d float64) {
	qt.mu.Lock()
	defer qt.mu.Unlock()
	qt.decay = d
}

// DecayEpsilon applies one decay step.
func (qt *QTable) DecayEpsilon() {
	qt.mu.Lock()
	defer qt.mu.Unlock()
	qt.epsilon *= qt.decay
	if qt.epsilon < 0.05 {
		qt.epsilon = 0.05
	}
}

// EGreedy selects an action using epsilon-greedy policy.
func (qt *QTable) EGreedy(state string, actions []string, epsilon float64, rng *rand.Rand) string {
	if rng.Float64() < epsilon {
		// Explore: random action
		return actions[rng.Intn(len(actions))]
	}

	// Exploit: argmax Q(s,a)
	bestAction := actions[0]
	bestQ := qt.Get(state, bestAction)
	for _, a := range actions[1:] {
		q := qt.Get(state, a)
		if q > bestQ {
			bestQ = q
			bestAction = a
		}
	}
	return bestAction
}

// Update performs Q-learning update: Q(s,a) += alpha * (reward + gamma * maxQ(s') - Q(s,a)).
func (qt *QTable) Update(state, action string, reward float64, nextState string, alpha, gamma float64) {
	qt.mu.Lock()
	defer qt.mu.Unlock()

	if qt.values[state] == nil {
		qt.values[state] = make(map[string]float64)
	}

	currentQ := qt.values[state][action]

	// Compute maxQ(s')
	maxNextQ := 0.0
	if nextActions, ok := qt.values[nextState]; ok {
		for _, v := range nextActions {
			if v > maxNextQ {
				maxNextQ = v
			}
		}
	}

	qt.values[state][action] = currentQ + alpha*(reward+gamma*maxNextQ-currentQ)
}

// Values returns the raw Q-table map for persistence.
func (qt *QTable) Values() map[string]map[string]float64 {
	qt.mu.RLock()
	defer qt.mu.RUnlock()
	result := make(map[string]map[string]float64)
	for state, actions := range qt.values {
		actionMap := make(map[string]float64)
		for action, value := range actions {
			actionMap[action] = value
		}
		result[state] = actionMap
	}
	return result
}

// LoadValues replaces the Q-table contents with the given data.
func (qt *QTable) LoadValues(data map[string]map[string]float64) {
	qt.mu.Lock()
	defer qt.mu.Unlock()
	qt.values = make(map[string]map[string]float64)
	for state, actions := range data {
		actionMap := make(map[string]float64)
		for action, value := range actions {
			actionMap[action] = value
		}
		qt.values[state] = actionMap
	}
}

// ShouldFallback returns true if all Q-values for the given actions are below the threshold.
func (qt *QTable) ShouldFallback(state string, actions []string, threshold float64) bool {
	qt.mu.RLock()
	defer qt.mu.RUnlock()
	stateMap, ok := qt.values[state]
	if !ok {
		return true
	}
	for _, a := range actions {
		if v, ok := stateMap[a]; ok && v >= threshold {
			return false
		}
	}
	return true
}
