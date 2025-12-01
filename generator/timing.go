package generator

import (
	"math/rand"
	"time"
)

// RateLimiter controls the rate of CDR generation with optional jitter
type RateLimiter struct {
	callsPerMinute float64
	jitterPercent  float64
	random         *rand.Rand
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(callsPerMinute float64, jitterPercent float64) *RateLimiter {
	return &RateLimiter{
		callsPerMinute: callsPerMinute,
		jitterPercent:  jitterPercent,
		random:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NextInterval returns the duration to wait before the next CDR
func (r *RateLimiter) NextInterval() time.Duration {
	if r.callsPerMinute <= 0 {
		return time.Minute // Default to 1 per minute if not set
	}

	// Base interval in nanoseconds
	baseInterval := time.Minute / time.Duration(r.callsPerMinute)

	// Apply jitter if configured
	if r.jitterPercent > 0 {
		// Generate random value between -jitter% and +jitter%
		jitterFactor := (r.random.Float64()*2 - 1) * (r.jitterPercent / 100)
		jitterAmount := time.Duration(float64(baseInterval) * jitterFactor)
		return baseInterval + jitterAmount
	}

	return baseInterval
}

// SetCallsPerMinute updates the rate
func (r *RateLimiter) SetCallsPerMinute(cpm float64) {
	r.callsPerMinute = cpm
}

// SetJitterPercent updates the jitter percentage
func (r *RateLimiter) SetJitterPercent(jp float64) {
	r.jitterPercent = jp
}

// Ticker creates a channel that sends at the configured rate with jitter
type Ticker struct {
	limiter *RateLimiter
	C       chan time.Time
	done    chan struct{}
}

// NewTicker creates a new ticker that fires at the rate limiter's interval
func NewTicker(limiter *RateLimiter) *Ticker {
	t := &Ticker{
		limiter: limiter,
		C:       make(chan time.Time, 1),
		done:    make(chan struct{}),
	}
	go t.run()
	return t
}

func (t *Ticker) run() {
	for {
		interval := t.limiter.NextInterval()
		select {
		case <-time.After(interval):
			select {
			case t.C <- time.Now():
			default:
				// Channel full, skip this tick
			}
		case <-t.done:
			return
		}
	}
}

// Stop stops the ticker
func (t *Ticker) Stop() {
	close(t.done)
}
