package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"xisu/backend-go/internal/database"
	"xisu/backend-go/internal/http/response"
	"xisu/backend-go/internal/service"
)

type AdminHandler struct {
	db           *gorm.DB
	adminService *service.AdminService
	logService   *service.SystemLogService
	mailService  *service.MailService
}

func NewAdminHandler(db *gorm.DB, mailSvc *service.MailService) *AdminHandler {
	return &AdminHandler{
		db:           db,
		adminService: service.NewAdminService(db),
		logService:   service.NewSystemLogService(db),
		mailService:  mailSvc,
	}
}

func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.adminService.GetDashboardStats()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get dashboard stats")
		return
	}
	response.OK(c, stats)
}

func (h *AdminHandler) GetPendingItems(c *gin.Context) {
	items, err := h.adminService.GetPendingItems()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get pending items")
		return
	}
	response.OK(c, items)
}

func (h *AdminHandler) GetSystemLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	query := service.SystemLogQuery{
		Page:     page,
		PageSize: pageSize,
		Level:    c.Query("level"),
		Action:   c.Query("action"),
		Module:   c.Query("module"),
		UserID:   c.Query("userId"),
	}

	if startAt := c.Query("startAt"); startAt != "" {
		if t, err := time.Parse(time.RFC3339, startAt); err == nil {
			query.StartAt = &t
		}
	}
	if endAt := c.Query("endAt"); endAt != "" {
		if t, err := time.Parse(time.RFC3339, endAt); err == nil {
			query.EndAt = &t
		}
	}

	result, err := h.logService.FindAll(query)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get system logs")
		return
	}
	response.OK(c, result)
}

func (h *AdminHandler) GetActionTypes(c *gin.Context) {
	actions, err := h.logService.GetActionTypes()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get action types")
		return
	}
	response.OK(c, actions)
}

func (h *AdminHandler) GetLogStats(c *gin.Context) {
	stats, err := h.logService.GetStats()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get log stats")
		return
	}
	response.OK(c, stats)
}

func (h *AdminHandler) GetFeatureFlags(c *gin.Context) {
	flags, err := h.adminService.GetFeatureFlags()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get feature flags")
		return
	}
	response.OK(c, flags)
}

type UpdateFeatureFlagRequest struct {
	IsEnabled bool `json:"isEnabled"`
}

func (h *AdminHandler) UpdateFeatureFlag(c *gin.Context) {
	name := c.Param("name")
	var req UpdateFeatureFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request")
		return
	}

	if err := h.adminService.UpdateFeatureFlag(name, req.IsEnabled); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to update feature flag")
		return
	}
	response.OK(c, gin.H{"success": true})
}

func (h *AdminHandler) GetConfig(c *gin.Context) {
	config, err := h.adminService.GetConfig()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get config")
		return
	}
	response.OK(c, config)
}

type UpdateConfigRequest struct {
	Configs map[string]string `json:"configs"`
}

func (h *AdminHandler) UpdateConfig(c *gin.Context) {
	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request")
		return
	}

	result, err := h.adminService.UpdateConfig(req.Configs)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to update config")
		return
	}
	response.OK(c, result)
}

type SendBulkEmailRequest struct {
	Subject   string `json:"subject" binding:"required,min=1,max=200"`
	Content   string `json:"content" binding:"required,min=1"`
	Target    string `json:"target"` // "all" | "active" | "inactive"
	Role      string `json:"role"`   // filter by role, empty = all
}

type SendBulkEmailResult struct {
	Total   int64  `json:"total"`
	Success int    `json:"success"`
	Failed  int    `json:"failed"`
	Skipped int    `json:"skipped"`
}

func (h *AdminHandler) SendBulkEmail(c *gin.Context) {
	var req SendBulkEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	// Build query
	query := h.db.Model(&database.User{})
	switch req.Target {
	case "active":
		query = query.Where("status = ?", "ACTIVE")
	case "inactive":
		query = query.Where("status != ?", "ACTIVE")
	case "all":
		// no filter
	default:
		query = query.Where("status = ?", "ACTIVE")
	}

	if req.Role != "" {
		query = query.Where("role = ?", req.Role)
	}

	// Get users with email
	var users []database.User
	if err := query.Where("email IS NOT NULL AND email != ''").Find(&users).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to query users")
		return
	}

	if len(users) == 0 {
		response.OK(c, SendBulkEmailResult{Total: 0, Success: 0, Failed: 0, Skipped: 0})
		return
	}

	success := 0
	failed := 0
	skipped := 0

	for _, user := range users {
		if user.Email == nil || *user.Email == "" {
			skipped++
			continue
		}

		// Personalize greeting
		displayName := ""
		if user.RealName != nil && *user.RealName != "" {
			displayName = *user.RealName
		} else {
			displayName = user.Username
		}

		personalizedContent := `
			<div style="max-width: 600px; margin: 0 auto; padding: 20px; font-family: 'Microsoft YaHei', sans-serif;">
				<div style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 30px; border-radius: 10px 10px 0 0;">
					<h1 style="color: white; margin: 0; text-align: center;">知外助手</h1>
				</div>
				<div style="background: #ffffff; padding: 40px; border: 1px solid #e5e7eb; border-top: none; border-radius: 0 0 10px 10px;">
					<p style="color: #6b7280; font-size: 16px; line-height: 1.6;">
						尊敬的 <strong>{{USER_NAME}}</strong>，您好！
					</p>
					<hr style="border: none; border-top: 1px solid #e5e7eb; margin: 24px 0;" />
					{{CONTENT}}
					<hr style="border: none; border-top: 1px solid #e5e7eb; margin: 24px 0;" />
					<p style="color: #9ca3af; font-size: 12px; text-align: center;">
						此邮件由知外助手管理后台发送，请勿直接回复
					</p>
				</div>
			</div>
		`
		personalizedContent = strings.Replace(personalizedContent, "{{USER_NAME}}", displayName, 1)
		personalizedContent = strings.Replace(personalizedContent, "{{CONTENT}}", req.Content, 1)

		if err := h.mailService.SendCustomMail(*user.Email, req.Subject, personalizedContent); err != nil {
			log.Printf("[SendBulkEmail] Failed to send to %s: %v", *user.Email, err)
			failed++
		} else {
			success++
		}

		// Avoid overwhelming the SMTP server
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("[SendBulkEmail] Done: total=%d success=%d failed=%d skipped=%d",
		len(users), success, failed, skipped)

	response.OK(c, SendBulkEmailResult{
		Total:   int64(len(users)),
		Success: success,
		Failed:  failed,
		Skipped: skipped,
	})
}
