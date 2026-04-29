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
