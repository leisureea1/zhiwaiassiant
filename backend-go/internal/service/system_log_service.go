package service

import (
	"time"

	"gorm.io/gorm"
	"xisu/backend-go/internal/database"
)

type SystemLogService struct {
	db *gorm.DB
}

func NewSystemLogService(db *gorm.DB) *SystemLogService {
	return &SystemLogService{db: db}
}

type SystemLogQuery struct {
	Page     int
	PageSize int
	Level    string
	Action   string
	Module   string
	UserID   string
	StartAt  *time.Time
	EndAt    *time.Time
}

type SystemLogResult struct {
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
	Data  []SystemLogItem `json:"data"`
}

type LogUser struct {
	ID       string  `json:"id"`
	Username string  `json:"username"`
	RealName *string `json:"realName,omitempty"`
}

type SystemLogItem struct {
	database.SystemLog
	User *LogUser `json:"user,omitempty"`
}

func (s *SystemLogService) FindAll(query SystemLogQuery) (*SystemLogResult, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}

	db := s.db.Model(&database.SystemLog{})

	if query.Level != "" {
		db = db.Where("level = ?", query.Level)
	}
	if query.Action != "" {
		db = db.Where("action = ?", query.Action)
	}
	if query.Module != "" {
		db = db.Where("module = ?", query.Module)
	}
	if query.UserID != "" {
		db = db.Where("user_id = ?", query.UserID)
	}
	if query.StartAt != nil {
		db = db.Where("created_at >= ?", query.StartAt)
	}
	if query.EndAt != nil {
		db = db.Where("created_at <= ?", query.EndAt)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	var logs []database.SystemLog
	offset := (query.Page - 1) * query.PageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&logs).Error; err != nil {
		return nil, err
	}

	userIDSet := make(map[string]struct{})
	for _, log := range logs {
		if log.UserID != nil && *log.UserID != "" {
			userIDSet[*log.UserID] = struct{}{}
		}
	}

	userIDs := make([]string, 0, len(userIDSet))
	for id := range userIDSet {
		userIDs = append(userIDs, id)
	}

	userMap := make(map[string]LogUser)
	if len(userIDs) > 0 {
		type userLite struct {
			ID       string  `gorm:"column:id"`
			Username string  `gorm:"column:username"`
			RealName *string `gorm:"column:real_name"`
		}

		var users []userLite
		if err := s.db.Model(&database.User{}).
			Select("id, username, real_name").
			Where("id IN ?", userIDs).
			Find(&users).Error; err != nil {
			return nil, err
		}

		for _, u := range users {
			userMap[u.ID] = LogUser{
				ID:       u.ID,
				Username: u.Username,
				RealName: u.RealName,
			}
		}
	}

	items := make([]SystemLogItem, 0, len(logs))
	for _, log := range logs {
		item := SystemLogItem{SystemLog: log}
		if log.UserID != nil {
			if user, ok := userMap[*log.UserID]; ok {
				u := user
				item.User = &u
			}
		}
		items = append(items, item)
	}

	return &SystemLogResult{
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
		Data:  items,
	}, nil
}

func (s *SystemLogService) GetActionTypes() ([]string, error) {
	var actions []string
	if err := s.db.Model(&database.SystemLog{}).Distinct("action").Pluck("action", &actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

type LogStats struct {
	TotalLogs    int64            `json:"totalLogs"`
	TodayLogs    int64            `json:"todayLogs"`
	ErrorLogs    int64            `json:"errorLogs"`
	ByLevel      map[string]int64 `json:"byLevel"`
	ByAction     map[string]int64 `json:"byAction"`
	TopUsers     []UserLogCount   `json:"topUsers"`
	RecentErrors []database.SystemLog `json:"recentErrors"`
}

type UserLogCount struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Count    int64  `json:"count"`
}

func (s *SystemLogService) GetStats() (*LogStats, error) {
	stats := &LogStats{
		ByLevel:  make(map[string]int64),
		ByAction: make(map[string]int64),
	}

	if err := s.db.Model(&database.SystemLog{}).Count(&stats.TotalLogs).Error; err != nil {
		return nil, err
	}

	today := time.Now().Truncate(24 * time.Hour)
	if err := s.db.Model(&database.SystemLog{}).Where("created_at >= ?", today).Count(&stats.TodayLogs).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&database.SystemLog{}).Where("level = ?", "ERROR").Count(&stats.ErrorLogs).Error; err != nil {
		return nil, err
	}

	type LevelCount struct {
		Level string
		Count int64
	}
	var levelCounts []LevelCount
	if err := s.db.Model(&database.SystemLog{}).Select("level, COUNT(*) as count").Group("level").Find(&levelCounts).Error; err != nil {
		return nil, err
	}
	for _, lc := range levelCounts {
		stats.ByLevel[lc.Level] = lc.Count
	}

	type ActionCount struct {
		Action string
		Count  int64
	}
	var actionCounts []ActionCount
	if err := s.db.Model(&database.SystemLog{}).Select("action, COUNT(*) as count").Group("action").Find(&actionCounts).Error; err != nil {
		return nil, err
	}
	for _, ac := range actionCounts {
		stats.ByAction[ac.Action] = ac.Count
	}

	if err := s.db.Model(&database.SystemLog{}).Where("level = ?", "ERROR").Order("created_at DESC").Limit(10).Find(&stats.RecentErrors).Error; err != nil {
		return nil, err
	}

	return stats, nil
}
