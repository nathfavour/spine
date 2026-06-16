package core

import (
	"math"
	"math/rand"
	"time"
)

// PoissonInterval calculates the next interval based on a Poisson distribution.
// lambda is the average number of events per unit time.
func PoissonInterval(lambda float64) time.Duration {
	if lambda <= 0 {
		return 0
	}
	// The time between events in a Poisson process follows an exponential distribution.
	// f(t) = lambda * e^(-lambda * t)
	// We can sample from this using: t = -ln(U) / lambda, where U is a uniform random number in (0, 1].
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

func (p *Pacer) Wait() {
	time.Sleep(p.Next())
}
