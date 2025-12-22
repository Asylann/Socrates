package ratelimit

import (
	"sync"
	"time"
)

// Bucket implements a simple token-bucket limiter.
type Bucket struct {
	capacity   float64   // max tokens
	tokens     float64   // current tokens (float to allow fractional refill)
	refillRate float64   // tokens per second
	last       time.Time // last refill moment
	mu         sync.Mutex
}

// NewBucket creates a bucket with given capacity and refill rate (tokens/sec).
func NewBucket(capacity int, refillRatePerSec float64) *Bucket {
	return &Bucket{
		capacity:   float64(capacity),
		tokens:     float64(capacity),
		refillRate: refillRatePerSec,
		last:       time.Now(),
	}
}

// Allow consumes n tokens (usually 1). Returns true if allowed.
func (b *Bucket) Allow(n int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.last).Seconds()
	if elapsed > 0 {
		b.tokens += elapsed * b.refillRate
		if b.tokens > b.capacity {
			b.tokens = b.capacity
		}
		b.last = now
	}

	required := float64(n)
	if b.tokens >= required {
		b.tokens -= required
		return true
	}
	return false
}
