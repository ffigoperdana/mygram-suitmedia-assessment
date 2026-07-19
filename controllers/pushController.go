package controllers

import (
	"finalproject/config"
	"finalproject/models"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type pushVAPIDPublicKeyResponse struct {
	Enabled   bool   `json:"enabled"`
	PublicKey string `json:"public_key,omitempty"`
}

type pushSubscriptionKeysRequest struct {
	P256DH string `json:"p256dh" binding:"required"`
	Auth   string `json:"auth" binding:"required"`
}

type pushSubscriptionRequest struct {
	Endpoint  string                      `json:"endpoint" binding:"required"`
	Keys      pushSubscriptionKeysRequest `json:"keys" binding:"required"`
	UserAgent string                      `json:"user_agent"`
}

type deletePushSubscriptionRequest struct {
	Endpoint string `json:"endpoint"`
}

func GetPushVAPIDPublicKey(c *gin.Context) {
	cfg := config.Load()
	response := pushVAPIDPublicKeyResponse{
		Enabled: cfg.PushNotificationsConfigured(),
	}

	if response.Enabled {
		response.PublicKey = cfg.VAPIDPublicKey
	}

	c.JSON(http.StatusOK, response)
}

func SavePushSubscription(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	claims, ok := requireClaims(c)
	if !ok {
		return
	}

	cfg := config.Load()
	if !cfg.PushNotificationsConfigured() {
		jsonError(c, http.StatusServiceUnavailable, "Service Unavailable", "push notifications are not configured")
		return
	}

	var request pushSubscriptionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", "endpoint, p256dh, and auth are required")
		return
	}

	request.Endpoint = strings.TrimSpace(request.Endpoint)
	if !isValidPushEndpoint(request.Endpoint) {
		jsonError(c, http.StatusBadRequest, "Bad Request", "push endpoint must be a valid HTTPS URL")
		return
	}

	now := time.Now().UTC()
	subscription := models.PushSubscription{Endpoint: request.Endpoint}
	updates := models.PushSubscription{
		UserID:     claims.ID,
		P256DH:     request.Keys.P256DH,
		Auth:       request.Keys.Auth,
		UserAgent:  truncatePushValue(request.UserAgent, 512),
		LastUsedAt: &now,
	}

	if err := db.Where(models.PushSubscription{Endpoint: request.Endpoint}).
		Assign(updates).
		FirstOrCreate(&subscription).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to save push subscription")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "subscription_saved",
	})
}

func DeletePushSubscription(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	claims, ok := requireClaims(c)
	if !ok {
		return
	}

	var request deletePushSubscriptionRequest
	_ = c.ShouldBindJSON(&request)

	endpoint := strings.TrimSpace(request.Endpoint)
	if endpoint == "" {
		endpoint = strings.TrimSpace(c.Query("endpoint"))
	}

	query := db.Where("user_id = ?", claims.ID)
	if endpoint != "" {
		query = query.Where("endpoint = ?", endpoint)
	}

	if err := query.Delete(&models.PushSubscription{}).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to delete push subscription")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "subscription_deleted",
	})
}

func isValidPushEndpoint(endpoint string) bool {
	parsed, err := url.ParseRequestURI(endpoint)
	if err != nil || parsed.Host == "" {
		return false
	}

	if parsed.Scheme == "https" {
		return true
	}

	if parsed.Scheme != "http" {
		return false
	}

	host := strings.Trim(parsed.Hostname(), "[]")
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

func truncatePushValue(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}

	return value[:max]
}
