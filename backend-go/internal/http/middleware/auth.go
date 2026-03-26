package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"xisu/backend-go/internal/http/response"
	"xisu/backend-go/internal/service"
)

const (
	ContextUserID  = "userID"
	ContextUserRole = "userRole"
)

func JWTAuth(tokenService *service.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(c, http.StatusUnauthorized, "invalid authorization format")
			c.Abort()
			return
		}

		claims, err := tokenService.ParseAccessToken(parts[1])
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid access token")
			c.Abort()
			return
		}

		if claims.Role == "TEMP" {
			response.Error(c, http.StatusForbidden, "invalid account role, please complete registration")
			c.Abort()
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUserRole, claims.Role)
		c.Next()
	}
}
