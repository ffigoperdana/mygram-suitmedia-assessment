package controllers

import (
	"context"
	"finalproject/database"
	"finalproject/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Check if the application is running
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Application is healthy"
// @Router /health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "mygram-api",
		"version":   "1.0.0",
	})
}

// ReadinessCheck godoc
// @Summary Readiness check endpoint
// @Description Check if the application is ready to serve requests (includes DB check)
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Application is ready"
// @Failure 503 {object} map[string]interface{} "Application is not ready"
// @Router /health/ready [get]
func ReadinessCheck(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not ready",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"service":   "mygram-api",
			"error":     "database is not initialized",
		})
		return
	}

	// Check database connectivity
	sqlDB, err := db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not ready",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"service":   "mygram-api",
			"error":     "database connection failed",
			"details":   err.Error(),
		})
		return
	}

	// Ping database
	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not ready",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"service":   "mygram-api",
			"error":     "database ping failed",
			"details":   err.Error(),
		})
		return
	}

	redisStatus := "disabled"
	if redisStore := services.GetRedisStore(); redisStore != nil {
		redisStatus = "connected"
		redisContext, cancel := context.WithTimeout(c.Request.Context(), time.Second)
		defer cancel()
		if err := redisStore.Ping(redisContext); err != nil {
			redisStatus = "degraded"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "mygram-api",
		"database":  "connected",
		"redis":     redisStatus,
		"version":   "1.0.0",
	})
}

// LivenessCheck godoc
// @Summary Liveness check endpoint
// @Description Check if the application is alive
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Application is alive"
// @Router /health/live [get]
func LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "mygram-api",
		"uptime":    time.Since(startTime).String(),
	})
}

var startTime = time.Now()
