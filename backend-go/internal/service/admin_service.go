package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
	"xisu/backend-go/internal/database"
)

type AdminService struct {
	db *gorm.DB
}

var editableConfigKeys = []string{
	"APP_PORT",
	"CORS_ORIGINS",
	"JWT_ACCESS_EXPIRES",
	"JWT_REFRESH_EXPIRES",
	"MAIL_HOST",
	"MAIL_PORT",
	"MAIL_USERNAME",
	"MAIL_FROM",
	"UPLOAD_DIR",
	"MAX_FILE_SIZE",
}

var sensitiveConfigKeys = map[string]bool{
	"JWT_SECRET":           true,
	"JWT_REFRESH_SECRET":   true,
	"MAIL_PASSWORD":        true,
	"JWXT_SERVICE_API_KEY": true,
	"DATABASE_URL":         true,
	"REDIS_PASSWORD":       true,
}

var configGroups = []map[string]any{
	{"label": "基础配置", "keys": []string{"APP_PORT", "CORS_ORIGINS"}},
	{"label": "JWT 认证", "keys": []string{"JWT_ACCESS_EXPIRES", "JWT_REFRESH_EXPIRES"}},
	{"label": "邮件配置", "keys": []string{"MAIL_HOST", "MAIL_PORT", "MAIL_USERNAME", "MAIL_PASSWORD", "MAIL_FROM"}},
	{"label": "教务服务", "keys": []string{"JWXT_SERVICE_API_KEY"}},
	{"label": "文件上传", "keys": []string{"UPLOAD_DIR", "MAX_FILE_SIZE"}},
}

func NewAdminService(db *gorm.DB) *AdminService {
	return &AdminService{db: db}
}

type DashboardStats struct {
	TotalUsers         int64 `json:"totalUsers"`
	ActiveUsers        int64 `json:"activeUsers"`
	TotalAnnouncements int64 `json:"totalAnnouncements"`
	TodayLogins        int64 `json:"todayLogins"`
}

func (s *AdminService) GetDashboardStats() (*DashboardStats, error) {
	stats := &DashboardStats{}

	if err := s.db.Model(&database.User{}).Count(&stats.TotalUsers).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&database.User{}).Where("status = ?", "ACTIVE").Count(&stats.ActiveUsers).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&database.Announcement{}).Count(&stats.TotalAnnouncements).Error; err != nil {
		return nil, err
	}

	today := time.Now().Truncate(24 * time.Hour)
	if err := s.db.Model(&database.User{}).Where("last_login_at >= ?", today).Count(&stats.TodayLogins).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

type PendingItems struct {
	DraftAnnouncements int64 `json:"draftAnnouncements"`
	InactiveUsers      int64 `json:"inactiveUsers"`
}

func (s *AdminService) GetPendingItems() (*PendingItems, error) {
	items := &PendingItems{}

	if err := s.db.Model(&database.Announcement{}).Where("status = ?", "DRAFT").Count(&items.DraftAnnouncements).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&database.User{}).Where("status = ?", "INACTIVE").Count(&items.InactiveUsers).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func (s *AdminService) GetFeatureFlags() ([]database.FeatureFlag, error) {
	var flags []database.FeatureFlag
	if err := s.db.Order("name").Find(&flags).Error; err != nil {
		return nil, err
	}
	return flags, nil
}

func (s *AdminService) UpdateFeatureFlag(name string, isEnabled bool) error {
	return s.db.Model(&database.FeatureFlag{}).Where("name = ?", name).Update("is_enabled", isEnabled).Error
}

func (s *AdminService) GetConfig() (map[string]any, error) {
	envMap, err := parseEnvFile()
	if err != nil {
		return nil, err
	}

	configs := map[string]string{}
	for _, key := range editableConfigKeys {
		value := envMap[key]
		if sensitiveConfigKeys[key] {
			configs[key] = maskValue(value)
		} else {
			configs[key] = value
		}
	}

	return map[string]any{
		"configs":      configs,
		"groups":       configGroups,
		"editableKeys": editableConfigKeys,
	}, nil
}

func (s *AdminService) UpdateConfig(newConfigs map[string]string) (map[string]any, error) {
	envMap, err := parseEnvFile()
	if err != nil {
		return nil, err
	}

	editableSet := map[string]bool{}
	for _, key := range editableConfigKeys {
		editableSet[key] = true
	}

	updated := make([]string, 0)
	for key, value := range newConfigs {
		if !editableSet[key] {
			continue
		}
		if isMasked(value) {
			continue
		}
		if envMap[key] != value {
			envMap[key] = value
			updated = append(updated, key)
		}
	}

	if len(updated) > 0 {
		if err := writeEnvFile(envMap); err != nil {
			return nil, err
		}
	}

	message := "没有需要更新的配置"
	if len(updated) > 0 {
		message = fmt.Sprintf("已更新 %d 项配置，部分配置需要重启服务后生效", len(updated))
	}

	return map[string]any{
		"updated": updated,
		"message": message,
	}, nil
}

func getEnvPath() string {
	return filepath.Join(".", ".env")
}

func parseEnvFile() (map[string]string, error) {
	path := getEnvPath()
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}

	result := map[string]string{}
	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		eq := strings.Index(trimmed, "=")
		if eq <= 0 {
			continue
		}
		key := strings.TrimSpace(trimmed[:eq])
		value := strings.TrimSpace(trimmed[eq+1:])
		result[key] = value
	}

	return result, nil
}

func writeEnvFile(envMap map[string]string) error {
	path := getEnvPath()
	original, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	updatedKeys := map[string]bool{}
	lines := strings.Split(string(original), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		eq := strings.Index(trimmed, "=")
		if eq <= 0 {
			continue
		}
		key := strings.TrimSpace(trimmed[:eq])
		value, ok := envMap[key]
		if !ok {
			continue
		}
		lines[i] = key + "=" + value
		updatedKeys[key] = true
	}

	for key, value := range envMap {
		if updatedKeys[key] {
			continue
		}
		lines = append(lines, key+"="+value)
	}

	content := strings.Join(lines, "\n")
	return os.WriteFile(path, []byte(content), 0644)
}

func maskValue(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
}

func isMasked(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	for _, r := range trimmed {
		if r != '*' {
			return false
		}
	}
	return true
}
