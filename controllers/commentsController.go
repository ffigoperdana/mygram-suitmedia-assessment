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

// CreateComments godoc
// @Summary Create comments
// @Description Create comments for photo identified by given id
// @Tags comments
// @Accept json
// @Produce json
// @Param photoId path int true "ID of the photo"
// @Param request body models.Comment true "Comment request body"
// @Security BearerAuth
// @Success 201 {object} models.Comment "Create comments success"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Photo not found"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /comments/create/{photoId} [post]
// @Router /api/v1/photos/{photoId}/comments [post]
func CreateComment(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	claims, ok := requireClaims(c)
	if !ok {
		return
	}

	photoID, err := helpers.ParseUintParam(c, "photoId")
	if err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", "invalid photo id")
		return
	}

	photo := models.Photo{}
	if err := db.Select("id", "user_id", "title").First(&photo, photoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			jsonError(c, http.StatusNotFound, "Photo Not Found", "photo does not exist, failed to create comment")
			return
		}

		jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to load photo")
		return
	}

	comment := models.Comment{}
	if err := helpers.BindRequest(c, &comment); err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	comment.UserID = claims.ID
	comment.PhotoID = photoID
	if err := db.Create(&comment).Error; err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	go services.NotifyNewComment(db, photo, comment)

	c.JSON(http.StatusCreated, comment)
}

// GetAllComments godoc
// @Summary Get all comments
// @Description Get all comments in finalproject
// @Tags comment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} []models.Comment "Get all comments success"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Database error"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /comments/getall [get]
// @Router /api/v1/comments [get]
func GetAllComments(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	allComments := []models.Comment{}
	if err := db.Find(&allComments).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to load comments")
		return
	}

	c.JSON(http.StatusOK, allComments)
}

// GetAllCommentsForPhoto godoc
// @Summary Get all comments for specific photo
// @Description Get all comments for photo with given id
// @Tags comment
// @Accept json
// @Produce json
// @Param photoId path int true "ID of the photo"
// @Security BearerAuth
// @Success 200 {object} []models.Comment "Get all comments success"
// @Failure 400 {object} ErrorResponse "Invalid photo id"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Database error"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /comments/getall/{photoId} [get]
// @Router /api/v1/photos/{photoId}/comments [get]
func GetAllCommentsForPhoto(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	photoID, err := helpers.ParseUintParam(c, "photoId")
	if err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", "invalid photo id")
		return
	}

	allComments := []models.Comment{}
	if err := db.Where("photo_id = ?", photoID).Find(&allComments).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to load comments")
		return
	}

	c.JSON(http.StatusOK, allComments)
}

// GetComments godoc
// @Summary Get comments
// @Description Get comments identified by given id
// @Tags comments
// @Accept json
// @Produce json
// @Param commentId path int true "ID of the comments"
// @Security BearerAuth
// @Success 200 {object} models.Comment "Get comment success"
// @Failure 400 {object} ErrorResponse "Invalid comment id"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Comment not found"
// @Failure 500 {object} ErrorResponse "Database error"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /comments/get/{commentId} [get]
// @Router /api/v1/comments/{commentId} [get]
func GetComment(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	commentID, err := helpers.ParseUintParam(c, "commentId")
	if err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", "invalid comment id")
		return
	}

	comment := models.Comment{}
	if err := db.First(&comment, commentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			jsonError(c, http.StatusNotFound, "Comment Not Found", "comment does not exist")
			return
		}

		jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to load comment")
		return
	}

	c.JSON(http.StatusOK, comment)
}

// UpdateComment godoc
// @Summary Update comment
// @Description Update comment identified by given id
// @Tags comment
// @Accept json
// @Produce json
// @Param commentId path int true "ID of the comment"
// @Param request body models.Comment true "Comment update request body"
// @Security BearerAuth
// @Success 200 {object} models.Comment "Update comment success"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 404 {object} ErrorResponse "Comment not found"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /comments/update/{commentId} [put]
// @Router /api/v1/comments/{commentId} [put]
func UpdateComment(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	commentID, err := helpers.ParseUintParam(c, "commentId")
	if err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", "invalid comment id")
		return
	}

	comment := models.Comment{}
	if err := helpers.BindRequest(c, &comment); err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	comment.ID = commentID
	if err := db.Model(&comment).Where("id = ?", commentID).Updates(map[string]interface{}{
		"message": comment.Message,
	}).Error; err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	if err := db.First(&comment, commentID).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to load updated comment")
		return
	}

	c.JSON(http.StatusOK, comment)
}

// DeleteComment godoc
// @Summary Delete comment
// @Description Delete comment identified by given ID
// @Tags comment
// @Accept json
// @Produce json
// @Param commentId path int true "ID of the comment"
// @Security BearerAuth
// @Success 200 {object} DeleteResponse "Delete comment success"
// @Failure 400 {object} ErrorResponse "Invalid comment id"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 404 {object} ErrorResponse "Comment not found"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /comments/delete/{commentId} [delete]
// @Router /api/v1/comments/{commentId} [delete]
func DeleteComment(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	commentID, err := helpers.ParseUintParam(c, "commentId")
	if err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", "invalid comment id")
		return
	}

	if err := db.Delete(&models.Comment{}, commentID).Error; err != nil {
		jsonError(c, http.StatusBadRequest, "Delete Error", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "delete_success",
		"message": "the comment has been successfully deleted",
	})
}
