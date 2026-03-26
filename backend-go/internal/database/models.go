package database

import (
	"encoding/json"
	"time"
)

type User struct {
	ID           string     `gorm:"column:id;primaryKey"`
	Username     string     `gorm:"column:username"`
	Email        *string    `gorm:"column:email"`
	Phone        *string    `gorm:"column:phone"`
	PasswordHash string     `gorm:"column:password_hash"`
	StudentID    *string    `gorm:"column:student_id"`
	RealName     *string    `gorm:"column:real_name"`
	Nickname     *string    `gorm:"column:nickname"`
	Avatar       *string    `gorm:"column:avatar"`
	College      *string    `gorm:"column:college"`
	Major        *string    `gorm:"column:major"`
	ClassName    *string    `gorm:"column:class_name"`
	Role         string     `gorm:"column:role"`
	Status       string     `gorm:"column:status"`
	JWXTUsername *string    `gorm:"column:jwxt_username"`
	JWXTPassword *string    `gorm:"column:jwxt_password"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at"`
	LastLoginIP  *string    `gorm:"column:last_login_ip"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

func (User) TableName() string {
	return "users"
}

type RefreshToken struct {
	ID        string    `gorm:"column:id;primaryKey"`
	Token     string    `gorm:"column:token"`
	UserID    string    `gorm:"column:user_id"`
	UserAgent *string   `gorm:"column:user_agent"`
	IPAddress *string   `gorm:"column:ip_address"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

type Announcement struct {
	ID          string     `gorm:"column:id;primaryKey"`
	Title       string     `gorm:"column:title"`
	Content     string     `gorm:"column:content"`
	Summary     *string    `gorm:"column:summary"`
	Type        string     `gorm:"column:type"`
	Status      string     `gorm:"column:status"`
	IsPinned    bool       `gorm:"column:is_pinned"`
	IsPopup     bool       `gorm:"column:is_popup"`
	AuthorID    string     `gorm:"column:author_id"`
	PublishedAt *time.Time `gorm:"column:published_at"`
	ExpiresAt   *time.Time `gorm:"column:expires_at"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
}

func (Announcement) TableName() string {
	return "announcements"
}

type AnnouncementView struct {
	ID             string    `gorm:"column:id;primaryKey"`
	AnnouncementID string    `gorm:"column:announcement_id"`
	UserID         string    `gorm:"column:user_id"`
	ViewedAt       time.Time `gorm:"column:viewed_at"`
}

func (AnnouncementView) TableName() string {
	return "announcement_views"
}

type SystemLog struct {
	ID            string          `gorm:"column:id;primaryKey"`
	Level         string          `gorm:"column:level"`
	Action        string          `gorm:"column:action"`
	Module        string          `gorm:"column:module"`
	Message       string          `gorm:"column:message"`
	Details       json.RawMessage `gorm:"column:details;type:json"`
	UserID        *string         `gorm:"column:user_id"`
	IPAddress     *string         `gorm:"column:ip_address"`
	UserAgent     *string         `gorm:"column:user_agent"`
	RequestPath   *string         `gorm:"column:request_path"`
	RequestMethod *string         `gorm:"column:request_method"`
	ResponseCode  *int            `gorm:"column:response_code"`
	Duration      *int            `gorm:"column:duration"`
	CreatedAt     time.Time       `gorm:"column:created_at"`
}

func (SystemLog) TableName() string {
	return "system_logs"
}

type FeatureFlag struct {
	ID          string          `gorm:"column:id;primaryKey"`
	Name        string          `gorm:"column:name"`
	Description *string         `gorm:"column:description"`
	IsEnabled   bool            `gorm:"column:is_enabled"`
	Metadata    json.RawMessage `gorm:"column:metadata;type:json"`
	CreatedAt   time.Time       `gorm:"column:created_at"`
	UpdatedAt   time.Time       `gorm:"column:updated_at"`
}

func (FeatureFlag) TableName() string {
	return "feature_flags"
}

type NotificationSetting struct {
	ID                 string    `gorm:"column:id;primaryKey"`
	UserID             string    `gorm:"column:user_id"`
	EmailEnabled       bool      `gorm:"column:email_enabled"`
	PushEnabled        bool      `gorm:"column:push_enabled"`
	BarkKey            *string   `gorm:"column:bark_key"`
	GradeNotify        bool      `gorm:"column:grade_notify"`
	ExamNotify         bool      `gorm:"column:exam_notify"`
	AnnouncementNotify bool      `gorm:"column:announcement_notify"`
	CreatedAt          time.Time `gorm:"column:created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at"`
}

func (NotificationSetting) TableName() string {
	return "notification_settings"
}

type UploadedFile struct {
	ID           string    `gorm:"column:id;primaryKey"`
	FileName     string    `gorm:"column:file_name"`
	OriginalName string    `gorm:"column:original_name"`
	FilePath     string    `gorm:"column:file_path"`
	FileSize     int       `gorm:"column:file_size"`
	MimeType     string    `gorm:"column:mime_type"`
	UploaderID   *string   `gorm:"column:uploader_id"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (UploadedFile) TableName() string {
	return "uploaded_files"
}

type GradeSubscription struct {
	ID               string     `gorm:"column:id;primaryKey"`
	UserID           string     `gorm:"column:user_id;uniqueIndex"`
	Enabled          bool       `gorm:"column:enabled;default:false"`
	LastCheckedAt    *time.Time `gorm:"column:last_checked_at"`
	LastGradeHash    *string    `gorm:"column:last_grade_hash"`
	LastNotifiedAt   *time.Time `gorm:"column:last_notified_at"`
	TotalNotified    int        `gorm:"column:total_notified;default:0"`
	SemesterID       *string    `gorm:"column:semester_id"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

func (GradeSubscription) TableName() string {
	return "grade_subscriptions"
}
