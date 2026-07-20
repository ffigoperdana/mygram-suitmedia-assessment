package middlewares

import (
	"context"
	"errors"
	"finalproject/models"
	"finalproject/services"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type fakeRateLimitStore struct {
	result services.RateLimitResult
	err    error
}

func (store *fakeRateLimitStore) GetPhotos(context.Context) ([]models.Photo, error) { return nil, nil }
func (store *fakeRateLimitStore) SetPhotos(context.Context, []models.Photo) error   { return nil }
func (store *fakeRateLimitStore) InvalidatePhotos(context.Context) error            { return nil }
func (store *fakeRateLimitStore) Allow(context.Context, string, int, time.Duration) (services.RateLimitResult, error) {
	return store.result, store.err
}
func (store *fakeRateLimitStore) Ping(context.Context) error { return nil }
func (store *fakeRateLimitStore) Close() error               { return nil }

func TestRedisRateLimitRejectsExceededLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	restore := services.SetRedisStore(&fakeRateLimitStore{result: services.RateLimitResult{
		Allowed: false, RetryAfter: 5 * time.Second,
	}})
	defer restore()

	router := gin.New()
	router.Use(RedisRateLimit("test", 1, time.Minute))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", recorder.Code)
	}
}

func TestRedisRateLimitFailureIsFailOpen(t *testing.T) {
	gin.SetMode(gin.TestMode)
	restore := services.SetRedisStore(&fakeRateLimitStore{err: errors.New("redis unavailable")})
	defer restore()

	router := gin.New()
	router.Use(RedisRateLimit("test", 1, time.Minute))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("Redis failure must fail open; got %d", recorder.Code)
	}
}
