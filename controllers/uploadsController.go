package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"time"

	"finalproject/config"
	"finalproject/services"

	"github.com/gin-gonic/gin"
)

var allowedUploadImageTypes = map[string]string{
	"image/gif":  ".gif",
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// UploadPhotoImage godoc
// @Summary Upload photo image
// @Description Upload an image file to configured S3-compatible object storage and return its public URL
// @Tags uploads
// @Accept mpfd
// @Produce json
// @Param file formData file true "Image file"
// @Security BearerAuth
// @Success 201 {object} UploadPhotoResponse "Upload success"
// @Failure 400 {object} ErrorResponse "Invalid file"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 413 {object} ErrorResponse "File too large"
// @Failure 502 {object} ErrorResponse "Object storage upload failed"
// @Failure 503 {object} ErrorResponse "Object storage is not configured"
// @Router /api/v1/uploads/photos [post]
func UploadPhotoImage(c *gin.Context) {
	cfg := config.Load()
	claims, ok := requireClaims(c)
	if !ok {
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", "image file is required")
		return
	}

	if fileHeader.Size <= 0 {
		jsonError(c, http.StatusBadRequest, "Bad Request", "image file is empty")
		return
	}

	if fileHeader.Size > cfg.S3UploadMaxBytes {
		jsonError(c, http.StatusRequestEntityTooLarge, "Payload Too Large", fmt.Sprintf("image must be %d MB or smaller", cfg.S3UploadMaxBytes/(1024*1024)))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", "failed to read image file")
		return
	}
	defer file.Close()

	contentType, extension, err := detectUploadImageType(file)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	key, err := photoUploadObjectKey(claims.ID, extension)
	if err != nil {
		jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to prepare image upload")
		return
	}

	result, err := services.UploadObject(c.Request.Context(), cfg, services.ObjectUploadInput{
		Key:         key,
		ContentType: contentType,
		Body:        file,
		Size:        fileHeader.Size,
	})
	if errors.Is(err, services.ErrObjectStorageNotConfigured) {
		jsonError(c, http.StatusServiceUnavailable, "Service Unavailable", "object storage is not configured")
		return
	}
	if err != nil {
		log.Printf(
			"object storage upload failed: %v; endpoint=%q region=%q bucket=%q force_path_style=%t public_base_url_set=%t access_key_prefix=%q secret_key_set=%t",
			err,
			cfg.S3Endpoint,
			cfg.S3Region,
			cfg.S3Bucket,
			cfg.S3ForcePathStyle,
			cfg.S3PublicBaseURL != "",
			accessKeyPrefix(cfg.S3AccessKeyID),
			cfg.S3SecretAccessKey != "",
		)
		jsonError(c, http.StatusBadGateway, "Bad Gateway", "failed to upload image")
		return
	}

	c.JSON(http.StatusCreated, UploadPhotoResponse{
		URL:         result.URL,
		Key:         result.Key,
		Bucket:      result.Bucket,
		ContentType: result.ContentType,
		Size:        result.Size,
	})
}

func detectUploadImageType(file multipart.File) (string, string, error) {
	header := make([]byte, 512)
	bytesRead, err := file.Read(header)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", "", errors.New("failed to read image file")
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", "", errors.New("failed to reset image file")
	}

	contentType := http.DetectContentType(header[:bytesRead])
	extension, ok := allowedUploadImageTypes[contentType]
	if !ok {
		return "", "", errors.New("file must be a jpeg, png, gif, or webp image")
	}

	return contentType, extension, nil
}

func photoUploadObjectKey(userID uint, extension string) (string, error) {
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"uploads/photos/%d/%d-%s%s",
		userID,
		time.Now().UTC().UnixNano(),
		hex.EncodeToString(randomBytes),
		extension,
	), nil
}

func accessKeyPrefix(accessKeyID string) string {
	if len(accessKeyID) <= 6 {
		return accessKeyID
	}

	return accessKeyID[:6] + "..."
}
