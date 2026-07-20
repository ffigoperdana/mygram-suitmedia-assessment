package services

import (
	"context"
	"encoding/json"
	"errors"
	"finalproject/config"
	"finalproject/models"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const PhotosListCacheKey = "mygram:photos:list:v1"

var ErrCacheMiss = errors.New("cache miss")

type RateLimitResult struct {
	Allowed    bool
	Remaining  int
	RetryAfter time.Duration
}

type RedisStore interface {
	GetPhotos(context.Context) ([]models.Photo, error)
	SetPhotos(context.Context, []models.Photo) error
	InvalidatePhotos(context.Context) error
	Allow(context.Context, string, int, time.Duration) (RateLimitResult, error)
	Ping(context.Context) error
	Close() error
}

var (
	redisStoreMu sync.RWMutex
	redisStore   RedisStore
)

func SetRedisStore(store RedisStore) func() {
	redisStoreMu.Lock()
	previous := redisStore
	redisStore = store
	redisStoreMu.Unlock()

	return func() {
		redisStoreMu.Lock()
		redisStore = previous
		redisStoreMu.Unlock()
	}
}

func GetRedisStore() RedisStore {
	redisStoreMu.RLock()
	defer redisStoreMu.RUnlock()
	return redisStore
}

type RedisClient struct {
	client   *redis.Client
	cacheTTL time.Duration
}

func NewRedisClient(cfg config.Config) (*RedisClient, error) {
	if !cfg.RedisEnabled {
		return nil, nil
	}
	if cfg.RedisAddr == "" {
		return nil, errors.New("REDIS_ADDR is required when REDIS_ENABLED=true")
	}

	client := redis.NewClient(&redis.Options{
		Addr:                  cfg.RedisAddr,
		Password:              cfg.RedisPassword,
		DB:                    cfg.RedisDB,
		DialTimeout:           2 * time.Second,
		ReadTimeout:           time.Second,
		WriteTimeout:          time.Second,
		ContextTimeoutEnabled: true,
		PoolSize:              20,
		MinIdleConns:          2,
		PoolTimeout:           2 * time.Second,
		ConnMaxIdleTime:       5 * time.Minute,
		MaxRetries:            2,
		MinRetryBackoff:       50 * time.Millisecond,
		MaxRetryBackoff:       200 * time.Millisecond,
	})

	cache := &RedisClient{
		client:   client,
		cacheTTL: time.Duration(cfg.RedisCacheTTLSeconds) * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := cache.Ping(ctx); err != nil {
		return cache, fmt.Errorf("ping Redis: %w", err)
	}

	return cache, nil
}

func (cache *RedisClient) GetPhotos(ctx context.Context) ([]models.Photo, error) {
	payload, err := cache.client.Get(ctx, PhotosListCacheKey).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}

	photos := []models.Photo{}
	if err := json.Unmarshal(payload, &photos); err != nil {
		_ = cache.InvalidatePhotos(ctx)
		return nil, fmt.Errorf("decode cached photos: %w", err)
	}

	return photos, nil
}

func (cache *RedisClient) SetPhotos(ctx context.Context, photos []models.Photo) error {
	payload, err := json.Marshal(photos)
	if err != nil {
		return fmt.Errorf("encode photos for cache: %w", err)
	}

	return cache.client.Set(ctx, PhotosListCacheKey, payload, cache.cacheTTL).Err()
}

func (cache *RedisClient) InvalidatePhotos(ctx context.Context) error {
	return cache.client.Del(ctx, PhotosListCacheKey).Err()
}

var rateLimitScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
local ttl = redis.call("PTTL", KEYS[1])
if current == 1 or ttl < 0 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
  ttl = tonumber(ARGV[1])
end
return {current, ttl}
`)

func (cache *RedisClient) Allow(
	ctx context.Context,
	key string,
	limit int,
	window time.Duration,
) (RateLimitResult, error) {
	result, err := rateLimitScript.Run(ctx, cache.client, []string{key}, window.Milliseconds()).Slice()
	if err != nil {
		return RateLimitResult{}, err
	}
	if len(result) != 2 {
		return RateLimitResult{}, errors.New("unexpected Redis rate-limit response")
	}

	current, ok := result[0].(int64)
	if !ok {
		return RateLimitResult{}, errors.New("invalid Redis rate-limit counter")
	}
	ttlMilliseconds, ok := result[1].(int64)
	if !ok {
		return RateLimitResult{}, errors.New("invalid Redis rate-limit TTL")
	}

	remaining := limit - int(current)
	if remaining < 0 {
		remaining = 0
	}

	return RateLimitResult{
		Allowed:    current <= int64(limit),
		Remaining:  remaining,
		RetryAfter: time.Duration(ttlMilliseconds) * time.Millisecond,
	}, nil
}

func (cache *RedisClient) Ping(ctx context.Context) error {
	return cache.client.Ping(ctx).Err()
}

func (cache *RedisClient) Close() error {
	return cache.client.Close()
}
