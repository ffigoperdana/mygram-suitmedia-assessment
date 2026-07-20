package controllers

import (
	"context"
	"errors"
	"finalproject/models"
	"finalproject/services"
	"testing"
)

type fakePhotoCache struct {
	photos          []models.Photo
	getErr          error
	setErr          error
	invalidateErr   error
	setCalls        int
	invalidateCalls int
}

func (cache *fakePhotoCache) GetPhotos(context.Context) ([]models.Photo, error) {
	return cache.photos, cache.getErr
}

func (cache *fakePhotoCache) SetPhotos(_ context.Context, photos []models.Photo) error {
	cache.setCalls++
	cache.photos = photos
	return cache.setErr
}

func (cache *fakePhotoCache) InvalidatePhotos(context.Context) error {
	cache.invalidateCalls++
	return cache.invalidateErr
}

func TestLoadPhotosCacheAsideHit(t *testing.T) {
	want := []models.Photo{{Title: "cached"}}
	cache := &fakePhotoCache{photos: want}
	loaderCalls := 0

	got, hit, err := loadPhotosCacheAside(context.Background(), cache, func(context.Context) ([]models.Photo, error) {
		loaderCalls++
		return nil, nil
	})
	if err != nil || !hit || loaderCalls != 0 || len(got) != 1 || got[0].Title != "cached" {
		t.Fatalf("unexpected cache hit result: got=%v hit=%v err=%v loaderCalls=%d", got, hit, err, loaderCalls)
	}
}

func TestLoadPhotosCacheAsideMiss(t *testing.T) {
	want := []models.Photo{{Title: "database"}}
	cache := &fakePhotoCache{getErr: services.ErrCacheMiss}

	got, hit, err := loadPhotosCacheAside(context.Background(), cache, func(context.Context) ([]models.Photo, error) {
		return want, nil
	})
	if err != nil || hit || cache.setCalls != 1 || len(got) != 1 || got[0].Title != "database" {
		t.Fatalf("unexpected cache miss result: got=%v hit=%v err=%v setCalls=%d", got, hit, err, cache.setCalls)
	}
}

func TestLoadPhotosCacheAsideRedisFailureFallsBackToPostgreSQL(t *testing.T) {
	want := []models.Photo{{Title: "database fallback"}}
	cache := &fakePhotoCache{getErr: errors.New("redis unavailable"), setErr: errors.New("redis unavailable")}

	got, hit, err := loadPhotosCacheAside(context.Background(), cache, func(context.Context) ([]models.Photo, error) {
		return want, nil
	})
	if err != nil || hit || len(got) != 1 || got[0].Title != "database fallback" {
		t.Fatalf("Redis failure must remain fail-open: got=%v hit=%v err=%v", got, hit, err)
	}
}

func TestInvalidatePhotoListCache(t *testing.T) {
	cache := &fakePhotoCache{invalidateErr: errors.New("redis unavailable")}
	invalidatePhotoListCache(context.Background(), cache)
	if cache.invalidateCalls != 1 {
		t.Fatalf("expected one invalidation, got %d", cache.invalidateCalls)
	}
}
