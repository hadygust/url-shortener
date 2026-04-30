package cache

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/hadygust/url-shortener/internal/dto"
	"github.com/hadygust/url-shortener/internal/model"
	"github.com/redis/go-redis/v9"
)

func (r *RedisCache) GetUrl(shortCode string) (*dto.UrlCache, error) {
	key := "url:" + shortCode

	val, err := r.Get(key)
	if err != nil || val == nil {
		return nil, err
	}

	bytes, _ := json.Marshal(val)

	var result dto.UrlCache
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *RedisCache) SetUrl(url model.Url) error {
	key := "url:" + url.ShortCode

	data := dto.UrlCache{
		ID:          url.ID.String(),
		OriginalUrl: url.OriginalUrl,
	}

	ttl := time.Until(url.ExpiresAt.Time)
	if ttl <= 0 {
		return nil
	}

	return r.Set(key, data, ttl)
}

type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisCache(addr string) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &RedisCache{
		client: rdb,
		ctx:    context.Background(),
	}
}

func (r *RedisCache) Get(key string) (any, error) {
	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		log.Println("error fetching redis")
		return nil, err
	}

	var data any
	err = json.Unmarshal([]byte(val), &data)
	if err != nil {
		log.Println("error unmarshalling cache")
		return nil, err
	}

	log.Printf("Got from cache: %#v", data)
	return data, nil
}

func (r *RedisCache) Set(key string, value any, ttl time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(r.ctx, key, bytes, ttl).Err()
}

func (r *RedisCache) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// GetUrlPostLimit retrieves the request counter for URL POST (creation) operations by user
func (r *RedisCache) GetUrlPostLimit(userID string) (*RequestCounter, error) {
	key := "ratelimit:user:post:" + userID

	val, err := r.Get(key)
	if err != nil || val == nil {
		return nil, err
	}

	bytes, _ := json.Marshal(val)

	var counter RequestCounter
	if err := json.Unmarshal(bytes, &counter); err != nil {
		return nil, err
	}

	return &counter, nil
}

// SetUrlPostLimit stores the request counter for URL POST (creation) operations by user
func (r *RedisCache) SetUrlPostLimit(userID string, counter *RequestCounter, ttl time.Duration) error {
	key := "ratelimit:user:post:" + userID
	return r.Set(key, counter, ttl)
}

// GetUrlGetLimit retrieves the request counter for URL GET (access) operations
func (r *RedisCache) GetUrlGetLimit(key string) (*RequestCounter, error) {
	cacheKey := "ratelimit:url:get:" + key

	val, err := r.Get(cacheKey)
	if err != nil || val == nil {
		return nil, err
	}

	bytes, _ := json.Marshal(val)

	var counter RequestCounter
	if err := json.Unmarshal(bytes, &counter); err != nil {
		return nil, err
	}

	return &counter, nil
}

// SetUrlGetLimit stores the request counter for URL GET (access) operations
func (r *RedisCache) SetUrlGetLimit(key string, counter *RequestCounter, ttl time.Duration) error {
	cacheKey := "ratelimit:url:get:" + key
	return r.Set(cacheKey, counter, ttl)
}

// GetIPRateLimit retrieves the request counter for IP-based rate limiting
func (r *RedisCache) GetIPRateLimit(ip string) (*RequestCounter, error) {
	key := "ratelimit:ip:" + ip

	val, err := r.Get(key)
	if err != nil || val == nil {
		return nil, err
	}

	bytes, _ := json.Marshal(val)

	var counter RequestCounter
	if err := json.Unmarshal(bytes, &counter); err != nil {
		return nil, err
	}

	return &counter, nil
}

// SetIPRateLimit stores the request counter for IP-based rate limiting
func (r *RedisCache) SetIPRateLimit(ip string, counter *RequestCounter, ttl time.Duration) error {
	key := "ratelimit:ip:" + ip
	return r.Set(key, counter, ttl)
}

// GetRedirectLimit retrieves the request counter for redirect tracking per URL
func (r *RedisCache) GetRedirectLimit(urlID string) (*RequestCounter, error) {
	key := "ratelimit:redirect:" + urlID

	val, err := r.Get(key)
	if err != nil || val == nil {
		return nil, err
	}

	bytes, _ := json.Marshal(val)

	var counter RequestCounter
	if err := json.Unmarshal(bytes, &counter); err != nil {
		return nil, err
	}

	return &counter, nil
}

// SetRedirectLimit stores the request counter for redirect tracking per URL
func (r *RedisCache) SetRedirectLimit(urlID string, counter *RequestCounter, ttl time.Duration) error {
	key := "ratelimit:redirect:" + urlID
	return r.Set(key, counter, ttl)
}
