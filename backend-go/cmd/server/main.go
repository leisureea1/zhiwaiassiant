package main

import (
	"log"
	"strings"

	"xisu/backend-go/internal/config"
	"xisu/backend-go/internal/database"
	"xisu/backend-go/internal/http"
)

// Version is injected at build time via -ldflags, defaults to dev for local runs.
var Version = "dev"

func main() {
	cfg := config.Load()
	if len(strings.TrimSpace(cfg.JWTSecret)) < 16 {
		log.Fatalf("JWT_SECRET is required and must be at least 16 characters")
	}
	if len(strings.TrimSpace(cfg.JWTRefreshSecret)) < 16 {
		log.Fatalf("JWT_REFRESH_SECRET is required and must be at least 16 characters")
	}

	db, err := database.NewMySQL(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect mysql failed: %v", err)
	}
	log.Printf("✅ MySQL connected successfully")

	// Redis 是可选的，如果连接失败不会导致程序退出
	redisClient := database.NewRedisOptional(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)

	r := http.NewRouter(cfg, db, redisClient)

	log.Printf("🚀 %s started at :%s", cfg.AppName, cfg.AppPort)
	log.Printf("🏷️ Version: %s", Version)
	log.Printf("📝 API Prefix: %s", cfg.APIPrefix)
	log.Printf("🌍 Environment: %s", cfg.AppEnv)

	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("server run failed: %v", err)
	}
}
