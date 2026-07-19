package middlewares

import (
	"finalproject/database"
	"finalproject/helpers"
	"finalproject/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := helpers.VerifyToken(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthenticated",
				"message": err.Error(),
			})
			return
		}

		db := database.GetDB()
		if db == nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service Unavailable",
				"message": "database is not ready",
			})
			return
		}

		user := models.User{}
		if err := db.Select("id", "role", "status").First(&user, claims.ID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthenticated",
				"message": "user no longer exists",
			})
			return
		}

		if user.Status == models.UserStatusBanned {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "user account is banned",
			})
			return
		}

		now := time.Now().UTC()
		_ = db.Model(&models.User{}).Where("id = ?", user.ID).Update("last_seen_at", &now).Error
		claims.Role = user.Role
		c.Set(helpers.UserDataKey, claims)
		c.Next()
	}
}
