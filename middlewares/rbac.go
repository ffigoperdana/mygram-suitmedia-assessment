package middlewares

import (
	"finalproject/helpers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, role := range allowedRoles {
		allowed[role] = true
	}

	return func(c *gin.Context) {
		claims, ok := helpers.GetUserClaims(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthenticated",
				"message": "sign in to proceed",
			})
			return
		}

		if !allowed[claims.Role] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "insufficient permissions",
			})
			return
		}

		c.Next()
	}
}
