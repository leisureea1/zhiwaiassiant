package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"xisu/backend-go/internal/database"
	"xisu/backend-go/internal/service/jwxt"
)

const (
	maxConcurrentChecks = 3
	maxRetries          = 2
	retryDelay          = 5 * time.Second
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

// RunOnce triggers a single check cycle. Safe to call from multiple goroutines.
func (s *GradeSubscriptionService) RunOnce() {
	s.runOnce()
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

	log.Printf("[GradeSubscription] Processing %d active subscriptions (concurrency: %d)", len(subscriptions), maxConcurrentChecks)

	var (
		mu          sync.Mutex
		successCnt  int
		failCnt     int
		wg          sync.WaitGroup
		sem         = make(chan struct{}, maxConcurrentChecks)
	)

	for i := range subscriptions {
		wg.Add(1)
		sem <- struct{}{} // acquire

		go func(sub *database.GradeSubscription) {
			defer wg.Done()
			defer func() { <-sem }() // release

			var err error
			for attempt := 0; attempt <= maxRetries; attempt++ {
				if attempt > 0 {
					log.Printf("[GradeSubscription] User %s: retry %d/%d", sub.UserID, attempt, maxRetries)
					time.Sleep(retryDelay)
				}
				err = s.checkUserGrades(ctx, sub)
				if err == nil {
					break
				}
			}

			mu.Lock()
			if err != nil {
				log.Printf("[GradeSubscription] User %s check failed after retries: %v", sub.UserID, err)
				failCnt++
			} else {
				successCnt++
			}
			mu.Unlock()
		}(&subscriptions[i])
	}

	wg.Wait()
	log.Printf("[GradeSubscription] Check cycle completed: %d success, %d failed", successCnt, failCnt)
}

// getOrCreateSession tries to reuse cached session from Redis first, only logs in if needed.
func (s *GradeSubscriptionService) getOrCreateSession(ctx context.Context, userID, username, password string) (*jwxt.CachedJWXTSession, error) {
	// Try loading cached session
	sess, err := s.jwxtSvc.LoadSession(ctx, userID)
	if err == nil && sess != nil {
		// Check if session is still valid (validated within last 30 minutes)
		if sess.ValidatedAt > 0 && time.Since(time.UnixMilli(sess.ValidatedAt)) < 30*time.Minute {
			log.Printf("[GradeSubscription] User %s: using cached session", userID)
			return sess, nil
		}
		// Validate the cached session
		if s.jwxtSvc.ValidateSession(sess) {
			sess.ValidatedAt = time.Now().UnixMilli()
			// Update validated timestamp in Redis
			_ = s.jwxtSvc.SaveSession(ctx, userID, sess, 50*time.Minute)
			log.Printf("[GradeSubscription] User %s: cached session validated", userID)
			return sess, nil
		}
		log.Printf("[GradeSubscription] User %s: cached session expired, logging in", userID)
	}

	// No valid cache, login
	newSess, err := s.jwxtSvc.Login(username, password)
	if err != nil {
		return nil, fmt.Errorf("jwxt login failed: %w", err)
	}
	newSess.ValidatedAt = time.Now().UnixMilli()

	// Save to Redis so normal requests can reuse it
	_ = s.jwxtSvc.SaveSession(ctx, userID, newSess, 50*time.Minute)
	log.Printf("[GradeSubscription] User %s: logged in and cached session", userID)

	return newSess, nil
}

func (s *GradeSubscriptionService) checkUserGrades(ctx context.Context, sub *database.GradeSubscription) error {
	// Get user info
	var user database.User
	if err := s.db.Where("id = ?", sub.UserID).First(&user).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if user.JWXTUsername == nil || user.JWXTPassword == nil || user.Email == nil || *user.Email == "" {
		return fmt.Errorf("user missing jwxt credentials or email")
	}

	// Get or create session (reuse cached, login only if needed)
	sess, err := s.getOrCreateSession(ctx, sub.UserID, *user.JWXTUsername, *user.JWXTPassword)
	if err != nil {
		return err
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

	// First check: only record the hash, don't send notification
	if sub.LastGradeHash == nil {
		log.Printf("[GradeSubscription] User %s: first check, recording baseline hash", sub.UserID)
		s.db.Model(sub).Updates(map[string]any{
			"last_checked_at": time.Now(),
			"last_grade_hash": &newHash,
			"semester_id":     currentSemesterID,
			"updated_at":      time.Now(),
		})
		return nil
	}

	// Check if grades changed
	if *sub.LastGradeHash == newHash {
		log.Printf("[GradeSubscription] User %s: no grade changes", sub.UserID)
		s.db.Model(sub).Updates(map[string]any{
			"last_checked_at": time.Now(),
			"updated_at":      time.Now(),
		})
		return nil
	}

	// Grades changed! Find semester name
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
	changeCount := len(gradesRaw.([]map[string]any))

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
		"semester_id":      currentSemesterID,
		"updated_at":       now,
	})

	return nil
}

func (s *GradeSubscriptionService) buildGradeTableHTML(gradesRaw any) string {
	grades, ok := gradesRaw.([]map[string]any)
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
	for i, row := range grades {
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
