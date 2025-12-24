package infrastructure

import (
	"sync"
	"time"
)

// MessageRateLimiter implements token bucket rate limiting per user
type MessageRateLimiter struct {
	mu          sync.RWMutex
	buckets     map[int]*tokenBucket
	rate        float64 // tokens per second
	maxTokens   float64 // burst capacity
	cleanupTick time.Duration
}

type tokenBucket struct {
	tokens     float64
	lastUpdate time.Time
}

// NewMessageRateLimiter creates a rate limiter with specified rate and burst
// rate: messages per second allowed
// burst: maximum burst capacity
func NewMessageRateLimiter(rate float64, burst int) *MessageRateLimiter {
	rl := &MessageRateLimiter{
		buckets:     make(map[int]*tokenBucket),
		rate:        rate,
		maxTokens:   float64(burst),
		cleanupTick: 5 * time.Minute,
	}
	
	// Start cleanup goroutine
	go rl.cleanup()
	
	return rl
}

// Allow checks if user can send a message (consumes 1 token if allowed)
func (rl *MessageRateLimiter) Allow(userID int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	bucket, exists := rl.buckets[userID]
	now := time.Now()
	
	if !exists {
		// Create new bucket with full tokens
		rl.buckets[userID] = &tokenBucket{
			tokens:     rl.maxTokens - 1, // Consume 1 token
			lastUpdate: now,
		}
		return true
	}
	
	// Refill tokens based on time elapsed
	elapsed := now.Sub(bucket.lastUpdate).Seconds()
	bucket.tokens += elapsed * rl.rate
	if bucket.tokens > rl.maxTokens {
		bucket.tokens = rl.maxTokens
	}
	bucket.lastUpdate = now
	
	// Check if we have a token
	if bucket.tokens >= 1 {
		bucket.tokens -= 1
		return true
	}
	
	return false
}

// WaitTime returns how long to wait before next message is allowed
func (rl *MessageRateLimiter) WaitTime(userID int) time.Duration {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	
	bucket, exists := rl.buckets[userID]
	if !exists {
		return 0
	}
	
	now := time.Now()
	elapsed := now.Sub(bucket.lastUpdate).Seconds()
	currentTokens := bucket.tokens + elapsed*rl.rate
	
	if currentTokens >= 1 {
		return 0
	}
	
	// Calculate wait time for 1 token
	needed := 1 - currentTokens
	waitSeconds := needed / rl.rate
	return time.Duration(waitSeconds * float64(time.Second))
}

// Reset removes rate limit state for a user (useful after quota reset)
func (rl *MessageRateLimiter) Reset(userID int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.buckets, userID)
}

// cleanup removes stale buckets periodically
func (rl *MessageRateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupTick)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for userID, bucket := range rl.buckets {
			// Remove buckets not used in last 10 minutes
			if now.Sub(bucket.lastUpdate) > 10*time.Minute {
				delete(rl.buckets, userID)
			}
		}
		rl.mu.Unlock()
	}
}

// GetStats returns rate limiter statistics
func (rl *MessageRateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	
	return map[string]interface{}{
		"active_users": len(rl.buckets),
		"rate":         rl.rate,
		"burst":        rl.maxTokens,
	}
}
