package ratelimit

import (
	"sync"
	"time"
)

// Manager keeps per-user buckets and performs optional cleanup.
type Manager struct {
	buckets  sync.Map      // map[int64]*Bucket
	lastSeen sync.Map      // map[int64]time.Time
	ttl      time.Duration // cleanup TTL for idle users
}

// NewManager creates a Manager. If ttl>0 it launches a cleanup goroutine.
func NewManager(ttl time.Duration) *Manager {
	m := &Manager{ttl: ttl}
	if ttl > 0 {
		go m.cleanupLoop()
	}
	return m
}

// GetBucket returns existing or newly created bucket for userID.
func (m *Manager) GetBucket(userID int64, capacity int, refillPerSec float64) *Bucket {
	raw, ok := m.buckets.Load(userID)
	if ok {
		m.lastSeen.Store(userID, time.Now())
		return raw.(*Bucket)
	}
	b := NewBucket(capacity, refillPerSec)
	m.buckets.Store(userID, b)
	m.lastSeen.Store(userID, time.Now())
	return b
}

func (m *Manager) cleanupLoop() {
	t := time.NewTicker(m.ttl / 2)
	defer t.Stop()
	for range t.C {
		now := time.Now()
		m.lastSeen.Range(func(k, v any) bool {
			uid := k.(int64)
			seen := v.(time.Time)
			if now.Sub(seen) > m.ttl {
				m.buckets.Delete(uid)
				m.lastSeen.Delete(uid)
			}
			return true
		})
	}
}
