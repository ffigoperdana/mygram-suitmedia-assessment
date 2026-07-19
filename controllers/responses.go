package controllers

import (
	"finalproject/database"
	"finalproject/helpers"
	"finalproject/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type TokenResponse struct {
	Token string             `json:"token"`
	User  PublicUserResponse `json:"user"`
}

type RegisterResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

type DeleteResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type UploadPhotoResponse struct {
	URL         string `json:"url"`
	Key         string `json:"key"`
	Bucket      string `json:"bucket"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

type PublicUserResponse struct {
	ID          uint       `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Age         int        `json:"age"`
	Role        string     `json:"role"`
	Status      string     `json:"status"`
	BannedAt    *time.Time `json:"banned_at,omitempty"`
	BanReason   string     `json:"ban_reason,omitempty"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	LastSeenAt  *time.Time `json:"last_seen_at,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type CurrentUserResponse struct {
	User PublicUserResponse `json:"user"`
}

type UsersListResponse struct {
	Users []PublicUserResponse `json:"users"`
	Total int64                `json:"total"`
	Page  int                  `json:"page"`
	Limit int                  `json:"limit"`
}

type AdminStatsResponse struct {
	TotalUsers       int64                `json:"total_users"`
	ActiveUsers      int64                `json:"active_users"`
	BannedUsers      int64                `json:"banned_users"`
	AdminUsers       int64                `json:"admin_users"`
	UsersSeenLast24h int64                `json:"users_seen_last_24h"`
	TotalPhotos      int64                `json:"total_photos"`
	TotalComments    int64                `json:"total_comments"`
	TotalSocialMedia int64                `json:"total_social_media"`
	RecentUsers      []PublicUserResponse `json:"recent_users"`
	GeneratedAt      time.Time            `json:"generated_at"`
}

type AdminUserUpdateRequest struct {
	Username  *string `json:"username" form:"username"`
	Email     *string `json:"email" form:"email"`
	Age       *int    `json:"age" form:"age"`
	Role      *string `json:"role" form:"role"`
	Status    *string `json:"status" form:"status"`
	BanReason *string `json:"ban_reason" form:"ban_reason"`
}

type BanUserRequest struct {
	Reason string `json:"reason" form:"reason"`
}

type AdminDeleteUserResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	UserID  uint   `json:"user_id"`
}

func toPublicUserResponse(user models.User) PublicUserResponse {
	return PublicUserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Age:         user.Age,
		Role:        user.Role,
		Status:      user.Status,
		BannedAt:    user.BannedAt,
		BanReason:   user.BanReason,
		LastLoginAt: user.LastLoginAt,
		LastSeenAt:  user.LastSeenAt,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

func toPublicUserResponses(users []models.User) []PublicUserResponse {
	responses := make([]PublicUserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, toPublicUserResponse(user))
	}

	return responses
}

func jsonError(c *gin.Context, status int, label string, message string) {
	c.JSON(status, gin.H{
		"error":   label,
		"message": message,
	})
}

func requireDB(c *gin.Context) (*gorm.DB, bool) {
	db := database.GetDB()
	if db == nil {
		jsonError(c, http.StatusServiceUnavailable, "Service Unavailable", "database is not ready")
		return nil, false
	}

	return db, true
}

func requireClaims(c *gin.Context) (*helpers.Claims, bool) {
	claims, ok := helpers.GetUserClaims(c)
	if !ok {
		jsonError(c, http.StatusUnauthorized, "Unauthenticated", "sign in to proceed")
		return nil, false
	}

	return claims, true
}
