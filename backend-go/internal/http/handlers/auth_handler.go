package handlers

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"xisu/backend-go/internal/database"
	"xisu/backend-go/internal/http/middleware"
	"xisu/backend-go/internal/http/response"
	"xisu/backend-go/internal/service"
)

type AuthHandler struct {
	db           *gorm.DB
	tokenService *service.TokenService
	loginMu      sync.Mutex
	loginAttempts map[string][]time.Time
}

func NewAuthHandler(db *gorm.DB, tokenService *service.TokenService) *AuthHandler {
	return &AuthHandler{db: db, tokenService: tokenService, loginAttempts: map[string][]time.Time{}}
}

const (
	maxLoginAttempts = 8
	loginWindow      = 10 * time.Minute
)

func (h *AuthHandler) loginAttemptKey(ip, identifier string) string {
	return ip + "|" + strings.ToLower(strings.TrimSpace(identifier))
}

func (h *AuthHandler) tooManyLoginAttempts(ip, identifier string) bool {
	now := time.Now()
	key := h.loginAttemptKey(ip, identifier)

	h.loginMu.Lock()
	defer h.loginMu.Unlock()

	history := h.loginAttempts[key]
	kept := make([]time.Time, 0, len(history)+1)
	for _, t := range history {
		if now.Sub(t) <= loginWindow {
			kept = append(kept, t)
		}
	}
	h.loginAttempts[key] = kept
	return len(kept) >= maxLoginAttempts
}

func (h *AuthHandler) recordLoginFailure(ip, identifier string) {
	now := time.Now()
	key := h.loginAttemptKey(ip, identifier)

	h.loginMu.Lock()
	defer h.loginMu.Unlock()

	history := h.loginAttempts[key]
	kept := make([]time.Time, 0, len(history)+1)
	for _, t := range history {
		if now.Sub(t) <= loginWindow {
			kept = append(kept, t)
		}
	}
	h.loginAttempts[key] = append(kept, now)
}

func (h *AuthHandler) resetLoginAttempts(ip, identifier string) {
	key := h.loginAttemptKey(ip, identifier)
	h.loginMu.Lock()
	delete(h.loginAttempts, key)
	h.loginMu.Unlock()
}

type LoginRequest struct {
	Identifier string `json:"identifier"`
	StudentID  string `json:"studentId"`
	Password   string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid login payload")
		return
	}

	identifier := strings.TrimSpace(req.Identifier)
	if identifier == "" {
		identifier = strings.TrimSpace(req.StudentID)
	}
	if identifier == "" {
		response.Error(c, http.StatusBadRequest, "identifier is required")
		return
	}
	ip := c.ClientIP()
	if h.tooManyLoginAttempts(ip, identifier) {
		response.Error(c, http.StatusTooManyRequests, "登录失败次数过多，请稍后重试")
		return
	}

	var user database.User
	err := h.db.Where("username = ? OR email = ? OR phone = ? OR student_id = ?", identifier, identifier, identifier, identifier).First(&user).Error
	if err != nil {
		h.recordLoginFailure(ip, identifier)
		response.Error(c, http.StatusUnauthorized, "username or password incorrect")
		return
	}

	if user.Status != "ACTIVE" {
		h.recordLoginFailure(ip, identifier)
		response.Error(c, http.StatusForbidden, "user is not active")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		h.recordLoginFailure(ip, identifier)
		response.Error(c, http.StatusUnauthorized, "username or password incorrect")
		return
	}
	h.resetLoginAttempts(ip, identifier)

	accessToken, accessExpiresAt, err := h.tokenService.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to generate access token")
		return
	}

	refreshToken, refreshExpiresAt, err := h.tokenService.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to generate refresh token")
		return
	}
	refreshDigest := refreshTokenDigest(refreshToken)

	ua := c.GetHeader("User-Agent")

	refreshRow := database.RefreshToken{
		ID:        uuid.NewString(),
		Token:     refreshDigest,
		UserID:    user.ID,
		UserAgent: &ua,
		IPAddress: &ip,
		ExpiresAt: refreshExpiresAt,
	}

	now := time.Now()
	_ = h.db.Model(&database.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"last_login_at": now,
		"last_login_ip": ip,
	}).Error

	if err := h.db.Create(&refreshRow).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to persist refresh token")
		return
	}

	response.OK(c, gin.H{
		"accessToken":      accessToken,
		"refreshToken":     refreshToken,
		"accessExpiresAt":  accessExpiresAt,
		"refreshExpiresAt": refreshExpiresAt,
		"user": gin.H{
			"id":        user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"phone":     user.Phone,
			"studentId": user.StudentID,
			"realName":  user.RealName,
			"nickname":  user.Nickname,
			"avatar":    user.Avatar,
			"role":      user.Role,
			"college":   user.College,
			"major":     user.Major,
			"className": user.ClassName,
			"jwxtBound": user.JWXTUsername != nil && user.JWXTPassword != nil,
		},
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid refresh payload")
		return
	}

	claims, err := h.tokenService.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	var row database.RefreshToken
	digest := refreshTokenDigest(req.RefreshToken)
	if err := h.db.Where("token = ? OR token = ?", digest, req.RefreshToken).First(&row).Error; err != nil {
		response.Error(c, http.StatusUnauthorized, "refresh token not found")
		return
	}
	if row.ExpiresAt.Before(time.Now()) {
		response.Error(c, http.StatusUnauthorized, "refresh token expired")
		return
	}

	var user database.User
	if err := h.db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		response.Error(c, http.StatusUnauthorized, "user not found")
		return
	}

	accessToken, accessExpiresAt, err := h.tokenService.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to generate access token")
		return
	}

	newRefreshToken, refreshExpiresAt, err := h.tokenService.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to generate refresh token")
		return
	}

	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&database.RefreshToken{}, "token = ? OR token = ?", digest, req.RefreshToken).Error; err != nil {
			return err
		}
		newDigest := refreshTokenDigest(newRefreshToken)

		newRow := database.RefreshToken{
			ID:        uuid.NewString(),
			Token:     newDigest,
			UserID:    user.ID,
			UserAgent: &ua,
			IPAddress: &ip,
			ExpiresAt: refreshExpiresAt,
		}
		return tx.Create(&newRow).Error
	}); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to rotate refresh token")
		return
	}

	response.OK(c, gin.H{
		"accessToken":      accessToken,
		"refreshToken":     newRefreshToken,
		"accessExpiresAt":  accessExpiresAt,
		"refreshExpiresAt": refreshExpiresAt,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextUserID)
	if userID == nil {
		response.Error(c, http.StatusUnauthorized, "invalid user context")
		return
	}

	if err := h.db.Delete(&database.RefreshToken{}, "user_id = ?", userID.(string)).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "logout failed")
		return
	}

	response.OK(c, gin.H{"success": true})
}
