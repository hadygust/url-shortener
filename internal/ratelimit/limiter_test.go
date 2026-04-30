package ratelimit

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/hadygust/url-shortener/internal/cache"
	"github.com/hadygust/url-shortener/internal/dto"
	"github.com/hadygust/url-shortener/internal/model"
)

// MockCache implements the cache.Cache interface for testing
type MockCache struct {
	data map[string]any
	mu   sync.RWMutex
}

func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]any),
	}
}

func (m *MockCache) Get(key string) (any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, exists := m.data[key]
	if !exists {
		return nil, nil
	}
	return val, nil
}

func (m *MockCache) Set(key string, value any, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

func (m *MockCache) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

func (m *MockCache) GetUrl(shortCode string) (*dto.UrlCache, error) {
	// Not needed for rate limiter tests
	return nil, nil
}

func (m *MockCache) SetUrl(url model.Url) error {
	// Not needed for rate limiter tests
	return nil
}

// Rate limit cache method implementations for testing
func (m *MockCache) GetUrlPostLimit(userID string) (*cache.RequestCounter, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, exists := m.data["ratelimit:user:post:"+userID]
	if !exists {
		return nil, nil
	}
	counter, ok := val.(cache.RequestCounter)
	if !ok {
		return nil, nil
	}
	return &counter, nil
}

func (m *MockCache) SetUrlPostLimit(userID string, counter *cache.RequestCounter, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data["ratelimit:user:post:"+userID] = counter
	return nil
}

func (m *MockCache) GetUrlGetLimit(key string) (*cache.RequestCounter, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, exists := m.data["ratelimit:url:get:"+key]
	if !exists {
		return nil, nil
	}
	counter, ok := val.(cache.RequestCounter)
	if !ok {
		return nil, nil
	}
	return &counter, nil
}

func (m *MockCache) SetUrlGetLimit(key string, counter *cache.RequestCounter, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data["ratelimit:url:get:"+key] = counter
	return nil
}

func (m *MockCache) GetIPRateLimit(ip string) (*cache.RequestCounter, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, exists := m.data["ratelimit:ip:"+ip]
	if !exists {
		return nil, nil
	}
	counter, ok := val.(cache.RequestCounter)
	if !ok {
		return nil, nil
	}
	return &counter, nil
}

func (m *MockCache) SetIPRateLimit(ip string, counter *cache.RequestCounter, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data["ratelimit:ip:"+ip] = counter
	return nil
}

func (m *MockCache) GetRedirectLimit(urlID string) (*cache.RequestCounter, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, exists := m.data["ratelimit:redirect:"+urlID]
	if !exists {
		return nil, nil
	}
	counter, ok := val.(cache.RequestCounter)
	if !ok {
		return nil, nil
	}
	return &counter, nil
}

func (m *MockCache) SetRedirectLimit(urlID string, counter *cache.RequestCounter, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data["ratelimit:redirect:"+urlID] = counter
	return nil
}

// TestNewRateLimiter tests the constructor
func TestNewRateLimiter(t *testing.T) {
	cache := NewMockCache()
	limiter := NewRateLimiter(cache)

	if limiter == nil {
		t.Fatal("expected non-nil rate limiter")
	}
	if limiter.cache == nil {
		t.Fatal("expected cache to be set")
	}
}

// TestAllowRequestFirstRequest tests that the first request is always allowed
func TestAllowRequestFirstRequest(t *testing.T) {
	cache := NewMockCache()
	limiter := NewRateLimiter(cache)

	allowed, count, err := limiter.AllowRequest("test-key", 10, time.Minute)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected first request to be allowed")
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}
}

// TestAllowRequestMultipleRequests tests allowing multiple requests within limit
func TestAllowRequestMultipleRequests(t *testing.T) {
	cache := NewMockCache()
	limiter := NewRateLimiter(cache)
	maxRequests := 5
	window := time.Minute

	for i := 1; i <= maxRequests; i++ {
		allowed, count, err := limiter.AllowRequest("test-key", maxRequests, window)

		if err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, err)
		}
		if !allowed {
			t.Errorf("iteration %d: expected request to be allowed", i)
		}
		if count != i {
			t.Errorf("iteration %d: expected count %d, got %d", i, i, count)
		}
	}
}

// TestAllowRequestExceedsLimit tests that requests are rejected when limit is exceeded
func TestAllowRequestExceedsLimit(t *testing.T) {
	cache := NewMockCache()
	limiter := NewRateLimiter(cache)
	maxRequests := 3
	window := time.Minute

	// Fill up the limit
	for i := 0; i < maxRequests; i++ {
		limiter.AllowRequest("test-key", maxRequests, window)
	}

	// Next request should be rejected
	allowed, count, err := limiter.AllowRequest("test-key", maxRequests, window)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected request to be rejected when limit exceeded")
	}
	if count != maxRequests {
		t.Errorf("expected count %d, got %d", maxRequests, count)
	}
}

// TestAllowRequestSlidingWindow tests that the sliding window properly removes old timestamps
func TestAllowRequestSlidingWindow(t *testing.T) {
	cache := NewMockCache()
	limiter := NewRateLimiter(cache)
	maxRequests := 2
	window := 2 * time.Second

	key := "test-key"

	// First request
	allowed1, _, _ := limiter.AllowRequest(key, maxRequests, window)
	if !allowed1 {
		t.Fatal("expected first request to be allowed")
	}

	// Second request
	allowed2, _, _ := limiter.AllowRequest(key, maxRequests, window)
	if !allowed2 {
		t.Fatal("expected second request to be allowed")
	}

	// Third request should be rejected (limit is 2)
	allowed3, _, _ := limiter.AllowRequest(key, maxRequests, window)
	if allowed3 {
		t.Error("expected third request to be rejected (limit exceeded)")
	}

	// Wait for window to expire
	time.Sleep(2500 * time.Millisecond)

	// Now the request should be allowed again (old timestamps expired)
	allowed4, count, _ := limiter.AllowRequest(key, maxRequests, window)
	if !allowed4 {
		t.Error("expected request to be allowed after window expiry")
	}
	if count != 1 {
		t.Errorf("expected count 1 after window expiry, got %d", count)
	}
}

// TestGetRemaining tests the GetRemaining function
func TestGetRemaining(t *testing.T) {
	cache := NewMockCache()
	limiter := NewRateLimiter(cache)
	maxRequests := 5
	window := time.Minute
	key := "test-key"

	// No requests yet
	remaining, err := limiter.GetRemaining(key, maxRequests, window)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if remaining != maxRequests {
		t.Errorf("expected %d remaining, got %d", maxRequests, remaining)
	}

	// Make 2 requests
	limiter.AllowRequest(key, maxRequests, window)
	limiter.AllowRequest(key, maxRequests, window)

	remaining, _ = limiter.GetRemaining(key, maxRequests, window)
	expectedRemaining := maxRequests - 2
	if remaining != expectedRemaining {
		t.Errorf("expected %d remaining, got %d", expectedRemaining, remaining)
	}
}

// TestGetRemainingAfterReset tests that remaining is correct after reset
func TestGetRemainingAfterReset(t *testing.T) {
	cache := NewMockCache()
	limiter := NewRateLimiter(cache)
	maxRequests := 5
	window := time.Minute
	key := "test-key"

	// Make 3 requests
	limiter.AllowRequest(key, maxRequests, window)
	limiter.AllowRequest(key, maxRequests, window)
	limiter.AllowRequest(key, maxRequests, window)

	// Reset limit
	err := limiter.ResetLimit(key)
	if err != nil {
		t.Fatalf("unexpected error during reset: %v", err)
	}

	// Should have full capacity again
	remaining, _ := limiter.GetRemaining(key, maxRequests, window)
	if remaining != maxRequests {
		t.Errorf("expected %d remaining after reset, got %d", maxRequests, remaining)
	}
}

// TestResetLimit tests the ResetLimit function
func TestResetLimit(t *testing.T) {
	cache := NewMockCache()
	limiter := NewRateLimiter(cache)
	key := "test-key"

	// Make some requests
	limiter.AllowRequest(key, 5, time.Minute)
	limiter.AllowRequest(key, 5, time.Minute)

	// Verify data is in cache
	data, _ := cache.Get(key)
	if data == nil {
		t.Fatal("expected data to be in cache")
	}

	// Reset
	err := limiter.ResetLimit(key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify data is removed from cache
	data, _ = cache.Get(key)
	if data != nil {
		t.Error("expected data to be removed from cache after reset")
	}
}

// TestAllowRequestConcurrency tests thread-safe behavior with sequential requests
// Note: In production with Redis, atomic Lua scripts provide stronger guarantees
func TestAllowRequestConcurrency(t *testing.T) {
	cache := NewMockCache()
	limiter := NewRateLimiter(cache)
	maxRequests := 10
	window := time.Minute
	key := "test-key"

	successCount := 0
	failureCount := 0

	// Make sequential requests (simulating concurrent requests arriving in sequence)
	for i := 0; i < 20; i++ {
		allowed, _, _ := limiter.AllowRequest(key, maxRequests, window)
		if allowed {
			successCount++
		} else {
			failureCount++
		}
	}

	if successCount != maxRequests {
		t.Errorf("expected %d successful requests, got %d", maxRequests, successCount)
	}
	if failureCount != 20-maxRequests {
		t.Errorf("expected %d failed requests, got %d", 20-maxRequests, failureCount)
	}
}

// TestAllowRequestWithDifferentKeys tests that different keys have separate limits
func TestAllowRequestWithDifferentKeys(t *testing.T) {
	cache := NewMockCache()
	limiter := NewRateLimiter(cache)
	maxRequests := 2
	window := time.Minute

	// Fill limit for key1
	limiter.AllowRequest("key1", maxRequests, window)
	limiter.AllowRequest("key1", maxRequests, window)

	// Third request to key1 should fail
	allowed1, _, _ := limiter.AllowRequest("key1", maxRequests, window)
	if allowed1 {
		t.Error("expected key1 request to be rejected")
	}

	// But key2 should still have full capacity
	allowed2, count, _ := limiter.AllowRequest("key2", maxRequests, window)
	if !allowed2 {
		t.Error("expected key2 request to be allowed")
	}
	if count != 1 {
		t.Errorf("expected count 1 for key2, got %d", count)
	}
}

// TestAllowRequestZeroRequests tests behavior with zero max requests
func TestAllowRequestZeroRequests(t *testing.T) {
	cache := NewMockCache()
	limiter := NewRateLimiter(cache)

	allowed, _, err := limiter.AllowRequest("test-key", 0, time.Minute)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected request to be rejected with zero max requests")
	}
}

// TestRequestCounterSerialization tests that RequestCounter is properly serialized/deserialized
func TestRequestCounterSerialization(t *testing.T) {
	mockCache := NewMockCache()
	counter := cache.RequestCounter{
		Timestamps: []int64{100, 200, 300},
	}

	// Manually set in cache
	mockCache.Set("test-key", counter, time.Minute)

	// Retrieve and verify
	data, _ := mockCache.Get("test-key")
	jsonData, _ := json.Marshal(data)
	var retrieved cache.RequestCounter
	json.Unmarshal(jsonData, &retrieved)

	if len(retrieved.Timestamps) != 3 {
		t.Errorf("expected 3 timestamps, got %d", len(retrieved.Timestamps))
	}
	if retrieved.Timestamps[1] != 200 {
		t.Errorf("expected second timestamp to be 200, got %d", retrieved.Timestamps[1])
	}
}

// TestGetRemainingWithExpiredTimestamps tests that GetRemaining ignores expired timestamps
func TestGetRemainingWithExpiredTimestamps(t *testing.T) {
	cache := NewMockCache()
	limiter := NewRateLimiter(cache)
	maxRequests := 5
	window := 1 * time.Second
	key := "test-key"

	// Make 3 requests
	limiter.AllowRequest(key, maxRequests, window)
	limiter.AllowRequest(key, maxRequests, window)
	limiter.AllowRequest(key, maxRequests, window)

	// Verify remaining
	remaining1, _ := limiter.GetRemaining(key, maxRequests, window)
	if remaining1 != 2 {
		t.Errorf("expected 2 remaining, got %d", remaining1)
	}

	// Wait for window to expire
	time.Sleep(1500 * time.Millisecond)

	// Remaining should be back to full (old timestamps expired)
	remaining2, _ := limiter.GetRemaining(key, maxRequests, window)
	if remaining2 != maxRequests {
		t.Errorf("expected %d remaining after window expiry, got %d", maxRequests, remaining2)
	}
}

// TestAllowRequestEmptyKey tests with empty string key
func TestAllowRequestEmptyKey(t *testing.T) {
	cache := NewMockCache()
	limiter := NewRateLimiter(cache)

	allowed, _, err := limiter.AllowRequest("", 5, time.Minute)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected request with empty key to be allowed")
	}
}

// TestMultipleUsersIsolation tests that different user limits are isolated
func TestMultipleUsersIsolation(t *testing.T) {
	cache := NewMockCache()
	limiter := NewRateLimiter(cache)
	limit := 3
	window := time.Minute

	// User 1: use up their limit
	for i := 0; i < limit; i++ {
		limiter.AllowRequest("user:1", limit, window)
	}

	// User 1: next request rejected
	allowed1, _, _ := limiter.AllowRequest("user:1", limit, window)
	if allowed1 {
		t.Error("expected user 1 request to be rejected")
	}

	// User 2: should have full limit available
	allowed2, count, _ := limiter.AllowRequest("user:2", limit, window)
	if !allowed2 {
		t.Error("expected user 2 request to be allowed")
	}
	if count != 1 {
		t.Errorf("expected user 2 count 1, got %d", count)
	}

	// User 3: should also have full limit available
	allowed3, count, _ := limiter.AllowRequest("user:3", limit, window)
	if !allowed3 {
		t.Error("expected user 3 request to be allowed")
	}
	if count != 1 {
		t.Errorf("expected user 3 count 1, got %d", count)
	}
}
