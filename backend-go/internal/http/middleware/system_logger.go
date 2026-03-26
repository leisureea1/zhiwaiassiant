package middleware

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"xisu/backend-go/internal/database"
)

func SystemLogger(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		if db == nil {
			return
		}

		status := c.Writer.Status()
		durationMs := int(time.Since(start).Milliseconds())
		ip := c.ClientIP()
		userAgent := c.Request.UserAgent()
		path := c.Request.URL.RequestURI()
		method := c.Request.Method
		message := fmt.Sprintf("%s %s %d", method, path, status)

		level := inferLogLevel(status)
		action := inferAction(method)

		var userID *string
		if v, ok := c.Get(ContextUserID); ok {
			switch t := v.(type) {
			case string:
				if strings.TrimSpace(t) != "" {
					u := t
					userID = &u
				}
			case fmt.Stringer:
				s := t.String()
				if strings.TrimSpace(s) != "" {
					u := s
					userID = &u
				}
			}
		}

		details, _ := json.Marshal(map[string]any{
			"ip":         ip,
			"duration":   durationMs,
			"statusCode": status,
		})

		respCode := status
		dur := durationMs
		ipAddr := truncateString(ip, 191)
		routePath := truncateString(path, 191)
		methodVal := truncateString(method, 32)
		msg := truncateString(message, 191)

		logRow := database.SystemLog{
			ID:            uuid.NewString(),
			Level:         level,
			Action:        action,
			Module:        "http",
			Message:       msg,
			Details:       details,
			UserID:        userID,
			IPAddress:     &ipAddr,
			UserAgent:     &userAgent,
			RequestPath:   &routePath,
			RequestMethod: &methodVal,
			ResponseCode:  &respCode,
			Duration:      &dur,
			CreatedAt:     time.Now(),
		}

		_ = db.Create(&logRow).Error
	}
}

func inferLogLevel(status int) string {
	switch {
	case status >= 500:
		return "ERROR"
	case status >= 400:
		return "WARN"
	default:
		return "INFO"
	}
}

func inferAction(method string) string {
	switch strings.ToUpper(method) {
	case "GET", "HEAD", "OPTIONS":
		return "VIEW"
	case "POST":
		return "CREATE"
	case "PUT", "PATCH":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	default:
		return "OTHER"
	}
}

func truncateString(v string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(v) <= max {
		return v
	}
	return v[:max]
}
