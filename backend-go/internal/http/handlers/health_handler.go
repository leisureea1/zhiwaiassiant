package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"xisu/backend-go/internal/http/response"
)

type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewHealthHandler(db *gorm.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

func (h *HealthHandler) Get(c *gin.Context) {
	status := gin.H{
		"status": "ok",
	}

	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		response.Error(c, http.StatusServiceUnavailable, "mysql unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := h.redis.Ping(ctx).Err(); err != nil {
		response.Error(c, http.StatusServiceUnavailable, "redis unavailable")
		return
	}

	response.OK(c, status)
}
