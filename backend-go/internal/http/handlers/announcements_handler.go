package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"xisu/backend-go/internal/database"
	"xisu/backend-go/internal/http/middleware"
	"xisu/backend-go/internal/http/response"
)

type AnnouncementsHandler struct {
	db *gorm.DB
}

func NewAnnouncementsHandler(db *gorm.DB) *AnnouncementsHandler {
	return &AnnouncementsHandler{db: db}
}

func (h *AnnouncementsHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	status := c.Query("status")
	announcementType := c.Query("type")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	db := h.db.Model(&database.Announcement{})

	userRole, exists := c.Get(middleware.ContextUserRole)
	if !exists || userRole == "USER" {
		db = db.Where("status = ?", "PUBLISHED")
		now := time.Now()
		db = db.Where("(published_at IS NULL OR published_at <= ?)", now)
		db = db.Where("(expires_at IS NULL OR expires_at > ?)", now)
	} else {
		if status != "" {
			db = db.Where("status = ?", status)
		}
	}

	if announcementType != "" {
		db = db.Where("type = ?", announcementType)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to count announcements")
		return
	}

	var announcements []database.Announcement
	offset := (page - 1) * pageSize
	if err := db.Order("is_pinned DESC, published_at DESC, created_at DESC").
		Offset(offset).Limit(pageSize).Find(&announcements).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get announcements")
		return
	}

	response.OK(c, gin.H{
		"total": total,
		"page":  page,
		"size":  pageSize,
		"data":  announcements,
	})
}

func (h *AnnouncementsHandler) Detail(c *gin.Context) {
	id := c.Param("id")
	var announcement database.Announcement
	if err := h.db.Where("id = ?", id).First(&announcement).Error; err != nil {
		response.Error(c, http.StatusNotFound, "announcement not found")
		return
	}

	userID, exists := c.Get(middleware.ContextUserID)
	if exists && userID != nil {
		view := database.AnnouncementView{
			ID:             uuid.NewString(),
			AnnouncementID: id,
			UserID:         userID.(string),
			ViewedAt:       time.Now(),
		}
		h.db.FirstOrCreate(&view, database.AnnouncementView{
			AnnouncementID: id,
			UserID:         userID.(string),
		})
	}

	response.OK(c, announcement)
}

func (h *AnnouncementsHandler) GetUnviewedCount(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextUserID)

	var count int64
	subQuery := h.db.Model(&database.AnnouncementView{}).
		Select("announcement_id").
		Where("user_id = ?", userID.(string))

	if err := h.db.Model(&database.Announcement{}).
		Where("status = ?", "PUBLISHED").
		Where("id NOT IN (?)", subQuery).
		Count(&count).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to count unviewed announcements")
		return
	}

	response.OK(c, gin.H{"count": count})
}

type CreateAnnouncementRequest struct {
	Title       string  `json:"title" binding:"required,max=200"`
	Content     string  `json:"content" binding:"required,max=50000"`
	Summary     *string `json:"summary" binding:"omitempty,max=500"`
	Type        string  `json:"type" binding:"omitempty,oneof=NORMAL IMPORTANT URGENT"`
	IsPinned    bool    `json:"isPinned"`
	IsPopup     bool    `json:"isPopup"`
	PublishedAt *string `json:"publishedAt"`
	ExpiresAt   *string `json:"expiresAt"`
}

func (h *AnnouncementsHandler) Create(c *gin.Context) {
	authorID, _ := c.Get(middleware.ContextUserID)
	var req CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request")
		return
	}

	announcement := database.Announcement{
		ID:       uuid.NewString(),
		Title:    req.Title,
		Content:  req.Content,
		Summary:  req.Summary,
		Type:     req.Type,
		Status:   "DRAFT",
		IsPinned: req.IsPinned,
		IsPopup:  req.IsPopup,
		AuthorID: authorID.(string),
	}

	if req.Type == "" {
		announcement.Type = "NORMAL"
	}

	if req.PublishedAt != nil {
		if t, err := time.Parse(time.RFC3339, *req.PublishedAt); err == nil {
			announcement.PublishedAt = &t
		}
	}
	if req.ExpiresAt != nil {
		if t, err := time.Parse(time.RFC3339, *req.ExpiresAt); err == nil {
			announcement.ExpiresAt = &t
		}
	}

	if err := h.db.Create(&announcement).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to create announcement")
		return
	}

	response.OK(c, announcement)
}

type UpdateAnnouncementRequest struct {
	Title       *string `json:"title" binding:"omitempty,max=200"`
	Content     *string `json:"content" binding:"omitempty,max=50000"`
	Summary     *string `json:"summary" binding:"omitempty,max=500"`
	Type        *string `json:"type" binding:"omitempty,oneof=NORMAL IMPORTANT URGENT"`
	IsPinned    *bool   `json:"isPinned"`
	IsPopup     *bool   `json:"isPopup"`
	PublishedAt *string `json:"publishedAt"`
	ExpiresAt   *string `json:"expiresAt"`
}

func (h *AnnouncementsHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req UpdateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request")
		return
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Summary != nil {
		updates["summary"] = *req.Summary
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.IsPinned != nil {
		updates["is_pinned"] = *req.IsPinned
	}
	if req.IsPopup != nil {
		updates["is_popup"] = *req.IsPopup
	}
	if req.PublishedAt != nil {
		if t, err := time.Parse(time.RFC3339, *req.PublishedAt); err == nil {
			updates["published_at"] = t
		}
	}
	if req.ExpiresAt != nil {
		if t, err := time.Parse(time.RFC3339, *req.ExpiresAt); err == nil {
			updates["expires_at"] = t
		}
	}

	if err := h.db.Model(&database.Announcement{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to update announcement")
		return
	}

	var announcement database.Announcement
	h.db.Where("id = ?", id).First(&announcement)
	response.OK(c, announcement)
}

func (h *AnnouncementsHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.db.Delete(&database.Announcement{}, "id = ?", id).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to delete announcement")
		return
	}
	response.OK(c, gin.H{"success": true})
}

func (h *AnnouncementsHandler) Publish(c *gin.Context) {
	id := c.Param("id")
	now := time.Now()
	if err := h.db.Model(&database.Announcement{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       "PUBLISHED",
		"published_at": now,
	}).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to publish announcement")
		return
	}

	var announcement database.Announcement
	h.db.Where("id = ?", id).First(&announcement)
	response.OK(c, announcement)
}

func (h *AnnouncementsHandler) TogglePin(c *gin.Context) {
	id := c.Param("id")
	var announcement database.Announcement
	if err := h.db.Where("id = ?", id).First(&announcement).Error; err != nil {
		response.Error(c, http.StatusNotFound, "announcement not found")
		return
	}

	if err := h.db.Model(&announcement).Update("is_pinned", !announcement.IsPinned).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to toggle pin")
		return
	}

	h.db.Where("id = ?", id).First(&announcement)
	response.OK(c, announcement)
}

func (h *AnnouncementsHandler) MarkViewed(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get(middleware.ContextUserID)

	view := database.AnnouncementView{
		ID:             uuid.NewString(),
		AnnouncementID: id,
		UserID:         userID.(string),
		ViewedAt:       time.Now(),
	}

	if err := h.db.FirstOrCreate(&view, database.AnnouncementView{
		AnnouncementID: id,
		UserID:         userID.(string),
	}).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to mark as viewed")
		return
	}

	response.OK(c, gin.H{"success": true})
}
