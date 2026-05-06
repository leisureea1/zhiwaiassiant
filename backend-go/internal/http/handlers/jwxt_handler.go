package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"xisu/backend-go/internal/database"
	"xisu/backend-go/internal/http/middleware"
	"xisu/backend-go/internal/http/response"
	"xisu/backend-go/internal/service"
)

const (
	jwxtSessionTTL       = 50 * time.Minute
	jwxtValidateInterval = 10 * time.Minute
)

type JWXTHandler struct {
	db      *gorm.DB
	service *service.JwxtDirectService
}

func NewJWXTHandler(db *gorm.DB, svc *service.JwxtDirectService) *JWXTHandler {
	return &JWXTHandler{db: db, service: svc}
}

func (h *JWXTHandler) Course(c *gin.Context) {
	userID := userIDFromContext(c)
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "invalid user context")
		return
	}

	sess, err := h.getOrCreateSession(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	data, err := h.service.GetCourse(sess, c.Query("semester_id"), "")
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error())
		return
	}

	_ = h.service.SaveSession(context.Background(), userID, sess, jwxtSessionTTL)
	respondNestSuccess(c, toLegacyJwxtEnvelope(data))
}

func (h *JWXTHandler) CourseRefresh(c *gin.Context) {
	userID := userIDFromContext(c)
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "invalid user context")
		return
	}

	h.service.ClearSession(context.Background(), userID)
	sess, err := h.getOrCreateSession(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.service.GetCourse(sess, c.Query("semester_id"), "")
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error())
		return
	}
	_ = h.service.SaveSession(context.Background(), userID, sess, jwxtSessionTTL)
	respondNestSuccess(c, toLegacyJwxtEnvelope(data))
}

func toLegacyJwxtEnvelope(data map[string]any) gin.H {
	wrapped := gin.H{
		"success": true,
		"data":    data,
	}

	if success, ok := data["success"].(bool); ok {
		wrapped["success"] = success
	}

	if errMsg, ok := data["error"]; ok {
		wrapped["error"] = errMsg
	}

	return wrapped
}

func respondNestSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "ok",
		"data":      data,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *JWXTHandler) Grade(c *gin.Context) {
	userID := userIDFromContext(c)
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "invalid user context")
		return
	}
	sess, err := h.getOrCreateSession(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.service.GetGrade(sess, c.Query("semester_id"))
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error())
		return
	}
	_ = h.service.SaveSession(context.Background(), userID, sess, jwxtSessionTTL)
	respondNestSuccess(c, toLegacyJwxtEnvelope(data))
}

func (h *JWXTHandler) Exam(c *gin.Context) {
	userID := userIDFromContext(c)
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "invalid user context")
		return
	}
	sess, err := h.getOrCreateSession(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.service.GetExam(sess, c.Query("semester_id"))
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error())
		return
	}
	if success, ok := data["success"].(bool); ok && !success {
		h.service.ClearSession(context.Background(), userID)
		sess, err = h.getOrCreateSession(c.Request.Context(), userID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		data, err = h.service.GetExam(sess, c.Query("semester_id"))
		if err != nil {
			response.Error(c, http.StatusBadGateway, err.Error())
			return
		}
	}
	success := true
	if v, ok := data["success"].(bool); ok {
		success = v
	}
	if success {
		_ = h.service.SaveSession(context.Background(), userID, sess, jwxtSessionTTL)
	} else {
		h.service.ClearSession(context.Background(), userID)
	}

	// 包装成前端期望的格式 - 匹配 Python backend 的响应结构
	exams, ok := data["exams"].([]map[string]any)
	if !ok {
		exams = []map[string]any{}
	}

	innerData := gin.H{
		"success":     success,
		"exams":       exams,
		"total":       len(exams),
		"semester_id": data["semester"],
	}
	if errMsg, ok := data["error"]; ok {
		innerData["error"] = errMsg
	}

	outerData := gin.H{
		"success": success,
		"data":    innerData,
	}
	if errMsg, ok := data["error"]; ok {
		outerData["error"] = errMsg
	}

	respondNestSuccess(c, outerData)
}

func (h *JWXTHandler) Semester(c *gin.Context) {
	userID := userIDFromContext(c)
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "invalid user context")
		return
	}
	sess, err := h.getOrCreateSession(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.service.GetSemester(sess)
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error())
		return
	}
	_ = h.service.SaveSession(context.Background(), userID, sess, jwxtSessionTTL)
	respondNestSuccess(c, toLegacyJwxtEnvelope(data))
}

func (h *JWXTHandler) User(c *gin.Context) {
	userID := userIDFromContext(c)
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "invalid user context")
		return
	}
	sess, err := h.getOrCreateSession(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.service.GetUser(sess)
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error())
		return
	}
	_ = h.service.SaveSession(context.Background(), userID, sess, jwxtSessionTTL)
	response.OK(c, data)
}

func (h *JWXTHandler) EvaluationPending(c *gin.Context) {
	userID := userIDFromContext(c)
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "invalid user context")
		return
	}
	sess, err := h.getOrCreateSession(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.service.GetEvaluationPending(sess)
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error())
		return
	}
	_ = h.service.SaveSession(context.Background(), userID, sess, jwxtSessionTTL)
	respondNestSuccess(c, toLegacyJwxtEnvelope(data))
}

func (h *JWXTHandler) EvaluationAuto(c *gin.Context) {
	userID := userIDFromContext(c)
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "invalid user context")
		return
	}
	sess, err := h.getOrCreateSession(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := h.service.AutoEvaluation(sess)
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error())
		return
	}
	_ = h.service.SaveSession(context.Background(), userID, sess, jwxtSessionTTL)
	respondNestSuccess(c, toLegacyJwxtEnvelope(data))
}

func (h *JWXTHandler) Bind(c *gin.Context) {
	userID := userIDFromContext(c)
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "invalid user context")
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" {
		response.Error(c, http.StatusBadRequest, "username/password required")
		return
	}

	sess, err := h.service.Login(strings.TrimSpace(req.Username), strings.TrimSpace(req.Password))
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	updates := map[string]any{
		"jwxt_username": req.Username,
		"jwxt_password": req.Password,
	}
	if info, infoErr := h.service.GetUser(sess); infoErr == nil {
		if v, ok := pickNonEmptyString(info, "name"); ok {
			updates["real_name"] = v
		}
		if v, ok := pickNonEmptyString(info, "department"); ok {
			updates["college"] = v
		}
		if v, ok := pickNonEmptyString(info, "major"); ok {
			updates["major"] = v
		}
		if v, ok := pickNonEmptyString(info, "class_name"); ok {
			updates["class_name"] = v
		}
	}

	if err := h.db.Model(&database.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "保存绑定信息失败")
		return
	}

	_ = h.service.SaveSession(context.Background(), userID, sess, jwxtSessionTTL)
	response.OK(c, gin.H{"message": "绑定成功"})
}

func (h *JWXTHandler) Unbind(c *gin.Context) {
	userID := userIDFromContext(c)
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "invalid user context")
		return
	}
	if err := h.db.Model(&database.User{}).Where("id = ?", userID).Updates(map[string]any{
		"jwxt_username": nil,
		"jwxt_password": nil,
	}).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "解绑失败")
		return
	}
	h.service.ClearSession(context.Background(), userID)
	response.OK(c, gin.H{"message": "解绑成功"})
}

func (h *JWXTHandler) getOrCreateSession(ctx context.Context, userID string) (*service.CachedJWXTSession, error) {
	sess, err := h.service.LoadSession(ctx, userID)
	if err == nil && sess != nil {
		if sess.ValidatedAt > 0 && time.Since(time.UnixMilli(sess.ValidatedAt)) < jwxtValidateInterval {
			return sess, nil
		}
		if h.service.ValidateSession(sess) {
			sess.ValidatedAt = time.Now().UnixMilli()
			return sess, nil
		}
	}

	var user database.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	if user.JWXTUsername == nil || user.JWXTPassword == nil || strings.TrimSpace(*user.JWXTUsername) == "" || strings.TrimSpace(*user.JWXTPassword) == "" {
		return nil, fmt.Errorf("请先绑定教务系统账号")
	}

	newsess, err := h.service.Login(*user.JWXTUsername, *user.JWXTPassword)
	if err != nil {
		return nil, err
	}
	newsess.ValidatedAt = time.Now().UnixMilli()
	return newsess, nil
}

func userIDFromContext(c *gin.Context) string {
	v, _ := c.Get(middleware.ContextUserID)
	if v == nil {
		return ""
	}
	id, _ := v.(string)
	return id
}

func pickNonEmptyString(data map[string]any, key string) (string, bool) {
	if data == nil {
		return "", false
	}
	v, ok := data[key]
	if !ok || v == nil {
		return "", false
	}
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	if s == "" || strings.EqualFold(s, "<nil>") {
		return "", false
	}
	return s, true
}
