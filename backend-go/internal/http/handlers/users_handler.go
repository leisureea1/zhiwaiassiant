package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"xisu/backend-go/internal/database"
	"xisu/backend-go/internal/http/middleware"
	"xisu/backend-go/internal/http/response"
)

type UsersHandler struct {
	db *gorm.DB
}

type userSafeResponse struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	Email        *string    `json:"email"`
	Phone        *string    `json:"phone"`
	StudentID    *string    `json:"studentId"`
	RealName     *string    `json:"realName"`
	Nickname     *string    `json:"nickname"`
	Avatar       *string    `json:"avatar"`
	College      *string    `json:"college"`
	Major        *string    `json:"major"`
	ClassName    *string    `json:"className"`
	Role         string     `json:"role"`
	Status       string     `json:"status"`
	JWXTUsername *string    `json:"jwxtUsername,omitempty"`
	JWXTBound    bool       `json:"jwxtBound"`
	LastLoginAt  *time.Time `json:"lastLoginAt"`
	LastLoginIP  *string    `json:"lastLoginIp"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

func toSafeUser(user database.User, includeJWXT bool) userSafeResponse {
	out := userSafeResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Phone:       user.Phone,
		StudentID:   user.StudentID,
		RealName:    user.RealName,
		Nickname:    user.Nickname,
		Avatar:      user.Avatar,
		College:     user.College,
		Major:       user.Major,
		ClassName:   user.ClassName,
		Role:        user.Role,
		Status:      user.Status,
		JWXTBound:   user.JWXTUsername != nil && user.JWXTPassword != nil,
		LastLoginAt: user.LastLoginAt,
		LastLoginIP: user.LastLoginIP,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
	if includeJWXT {
		out.JWXTUsername = user.JWXTUsername
	}
	return out
}

func NewUsersHandler(db *gorm.DB) *UsersHandler {
	return &UsersHandler{db: db}
}

func (h *UsersHandler) Me(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextUserID)
	if userID == nil {
		response.Error(c, http.StatusUnauthorized, "invalid user context")
		return
	}

	var user database.User
	if err := h.db.Where("id = ?", userID.(string)).First(&user).Error; err != nil {
		response.Error(c, http.StatusNotFound, "user not found")
		return
	}

	response.OK(c, toSafeUser(user, true))
}

func (h *UsersHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	currentUserID, _ := c.Get(middleware.ContextUserID)
	currentUserRole, _ := c.Get(middleware.ContextUserRole)

	// 仅允许本人或管理员查看
	userID, _ := currentUserID.(string)
	role, _ := currentUserRole.(string)
	if userID != id && role != "ADMIN" && role != "SUPER_ADMIN" {
		response.Error(c, http.StatusForbidden, "no permission to view other users")
		return
	}

	var user database.User
	if err := h.db.Where("id = ?", id).First(&user).Error; err != nil {
		response.Error(c, http.StatusNotFound, "user not found")
		return
	}

	response.OK(c, toSafeUser(user, false))
}

type UserListQuery struct {
	Page     int
	PageSize int
	Role     string
	Status   string
	Search   string
}

func (h *UsersHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	query := UserListQuery{
		Page:     page,
		PageSize: pageSize,
		Role:     c.Query("role"),
		Status:   c.Query("status"),
		Search:   c.Query("search"),
	}

	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}

	db := h.db.Model(&database.User{})

	if query.Role != "" {
		db = db.Where("role = ?", query.Role)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Search != "" {
		db = db.Where("username LIKE ? OR real_name LIKE ? OR student_id LIKE ?",
			"%"+query.Search+"%", "%"+query.Search+"%", "%"+query.Search+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to count users")
		return
	}

	var users []database.User
	offset := (query.Page - 1) * query.PageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&users).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get users")
		return
	}

	userList := make([]userSafeResponse, 0, len(users))
	for _, u := range users {
		userList = append(userList, toSafeUser(u, false))
	}

	response.OK(c, gin.H{
		"total": total,
		"page":  query.Page,
		"size":  query.PageSize,
		"data":  userList,
	})
}

type UpdateUserRequest struct {
	RealName  *string `json:"realName"`
	Nickname  *string `json:"nickname"`
	Avatar    *string `json:"avatar"`
	College   *string `json:"college"`
	Major     *string `json:"major"`
	ClassName *string `json:"className"`
}

func (h *UsersHandler) Update(c *gin.Context) {
	id := c.Param("id")
	currentUserID, _ := c.Get(middleware.ContextUserID)
	currentUserRole, _ := c.Get(middleware.ContextUserRole)

	userID, _ := currentUserID.(string)
	role, _ := currentUserRole.(string)

	// 仅允许本人更新，或管理员更新他人
	if userID != id && role != "ADMIN" && role != "SUPER_ADMIN" {
		response.Error(c, http.StatusForbidden, "no permission to update other users")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request")
		return
	}

	updates := make(map[string]interface{})
	if req.RealName != nil {
		updates["real_name"] = *req.RealName
	}
	if req.Nickname != nil {
		updates["nickname"] = *req.Nickname
	}
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}
	if req.College != nil {
		updates["college"] = *req.College
	}
	if req.Major != nil {
		updates["major"] = *req.Major
	}
	if req.ClassName != nil {
		updates["class_name"] = *req.ClassName
	}

	if err := h.db.Model(&database.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to update user")
		return
	}

	var user database.User
	h.db.Where("id = ?", id).First(&user)
	response.OK(c, toSafeUser(user, false))
}

type AdminUpdateUserRequest struct {
	Role   *string `json:"role"`
	Status *string `json:"status"`
}

func (h *UsersHandler) AdminUpdate(c *gin.Context) {
	id := c.Param("id")
	var req AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request")
		return
	}

	updates := make(map[string]interface{})
	if req.Role != nil {
		allowedRoles := map[string]bool{"USER": true, "ADMIN": true, "SUPER_ADMIN": true}
		if !allowedRoles[*req.Role] {
			response.Error(c, http.StatusBadRequest, "invalid role")
			return
		}
		// 只有 SUPER_ADMIN 才能设置 SUPER_ADMIN 角色
		if *req.Role == "SUPER_ADMIN" {
			currentRole, exists := c.Get(middleware.ContextUserRole)
			if !exists || currentRole.(string) != "SUPER_ADMIN" {
				response.Error(c, http.StatusForbidden, "只有超级管理员才能设置超级管理员角色")
				return
			}
		}
		updates["role"] = *req.Role
	}
	if req.Status != nil {
		allowedStatus := map[string]bool{"ACTIVE": true, "INACTIVE": true, "BANNED": true}
		if !allowedStatus[*req.Status] {
			response.Error(c, http.StatusBadRequest, "invalid status")
			return
		}
		updates["status"] = *req.Status
	}

	if err := h.db.Model(&database.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to update user")
		return
	}

	var user database.User
	h.db.Where("id = ?", id).First(&user)
	response.OK(c, toSafeUser(user, false))
}

func (h *UsersHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.db.Delete(&database.User{}, "id = ?", id).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to delete user")
		return
	}
	response.OK(c, gin.H{"success": true})
}

func (h *UsersHandler) GetNotificationSettings(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextUserID)
	var settings database.NotificationSetting
	if err := h.db.Where("user_id = ?", userID.(string)).First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.OK(c, gin.H{
				"emailEnabled":       true,
				"pushEnabled":        true,
				"gradeNotify":        true,
				"examNotify":         true,
				"announcementNotify": true,
			})
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get notification settings")
		return
	}
	response.OK(c, settings)
}

type UpdateNotificationSettingsRequest struct {
	EmailEnabled       *bool   `json:"emailEnabled"`
	PushEnabled        *bool   `json:"pushEnabled"`
	BarkKey            *string `json:"barkKey"`
	GradeNotify        *bool   `json:"gradeNotify"`
	ExamNotify         *bool   `json:"examNotify"`
	AnnouncementNotify *bool   `json:"announcementNotify"`
}

func (h *UsersHandler) UpdateNotificationSettings(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextUserID)
	var req UpdateNotificationSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request")
		return
	}

	updates := make(map[string]interface{})
	if req.EmailEnabled != nil {
		updates["email_enabled"] = *req.EmailEnabled
	}
	if req.PushEnabled != nil {
		updates["push_enabled"] = *req.PushEnabled
	}
	if req.BarkKey != nil {
		updates["bark_key"] = *req.BarkKey
	}
	if req.GradeNotify != nil {
		updates["grade_notify"] = *req.GradeNotify
	}
	if req.ExamNotify != nil {
		updates["exam_notify"] = *req.ExamNotify
	}
	if req.AnnouncementNotify != nil {
		updates["announcement_notify"] = *req.AnnouncementNotify
	}

	if err := h.db.Model(&database.NotificationSetting{}).Where("user_id = ?", userID.(string)).Updates(updates).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "notification settings not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to update notification settings")
		return
	}

	var settings database.NotificationSetting
	h.db.Where("user_id = ?", userID.(string)).First(&settings)
	response.OK(c, settings)
}
