package controllers

import (
	"errors"
	"finalproject/helpers"
	"finalproject/models"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AdminStats godoc
// @Summary Get admin dashboard stats
// @Description Get basic user and content metrics for the admin dashboard
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} AdminStatsResponse "Admin stats response"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /api/v1/admin/stats [get]
func AdminStats(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	now := time.Now().UTC()
	stats := AdminStatsResponse{GeneratedAt: now}
	since := now.Add(-24 * time.Hour)

	counts := []struct {
		target *int64
		model  interface{}
		query  string
		args   []interface{}
	}{
		{target: &stats.TotalUsers, model: &models.User{}},
		{target: &stats.ActiveUsers, model: &models.User{}, query: "status = ?", args: []interface{}{models.UserStatusActive}},
		{target: &stats.BannedUsers, model: &models.User{}, query: "status = ?", args: []interface{}{models.UserStatusBanned}},
		{target: &stats.AdminUsers, model: &models.User{}, query: "role = ?", args: []interface{}{models.RoleAdmin}},
		{target: &stats.UsersSeenLast24h, model: &models.User{}, query: "last_seen_at >= ?", args: []interface{}{since}},
		{target: &stats.TotalPhotos, model: &models.Photo{}},
		{target: &stats.TotalComments, model: &models.Comment{}},
		{target: &stats.TotalSocialMedia, model: &models.SocialMedia{}},
	}

	for _, count := range counts {
		query := db.Model(count.model)
		if count.query != "" {
			query = query.Where(count.query, count.args...)
		}
		if err := query.Count(count.target).Error; err != nil {
			jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to load admin stats")
			return
		}
	}

	recentUsers := []models.User{}
	if err := db.Order("id DESC").Limit(5).Find(&recentUsers).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to load recent users")
		return
	}
	stats.RecentUsers = toPublicUserResponses(recentUsers)

	c.JSON(http.StatusOK, stats)
}

// AdminListUsers godoc
// @Summary List users
// @Description List users for admin management
// @Tags admin
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Param search query string false "Search username or email"
// @Param role query string false "Filter by role"
// @Param status query string false "Filter by status"
// @Security BearerAuth
// @Success 200 {object} UsersListResponse "Users list response"
// @Failure 400 {object} ErrorResponse "Invalid query"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /api/v1/admin/users [get]
func AdminListUsers(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	page := queryInt(c, "page", 1, 1, 100000)
	limit := queryInt(c, "limit", 20, 1, 100)
	offset := (page - 1) * limit

	query := db.Model(&models.User{})
	search := strings.TrimSpace(c.Query("search"))
	if search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(username) LIKE ? OR LOWER(email) LIKE ?", like, like)
	}

	role := strings.TrimSpace(c.Query("role"))
	if role != "" {
		if !models.IsValidUserRole(role) {
			jsonError(c, http.StatusBadRequest, "Bad Request", "invalid role filter")
			return
		}
		query = query.Where("role = ?", role)
	}

	status := strings.TrimSpace(c.Query("status"))
	if status != "" {
		if !models.IsValidUserStatus(status) {
			jsonError(c, http.StatusBadRequest, "Bad Request", "invalid status filter")
			return
		}
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to count users")
		return
	}

	users := []models.User{}
	if err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to load users")
		return
	}

	c.JSON(http.StatusOK, UsersListResponse{
		Users: toPublicUserResponses(users),
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// AdminGetUser godoc
// @Summary Get user
// @Description Get a user by ID for admin management
// @Tags admin
// @Accept json
// @Produce json
// @Param userId path int true "ID of the user"
// @Security BearerAuth
// @Success 200 {object} PublicUserResponse "User response"
// @Failure 400 {object} ErrorResponse "Invalid user id"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /api/v1/admin/users/{userId} [get]
func AdminGetUser(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	userID, err := parseUserID(c)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", "invalid user id")
		return
	}

	user, ok := findUserByID(c, db, userID)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, toPublicUserResponse(user))
}

// AdminUpdateUser godoc
// @Summary Update user
// @Description Update a user profile, role, or status as an admin
// @Tags admin
// @Accept json
// @Produce json
// @Param userId path int true "ID of the user"
// @Param request body AdminUserUpdateRequest true "User update request body"
// @Security BearerAuth
// @Success 200 {object} PublicUserResponse "Updated user response"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /api/v1/admin/users/{userId} [patch]
func AdminUpdateUser(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	claims, ok := requireClaims(c)
	if !ok {
		return
	}

	userID, err := parseUserID(c)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", "invalid user id")
		return
	}

	request := AdminUserUpdateRequest{}
	if err := bindOptionalAdminRequest(c, &request); err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	updates, err := buildAdminUserUpdates(request)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	if len(updates) == 0 {
		jsonError(c, http.StatusBadRequest, "Bad Request", "at least one field is required")
		return
	}

	if userID == claims.ID {
		if role, ok := updates["role"].(string); ok && role != models.RoleAdmin {
			jsonError(c, http.StatusForbidden, "Forbidden", "admin cannot remove their own admin role")
			return
		}
		if status, ok := updates["status"].(string); ok && status == models.UserStatusBanned {
			jsonError(c, http.StatusForbidden, "Forbidden", "admin cannot ban their own account")
			return
		}
	}

	if _, ok := findUserByID(c, db, userID); !ok {
		return
	}

	if status, ok := updates["status"].(string); ok && status == models.UserStatusBanned {
		now := time.Now().UTC()
		updates["banned_at"] = &now
	} else if status, ok := updates["status"].(string); ok && status == models.UserStatusActive {
		updates["banned_at"] = nil
		if _, exists := updates["ban_reason"]; !exists {
			updates["ban_reason"] = ""
		}
	}

	if err := db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	user, ok := findUserByID(c, db, userID)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, toPublicUserResponse(user))
}

// AdminBanUser godoc
// @Summary Ban user
// @Description Ban a user account
// @Tags admin
// @Accept json
// @Produce json
// @Param userId path int true "ID of the user"
// @Param request body BanUserRequest false "Ban reason"
// @Security BearerAuth
// @Success 200 {object} PublicUserResponse "Banned user response"
// @Failure 400 {object} ErrorResponse "Invalid user id"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /api/v1/admin/users/{userId}/ban [post]
func AdminBanUser(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	claims, ok := requireClaims(c)
	if !ok {
		return
	}

	userID, err := parseUserID(c)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", "invalid user id")
		return
	}
	if userID == claims.ID {
		jsonError(c, http.StatusForbidden, "Forbidden", "admin cannot ban their own account")
		return
	}

	request := BanUserRequest{}
	if err := bindOptionalAdminRequest(c, &request); err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	if _, ok := findUserByID(c, db, userID); !ok {
		return
	}

	now := time.Now().UTC()
	if err := db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"status":     models.UserStatusBanned,
		"banned_at":  &now,
		"ban_reason": strings.TrimSpace(request.Reason),
	}).Error; err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	user, ok := findUserByID(c, db, userID)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, toPublicUserResponse(user))
}

// AdminUnbanUser godoc
// @Summary Unban user
// @Description Restore a banned user account
// @Tags admin
// @Accept json
// @Produce json
// @Param userId path int true "ID of the user"
// @Security BearerAuth
// @Success 200 {object} PublicUserResponse "Unbanned user response"
// @Failure 400 {object} ErrorResponse "Invalid user id"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /api/v1/admin/users/{userId}/unban [post]
func AdminUnbanUser(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	userID, err := parseUserID(c)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", "invalid user id")
		return
	}

	if _, ok := findUserByID(c, db, userID); !ok {
		return
	}

	if err := db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"status":     models.UserStatusActive,
		"banned_at":  nil,
		"ban_reason": "",
	}).Error; err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	user, ok := findUserByID(c, db, userID)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, toPublicUserResponse(user))
}

// AdminDeleteUser godoc
// @Summary Delete user
// @Description Delete a user account as an admin
// @Tags admin
// @Accept json
// @Produce json
// @Param userId path int true "ID of the user"
// @Security BearerAuth
// @Success 200 {object} AdminDeleteUserResponse "Delete user response"
// @Failure 400 {object} ErrorResponse "Invalid user id"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 503 {object} ErrorResponse "Database is not ready"
// @Router /api/v1/admin/users/{userId} [delete]
func AdminDeleteUser(c *gin.Context) {
	db, ok := requireDB(c)
	if !ok {
		return
	}

	claims, ok := requireClaims(c)
	if !ok {
		return
	}

	userID, err := parseUserID(c)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", "invalid user id")
		return
	}
	if userID == claims.ID {
		jsonError(c, http.StatusForbidden, "Forbidden", "admin cannot delete their own account")
		return
	}

	if _, ok := findUserByID(c, db, userID); !ok {
		return
	}

	if err := db.Delete(&models.User{}, userID).Error; err != nil {
		jsonError(c, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	c.JSON(http.StatusOK, AdminDeleteUserResponse{
		Status:  "delete_success",
		Message: "the user has been successfully deleted",
		UserID:  userID,
	})
}

func parseUserID(c *gin.Context) (uint, error) {
	value, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil || value == 0 {
		return 0, errors.New("invalid user id")
	}

	return uint(value), nil
}

func findUserByID(c *gin.Context, db *gorm.DB, userID uint) (models.User, bool) {
	user := models.User{}
	if err := db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			jsonError(c, http.StatusNotFound, "User Not Found", "user does not exist")
			return models.User{}, false
		}

		jsonError(c, http.StatusInternalServerError, "Internal Server Error", "failed to load user")
		return models.User{}, false
	}

	return user, true
}

func buildAdminUserUpdates(request AdminUserUpdateRequest) (map[string]interface{}, error) {
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
		if _, err := mail.ParseAddress(email); err != nil {
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

	if request.Role != nil {
		role := strings.TrimSpace(*request.Role)
		if !models.IsValidUserRole(role) {
			return nil, errors.New("invalid role")
		}
		updates["role"] = role
	}

	if request.Status != nil {
		status := strings.TrimSpace(*request.Status)
		if !models.IsValidUserStatus(status) {
			return nil, errors.New("invalid status")
		}
		updates["status"] = status
	}

	if request.BanReason != nil {
		updates["ban_reason"] = strings.TrimSpace(*request.BanReason)
	}

	return updates, nil
}

func queryInt(c *gin.Context, key string, fallback int, min int, max int) int {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	if parsed < min {
		return min
	}
	if parsed > max {
		return max
	}

	return parsed
}

func bindOptionalAdminRequest(c *gin.Context, destination interface{}) error {
	if c.Request.Body == nil || c.Request.ContentLength == 0 {
		return nil
	}

	return helpers.BindRequest(c, destination)
}
