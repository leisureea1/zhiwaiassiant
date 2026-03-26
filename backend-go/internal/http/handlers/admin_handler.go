package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"xisu/backend-go/internal/http/response"
	"xisu/backend-go/internal/service"
)

type AdminHandler struct {
	db             *gorm.DB
	adminService   *service.AdminService
	logService     *service.SystemLogService
}

func NewAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{
		db:           db,
		adminService: service.NewAdminService(db),
		logService:   service.NewSystemLogService(db),
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
