package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"xisu/backend-go/internal/http/response"
)

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get(ContextUserRole)
		if !exists {
			response.Error(c, http.StatusForbidden, "no role found")
			c.Abort()
			return
		}

		role := userRole.(string)
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}

		response.Error(c, http.StatusForbidden, "insufficient permissions")
		c.Abort()
	}
}
