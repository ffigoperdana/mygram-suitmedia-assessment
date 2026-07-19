package controllers

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"finalproject/config"
	"finalproject/services"

	"github.com/gin-gonic/gin"
)

const mediaPhotoPrefix = "uploads/photos/"

// ServeMediaObject streams uploaded photo media through the API domain.
func ServeMediaObject(c *gin.Context) {
	key, ok := mediaObjectKey(c.Param("objectKey"))
	if !ok {
		jsonError(c, http.StatusNotFound, "Not Found", "media does not exist")
		return
	}

	cfg := config.Load()
	result, err := services.GetObject(c.Request.Context(), cfg, key)
	if errors.Is(err, services.ErrObjectStorageNotConfigured) {
		jsonError(c, http.StatusServiceUnavailable, "Service Unavailable", "object storage is not configured")
		return
	}
	if mediaObjectNotFound(err) {
		jsonError(c, http.StatusNotFound, "Not Found", "media does not exist")
		return
	}
	if err != nil {
		log.Printf(
			"object storage download failed: %v; endpoint=%q region=%q bucket=%q key=%q force_path_style=%t",
			err,
			cfg.S3Endpoint,
			cfg.S3Region,
			cfg.S3Bucket,
			key,
			cfg.S3ForcePathStyle,
		)
		jsonError(c, http.StatusBadGateway, "Bad Gateway", "failed to load media")
		return
	}
	defer result.Body.Close()

	contentType := result.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	headers := map[string]string{
		"Cache-Control": "public, max-age=31536000, immutable",
	}
	if result.CacheControl != "" {
		headers["Cache-Control"] = result.CacheControl
	}
	if result.ETag != "" {
		headers["ETag"] = result.ETag
	}

	c.DataFromReader(http.StatusOK, result.ContentLength, contentType, result.Body, headers)
}

func mediaObjectKey(rawObjectKey string) (string, bool) {
	decodedKey, err := url.PathUnescape(strings.TrimPrefix(rawObjectKey, "/"))
	if err != nil {
		return "", false
	}

	key := strings.TrimPrefix(path.Clean("/"+decodedKey), "/")
	if key == "." || key == "" {
		return "", false
	}
	if !strings.HasPrefix(key, mediaPhotoPrefix) {
		return "", false
	}

	return key, true
}

func mediaObjectNotFound(err error) bool {
	if err == nil {
		return false
	}

	var storageErr *services.ObjectStorageError
	if !errors.As(err, &storageErr) {
		return false
	}

	switch storageErr.Code {
	case "NoSuchKey", "NotFound", "NoSuchBucket":
		return true
	}

	return storageErr.StatusCode == http.StatusNotFound
}
