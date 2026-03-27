package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"xisu/backend-go/internal/database"
	"xisu/backend-go/internal/http/middleware"
	"xisu/backend-go/internal/http/response"
	"xisu/backend-go/internal/service"
)

type GradeSubscriptionHandler struct {
	db      *gorm.DB
	subSvc  *service.GradeSubscriptionService
}

func NewGradeSubscriptionHandler(db *gorm.DB, subSvc *service.GradeSubscriptionService) *GradeSubscriptionHandler {
	return &GradeSubscriptionHandler{db: db, subSvc: subSvc}
}

func (h *GradeSubscriptionHandler) GetSubscription(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextUserID)
	if userID == nil {
		response.Error(c, http.StatusUnauthorized, "invalid user context")
		return
	}

	var sub database.GradeSubscription
	if err := h.db.Where("user_id = ?", userID.(string)).First(&sub).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.OK(c, gin.H{
				"enabled": false,
			})
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get subscription")
		return
	}

	response.OK(c, gin.H{
		"id":             sub.ID,
		"enabled":        sub.Enabled,
		"lastCheckedAt":  sub.LastCheckedAt,
		"lastNotifiedAt": sub.LastNotifiedAt,
		"totalNotified":  sub.TotalNotified,
		"semesterId":     sub.SemesterID,
		"createdAt":      sub.CreatedAt,
	})
}

type UpdateGradeSubscriptionRequest struct {
	Enabled *bool `json:"enabled"`
}

func (h *GradeSubscriptionHandler) UpdateSubscription(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextUserID)
	if userID == nil {
		response.Error(c, http.StatusUnauthorized, "invalid user context")
		return
	}

	var req UpdateGradeSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Enabled == nil {
		response.Error(c, http.StatusBadRequest, "invalid request")
		return
	}

	uid := userID.(string)

	// Check if user has JWXT bound
	var user database.User
	if err := h.db.Where("id = ?", uid).First(&user).Error; err != nil {
		response.Error(c, http.StatusNotFound, "user not found")
		return
	}
	if user.JWXTUsername == nil || user.JWXTPassword == nil {
		response.Error(c, http.StatusBadRequest, "请先绑定教务系统账号")
		return
	}

	// Check if user has email
	if user.Email == nil || *user.Email == "" {
		response.Error(c, http.StatusBadRequest, "请先设置邮箱，成绩变化通知将通过邮箱发送")
		return
	}

	var sub database.GradeSubscription
	err := h.db.Where("user_id = ?", uid).First(&sub).Error
	if err == gorm.ErrRecordNotFound {
		sub = database.GradeSubscription{
			ID:        uuid.NewString(),
			UserID:    uid,
			Enabled:   *req.Enabled,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := h.db.Create(&sub).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, "failed to create subscription")
			return
		}
	} else if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get subscription")
		return
	} else {
		if err := h.db.Model(&sub).Updates(map[string]any{
			"enabled":    *req.Enabled,
			"updated_at": time.Now(),
		}).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, "failed to update subscription")
			return
		}
	}

	status := "已开启"
	if !*req.Enabled {
		status = "已关闭"
	}

	response.OK(c, gin.H{
		"enabled": *req.Enabled,
		"message": "成绩订阅" + status + "，系统将每小时检查一次成绩变化并通过邮箱通知您",
	})
}

// TriggerCheck manually triggers a grade check cycle (admin only).
func (h *GradeSubscriptionHandler) TriggerCheck(c *gin.Context) {
	log.Printf("[GradeSubscription] Manual trigger received")
	if h.subSvc == nil {
		response.Error(c, http.StatusServiceUnavailable, "subscription service not available")
		return
	}
	go h.subSvc.RunOnce()
	response.OK(c, gin.H{"message": "成绩检查已触发，请查看服务日志了解结果"})
}
