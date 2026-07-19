package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"finalproject/database"
	"finalproject/helpers"
	"finalproject/models"
	"finalproject/router"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthEndpointsWithoutDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	database.SetDB(nil)
	r := router.StartApp()

	response := performJSONRequest(r, http.MethodGet, "/health", nil, "")
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Contains(t, response.Body.String(), "healthy")

	response = performJSONRequest(r, http.MethodGet, "/health/live", nil, "")
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Contains(t, response.Body.String(), "alive")

	response = performJSONRequest(r, http.MethodGet, "/health/ready", nil, "")
	assert.Equal(t, http.StatusServiceUnavailable, response.Code)
	assert.Contains(t, response.Body.String(), "database is not initialized")
}

func TestCORSPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")

	r := router.StartApp()
	request := httptest.NewRequest(http.MethodOptions, "/photos/getall", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	request.Header.Set("Access-Control-Request-Method", http.MethodGet)
	response := httptest.NewRecorder()

	r.ServeHTTP(response, request)

	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Equal(t, "http://localhost:3000", response.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, response.Header().Get("Access-Control-Allow-Headers"), "Authorization")
}

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	database.SetDB(nil)
	r := router.StartApp()

	response := performJSONRequest(r, http.MethodGet, "/health", nil, "")
	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "nosniff", response.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", response.Header().Get("X-Frame-Options"))
	assert.Contains(t, response.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'")
}

func TestPublicOpenAPISpecFiltersAdminAndLegacyRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	database.SetDB(nil)
	t.Setenv("PUBLIC_OPENAPI_ENABLED", "true")
	t.Setenv("SWAGGER_UI_MODE", "public")

	r := router.StartApp()
	response := performJSONRequest(r, http.MethodGet, "/openapi/public.json", nil, "")
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())

	var spec struct {
		Paths       map[string]interface{} `json:"paths"`
		Definitions map[string]interface{} `json:"definitions"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &spec))

	require.Contains(t, spec.Paths, "/api/v1/auth/register")
	require.Contains(t, spec.Paths, "/api/v1/auth/login")
	require.Contains(t, spec.Paths, "/api/v1/me")
	require.Contains(t, spec.Paths, "/api/v1/photos")
	require.Contains(t, spec.Paths, "/api/v1/uploads/photos")
	require.Contains(t, spec.Paths, "/api/v1/social-media")

	mePath, ok := spec.Paths["/api/v1/me"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, mePath, "get")
	assert.Contains(t, mePath, "patch")

	for path := range spec.Paths {
		assert.NotContains(t, path, "/api/v1/admin")
		assert.False(t, strings.HasPrefix(path, "/users/"), path)
		assert.False(t, strings.HasPrefix(path, "/photos/"), path)
		assert.False(t, strings.HasPrefix(path, "/comments/"), path)
		assert.False(t, strings.HasPrefix(path, "/socialmedia/"), path)
	}

	for definition := range spec.Definitions {
		assert.NotContains(t, definition, "Admin")
	}
}

func TestSwaggerUICanBeDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	database.SetDB(nil)
	t.Setenv("PUBLIC_OPENAPI_ENABLED", "true")
	t.Setenv("SWAGGER_UI_MODE", "disabled")

	r := router.StartApp()

	publicSpecResponse := performJSONRequest(r, http.MethodGet, "/openapi/public.json", nil, "")
	assert.Equal(t, http.StatusOK, publicSpecResponse.Code)

	swaggerResponse := performJSONRequest(r, http.MethodGet, "/swagger/index.html", nil, "")
	assert.Equal(t, http.StatusNotFound, swaggerResponse.Code)
}

func TestSwaggerUIAllowsSameOriginEmbedding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	database.SetDB(nil)
	t.Setenv("PUBLIC_OPENAPI_ENABLED", "true")
	t.Setenv("SWAGGER_UI_MODE", "public")

	r := router.StartApp()

	swaggerResponse := performJSONRequest(r, http.MethodGet, "/swagger/index.html", nil, "")
	require.Equal(t, http.StatusOK, swaggerResponse.Code, swaggerResponse.Body.String())
	assert.Equal(t, "SAMEORIGIN", swaggerResponse.Header().Get("X-Frame-Options"))
	assert.Contains(t, swaggerResponse.Header().Get("Content-Security-Policy"), "frame-ancestors 'self'")
}

func TestMediaProxyRejectsInvalidPathsAndRequiresStorage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	database.SetDB(nil)
	t.Setenv("S3_ENDPOINT", " ")
	t.Setenv("S3_REGION", " ")
	t.Setenv("S3_BUCKET", " ")
	t.Setenv("S3_ACCESS_KEY_ID", " ")
	t.Setenv("S3_SECRET_ACCESS_KEY", " ")

	r := router.StartApp()

	invalidPathResponse := performJSONRequest(r, http.MethodGet, "/media/users/private.jpg", nil, "")
	assert.Equal(t, http.StatusNotFound, invalidPathResponse.Code)

	validPathResponse := performJSONRequest(r, http.MethodGet, "/media/uploads/photos/1/photo.jpg", nil, "")
	assert.Equal(t, http.StatusServiceUnavailable, validPathResponse.Code)
}

func TestUserRegistrationLoginAndValidation(t *testing.T) {
	r := setupDatabaseBackedTest(t)

	registerResponse := performJSONRequest(r, http.MethodPost, "/users/register", map[string]interface{}{
		"username": "figo",
		"email":    "figo@example.com",
		"password": "secret123",
		"age":      20,
	}, "")
	require.Equal(t, http.StatusCreated, registerResponse.Code, registerResponse.Body.String())
	assert.NotContains(t, registerResponse.Body.String(), "secret123")
	assert.NotContains(t, registerResponse.Body.String(), "password")

	duplicateResponse := performJSONRequest(r, http.MethodPost, "/users/register", map[string]interface{}{
		"username": "figo2",
		"email":    "figo@example.com",
		"password": "secret123",
		"age":      20,
	}, "")
	assert.Equal(t, http.StatusBadRequest, duplicateResponse.Code)

	invalidEmailResponse := performJSONRequest(r, http.MethodPost, "/users/register", map[string]interface{}{
		"username": "invalid",
		"email":    "not-email",
		"password": "secret123",
		"age":      20,
	}, "")
	assert.Equal(t, http.StatusBadRequest, invalidEmailResponse.Code)

	lowAgeResponse := performJSONRequest(r, http.MethodPost, "/users/register", map[string]interface{}{
		"username": "young",
		"email":    "young@example.com",
		"password": "secret123",
		"age":      8,
	}, "")
	assert.Equal(t, http.StatusBadRequest, lowAgeResponse.Code)

	shortPasswordResponse := performJSONRequest(r, http.MethodPost, "/users/register", map[string]interface{}{
		"username": "short",
		"email":    "short@example.com",
		"password": "123",
		"age":      18,
	}, "")
	assert.Equal(t, http.StatusBadRequest, shortPasswordResponse.Code)

	loginResponse := performJSONRequest(r, http.MethodPost, "/users/login", map[string]interface{}{
		"email":    "figo@example.com",
		"password": "secret123",
	}, "")
	require.Equal(t, http.StatusOK, loginResponse.Code, loginResponse.Body.String())

	var loginBody struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(loginResponse.Body.Bytes(), &loginBody))
	require.NotEmpty(t, loginBody.Token)

	claims, err := helpers.VerifyTokenString(loginBody.Token)
	require.NoError(t, err)
	assert.Equal(t, "figo", claims.Username)
	assert.Equal(t, "figo@example.com", claims.Email)
	assert.Equal(t, models.RoleUser, claims.Role)
	require.NotNil(t, claims.ExpiresAt)
	assert.WithinDuration(t, time.Now().Add(24*time.Hour), claims.ExpiresAt.Time, 2*time.Minute)

	invalidLoginResponse := performJSONRequest(r, http.MethodPost, "/users/login", map[string]interface{}{
		"email":    "figo@example.com",
		"password": "wrong-password",
	}, "")
	assert.Equal(t, http.StatusUnauthorized, invalidLoginResponse.Code)
}

func TestAuthenticationMiddleware(t *testing.T) {
	r := setupDatabaseBackedTest(t)
	_, token := createUserAndToken(t, r, "owner", "owner@example.com")

	noTokenResponse := performJSONRequest(r, http.MethodGet, "/photos/getall", nil, "")
	assert.Equal(t, http.StatusUnauthorized, noTokenResponse.Code)

	malformedResponse := performJSONRequest(r, http.MethodGet, "/photos/getall", nil, "not-a-jwt")
	assert.Equal(t, http.StatusUnauthorized, malformedResponse.Code)

	expiredToken, err := helpers.GenerateTokenWithDuration(1, "expired", "expired@example.com", models.RoleUser, -time.Hour)
	require.NoError(t, err)
	expiredResponse := performJSONRequest(r, http.MethodGet, "/photos/getall", nil, expiredToken)
	assert.Equal(t, http.StatusUnauthorized, expiredResponse.Code)

	validResponse := performJSONRequest(r, http.MethodGet, "/photos/getall", nil, token)
	assert.Equal(t, http.StatusOK, validResponse.Code)
	assert.JSONEq(t, "[]", validResponse.Body.String())
}

func TestCurrentUserAndAdminRBAC(t *testing.T) {
	r := setupDatabaseBackedTest(t)
	db := database.GetDB()
	require.NotNil(t, db)

	userID, userToken := createUserAndToken(t, r, "member", "member@example.com")
	adminID, adminToken := createUserAndToken(t, r, "admin", "admin@example.com")
	require.NoError(t, db.Model(&models.User{}).Where("id = ?", adminID).Update("role", models.RoleAdmin).Error)

	meResponse := performJSONRequest(r, http.MethodGet, "/api/v1/me", nil, userToken)
	require.Equal(t, http.StatusOK, meResponse.Code, meResponse.Body.String())
	assert.Contains(t, meResponse.Body.String(), "member@example.com")
	assert.NotContains(t, meResponse.Body.String(), "secret123")
	assert.NotContains(t, meResponse.Body.String(), "password")

	updateMeResponse := performJSONRequest(r, http.MethodPatch, "/api/v1/me", map[string]interface{}{
		"username": "member-updated",
		"email":    "member-updated@example.com",
		"age":      26,
		"role":     models.RoleAdmin,
		"status":   models.UserStatusBanned,
	}, userToken)
	require.Equal(t, http.StatusOK, updateMeResponse.Code, updateMeResponse.Body.String())
	assert.Contains(t, updateMeResponse.Body.String(), "member-updated@example.com")
	assert.Contains(t, updateMeResponse.Body.String(), `"age":26`)
	assert.Contains(t, updateMeResponse.Body.String(), `"role":"user"`)
	assert.Contains(t, updateMeResponse.Body.String(), `"status":"active"`)

	invalidMeResponse := performJSONRequest(r, http.MethodPatch, "/api/v1/me", map[string]interface{}{
		"email": "not-email",
	}, userToken)
	assert.Equal(t, http.StatusBadRequest, invalidMeResponse.Code)

	regularAdminResponse := performJSONRequest(r, http.MethodGet, "/api/v1/admin/stats", nil, userToken)
	assert.Equal(t, http.StatusForbidden, regularAdminResponse.Code)

	statsResponse := performJSONRequest(r, http.MethodGet, "/api/v1/admin/stats", nil, adminToken)
	require.Equal(t, http.StatusOK, statsResponse.Code, statsResponse.Body.String())
	assert.Contains(t, statsResponse.Body.String(), "total_users")
	assert.Contains(t, statsResponse.Body.String(), "active_users")

	usersResponse := performJSONRequest(r, http.MethodGet, "/api/v1/admin/users?limit=10", nil, adminToken)
	require.Equal(t, http.StatusOK, usersResponse.Code, usersResponse.Body.String())
	assert.Contains(t, usersResponse.Body.String(), "member-updated@example.com")

	getUserResponse := performJSONRequest(r, http.MethodGet, fmt.Sprintf("/api/v1/admin/users/%d", userID), nil, adminToken)
	require.Equal(t, http.StatusOK, getUserResponse.Code, getUserResponse.Body.String())
	assert.Contains(t, getUserResponse.Body.String(), "member")

	updateUserResponse := performJSONRequest(r, http.MethodPatch, fmt.Sprintf("/api/v1/admin/users/%d", userID), map[string]interface{}{
		"username": "member-updated",
	}, adminToken)
	require.Equal(t, http.StatusOK, updateUserResponse.Code, updateUserResponse.Body.String())
	assert.Contains(t, updateUserResponse.Body.String(), "member-updated")

	selfBanResponse := performJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/admin/users/%d/ban", adminID), map[string]interface{}{
		"reason": "self-ban attempt",
	}, adminToken)
	assert.Equal(t, http.StatusForbidden, selfBanResponse.Code)

	banResponse := performJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/admin/users/%d/ban", userID), map[string]interface{}{
		"reason": "policy violation",
	}, adminToken)
	require.Equal(t, http.StatusOK, banResponse.Code, banResponse.Body.String())
	assert.Contains(t, banResponse.Body.String(), models.UserStatusBanned)

	bannedLoginResponse := performJSONRequest(r, http.MethodPost, "/users/login", map[string]interface{}{
		"email":    "member-updated@example.com",
		"password": "secret123",
	}, "")
	assert.Equal(t, http.StatusForbidden, bannedLoginResponse.Code)

	bannedTokenResponse := performJSONRequest(r, http.MethodGet, "/api/v1/me", nil, userToken)
	assert.Equal(t, http.StatusForbidden, bannedTokenResponse.Code)

	unbanResponse := performJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/admin/users/%d/unban", userID), nil, adminToken)
	require.Equal(t, http.StatusOK, unbanResponse.Code, unbanResponse.Body.String())
	assert.Contains(t, unbanResponse.Body.String(), models.UserStatusActive)

	restoredLoginResponse := performJSONRequest(r, http.MethodPost, "/users/login", map[string]interface{}{
		"email":    "member-updated@example.com",
		"password": "secret123",
	}, "")
	assert.Equal(t, http.StatusOK, restoredLoginResponse.Code)

	deleteCandidateID, _ := createUserAndToken(t, r, "delete-me", "delete-me@example.com")
	deleteResponse := performJSONRequest(r, http.MethodDelete, fmt.Sprintf("/api/v1/admin/users/%d", deleteCandidateID), nil, adminToken)
	require.Equal(t, http.StatusOK, deleteResponse.Code, deleteResponse.Body.String())

	deletedGetResponse := performJSONRequest(r, http.MethodGet, fmt.Sprintf("/api/v1/admin/users/%d", deleteCandidateID), nil, adminToken)
	assert.Equal(t, http.StatusNotFound, deletedGetResponse.Code)
}

func TestPushSubscriptionRoutes(t *testing.T) {
	r := setupDatabaseBackedTest(t)
	userID, token := createUserAndToken(t, r, "push-user", "push@example.com")

	disabledResponse := performJSONRequest(r, http.MethodGet, "/api/v1/push/vapid-public-key", nil, token)
	require.Equal(t, http.StatusOK, disabledResponse.Code, disabledResponse.Body.String())
	assert.Contains(t, disabledResponse.Body.String(), `"enabled":false`)

	t.Setenv("PUSH_NOTIFICATIONS_ENABLED", "true")
	t.Setenv("VAPID_PUBLIC_KEY", "test-public-key")
	t.Setenv("VAPID_PRIVATE_KEY", "test-private-key")
	t.Setenv("VAPID_SUBJECT", "mailto:test@example.com")

	keyResponse := performJSONRequest(r, http.MethodGet, "/api/v1/push/vapid-public-key", nil, token)
	require.Equal(t, http.StatusOK, keyResponse.Code, keyResponse.Body.String())
	assert.Contains(t, keyResponse.Body.String(), `"enabled":true`)
	assert.Contains(t, keyResponse.Body.String(), `"public_key":"test-public-key"`)

	endpoint := "https://updates.push.services.mozilla.com/wpush/v2/test"
	saveResponse := performJSONRequest(r, http.MethodPost, "/api/v1/push/subscriptions", map[string]interface{}{
		"endpoint": endpoint,
		"keys": map[string]interface{}{
			"p256dh": "p256dh-key",
			"auth":   "auth-key",
		},
		"user_agent": "go-test",
	}, token)
	require.Equal(t, http.StatusCreated, saveResponse.Code, saveResponse.Body.String())

	var count int64
	require.NoError(t, database.GetDB().Model(&models.PushSubscription{}).Where("user_id = ?", userID).Count(&count).Error)
	assert.Equal(t, int64(1), count)

	deleteResponse := performJSONRequest(r, http.MethodDelete, "/api/v1/push/subscriptions", map[string]interface{}{
		"endpoint": endpoint,
	}, token)
	require.Equal(t, http.StatusOK, deleteResponse.Code, deleteResponse.Body.String())

	require.NoError(t, database.GetDB().Model(&models.PushSubscription{}).Where("user_id = ?", userID).Count(&count).Error)
	assert.Equal(t, int64(0), count)
}

func TestPhotoEndpointsAndOwnership(t *testing.T) {
	r := setupDatabaseBackedTest(t)
	_, ownerToken := createUserAndToken(t, r, "owner", "owner@example.com")
	_, otherToken := createUserAndToken(t, r, "other", "other@example.com")

	createResponse := performJSONRequest(r, http.MethodPost, "/photos/create", map[string]interface{}{
		"title":     "First photo",
		"caption":   "A real caption",
		"photo_url": "https://example.com/photo.jpg",
	}, ownerToken)
	require.Equal(t, http.StatusCreated, createResponse.Code, createResponse.Body.String())

	javascriptPhotoResponse := performJSONRequest(r, http.MethodPost, "/photos/create", map[string]interface{}{
		"title":     "Unsafe photo",
		"caption":   "Nope",
		"photo_url": "javascript:alert(1)",
	}, ownerToken)
	assert.Equal(t, http.StatusBadRequest, javascriptPhotoResponse.Code)

	var createdPhoto models.Photo
	require.NoError(t, json.Unmarshal(createResponse.Body.Bytes(), &createdPhoto))
	require.NotZero(t, createdPhoto.ID)

	listResponse := performJSONRequest(r, http.MethodGet, "/photos/getall", nil, ownerToken)
	require.Equal(t, http.StatusOK, listResponse.Code)
	var photos []models.Photo
	require.NoError(t, json.Unmarshal(listResponse.Body.Bytes(), &photos))
	require.Len(t, photos, 1)

	getResponse := performJSONRequest(r, http.MethodGet, fmt.Sprintf("/photos/get/%d", createdPhoto.ID), nil, ownerToken)
	require.Equal(t, http.StatusOK, getResponse.Code)

	otherUpdateResponse := performJSONRequest(r, http.MethodPut, fmt.Sprintf("/photos/update/%d", createdPhoto.ID), map[string]interface{}{
		"title":     "Stolen edit",
		"caption":   "Nope",
		"photo_url": "https://example.com/other.jpg",
	}, otherToken)
	assert.Equal(t, http.StatusForbidden, otherUpdateResponse.Code)

	ownerUpdateResponse := performJSONRequest(r, http.MethodPut, fmt.Sprintf("/photos/update/%d", createdPhoto.ID), map[string]interface{}{
		"title":     "Updated photo",
		"caption":   "",
		"photo_url": "https://example.com/updated.jpg",
	}, ownerToken)
	require.Equal(t, http.StatusOK, ownerUpdateResponse.Code, ownerUpdateResponse.Body.String())
	assert.Contains(t, ownerUpdateResponse.Body.String(), "Updated photo")

	otherDeleteResponse := performJSONRequest(r, http.MethodDelete, fmt.Sprintf("/photos/delete/%d", createdPhoto.ID), nil, otherToken)
	assert.Equal(t, http.StatusForbidden, otherDeleteResponse.Code)

	ownerDeleteResponse := performJSONRequest(r, http.MethodDelete, fmt.Sprintf("/photos/delete/%d", createdPhoto.ID), nil, ownerToken)
	assert.Equal(t, http.StatusOK, ownerDeleteResponse.Code)

	emptyListResponse := performJSONRequest(r, http.MethodGet, "/photos/getall", nil, ownerToken)
	assert.Equal(t, http.StatusOK, emptyListResponse.Code)
	assert.JSONEq(t, "[]", emptyListResponse.Body.String())
}

func TestPhotoImageUploadValidationAndStorageConfig(t *testing.T) {
	r := setupDatabaseBackedTest(t)
	_, token := createUserAndToken(t, r, "uploader", "uploader@example.com")

	unauthenticatedResponse := performMultipartRequest(r, "/api/v1/uploads/photos", "file", "photo.png", tinyPNGBytes(), "")
	assert.Equal(t, http.StatusUnauthorized, unauthenticatedResponse.Code)

	invalidFileResponse := performMultipartRequest(r, "/api/v1/uploads/photos", "file", "notes.txt", []byte("not an image"), token)
	require.Equal(t, http.StatusBadRequest, invalidFileResponse.Code, invalidFileResponse.Body.String())
	assert.Contains(t, invalidFileResponse.Body.String(), "jpeg, png, gif, or webp")

	unconfiguredStorageResponse := performMultipartRequest(r, "/api/v1/uploads/photos", "file", "photo.png", tinyPNGBytes(), token)
	require.Equal(t, http.StatusServiceUnavailable, unconfiguredStorageResponse.Code, unconfiguredStorageResponse.Body.String())
	assert.Contains(t, unconfiguredStorageResponse.Body.String(), "object storage is not configured")
}

func TestCommentEndpointsAndOwnership(t *testing.T) {
	r := setupDatabaseBackedTest(t)
	_, ownerToken := createUserAndToken(t, r, "owner", "owner@example.com")
	_, otherToken := createUserAndToken(t, r, "other", "other@example.com")
	photo := createPhoto(t, r, ownerToken)

	missingPhotoResponse := performJSONRequest(r, http.MethodPost, "/comments/create/9999", map[string]interface{}{
		"message": "missing",
	}, ownerToken)
	assert.Equal(t, http.StatusNotFound, missingPhotoResponse.Code)

	createResponse := performJSONRequest(r, http.MethodPost, fmt.Sprintf("/comments/create/%d", photo.ID), map[string]interface{}{
		"message": "Nice photo",
	}, ownerToken)
	require.Equal(t, http.StatusCreated, createResponse.Code, createResponse.Body.String())

	var comment models.Comment
	require.NoError(t, json.Unmarshal(createResponse.Body.Bytes(), &comment))
	require.NotZero(t, comment.ID)

	listForPhotoResponse := performJSONRequest(r, http.MethodGet, fmt.Sprintf("/comments/getall/%d", photo.ID), nil, ownerToken)
	require.Equal(t, http.StatusOK, listForPhotoResponse.Code)
	var comments []models.Comment
	require.NoError(t, json.Unmarshal(listForPhotoResponse.Body.Bytes(), &comments))
	require.Len(t, comments, 1)

	otherUpdateResponse := performJSONRequest(r, http.MethodPut, fmt.Sprintf("/comments/update/%d", comment.ID), map[string]interface{}{
		"message": "Not mine",
	}, otherToken)
	assert.Equal(t, http.StatusForbidden, otherUpdateResponse.Code)

	ownerUpdateResponse := performJSONRequest(r, http.MethodPut, fmt.Sprintf("/comments/update/%d", comment.ID), map[string]interface{}{
		"message": "Edited comment",
	}, ownerToken)
	require.Equal(t, http.StatusOK, ownerUpdateResponse.Code)
	assert.Contains(t, ownerUpdateResponse.Body.String(), "Edited comment")

	ownerDeleteResponse := performJSONRequest(r, http.MethodDelete, fmt.Sprintf("/comments/delete/%d", comment.ID), nil, ownerToken)
	assert.Equal(t, http.StatusOK, ownerDeleteResponse.Code)

	emptyListResponse := performJSONRequest(r, http.MethodGet, fmt.Sprintf("/comments/getall/%d", photo.ID), nil, ownerToken)
	assert.Equal(t, http.StatusOK, emptyListResponse.Code)
	assert.JSONEq(t, "[]", emptyListResponse.Body.String())
}

func TestSocialMediaEndpointsAndOwnership(t *testing.T) {
	r := setupDatabaseBackedTest(t)
	_, ownerToken := createUserAndToken(t, r, "owner", "owner@example.com")
	_, otherToken := createUserAndToken(t, r, "other", "other@example.com")

	createResponse := performJSONRequest(r, http.MethodPost, "/socialmedia/create", map[string]interface{}{
		"name":             "GitHub",
		"social_media_url": "https://github.com/example",
	}, ownerToken)
	require.Equal(t, http.StatusCreated, createResponse.Code, createResponse.Body.String())

	var socialMedia models.SocialMedia
	require.NoError(t, json.Unmarshal(createResponse.Body.Bytes(), &socialMedia))
	require.NotZero(t, socialMedia.ID)

	invalidURLResponse := performJSONRequest(r, http.MethodPost, "/socialmedia/create", map[string]interface{}{
		"name":             "Broken",
		"social_media_url": "not-a-url",
	}, ownerToken)
	assert.Equal(t, http.StatusBadRequest, invalidURLResponse.Code)

	javascriptURLResponse := performJSONRequest(r, http.MethodPost, "/socialmedia/create", map[string]interface{}{
		"name":             "Unsafe",
		"social_media_url": "javascript:alert(1)",
	}, ownerToken)
	assert.Equal(t, http.StatusBadRequest, javascriptURLResponse.Code)

	searchURLResponse := performJSONRequest(r, http.MethodPost, "/socialmedia/create", map[string]interface{}{
		"name":             "TikTok search",
		"social_media_url": "https://www.tiktok.com/search?q=robby%20pantjoro",
	}, ownerToken)
	assert.Equal(t, http.StatusBadRequest, searchURLResponse.Code)

	genericURLResponse := performJSONRequest(r, http.MethodPost, "/socialmedia/create", map[string]interface{}{
		"name":             "Generic site",
		"social_media_url": "https://example.com",
	}, ownerToken)
	assert.Equal(t, http.StatusBadRequest, genericURLResponse.Code)

	listResponse := performJSONRequest(r, http.MethodGet, "/socialmedia/getall", nil, ownerToken)
	require.Equal(t, http.StatusOK, listResponse.Code)
	var socialMedias []models.SocialMedia
	require.NoError(t, json.Unmarshal(listResponse.Body.Bytes(), &socialMedias))
	require.Len(t, socialMedias, 1)

	otherUpdateResponse := performJSONRequest(r, http.MethodPut, fmt.Sprintf("/socialmedia/update/%d", socialMedia.ID), map[string]interface{}{
		"name":             "LinkedIn",
		"social_media_url": "https://linkedin.com/in/example",
	}, otherToken)
	assert.Equal(t, http.StatusForbidden, otherUpdateResponse.Code)

	ownerUpdateResponse := performJSONRequest(r, http.MethodPut, fmt.Sprintf("/socialmedia/update/%d", socialMedia.ID), map[string]interface{}{
		"name":             "LinkedIn",
		"social_media_url": "https://linkedin.com/in/example",
	}, ownerToken)
	require.Equal(t, http.StatusOK, ownerUpdateResponse.Code)
	assert.Contains(t, ownerUpdateResponse.Body.String(), "LinkedIn")

	ownerDeleteResponse := performJSONRequest(r, http.MethodDelete, fmt.Sprintf("/socialmedia/delete/%d", socialMedia.ID), nil, ownerToken)
	assert.Equal(t, http.StatusOK, ownerDeleteResponse.Code)

	emptyListResponse := performJSONRequest(r, http.MethodGet, "/socialmedia/getall", nil, ownerToken)
	assert.Equal(t, http.StatusOK, emptyListResponse.Code)
	assert.JSONEq(t, "[]", emptyListResponse.Body.String())
}

func TestReadinessWithDatabase(t *testing.T) {
	r := setupDatabaseBackedTest(t)

	response := performJSONRequest(r, http.MethodGet, "/health/ready", nil, "")
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Contains(t, response.Body.String(), "connected")
}

func setupDatabaseBackedTest(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	configureTestEnv(t)

	if database.GetDB() == nil {
		if err := database.StartDB(); err != nil {
			if strings.EqualFold(os.Getenv("REQUIRE_TEST_DATABASE"), "true") {
				t.Fatalf("required test database is unavailable: %v", err)
			}
			t.Skipf("test database is unavailable: %v", err)
		}
	}

	resetTestDatabase(t)
	return router.StartApp()
}

func configureTestEnv(t *testing.T) {
	t.Helper()
	setTestEnvDefault(t, "GIN_MODE", "test")
	setTestEnvDefault(t, "JWT_SECRET", "test-secret-that-is-long-enough-for-mygram")
	setTestEnvDefault(t, "JWT_EXPIRATION_HOURS", "24")
	setTestEnvDefault(t, "CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	setTestEnvDefault(t, "CAP_ENABLED", "false")
	setTestEnvDefault(t, "PUBLIC_OPENAPI_ENABLED", "true")
	setTestEnvDefault(t, "SWAGGER_UI_MODE", "internal")
	setTestEnvDefault(t, "S3_ENDPOINT", " ")
	setTestEnvDefault(t, "S3_REGION", "garage")
	setTestEnvDefault(t, "S3_BUCKET", " ")
	setTestEnvDefault(t, "S3_ACCESS_KEY_ID", " ")
	setTestEnvDefault(t, "S3_SECRET_ACCESS_KEY", " ")
	setTestEnvDefault(t, "S3_FORCE_PATH_STYLE", "true")
	setTestEnvDefault(t, "S3_UPLOAD_MAX_MB", "5")
	setTestEnvDefault(t, "PUSH_NOTIFICATIONS_ENABLED", "false")
	setTestEnvDefault(t, "VAPID_PUBLIC_KEY", "")
	setTestEnvDefault(t, "VAPID_PRIVATE_KEY", "")
	setTestEnvDefault(t, "VAPID_SUBJECT", "mailto:test@example.com")
	setTestEnvDefault(t, "DB_HOST", "localhost")
	setTestEnvDefault(t, "DB_USER", "postgres")
	setTestEnvDefault(t, "DB_PASSWORD", "admin")
	setTestEnvDefault(t, "DB_NAME", "finalproject_test")
	setTestEnvDefault(t, "DB_PORT", "5432")
	setTestEnvDefault(t, "DB_SSLMODE", "disable")
}

func setTestEnvDefault(t *testing.T, key string, value string) {
	t.Helper()
	if _, exists := os.LookupEnv(key); !exists {
		t.Setenv(key, value)
	}
}

func resetTestDatabase(t *testing.T) {
	t.Helper()
	db := database.GetDB()
	require.NotNil(t, db)

	require.NoError(t, db.Migrator().DropTable(
		&models.PushSubscription{},
		&models.Comment{},
		&models.Photo{},
		&models.SocialMedia{},
		&models.User{},
	))
	require.NoError(t, db.AutoMigrate(
		&models.User{},
		&models.Photo{},
		&models.Comment{},
		&models.SocialMedia{},
		&models.PushSubscription{},
	))
}

func createUserAndToken(t *testing.T, r http.Handler, username string, email string) (uint, string) {
	t.Helper()

	registerResponse := performJSONRequest(r, http.MethodPost, "/users/register", map[string]interface{}{
		"username": username,
		"email":    email,
		"password": "secret123",
		"age":      20,
	}, "")
	require.Equal(t, http.StatusCreated, registerResponse.Code, registerResponse.Body.String())

	var registeredUser struct {
		ID uint `json:"id"`
	}
	require.NoError(t, json.Unmarshal(registerResponse.Body.Bytes(), &registeredUser))
	require.NotZero(t, registeredUser.ID)

	loginResponse := performJSONRequest(r, http.MethodPost, "/users/login", map[string]interface{}{
		"email":    email,
		"password": "secret123",
	}, "")
	require.Equal(t, http.StatusOK, loginResponse.Code, loginResponse.Body.String())

	var loginBody struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(loginResponse.Body.Bytes(), &loginBody))
	require.NotEmpty(t, loginBody.Token)

	return registeredUser.ID, loginBody.Token
}

func createPhoto(t *testing.T, r http.Handler, token string) models.Photo {
	t.Helper()

	response := performJSONRequest(r, http.MethodPost, "/photos/create", map[string]interface{}{
		"title":     "Photo",
		"caption":   "Caption",
		"photo_url": "https://example.com/photo.jpg",
	}, token)
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())

	var photo models.Photo
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &photo))
	require.NotZero(t, photo.ID)

	return photo
}

func performJSONRequest(r http.Handler, method string, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		payload, _ := json.Marshal(body)
		reader = bytes.NewReader(payload)
	}

	request := httptest.NewRequest(method, path, reader)
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	return response
}

func performMultipartRequest(r http.Handler, path string, fieldName string, filename string, content []byte, token string) *httptest.ResponseRecorder {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile(fieldName, filename)
	_, _ = part.Write(content)
	_ = writer.Close()

	request := httptest.NewRequest(http.MethodPost, path, body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	return response
}

func tinyPNGBytes() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89,
	}
}
