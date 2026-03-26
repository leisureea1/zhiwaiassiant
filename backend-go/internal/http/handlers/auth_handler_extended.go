package handlers

import (
	"context"
	crand "crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"xisu/backend-go/internal/database"
	"xisu/backend-go/internal/http/middleware"
	"xisu/backend-go/internal/http/response"
	"xisu/backend-go/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	VerificationCodePrefix = "verify_code:"
	EmailVerifiedPrefix    = "email_verified:"
	ResetPasswordPrefix    = "reset_pwd_code:"
	VerifyAttemptPrefix    = "verify_attempt:"
	ResetAttemptPrefix     = "reset_attempt:"
	MaxCodeAttempts        = 5
	minPasswordLength      = 8
)

var passwordLowercaseRegex = regexp.MustCompile(`^.*[a-z].*$`)
var passwordUppercaseRegex = regexp.MustCompile(`^.*[A-Z].*$`)
var passwordDigitRegex = regexp.MustCompile(`^.*\d.*$`)

func validatePasswordComplexity(password string) string {
	if len(password) < minPasswordLength {
		return fmt.Sprintf("密码长度至少为%d位", minPasswordLength)
	}
	if !passwordLowercaseRegex.MatchString(password) {
		return "密码必须包含至少一个小写字母"
	}
	if !passwordUppercaseRegex.MatchString(password) {
		return "密码必须包含至少一个大写字母"
	}
	if !passwordDigitRegex.MatchString(password) {
		return "密码必须包含至少一个数字"
	}
	return ""
}

type resetTokenClaims struct {
	Email string `json:"email"`
	Use   string `json:"use"`
	jwt.RegisteredClaims
}

type ExtendedAuthHandler struct {
	db          *gorm.DB
	tokenSvc    *service.TokenService
	mailSvc     *service.MailService
	jwxtSvc     *service.JwxtDirectService
	redisClient *redis.Client
	storeMu     sync.RWMutex
	codeStore   map[string]memoryCodeEntry
}

type memoryCodeEntry struct {
	value     string
	expiresAt time.Time
}

func NewExtendedAuthHandler(db *gorm.DB, tokenSvc *service.TokenService, mailSvc *service.MailService, jwxtSvc *service.JwxtDirectService, redisClient *redis.Client) *ExtendedAuthHandler {
	return &ExtendedAuthHandler{
		db:          db,
		tokenSvc:    tokenSvc,
		mailSvc:     mailSvc,
		jwxtSvc:     jwxtSvc,
		redisClient: redisClient,
		codeStore:   make(map[string]memoryCodeEntry),
	}
}

func (h *ExtendedAuthHandler) setCode(ctx context.Context, key, value string, ttl time.Duration) error {
	if h.redisClient != nil {
		if err := h.redisClient.Set(ctx, key, value, ttl).Err(); err == nil {
			return nil
		}
	}

	h.storeMu.Lock()
	h.codeStore[key] = memoryCodeEntry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
	h.storeMu.Unlock()
	return nil
}

func (h *ExtendedAuthHandler) ttlCode(ctx context.Context, key string) (time.Duration, error) {
	if h.redisClient != nil {
		if ttl, err := h.redisClient.TTL(ctx, key).Result(); err == nil {
			return ttl, nil
		}
	}

	h.storeMu.RLock()
	entry, ok := h.codeStore[key]
	h.storeMu.RUnlock()
	if !ok {
		return -2 * time.Second, nil
	}

	remaining := time.Until(entry.expiresAt)
	if remaining <= 0 {
		h.storeMu.Lock()
		delete(h.codeStore, key)
		h.storeMu.Unlock()
		return -2 * time.Second, nil
	}

	return remaining, nil
}

func (h *ExtendedAuthHandler) getCode(ctx context.Context, key string) (string, error) {
	if h.redisClient != nil {
		if code, err := h.redisClient.Get(ctx, key).Result(); err == nil {
			return code, nil
		} else if err == redis.Nil {
			return "", redis.Nil
		}
	}

	h.storeMu.RLock()
	entry, ok := h.codeStore[key]
	h.storeMu.RUnlock()
	if !ok {
		return "", redis.Nil
	}
	if time.Now().After(entry.expiresAt) {
		h.storeMu.Lock()
		delete(h.codeStore, key)
		h.storeMu.Unlock()
		return "", redis.Nil
	}

	return entry.value, nil
}

func (h *ExtendedAuthHandler) delCode(ctx context.Context, key string) {
	if h.redisClient != nil {
		h.redisClient.Del(ctx, key)
	}
	h.storeMu.Lock()
	delete(h.codeStore, key)
	h.storeMu.Unlock()
}

func (h *ExtendedAuthHandler) getCounter(ctx context.Context, key string) int {
	value, err := h.getCode(ctx, key)
	if err != nil {
		return 0
	}
	count, _ := strconv.Atoi(value)
	return count
}

func (h *ExtendedAuthHandler) increaseCounter(ctx context.Context, key string, ttl time.Duration) int {
	count := h.getCounter(ctx, key) + 1
	_ = h.setCode(ctx, key, strconv.Itoa(count), ttl)
	return count
}

func (h *ExtendedAuthHandler) resetCounter(ctx context.Context, key string) {
	h.delCode(ctx, key)
}

func getResetTokenSecret() string {
	if secret := strings.TrimSpace(os.Getenv("JWT_RESET_SECRET")); secret != "" {
		return secret
	}
	if secret := strings.TrimSpace(os.Getenv("JWT_REFRESH_SECRET")); secret != "" {
		return secret
	}
	return strings.TrimSpace(os.Getenv("JWT_SECRET"))
}

func generateResetToken(email string) (string, error) {
	secret := getResetTokenSecret()
	if len(secret) < 16 {
		return "", fmt.Errorf("reset token secret not configured")
	}

	claims := resetTokenClaims{
		Email: email,
		Use:   "password-reset",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.NewString(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func parseResetToken(tokenString string) (*resetTokenClaims, error) {
	secret := getResetTokenSecret()
	if len(secret) < 16 {
		return nil, fmt.Errorf("reset token secret not configured")
	}

	token, err := jwt.ParseWithClaims(tokenString, &resetTokenClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*resetTokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid reset token")
	}
	if claims.Use != "password-reset" {
		return nil, fmt.Errorf("invalid reset token use")
	}
	if strings.TrimSpace(claims.Email) == "" {
		return nil, fmt.Errorf("invalid reset token email")
	}

	return claims, nil
}

type SendVerificationCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *ExtendedAuthHandler) SendVerificationCode(c *gin.Context) {
	var req SendVerificationCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request")
		return
	}

	ctx := context.Background()
	redisKey := VerificationCodePrefix + req.Email

	// 检查邮箱是否已注册，但不暴露结果以防用户枚举
	var count int64
	emailAlreadyRegistered := false
	if err := h.db.Model(&database.User{}).Where("email = ?", req.Email).Count(&count).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "database error")
		return
	}
	if count > 0 {
		emailAlreadyRegistered = true
	}

	// 检查是否频繁发送
	ttl, err := h.ttlCode(ctx, redisKey)
	if err == nil && ttl > 540*time.Second {
		remaining := int((ttl - 540*time.Second).Seconds())
		if remaining < 1 {
			remaining = 1
		}
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("请勿频繁发送验证码，请%d秒后再试", remaining))
		return
	}

	// 生成验证码
	code := generateCode()

	// 保存到 Redis (10分钟)
	if err := h.setCode(ctx, redisKey, code, 10*time.Minute); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to save verification code")
		return
	}

	// 发送邮件（仅未注册的邮箱才实际发送）
	if !emailAlreadyRegistered {
		if err := h.mailSvc.SendVerificationCode(req.Email, code); err != nil {
			if isMailServiceUnavailable(err) {
				response.Error(c, http.StatusServiceUnavailable, "邮件服务未配置或暂时不可用")
				return
			}
			response.Error(c, http.StatusInternalServerError, "验证码发送失败，请稍后再试")
			return
		}
	}

	if strings.ToLower(os.Getenv("APP_ENV")) != "production" {
		response.OK(c, gin.H{"message": "验证码已发送，请查收邮件"})
		return
	}

	response.OK(c, gin.H{"message": "验证码已发送，请查收邮件"})
}

type VerifyEmailCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,numeric,len=6"`
}

func (h *ExtendedAuthHandler) VerifyEmailCode(c *gin.Context) {
	var req VerifyEmailCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request")
		return
	}

	ctx := context.Background()
	redisKey := VerificationCodePrefix + req.Email

	storedCode, err := h.getCode(ctx, redisKey)
	if err == redis.Nil {
		response.Error(c, http.StatusBadRequest, "验证码不存在或已过期，请重新发送")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get verification code")
		return
	}

	if storedCode != req.Code {
		response.Error(c, http.StatusBadRequest, "验证码错误")
		return
	}

	// 验证成功，删除验证码
	h.delCode(ctx, redisKey)

	// 生成一次性邮箱验证 token，存入 Redis（30 分钟有效）
	verifyToken := uuid.NewString()
	verifyKey := EmailVerifiedPrefix + req.Email
	_ = h.setCode(ctx, verifyKey, verifyToken, 30*time.Minute)

	response.OK(c, gin.H{
		"verified":   true,
		"emailToken": verifyToken,
	})
}

type RegisterRequest struct {
	Username      string  `json:"username" binding:"required,min=3"`
	Password      string  `json:"password" binding:"required,min=8"`
	Email         string  `json:"email" binding:"required,email"`
	StudentID     string  `json:"studentId" binding:"required"`
	XiwaiPassword string  `json:"xiwaiPassword" binding:"required"`
	EmailToken    string  `json:"emailToken" binding:"required"`
	Avatar        *string `json:"avatar"`
}

func (h *ExtendedAuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request")
		return
	}

	// 验证用户名格式
	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(req.Username) {
		response.Error(c, http.StatusBadRequest, "用户名只能包含字母、数字和下划线")
		return
	}

	// 验证密码复杂度
	if msg := validatePasswordComplexity(req.Password); msg != "" {
		response.Error(c, http.StatusBadRequest, msg)
		return
	}

	// 校验邮箱验证 token
	ctx := context.Background()
	verifyKey := EmailVerifiedPrefix + req.Email
	storedToken, err := h.getCode(ctx, verifyKey)
	if err != nil || storedToken != req.EmailToken {
		response.Error(c, http.StatusUnauthorized, "邮箱未验证或验证已过期，请重新验证")
		return
	}
	h.delCode(ctx, verifyKey) // 一次性使用，验证后立即删除

	if h.jwxtSvc == nil {
		response.Error(c, http.StatusServiceUnavailable, "教务系统服务暂时不可用")
		return
	}

	studentID := strings.TrimSpace(req.StudentID)
	xiwaiPassword := strings.TrimSpace(req.XiwaiPassword)
	sess, err := h.jwxtSvc.Login(studentID, xiwaiPassword)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "教务系统验证失败，请检查学号和密码是否正确")
		return
	}

	var realName, college, major, className *string
	if info, infoErr := h.jwxtSvc.GetUser(sess); infoErr == nil {
		realName = mapValueStringPtr(info, "name")
		college = mapValueStringPtr(info, "department")
		major = mapValueStringPtr(info, "major")
		className = mapValueStringPtr(info, "class_name")
	}

	// 检查用户名、邮箱、学号是否已存在
	var count int64
	if err := h.db.Model(&database.User{}).Where("username = ? OR email = ? OR student_id = ?",
		req.Username, req.Email, studentID).Count(&count).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "database error")
		return
	}
	if count > 0 {
		response.Error(c, http.StatusConflict, "用户名、邮箱或学号已存在")
		return
	}

	// 密码加密
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	// 创建用户
	user := database.User{
		ID:           uuid.NewString(),
		Username:     req.Username,
		Email:        &req.Email,
		PasswordHash: string(passwordHash),
		StudentID:    &studentID,
		RealName:     realName,
		College:      college,
		Major:        major,
		ClassName:    className,
		Avatar:       req.Avatar,
		Role:         "USER",
		Status:       "ACTIVE",
		JWXTUsername: &studentID,
		JWXTPassword: &xiwaiPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.db.Create(&user).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to create user")
		return
	}

	// 生成令牌
	accessToken, accessExpiresAt, err := h.tokenSvc.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to generate access token")
		return
	}

	refreshToken, refreshExpiresAt, err := h.tokenSvc.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to generate refresh token")
		return
	}
	refreshDigest := refreshTokenDigest(refreshToken)

	// 保存刷新令牌
	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")
	refreshRow := database.RefreshToken{
		ID:        uuid.NewString(),
		Token:     refreshDigest,
		UserID:    user.ID,
		UserAgent: &ua,
		IPAddress: &ip,
		ExpiresAt: refreshExpiresAt,
		CreatedAt: time.Now(),
	}
	if err := h.db.Create(&refreshRow).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to persist refresh token")
		return
	}

	response.OK(c, gin.H{
		"accessToken":      accessToken,
		"refreshToken":     refreshToken,
		"accessExpiresAt":  accessExpiresAt,
		"refreshExpiresAt": refreshExpiresAt,
		"user": gin.H{
			"id":        user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"studentId": user.StudentID,
			"realName":  user.RealName,
			"avatar":    user.Avatar,
			"role":      user.Role,
			"college":   user.College,
			"major":     user.Major,
			"className": user.ClassName,
			"jwxtBound": true,
		},
	})
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}

func (h *ExtendedAuthHandler) ChangePassword(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextUserID)
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request")
		return
	}

	if msg := validatePasswordComplexity(req.NewPassword); msg != "" {
		response.Error(c, http.StatusBadRequest, msg)
		return
	}

	var user database.User
	if err := h.db.Where("id = ?", userID.(string)).First(&user).Error; err != nil {
		response.Error(c, http.StatusNotFound, "user not found")
		return
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		response.Error(c, http.StatusBadRequest, "旧密码错误")
		return
	}

	// 加密新密码
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	// 更新密码
	if err := h.db.Model(&user).Update("password_hash", string(newPasswordHash)).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to update password")
		return
	}

	// 删除所有刷新令牌
	h.db.Delete(&database.RefreshToken{}, "user_id = ?", userID.(string))

	response.OK(c, gin.H{"message": "密码修改成功，请重新登录"})
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *ExtendedAuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	ctx := context.Background()
	redisKey := ResetPasswordPrefix + req.Email
	resetToken, err := generateResetToken(req.Email)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to generate reset token")
		return
	}
	genericSuccess := gin.H{"message": "若邮箱已注册，验证码与重置令牌已发送", "token": resetToken}

	// 检查邮箱是否已注册
	var user database.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.OK(c, genericSuccess)
			return
		}
		response.Error(c, http.StatusInternalServerError, "database error")
		return
	}

	// 检查是否频繁发送
	ttl, err := h.ttlCode(ctx, redisKey)
	if err == nil && ttl > 540*time.Second {
		response.OK(c, genericSuccess)
		return
	}

	// 生成验证码
	code := generateCode()

	// 保存到 Redis (10分钟)
	if err := h.setCode(ctx, redisKey, code, 10*time.Minute); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to save verification code")
		return
	}
	h.resetCounter(ctx, ResetAttemptPrefix+req.Email)

	// 发送邮件
	if err := h.mailSvc.SendPasswordReset(req.Email, code); err != nil {
		if isMailServiceUnavailable(err) {
			response.Error(c, http.StatusServiceUnavailable, "邮件服务未配置或暂时不可用")
			return
		}
		response.Error(c, http.StatusInternalServerError, "验证码发送失败，请稍后再试")
		return
	}

	response.OK(c, genericSuccess)
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	Code        string `json:"code" binding:"required,len=6"`
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}

func (h *ExtendedAuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request")
		return
	}

	if msg := validatePasswordComplexity(req.NewPassword); msg != "" {
		response.Error(c, http.StatusBadRequest, msg)
		return
	}

	tokenData, err := parseResetToken(req.Token)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的重置令牌")
		return
	}

	ctx := context.Background()
	redisKey := ResetPasswordPrefix + tokenData.Email
	attemptKey := ResetAttemptPrefix + tokenData.Email
	req.Code = strings.TrimSpace(req.Code)

	if attempts := h.getCounter(ctx, attemptKey); attempts >= MaxCodeAttempts {
		response.Error(c, http.StatusTooManyRequests, "验证码错误次数过多，请重新获取")
		return
	}

	// 验证验证码
	storedCode, err := h.getCode(ctx, redisKey)
	if err == redis.Nil {
		response.Error(c, http.StatusBadRequest, "验证码不存在或已过期，请重新获取")
		return
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get verification code")
		return
	}

	if storedCode != req.Code {
		if h.increaseCounter(ctx, attemptKey, 10*time.Minute) >= MaxCodeAttempts {
			response.Error(c, http.StatusTooManyRequests, "验证码错误次数过多，请重新获取")
			return
		}
		response.Error(c, http.StatusBadRequest, "验证码错误")
		return
	}
	h.resetCounter(ctx, attemptKey)

	// 查找用户
	var user database.User
	if err := h.db.Where("email = ?", tokenData.Email).First(&user).Error; err != nil {
		response.Error(c, http.StatusBadRequest, "重置失败，请重新获取验证码")
		return
	}

	// 更新密码
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	if err := h.db.Model(&user).Update("password_hash", string(newPasswordHash)).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to update password")
		return
	}

	// 删除验证码
	h.delCode(ctx, redisKey)
	h.resetCounter(ctx, attemptKey)

	// 删除所有刷新令牌
	h.db.Delete(&database.RefreshToken{}, "user_id = ?", user.ID)

	response.OK(c, gin.H{"message": "密码重置成功，请使用新密码登录"})
}

func generateCode() string {
	n, err := crand.Int(crand.Reader, big.NewInt(1000000))
	if err != nil {
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	return fmt.Sprintf("%06d", n.Int64())
}

func isMailServiceUnavailable(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "smtp") || strings.Contains(msg, "connection")
}

func mapValueStringPtr(m map[string]any, key string) *string {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	if s == "" {
		return nil
	}
	return &s
}
