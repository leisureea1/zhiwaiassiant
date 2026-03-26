package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
	"xisu/backend-go/internal/database"
	"xisu/backend-go/internal/service/jwxt"
)

type GradeSubscriptionService struct {
	db      *gorm.DB
	jwxtSvc *jwxt.JwxtDirectService
	mailSvc *MailService
}

func NewGradeSubscriptionService(db *gorm.DB, jwxtSvc *jwxt.JwxtDirectService, mailSvc *MailService) *GradeSubscriptionService {
	return &GradeSubscriptionService{db: db, jwxtSvc: jwxtSvc, mailSvc: mailSvc}
}

// Start begins the hourly grade checking loop. It should be called in a goroutine.
func (s *GradeSubscriptionService) Start() {
	log.Printf("[GradeSubscription] Service started, will check every hour")
	s.runOnce()
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		s.runOnce()
	}
}

func (s *GradeSubscriptionService) runOnce() {
	log.Printf("[GradeSubscription] Running grade check cycle...")
	ctx := context.Background()

	var subscriptions []database.GradeSubscription
	if err := s.db.Where("enabled = ?", true).Find(&subscriptions).Error; err != nil {
		log.Printf("[GradeSubscription] Failed to query subscriptions: %v", err)
		return
	}

	if len(subscriptions) == 0 {
		log.Printf("[GradeSubscription] No active subscriptions")
		return
	}

	log.Printf("[GradeSubscription] Processing %d active subscriptions", len(subscriptions))
	successCount := 0
	failCount := 0

	for _, sub := range subscriptions {
		if err := s.checkUserGrades(ctx, &sub); err != nil {
			log.Printf("[GradeSubscription] User %s check failed: %v", sub.UserID, err)
			failCount++
		} else {
			successCount++
		}
		// Sleep briefly between users to avoid overwhelming the JWXT server
		time.Sleep(2 * time.Second)
	}

	log.Printf("[GradeSubscription] Check cycle completed: %d success, %d failed", successCount, failCount)
}

func (s *GradeSubscriptionService) checkUserGrades(_ context.Context, sub *database.GradeSubscription) error {
	// Get user info
	var user database.User
	if err := s.db.Where("id = ?", sub.UserID).First(&user).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if user.JWXTUsername == nil || user.JWXTPassword == nil || user.Email == nil || *user.Email == "" {
		return fmt.Errorf("user missing jwxt credentials or email")
	}

	// Login to JWXT
	sess, err := s.jwxtSvc.Login(*user.JWXTUsername, *user.JWXTPassword)
	if err != nil {
		return fmt.Errorf("jwxt login failed: %w", err)
	}

	// Get current semester
	semesterData, err := s.jwxtSvc.GetSemester(sess)
	if err != nil {
		return fmt.Errorf("failed to get semester: %w", err)
	}

	currentSemesterID := ""
	semesters := []map[string]any{}
	if sd, ok := semesterData["semesters"].([]map[string]any); ok {
		semesters = sd
	}
	if cs, ok := semesterData["current_semester_id"].(string); ok {
		currentSemesterID = cs
	}
	if currentSemesterID == "" && len(semesters) > 0 {
		for _, sem := range semesters {
			if cur, ok := sem["current"].(bool); ok && cur {
				if id, ok := sem["id"].(string); ok {
					currentSemesterID = id
					break
				}
			}
		}
	}
	if currentSemesterID == "" {
		return fmt.Errorf("no current semester found")
	}

	// Update subscription semester ID
	s.db.Model(sub).Update("semester_id", currentSemesterID)

	// Get grades
	gradeData, err := s.jwxtSvc.GetGrade(sess, currentSemesterID)
	if err != nil {
		return fmt.Errorf("failed to get grades: %w", err)
	}

	gradesRaw, ok := gradeData["grades"]
	if !ok {
		return fmt.Errorf("no grades in response")
	}

	// Serialize grades to JSON for hash comparison
	gradesJSON, err := json.Marshal(gradesRaw)
	if err != nil {
		return fmt.Errorf("failed to serialize grades: %w", err)
	}

	newHash := fmt.Sprintf("%x", sha256.Sum256(gradesJSON))

	// Check if grades changed
	if sub.LastGradeHash != nil && *sub.LastGradeHash == newHash {
		log.Printf("[GradeSubscription] User %s: no grade changes", sub.UserID)
		s.db.Model(sub).Updates(map[string]any{
			"last_checked_at": time.Now(),
			"updated_at":      time.Now(),
		})
		return nil
	}

	// Grades changed! Determine what's new
	semesterName := currentSemesterID
	for _, sem := range semesters {
		if id, ok := sem["id"].(string); ok && id == currentSemesterID {
			if name, ok := sem["name"].(string); ok {
				semesterName = name
			}
			break
		}
	}

	// Build grade table HTML for the email
	gradeTableHTML := s.buildGradeTableHTML(gradesRaw)

	// Calculate change count
	changeCount := len(gradesRaw.([]any))
	if sub.LastGradeHash != nil {
		// If we have previous data, we just report total courses as updated
		// In a more sophisticated implementation, we could diff the two
		changeCount = len(gradesRaw.([]any))
	} else {
		// First check - notify about all current grades
		changeCount = len(gradesRaw.([]any))
	}

	// Send notification email
	realName := "同学"
	if user.RealName != nil && *user.RealName != "" {
		realName = *user.RealName
	}

	if err := s.mailSvc.SendGradeNotification(*user.Email, realName, semesterName, gradeTableHTML, changeCount); err != nil {
		log.Printf("[GradeSubscription] Failed to send email to %s: %v", *user.Email, err)
	} else {
		log.Printf("[GradeSubscription] Sent grade notification to %s (%d changes)", *user.Email, changeCount)
	}

	// Update subscription
	now := time.Now()
	s.db.Model(sub).Updates(map[string]any{
		"last_checked_at":  &now,
		"last_grade_hash":  &newHash,
		"last_notified_at": &now,
		"total_notified":   sub.TotalNotified + 1,
		"updated_at":       now,
	})

	return nil
}

func (s *GradeSubscriptionService) buildGradeTableHTML(gradesRaw any) string {
	grades, ok := gradesRaw.([]any)
	if !ok || len(grades) == 0 {
		return "<p style='color: #6b7280;'>暂无成绩数据</p>"
	}

	html := `<table style="width: 100%; border-collapse: collapse; font-size: 14px;">
		<thead>
			<tr style="background-color: #f3f4f6;">
				<th style="padding: 10px; text-align: left; border-bottom: 2px solid #e5e7eb;">课程名称</th>
				<th style="padding: 10px; text-align: center; border-bottom: 2px solid #e5e7eb;">学分</th>
				<th style="padding: 10px; text-align: center; border-bottom: 2px solid #e5e7eb;">成绩</th>
			</tr>
		</thead>
		<tbody>`

	var rows strings.Builder
	for i, g := range grades {
		row, ok := g.(map[string]any)
		if !ok {
			continue
		}

		courseName := formatValue(row["课程名称"], row["课程"])
		credits := formatValue(row["学分"])
		score := formatScore(row)

		bgColor := ""
		if i%2 == 1 {
			bgColor = " background-color: #f9fafb;"
		}

		scoreColor := "#1f2937"
		if score != "" {
			scoreFloat := parseFloatStr(score)
			if scoreFloat >= 90 {
				scoreColor = "#059669"
			} else if scoreFloat >= 80 {
				scoreColor = "#2563eb"
			} else if scoreFloat >= 60 {
				scoreColor = "#d97706"
			} else if scoreFloat > 0 {
				scoreColor = "#dc2626"
			}
		}

		fmt.Fprintf(&rows, `
			<tr style="%s">
				<td style="padding: 10px; border-bottom: 1px solid #e5e7eb;">%s</td>
				<td style="padding: 10px; text-align: center; border-bottom: 1px solid #e5e7eb;">%s</td>
				<td style="padding: 10px; text-align: center; border-bottom: 1px solid #e5e7eb; color: %s; font-weight: bold;">%s</td>
			</tr>`,
			bgColor, courseName, credits, scoreColor, score)
	}

	html += rows.String() + `</tbody></table>`
	return html
}

func formatValue(values ...any) string {
	for _, v := range values {
		if v == nil {
			continue
		}
		s := strings.TrimSpace(fmt.Sprintf("%v", v))
		if s != "" && s != "<nil>" {
			return s
		}
	}
	return "-"
}

func formatScore(row map[string]any) string {
	fields := []string{"最终成绩", "总评成绩", "成绩", "总评"}
	for _, key := range fields {
		if v, ok := row[key]; ok {
			s := strings.TrimSpace(fmt.Sprintf("%v", v))
			if s != "" && s != "<nil>" {
				return s
			}
		}
	}
	return "-"
}

func parseFloatStr(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return -1
	}
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	if err != nil {
		return -1
	}
	return f
}
