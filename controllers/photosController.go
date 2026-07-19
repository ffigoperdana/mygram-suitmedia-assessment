package controllers

import (
	"errors"
	"finalproject/helpers"
	"finalproject/models"
	"finalproject/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreatePhoto godoc
// @Summary Create photo
// @Description Create photo to post in finalproject
// @Tags photo
// @Accept json
// @Produce json
// @Param request body models.Photo true "Photo request body"
// @Security BearerAuth
// @Success 201 {object} models.Photo "Create photo success"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /photos/create [post]
// @Router /api/v1/photos [post]
func CreatePhoto(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	claims, ok := requireClaims(c)
	if !ok {
		return
	}

	photo := models.Photo{}
	if err := helpers.BindRequest(c, &photo); err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	if err := helpers.ValidateHTTPURL(photo.PhotoURL); err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	photo.UserID = claims.ID
	if err := db.Create(&photo).Error; err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	go services.NotifyNewPhoto(db, photo)

	c.JSON(http.StatusCreated, photo)
}

// GetAllPhotos godoc
// @Summary Get all photos
// @Description Get all existing photos
// @Tags photo
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} []models.Photo{} "Get all photos success"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Database error"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /photos/getall [get]
// @Router /api/v1/photos [get]
func GetAllPhotos(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	allPhotos := []models.Photo{}
	if err := db.Find(&allPhotos).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to load photos")
		return
	}

	c.JSON(http.StatusOK, allPhotos)
}

// GetPhoto godoc
// @Summary Get photo
// @Description Get photo by ID
// @Tags photo
// @Accept json
// @Produce json
// @Param photoId path int true "ID of the photo"
// @Security BearerAuth
// @Success 200 {object} models.Photo{} "Get photo success"
// @Failure 400 {object} ErrorResponse "Invalid photo id"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Photo not found"
// @Failure 500 {object} ErrorResponse "Database error"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /photos/get/{photoId} [get]
// @Router /api/v1/photos/{photoId} [get]
func GetPhoto(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	photoID, err := helpers.ParseUintParam(c, "photoId")
	if err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", "invalid photo id")
		return
	}

	photo := models.Photo{}
	if err := db.First(&photo, photoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			jsonError(c, http.StatusNotFound, "Photo Not Found", "photo does not exist")
			return
		}

		jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to load photo")
		return
	}

	c.JSON(http.StatusOK, photo)
}

// UpdatePhoto godoc
// @Summary Update photo
// @Description Update photo identified by given ID
// @Tags photo
// @Accept json
// @Produce json
// @Param photoId path int true "ID of the photo"
// @Param request body models.Photo true "Photo update request body"
// @Security BearerAuth
// @Success 200 {object} models.Photo{} "Update photo success"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 404 {object} ErrorResponse "Photo not found"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /photos/update/{photoId} [put]
// @Router /api/v1/photos/{photoId} [put]
func UpdatePhoto(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	photoID, err := helpers.ParseUintParam(c, "photoId")
	if err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", "invalid photo id")
		return
	}

	photo := models.Photo{}
	if err := helpers.BindRequest(c, &photo); err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	if err := helpers.ValidateHTTPURL(photo.PhotoURL); err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	photo.ID = photoID
	if err := db.Model(&photo).Where("id = ?", photoID).Updates(map[string]interface{}{
		"title":     photo.Title,
		"caption":   photo.Caption,
		"photo_url": photo.PhotoURL,
	}).Error; err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	if err := db.First(&photo, photoID).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to load updated photo")
		return
	}

	c.JSON(http.StatusOK, photo)
}

// DeletePhoto godoc
// @Summary Delete photo
// @Description Delete photo identified by given ID
// @Tags photo
// @Accept json
// @Produce json
// @Param photoId path int true "ID of the photo"
// @Security BearerAuth
// @Success 200 {object} DeleteResponse "Delete photo success"
// @Failure 400 {object} ErrorResponse "Invalid photo id"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 404 {object} ErrorResponse "Photo not found"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /photos/delete/{photoId} [delete]
// @Router /api/v1/photos/{photoId} [delete]
func DeletePhoto(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	photoID, err := helpers.ParseUintParam(c, "photoId")
	if err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", "invalid photo id")
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("photo_id = ?", photoID).Delete(&models.Comment{}).Error; err != nil {
			return err
		}

		result := tx.Delete(&models.Photo{}, photoID)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		jsonError(c, http.StatusNotFound, "Photo Not Found", "photo does not exist")
		return
	}
	if err != nil {
		jsonError(c, http.StatusBadRequest, "Delete Error", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "delete_success",
		"message": "the photo has been successfully deleted",
	})
}
