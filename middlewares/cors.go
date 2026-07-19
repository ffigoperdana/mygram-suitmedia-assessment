package middlewares

import (
	"finalproject/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	cfg := config.Load()
	allowedOrigins := map[string]bool{}
	allowAnyOrigin := false

	for _, origin := range cfg.CORSAllowedOrigins {
		if origin == "*" {
			allowAnyOrigin = true
			continue
		}

		allowedOrigins[origin] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" && (allowAnyOrigin || allowedOrigins[origin]) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept, Origin")
			c.Header("Access-Control-Allow-Methods", strings.Join([]string{
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
				http.MethodOptions,
			}, ", "))
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
