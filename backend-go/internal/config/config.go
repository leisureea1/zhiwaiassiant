package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName string
	AppEnv  string
	AppPort string

	APIPrefix   string
	CORSOrigins []string

	DatabaseURL string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	JWTSecret        string
	JWTRefreshSecret string
	AccessTTL        time.Duration
	RefreshTTL       time.Duration

	// Mail configuration
	MailHost     string
	MailPort     string
	MailUsername string
	MailPassword string
	MailFrom     string
}

func Load() *Config {
	_ = godotenv.Load()

	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	accessTTL, err := time.ParseDuration(getEnv("JWT_ACCESS_EXPIRES", "15m"))
	if err != nil {
		log.Printf("invalid JWT_ACCESS_EXPIRES, fallback to 15m")
		accessTTL = 15 * time.Minute
	}

	refreshTTL, err := time.ParseDuration(getEnv("JWT_REFRESH_EXPIRES", "168h"))
	if err != nil {
		log.Printf("invalid JWT_REFRESH_EXPIRES, fallback to 168h")
		refreshTTL = 168 * time.Hour
	}

	return &Config{
		AppName: getEnv("APP_NAME", "XISU Go Backend"),
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "3000"),

		APIPrefix:   getEnv("API_PREFIX", "/api/v1"),
		CORSOrigins: splitCSV(getEnv("CORS_ORIGINS", "http://localhost:5173,http://localhost:3000")),

		DatabaseURL: getEnv("DATABASE_URL", ""),

		RedisAddr:     getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,

		JWTSecret:        getEnv("JWT_SECRET", ""),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", ""),
		AccessTTL:        accessTTL,
		RefreshTTL:       refreshTTL,

		MailHost:     getEnv("MAIL_HOST", "smtp.gmail.com"),
		MailPort:     getEnv("MAIL_PORT", "587"),
		MailUsername: firstNonEmpty(getEnv("MAIL_USERNAME", ""), getEnv("MAIL_USER", "")),
		MailPassword: getEnv("MAIL_PASSWORD", ""),
		MailFrom: firstNonEmpty(
			getEnv("MAIL_FROM", ""),
			getEnv("MAIL_USERNAME", ""),
			getEnv("MAIL_USER", ""),
		),
	}
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func splitCSV(v string) []string {
	items := strings.Split(v, ",")
	out := make([]string, 0, len(items))
	for _, s := range items {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
