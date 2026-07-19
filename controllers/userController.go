package controllers

import (
	"errors"
	"finalproject/config"
	"finalproject/database"
	"finalproject/helpers"
	"finalproject/models"
	"net/http"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LoginRequest struct {
	Email        string `json:"email" form:"email"`
	Password     string `json:"password" form:"password"`
	CaptchaToken string `json:"captcha_token,omitempty" form:"captcha_token"`
}

type RegisterRequest struct {
	Username     string `json:"username" form:"username"`
	Email        string `json:"email" form:"email"`
	Password     string `json:"password" form:"password"`
	Age          int    `json:"age" form:"age"`
	CaptchaToken string `json:"captcha_token,omitempty" form:"captcha_token"`
}

type ProfileUpdateRequest struct {
	Username *string `json:"username" form:"username"`
	Email    *string `json:"email" form:"email"`
	Age      *int    `json:"age" form:"age"`
}

// UserRegister godoc
// @Summary Register user
// @Description Register new user
// @Tags user
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register request body"
// @Success 201 {object} RegisterResponse "Register success response"
// @Failure 400 {object} ErrorResponse "Invalid request or duplicate data"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /users/register [post]
// @Router /api/v1/auth/register [post]
func UserRegister(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Service Unavailable",
			"message": "database is not ready",
		})
		return
	}

	cfg := config.Load()
	request := RegisterRequest{}
	if err := helpers.BindRequest(c, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": err.Error(),
		})
		return
	}

	if err := helpers.VerifyCaptchaToken(cfg, request.CaptchaToken); err != nil {
		handleCaptchaError(c, err)
		return
	}

	user := models.User{
		Username: request.Username,
		Email:    request.Email,
		Password: request.Password,
		Age:      request.Age,
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"age":      user.Age,
		"role":     user.Role,
		"status":   user.Status,
	})
}

// UserLogin godoc
// @Summary Login user
// @Description Login user by email
// @Tags user
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request body"
// @Success 200 {object} TokenResponse "Login response"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Invalid email/password"
// @Failure 500 {object} ErrorResponse "Token generation or database error"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /users/login [post]
// @Router /api/v1/auth/login [post]
func UserLogin(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Service Unavailable",
			"message": "database is not ready",
		})
		return
	}

	cfg := config.Load()
	request := LoginRequest{}
	if err := helpers.BindRequest(c, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": err.Error(),
		})
		return
	}

	if request.Email == "" || request.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "email and password are required",
		})
		return
	}

	if cfg.CapEnabled && cfg.CapRequiredOnLogin {
		if err := helpers.VerifyCaptchaToken(cfg, request.CaptchaToken); err != nil {
			handleCaptchaError(c, err)
			return
		}
	}

	user := models.User{}
	if err := db.Where("email = ?", request.Email).Take(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "invalid email/password",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "failed to load user",
		})
		return
	}

	if user.Status == models.UserStatusBanned {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "user account is banned",
		})
		return
	}

	if !helpers.ComparePass([]byte(user.Password), []byte(request.Password)) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "invalid email/password",
		})
		return
	}

	now := time.Now().UTC()
	if err := db.Model(&user).Updates(map[string]interface{}{
		"last_login_at": &now,
		"last_seen_at":  &now,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "failed to update login state",
		})
		return
	}
	user.LastLoginAt = &now
	user.LastSeenAt = &now

	token, err := helpers.GenerateToken(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  toPublicUserResponse(user),
	})
}

// GetMe godoc
// @Summary Get current user
// @Description Get the currently authenticated user's profile
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} CurrentUserResponse "Current user response"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /api/v1/me [get]
func GetMe(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	claims, ok := requireClaims(c)
	if !ok {
		return
	}

	user := models.User{}
	if err := db.First(&user, claims.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			jsonError(c, http.StatusNotFound, "User Not Found", "user does not exist")
			return
		}

		jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to load user")
		return
	}

	c.JSON(http.StatusOK, CurrentUserResponse{User: toPublicUserResponse(user)})
}

// UpdateMe godoc
// @Summary Update current user
// @Description Update the currently authenticated user's profile. Role, status, and ban fields are admin-only and cannot be changed here.
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ProfileUpdateRequest true "Profile update request body"
// @Success 200 {object} CurrentUserResponse "Updated current user response"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /api/v1/me [patch]
func UpdateMe(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	claims, ok := requireClaims(c)
	if !ok {
		return
	}

	request := ProfileUpdateRequest{}
	if err := helpers.BindRequest(c, &request); err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	updates, err := buildProfileUpdates(request)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	user := models.User{}
	if err := db.First(&user, claims.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			jsonError(c, http.StatusNotFound, "User Not Found", "user does not exist")
			return
		}

		jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to load user")
		return
	}

	if len(updates) > 0 {
		if err := db.Model(&models.User{}).Where("id = ?", claims.ID).Updates(updates).Error; err != nil {
			jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
			return
		}
	}

	if err := db.First(&user, claims.ID).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to load updated user")
		return
	}

	c.JSON(http.StatusOK, CurrentUserResponse{User: toPublicUserResponse(user)})
}

func buildProfileUpdates(request ProfileUpdateRequest) (map[string]interface{}, error) {
	updates := map[string]interface{}{}

	if request.Username != nil {
		username := strings.TrimSpace(*request.Username)
		if username == "" {
			return nil, errors.New("username cannot be empty")
		}
		updates["username"] = username
	}

	if request.Email != nil {
		email := strings.TrimSpace(*request.Email)
		if email == "" {
			return nil, errors.New("email cannot be empty")
		}
		if !govalidator.IsEmail(email) {
			return nil, errors.New("invalid email format")
		}
		updates["email"] = email
	}

	if request.Age != nil {
		if *request.Age < 9 || *request.Age > 100 {
			return nil, errors.New("age has to be between 9 and 100")
		}
		updates["age"] = *request.Age
	}

	return updates, nil
}

func handleCaptchaError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, helpers.ErrCaptchaRequired), errors.Is(err, helpers.ErrCaptchaFailed):
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
	case errors.Is(err, helpers.ErrCaptchaNotConfigured):
		jsonError(c, http.StatusInternalServerError, "Internal Server Error", err.Error())
	case errors.Is(err, helpers.ErrCaptchaUnavailable):
		jsonError(c, http.StatusServiceUnavailable, "Service Unavailable", err.Error())
	default:
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
	}
}
