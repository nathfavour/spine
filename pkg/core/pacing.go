package core

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/nathfavour/spine/pkg/types"
)

// PoissonInterval calculates the next interval based on a Poisson distribution.
func PoissonInterval(lambda float64) time.Duration {
	if lambda <= 0 {
		return 0
	}
	u := rand.Float64()
	for u == 0 {
		u = rand.Float64()
	}
	seconds := -math.Log(u) / lambda
	return time.Duration(seconds * float64(time.Second))
}

type Pacer struct {
	lambda float64
}

func NewPacer(avgInterval time.Duration) *Pacer {
	return &Pacer{
		lambda: 1.0 / avgInterval.Seconds(),
	}
}

func (p *Pacer) Next() time.Duration {
	return PoissonInterval(p.lambda)
}

// MimicManager handles multi-agent pacing profiles
type MimicManager struct {
	mu     sync.RWMutex
	agents map[string]types.Purpose
	pacers map[types.Purpose]*Pacer
}

func NewMimicManager() *MimicManager {
	m := &MimicManager{
		agents: make(map[string]types.Purpose),
		pacers: make(map[types.Purpose]*Pacer),
	}

	// Initialize default purpose pacers
	m.pacers[types.PurposeSocial] = NewPacer(30 * time.Second)   // Avg 30s for social
	m.pacers[types.PurposeResearch] = NewPacer(5 * time.Minute)  // Avg 5m for research
	m.pacers[types.PurposeBurst] = NewPacer(5 * time.Second)    // Avg 5s for bursts
	m.pacers[types.PurposeIdle] = NewPacer(1 * time.Hour)       // Avg 1h for idle

	return m
}

func (m *MimicManager) Register(agentID string, purpose types.Purpose) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.agents[agentID] = purpose
}

func (m *MimicManager) GetNextInterval(agentID string, weight float64) time.Duration {
	m.mu.RLock()
	purpose, ok := m.agents[agentID]
	m.mu.RUnlock()

	if !ok {
		purpose = types.PurposeIdle
	}

	m.mu.RLock()
	pacer := m.pacers[purpose]
	m.mu.RUnlock()

	interval := pacer.Next()
	if weight > 0 {
		// Adjust lambda by weight: interval = -ln(U) / (lambda * weight)
		// which is equivalent to dividing the base interval by weight
		interval = time.Duration(float64(interval) / weight)
	}

	return interval
}
