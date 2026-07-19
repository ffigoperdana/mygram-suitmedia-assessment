package controllers

import (
	"errors"
	"finalproject/helpers"
	"finalproject/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateSocialMedia godoc
// @Summary Create social media
// @Description Create social media of the user
// @Tags social media
// @Accept json
// @Produce json
// @Param request body models.SocialMedia true "Social media request body"
// @Security BearerAuth
// @Success 201 {object} models.SocialMedia "Create social media success"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /socialmedia/create [post]
// @Router /api/v1/social-media [post]
func CreateSocialMedia(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	claims, ok := requireClaims(c)
	if !ok {
		return
	}

	socialMedia := models.SocialMedia{}
	if err := helpers.BindRequest(c, &socialMedia); err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	if err := helpers.ValidateSocialProfileURL(socialMedia.SocialMediaURL); err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	socialMedia.UserID = claims.ID
	if err := db.Create(&socialMedia).Error; err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	c.JSON(http.StatusCreated, socialMedia)
}

// GetAllSocialMedia godoc
// @Summary Get all social media
// @Description Get all social media in finalproject
// @Tags social media
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} []models.SocialMedia "Get all social media success"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Database error"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /socialmedia/getall [get]
// @Router /api/v1/social-media [get]
func GetAllSocialMedias(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	allSocialMedias := []models.SocialMedia{}
	if err := db.Find(&allSocialMedias).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to load social media links")
		return
	}

	c.JSON(http.StatusOK, allSocialMedias)
}

// GetSocialMedia godoc
// @Summary Get social media
// @Description Get social media identified by given id
// @Tags social media
// @Accept json
// @Produce json
// @Param socialMediaId path int true "ID of the social media"
// @Security BearerAuth
// @Success 200 {object} models.SocialMedia "Get social media success"
// @Failure 400 {object} ErrorResponse "Invalid social media id"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Social media not found"
// @Failure 500 {object} ErrorResponse "Database error"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /socialmedia/get/{socialMediaId} [get]
// @Router /api/v1/social-media/{socialMediaId} [get]
func GetSocialMedia(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	socialMediaID, err := helpers.ParseUintParam(c, "socialMediaId")
	if err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", "invalid social media id")
		return
	}

	socialMedia := models.SocialMedia{}
	if err := db.First(&socialMedia, socialMediaID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			jsonError(c, http.StatusNotFound, "Social Media Not Found", "social media does not exist")
			return
		}

		jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to load social media")
		return
	}

	c.JSON(http.StatusOK, socialMedia)
}

// UpdateSocialMedia godoc
// @Summary Update social media
// @Description Update social media identified by given id
// @Tags social media
// @Accept json
// @Produce json
// @Param socialMediaId path int true "ID of the social media"
// @Param request body models.SocialMedia true "Social media update request body"
// @Security BearerAuth
// @Success 200 {object} models.SocialMedia "Update social media success"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 404 {object} ErrorResponse "Social media not found"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /socialmedia/update/{socialMediaId} [put]
// @Router /api/v1/social-media/{socialMediaId} [put]
func UpdateSocialMedia(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	socialMediaID, err := helpers.ParseUintParam(c, "socialMediaId")
	if err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", "invalid social media id")
		return
	}

	socialMedia := models.SocialMedia{}
	if err := helpers.BindRequest(c, &socialMedia); err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	if err := helpers.ValidateSocialProfileURL(socialMedia.SocialMediaURL); err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	socialMedia.ID = socialMediaID
	if err := db.Model(&socialMedia).Where("id = ?", socialMediaID).Updates(map[string]interface{}{
		"name":             socialMedia.Name,
		"social_media_url": socialMedia.SocialMediaURL,
	}).Error; err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	if err := db.First(&socialMedia, socialMediaID).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to load updated social media")
		return
	}

	c.JSON(http.StatusOK, socialMedia)
}

// DeleteSocialMedia godoc
// @Summary Delete social media
// @Description Delete social media identified by given ID
// @Tags social media
// @Accept json
// @Produce json
// @Param socialMediaId path int true "ID of the social media"
// @Security BearerAuth
// @Success 200 {object} DeleteResponse "Delete social media success"
// @Failure 400 {object} ErrorResponse "Invalid social media id"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 404 {object} ErrorResponse "Social media not found"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /socialmedia/delete/{socialMediaId} [delete]
// @Router /api/v1/social-media/{socialMediaId} [delete]
func DeleteSocialMedia(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	socialMediaID, err := helpers.ParseUintParam(c, "socialMediaId")
	if err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", "invalid social media id")
		return
	}

	if err := db.Delete(&models.SocialMedia{}, socialMediaID).Error; err != nil {
		jsonError(c, http.StatusBadRequest, "Delete Error", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "delete_success",
		"message": "the social media has been successfully deleted",
	})
}
