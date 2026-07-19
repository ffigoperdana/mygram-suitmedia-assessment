package middlewares

import (
	"finalproject/database"
	"finalproject/helpers"
	"finalproject/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func PhotoAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := database.GetDB()
		if db == nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service Unavailable",
				"message": "database is not ready",
			})
			return
		}

		photoID, err := helpers.ParseUintParam(c, "photoId")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "Bad Request",
				"message": "invalid photo id",
			})
			return
		}

		claims, ok := helpers.GetUserClaims(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthenticated",
				"message": "sign in to proceed",
			})
			return
		}

		photo := models.Photo{}
		if err := db.Select("user_id").First(&photo, photoID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error":   "Photo Not Found",
				"message": "photo does not exist",
			})
			return
		}

		if photo.UserID != claims.ID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "you are not the owner of the photo",
			})
			return
		}

		c.Next()
	}
}

func CommentAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := database.GetDB()
		if db == nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service Unavailable",
				"message": "database is not ready",
			})
			return
		}

		commentID, err := helpers.ParseUintParam(c, "commentId")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "Bad Request",
				"message": "invalid comment id",
			})
			return
		}

		claims, ok := helpers.GetUserClaims(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthenticated",
				"message": "sign in to proceed",
			})
			return
		}

		comment := models.Comment{}
		if err := db.Select("user_id").First(&comment, commentID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error":   "Comment Not Found",
				"message": "comment does not exist",
			})
			return
		}

		if comment.UserID != claims.ID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "you are not the writer of the comment",
			})
			return
		}

		c.Next()
	}
}

func SocialMediaAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := database.GetDB()
		if db == nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service Unavailable",
				"message": "database is not ready",
			})
			return
		}

		socialMediaID, err := helpers.ParseUintParam(c, "socialMediaId")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "Bad Request",
				"message": "invalid social media id",
			})
			return
		}

		claims, ok := helpers.GetUserClaims(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthenticated",
				"message": "sign in to proceed",
			})
			return
		}

		socialMedia := models.SocialMedia{}
		if err := db.Select("user_id").First(&socialMedia, socialMediaID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error":   "Social Media Not Found",
				"message": "social media does not exist",
			})
			return
		}

		if socialMedia.UserID != claims.ID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "you are not the owner of the social media",
			})
			return
		}

		c.Next()
	}
}
