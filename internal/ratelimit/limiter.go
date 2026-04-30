package ratelimit

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hadygust/url-shortener/internal/cache"
)

// RateLimiter handles rate limiting using cache with a sliding window algorithm
type RateLimiter struct {
	cache cache.Cache
}

// Config holds rate limit configuration
type Config struct {
	MaxRequests int           // Number of requests allowed
	Window      time.Duration // Time window for the limit
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(c cache.Cache) *RateLimiter {
	return &RateLimiter{
		cache: c,
	}
}

// AllowRequest checks if a request is allowed based on rate limiting rules
// Uses sliding window algorithm: removes old entries and counts current window
func (rl *RateLimiter) AllowRequest(key string, maxRequests int, window time.Duration) (bool, int, error) {
	now := time.Now().Unix()
	windowStart := now - int64(window.Seconds())

	// Get current counter from cache
	data, err := rl.cache.Get(key)
	var counter cache.RequestCounter

	if err != nil {
		return false, 0, fmt.Errorf("rate limit check failed: %w", err)
	}

	if data != nil {
		// Parse existing counter
		jsonData, err := json.Marshal(data)
		if err != nil {
			return false, 0, fmt.Errorf("failed to marshal counter: %w", err)
		}
		if err := json.Unmarshal(jsonData, &counter); err != nil {
			return false, 0, fmt.Errorf("failed to parse counter: %w", err)
		}
	}

	// Remove timestamps outside the sliding window
	var validTimestamps []int64
	for _, ts := range counter.Timestamps {
		if ts > windowStart {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	// Check if request is allowed
	if len(validTimestamps) < maxRequests {
		// Add current request
		validTimestamps = append(validTimestamps, now)
		counter.Timestamps = validTimestamps

		// Save updated counter back to cache
		err := rl.cache.Set(key, counter, window)
		if err != nil {
			return false, 0, fmt.Errorf("failed to update rate limit: %w", err)
		}

		return true, len(validTimestamps), nil
	}

	// Limit exceeded
	return false, len(validTimestamps), nil
}

// GetRemaining returns the remaining requests in the current window
func (rl *RateLimiter) GetRemaining(key string, maxRequests int, window time.Duration) (int, error) {
	now := time.Now().Unix()
	windowStart := now - int64(window.Seconds())

	// Get current counter from cache
	data, err := rl.cache.Get(key)
	if err != nil {
		return 0, fmt.Errorf("failed to get remaining: %w", err)
	}

	var counter cache.RequestCounter
	if data != nil {
		// Parse existing counter
		jsonData, err := json.Marshal(data)
		if err != nil {
			return 0, fmt.Errorf("failed to marshal counter: %w", err)
		}
		if err := json.Unmarshal(jsonData, &counter); err != nil {
			return 0, fmt.Errorf("failed to parse counter: %w", err)
		}
	}

	// Count valid timestamps in the window
	validCount := 0
	for _, ts := range counter.Timestamps {
		if ts > windowStart {
			validCount++
		}
	}

	remaining := maxRequests - validCount
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// ResetLimit clears rate limit data for a key
func (rl *RateLimiter) ResetLimit(key string) error {
	return rl.cache.Delete(key)
}

// AllowUserPostRequest checks if a user is allowed to create a URL (POST request)
// Uses the specialized cache method for URL creation rate limiting
func (rl *RateLimiter) AllowUserPostRequest(userID string, maxRequests int, window time.Duration) (bool, int, error) {
	now := time.Now().Unix()
	windowStart := now - int64(window.Seconds())

	// Get current counter from cache using specialized method
	counter, err := rl.cache.GetUrlPostLimit(userID)
	if err != nil {
		return false, 0, fmt.Errorf("rate limit check failed: %w", err)
	}

	var timestamps []int64
	if counter != nil {
		timestamps = counter.Timestamps
	}

	// Remove timestamps outside the sliding window
	var validTimestamps []int64
	for _, ts := range timestamps {
		if ts > windowStart {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	// Check if request is allowed
	if len(validTimestamps) < maxRequests {
		validTimestamps = append(validTimestamps, now)
		newCounter := &cache.RequestCounter{Timestamps: validTimestamps}

		err := rl.cache.SetUrlPostLimit(userID, newCounter, window)
		if err != nil {
			return false, 0, fmt.Errorf("failed to update rate limit: %w", err)
		}

		return true, len(validTimestamps), nil
	}

	return false, len(validTimestamps), nil
}

// AllowUrlGetRequest checks if a request is allowed to access a URL (GET request)
// Uses the specialized cache method for URL access rate limiting
func (rl *RateLimiter) AllowUrlGetRequest(shortCode string, maxRequests int, window time.Duration) (bool, int, error) {
	now := time.Now().Unix()
	windowStart := now - int64(window.Seconds())

	// Get current counter from cache using specialized method
	counter, err := rl.cache.GetUrlGetLimit(shortCode)
	if err != nil {
		return false, 0, fmt.Errorf("rate limit check failed: %w", err)
	}

	var timestamps []int64
	if counter != nil {
		timestamps = counter.Timestamps
	}

	// Remove timestamps outside the sliding window
	var validTimestamps []int64
	for _, ts := range timestamps {
		if ts > windowStart {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	// Check if request is allowed
	if len(validTimestamps) < maxRequests {
		validTimestamps = append(validTimestamps, now)
		newCounter := &cache.RequestCounter{Timestamps: validTimestamps}

		err := rl.cache.SetUrlGetLimit(shortCode, newCounter, window)
		if err != nil {
			return false, 0, fmt.Errorf("failed to update rate limit: %w", err)
		}

		return true, len(validTimestamps), nil
	}

	return false, len(validTimestamps), nil
}

// AllowIPRequest checks if a request is allowed from an IP address
// Uses the specialized cache method for IP-based rate limiting
func (rl *RateLimiter) AllowIPRequest(ip string, maxRequests int, window time.Duration) (bool, int, error) {
	now := time.Now().Unix()
	windowStart := now - int64(window.Seconds())

	// Get current counter from cache using specialized method
	counter, err := rl.cache.GetIPRateLimit(ip)
	if err != nil {
		return false, 0, fmt.Errorf("rate limit check failed: %w", err)
	}

	var timestamps []int64
	if counter != nil {
		timestamps = counter.Timestamps
	}

	// Remove timestamps outside the sliding window
	var validTimestamps []int64
	for _, ts := range timestamps {
		if ts > windowStart {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	// Check if request is allowed
	if len(validTimestamps) < maxRequests {
		validTimestamps = append(validTimestamps, now)
		newCounter := &cache.RequestCounter{Timestamps: validTimestamps}

		err := rl.cache.SetIPRateLimit(ip, newCounter, window)
		if err != nil {
			return false, 0, fmt.Errorf("failed to update rate limit: %w", err)
		}

		return true, len(validTimestamps), nil
	}

	return false, len(validTimestamps), nil
}

// AllowRedirectRequest checks if a redirect request is allowed for a URL
// Uses the specialized cache method for redirect tracking rate limiting
func (rl *RateLimiter) AllowRedirectRequest(urlID string, maxRequests int, window time.Duration) (bool, int, error) {
	now := time.Now().Unix()
	windowStart := now - int64(window.Seconds())

	// Get current counter from cache using specialized method
	counter, err := rl.cache.GetRedirectLimit(urlID)
	if err != nil {
		return false, 0, fmt.Errorf("rate limit check failed: %w", err)
	}

	var timestamps []int64
	if counter != nil {
		timestamps = counter.Timestamps
	}

	// Remove timestamps outside the sliding window
	var validTimestamps []int64
	for _, ts := range timestamps {
		if ts > windowStart {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	// Check if request is allowed
	if len(validTimestamps) < maxRequests {
		validTimestamps = append(validTimestamps, now)
		newCounter := &cache.RequestCounter{Timestamps: validTimestamps}

		err := rl.cache.SetRedirectLimit(urlID, newCounter, window)
		if err != nil {
			return false, 0, fmt.Errorf("failed to update rate limit: %w", err)
		}

		return true, len(validTimestamps), nil
	}

	return false, len(validTimestamps), nil
}
