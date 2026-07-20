package services

import (
	"context"
	"finalproject/config"
	"finalproject/models"
	"os"
	"testing"
	"time"
)

func TestRedisPhotoCacheIntegration(t *testing.T) {
	if os.Getenv("REQUIRE_TEST_REDIS") != "true" {
		t.Skip("set REQUIRE_TEST_REDIS=true to run the Redis integration test")
	}

	client, err := NewRedisClient(config.Config{
		RedisEnabled:         true,
		RedisAddr:            os.Getenv("REDIS_ADDR"),
		RedisPassword:        os.Getenv("REDIS_PASSWORD"),
		RedisDB:              0,
		RedisCacheTTLSeconds: 60,
	})
	if err != nil {
		t.Fatalf("connect to test Redis: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	_ = client.InvalidatePhotos(ctx)
	photos := []models.Photo{{Title: "integration"}}
	if err := client.SetPhotos(ctx, photos); err != nil {
		t.Fatalf("set cached photos: %v", err)
	}
	cached, err := client.GetPhotos(ctx)
	if err != nil || len(cached) != 1 || cached[0].Title != "integration" {
		t.Fatalf("unexpected cached photos: photos=%v err=%v", cached, err)
	}
	if err := client.InvalidatePhotos(ctx); err != nil {
		t.Fatalf("invalidate cached photos: %v", err)
	}
	if _, err := client.GetPhotos(ctx); err != ErrCacheMiss {
		t.Fatalf("expected cache miss after invalidation, got %v", err)
	}

	key := "mygram:ratelimit:integration:test"
	client.client.Del(ctx, key)
	first, err := client.Allow(ctx, key, 1, time.Minute)
	if err != nil || !first.Allowed {
		t.Fatalf("first rate-limited request should be allowed: result=%+v err=%v", first, err)
	}
	second, err := client.Allow(ctx, key, 1, time.Minute)
	if err != nil || second.Allowed {
		t.Fatalf("second rate-limited request should be rejected: result=%+v err=%v", second, err)
	}
}
