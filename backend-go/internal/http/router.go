package http

import (
	"xisu/backend-go/internal/config"
	"xisu/backend-go/internal/http/handlers"
	"xisu/backend-go/internal/http/middleware"
	"xisu/backend-go/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func NewRouter(cfg *config.Config, db *gorm.DB, redisClient *redis.Client) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS(cfg.CORSOrigins))
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.SystemLogger(db))

	// 仅开发环境暴露测试端点
	if cfg.AppEnv != "production" {
		r.StaticFile("/jwxt-test", "./static/jwxt_test.html")
	}
	r.Static("/uploads/avatars", "./uploads/avatars")

	uploadHandler := handlers.NewUploadHandler(db, "./uploads")

	// 附件下载 - 公开但文件名为UUID，难以猜测
	r.GET("/uploads/attachments/:filename", uploadHandler.ServeAttachment)

	tokenSvc := service.NewTokenService(cfg.JWTSecret, cfg.JWTRefreshSecret, cfg.AccessTTL, cfg.RefreshTTL)
	jwxtSvc := service.NewJwxtDirectService(redisClient)
	mailSvc := service.NewMailService(cfg.MailHost, cfg.MailPort, cfg.MailUsername, cfg.MailPassword, cfg.MailFrom)

	healthHandler := handlers.NewHealthHandler(db, redisClient)
	authHandler := handlers.NewAuthHandler(db, tokenSvc, redisClient)
	extAuthHandler := handlers.NewExtendedAuthHandler(db, tokenSvc, mailSvc, jwxtSvc, redisClient)
	usersHandler := handlers.NewUsersHandler(db)
	annHandler := handlers.NewAnnouncementsHandler(db)
	jwxtHandler := handlers.NewJWXTHandler(db, jwxtSvc)
	adminHandler := handlers.NewAdminHandler(db)

	api := r.Group(cfg.APIPrefix)
	{
		api.GET("/health", healthHandler.Get)

		auth := api.Group("/auth")
		{
			// 公开接口
			auth.POST("/send-code", extAuthHandler.SendVerificationCode)
			auth.POST("/verify-code", extAuthHandler.VerifyEmailCode)
			auth.POST("/register", extAuthHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/forgot-password", extAuthHandler.ForgotPassword)
			auth.POST("/reset-password", extAuthHandler.ResetPassword)
		}

		protected := api.Group("/")
		protected.Use(middleware.JWTAuth(tokenSvc))
		{
			protected.POST("/auth/logout", authHandler.Logout)
			protected.POST("/auth/change-password", extAuthHandler.ChangePassword)

			// 用户相关路由
			protected.GET("/users/me", usersHandler.Me)
			protected.GET("/users/notifications/settings", usersHandler.GetNotificationSettings)
			protected.POST("/users/notifications/settings", usersHandler.UpdateNotificationSettings)

			// 用户列表 - 需要管理员权限
			protected.GET("/users", middleware.RequireRole("ADMIN", "SUPER_ADMIN"), usersHandler.List)

			// 用户详情和更新
			protected.GET("/users/:id", usersHandler.GetUserByID)
			protected.PUT("/users/:id", usersHandler.Update)
			protected.POST("/users/:id/avatar/upload", uploadHandler.UploadAvatar)

			protected.POST("/upload", uploadHandler.UploadFile)

			protected.GET("/announcements", annHandler.List)
			protected.GET("/announcements/unviewed-count", annHandler.GetUnviewedCount)
			protected.GET("/announcements/:id", annHandler.Detail)
			protected.POST("/announcements/:id/mark-viewed", annHandler.MarkViewed)

			jwxt := protected.Group("/jwxt")
			{
				jwxt.GET("/course", jwxtHandler.Course)
				jwxt.GET("/course/refresh", jwxtHandler.CourseRefresh)
				jwxt.GET("/grade", jwxtHandler.Grade)
				jwxt.GET("/exam", jwxtHandler.Exam)
				jwxt.GET("/semester", jwxtHandler.Semester)
				jwxt.GET("/evaluation/pending", jwxtHandler.EvaluationPending)
				jwxt.POST("/evaluation/auto", jwxtHandler.EvaluationAuto)
				jwxt.GET("/user", jwxtHandler.User)
				jwxt.POST("/bind", jwxtHandler.Bind)
				jwxt.POST("/unbind", jwxtHandler.Unbind)
			}

			admin := protected.Group("/admin")
			admin.Use(middleware.RequireRole("ADMIN", "SUPER_ADMIN"))
			{
				admin.GET("/dashboard/stats", adminHandler.GetDashboardStats)
				admin.GET("/dashboard/pending-items", adminHandler.GetPendingItems)
				admin.GET("/system-logs", adminHandler.GetSystemLogs)
				admin.GET("/system-logs/action-types", adminHandler.GetActionTypes)
				admin.GET("/system-logs/stats", adminHandler.GetLogStats)
				admin.GET("/features", adminHandler.GetFeatureFlags)
				admin.POST("/features/:name", adminHandler.UpdateFeatureFlag)

				// 管理员更新用户信息
				admin.PUT("/users/:id/admin", usersHandler.AdminUpdate)

				admin.POST("/announcements", annHandler.Create)
				admin.PUT("/announcements/:id", annHandler.Update)
				admin.DELETE("/announcements/:id", annHandler.Delete)
				admin.POST("/announcements/:id/publish", annHandler.Publish)
				admin.POST("/announcements/:id/pin", annHandler.TogglePin)
			}

			superAdmin := protected.Group("/admin")
			superAdmin.Use(middleware.RequireRole("SUPER_ADMIN"))
			{
				superAdmin.GET("/config", adminHandler.GetConfig)
				superAdmin.POST("/config", adminHandler.UpdateConfig)
				superAdmin.DELETE("/users/:id", usersHandler.Delete)
			}
		}
	}

	return r
}
