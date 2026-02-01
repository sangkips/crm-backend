package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// TenantRateLimiter provides per-tenant rate limiting to prevent noisy neighbor issues
type TenantRateLimiter struct {
	limiters    map[uuid.UUID]*rateLimiterEntry
	mu          sync.RWMutex
	rate        rate.Limit // requests per second
	burst       int        // maximum burst size
	cleanupTick time.Duration
	entryTTL    time.Duration
}

type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiterConfig holds configuration for the rate limiter
type RateLimiterConfig struct {
	RequestsPerSecond float64       // Rate of requests allowed per second
	BurstSize         int           // Maximum burst size
	CleanupInterval   time.Duration // How often to clean up stale entries
	EntryTTL          time.Duration // How long to keep unused entries
}

// DefaultRateLimiterConfig returns sensible defaults
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerSecond: 10,               // 10 requests per second
		BurstSize:         20,               // Allow bursts of 20
		CleanupInterval:   5 * time.Minute,  // Clean up every 5 minutes
		EntryTTL:          10 * time.Minute, // Remove entries unused for 10 mins
	}
}

// NewTenantRateLimiter creates a new per-tenant rate limiter
func NewTenantRateLimiter(cfg RateLimiterConfig) *TenantRateLimiter {
	rl := &TenantRateLimiter{
		limiters:    make(map[uuid.UUID]*rateLimiterEntry),
		rate:        rate.Limit(cfg.RequestsPerSecond),
		burst:       cfg.BurstSize,
		cleanupTick: cfg.CleanupInterval,
		entryTTL:    cfg.EntryTTL,
	}

	// Start background cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// getLimiter returns the rate limiter for a specific tenant
func (rl *TenantRateLimiter) getLimiter(tenantID uuid.UUID) *rate.Limiter {
	rl.mu.RLock()
	entry, exists := rl.limiters[tenantID]
	rl.mu.RUnlock()

	if exists {
		// Update last seen time
		rl.mu.Lock()
		entry.lastSeen = time.Now()
		rl.mu.Unlock()
		return entry.limiter
	}

	// Create new limiter for this tenant
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double check after acquiring write lock
	if entry, exists := rl.limiters[tenantID]; exists {
		entry.lastSeen = time.Now()
		return entry.limiter
	}

	limiter := rate.NewLimiter(rl.rate, rl.burst)
	rl.limiters[tenantID] = &rateLimiterEntry{
		limiter:  limiter,
		lastSeen: time.Now(),
	}

	return limiter
}

// cleanupLoop periodically removes stale rate limiter entries
func (rl *TenantRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupTick)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes entries that haven't been used recently
func (rl *TenantRateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-rl.entryTTL)
	for tenantID, entry := range rl.limiters {
		if entry.lastSeen.Before(cutoff) {
			delete(rl.limiters, tenantID)
		}
	}
}

// Middleware returns a Gin middleware that applies per-tenant rate limiting
func (rl *TenantRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get tenant ID from context (set by TenantMiddleware or AuthMiddleware)
		tenantID := GetTenantID(c)

		// If no tenant ID, use a default limiter for unauthenticated requests
		// or skip rate limiting for public endpoints
		if tenantID == uuid.Nil {
			// For unauthenticated requests, use IP-based limiting
			// You could implement IP-based limiting here as a fallback
			c.Next()
			return
		}

		limiter := rl.getLimiter(tenantID)

		if !limiter.Allow() {
			c.Header("X-RateLimit-Limit", formatInt(rl.burst))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "Rate limit exceeded. Please try again later.",
				"error":   "too_many_requests",
			})
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", formatInt(rl.burst))
		c.Header("X-RateLimit-Remaining", formatInt(int(limiter.Tokens())))

		c.Next()
	}
}

// formatInt converts int to string for headers
func formatInt(n int) string {
	if n < 0 {
		return "-" + formatIntPositive(-n)
	}
	return formatIntPositive(n)
}

func formatIntPositive(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return formatIntPositive(n/10) + string(rune('0'+n%10))
}

// Stats returns current statistics about the rate limiter
func (rl *TenantRateLimiter) Stats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return map[string]interface{}{
		"active_tenants":      len(rl.limiters),
		"rate_per_second":     float64(rl.rate),
		"burst_size":          rl.burst,
		"cleanup_interval_ms": rl.cleanupTick.Milliseconds(),
		"entry_ttl_ms":        rl.entryTTL.Milliseconds(),
	}
}
