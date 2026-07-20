package middlewares

import (
	"crypto/sha256"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"finalproject/services"

	"github.com/gin-gonic/gin"
)

func RedisRateLimit(scope string, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		store := services.GetRedisStore()
		if store == nil {
			c.Next()
			return
		}

		identifierHash := sha256.Sum256([]byte(clientIdentifier(c)))
		identifier := fmt.Sprintf("%x", identifierHash[:16])
		key := fmt.Sprintf("mygram:ratelimit:%s:%s", scope, identifier)
		result, err := store.Allow(c.Request.Context(), key, limit, window)
		if err != nil {
			log.Printf("Redis rate limiter unavailable; allowing request: %v", err)
			c.Next()
			return
		}

		resetSeconds := int(math.Ceil(result.RetryAfter.Seconds()))
		if resetSeconds < 1 {
			resetSeconds = 1
		}
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		c.Header("X-RateLimit-Reset", strconv.Itoa(resetSeconds))

		if !result.Allowed {
			c.Header("Retry-After", strconv.Itoa(resetSeconds))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too Many Requests",
				"message": "rate limit exceeded; retry later",
			})
			return
		}

		c.Next()
	}
}

func clientIdentifier(c *gin.Context) string {
	forwardedFor := c.GetHeader("X-Forwarded-For")
	if forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		if len(parts) >= 2 {
			return strings.TrimSpace(parts[len(parts)-2])
		}
		return strings.TrimSpace(parts[0])
	}

	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return c.Request.RemoteAddr
}
