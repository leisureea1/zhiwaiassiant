package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"xisu/backend-go/internal/database"
	"xisu/backend-go/internal/http/middleware"
	"xisu/backend-go/internal/http/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadHandler struct {
	db        *gorm.DB
	uploadDir string
}

var allowedAttachmentExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".pdf":  true,
	".txt":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
	".ppt":  true,
	".pptx": true,
	".zip":  true,
	".rar":  true,
	".7z":   true,
}

var allowedAttachmentMIMEs = map[string]bool{
	"image/jpeg":                true,
	"image/png":                 true,
	"image/gif":                 true,
	"image/webp":                true,
	"application/pdf":           true,
	"text/plain; charset=utf-8": true,
	"application/msword":        true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
	"application/vnd.ms-powerpoint":                                             true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	"application/zip":              true,
	"application/x-rar-compressed": true,
	"application/vnd.rar":          true,
	"application/x-7z-compressed":  true,
}

func NewUploadHandler(db *gorm.DB, uploadDir string) *UploadHandler {
	// 确保上传目录存在
	os.MkdirAll(filepath.Join(uploadDir, "avatars"), 0755)
	os.MkdirAll(filepath.Join(uploadDir, "attachments"), 0755)

	return &UploadHandler{
		db:        db,
		uploadDir: uploadDir,
	}
}

func (h *UploadHandler) UploadAvatar(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextUserID)
	id := c.Param("id")

	// 只能上传自己的头像
	if userID.(string) != id {
		response.Error(c, http.StatusForbidden, "无权更新其他用户头像")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "请上传头像文件")
		return
	}

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	if !allowedExts[ext] {
		response.Error(c, http.StatusBadRequest, "仅支持 jpg/png/gif/webp 图片")
		return
	}

	detectedMime, err := detectMimeType(file)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无法识别文件类型")
		return
	}
	allowedImageMimes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	if !allowedImageMimes[detectedMime] {
		response.Error(c, http.StatusBadRequest, "头像文件内容类型不允许")
		return
	}

	// 验证文件大小 (5MB)
	if file.Size > 5*1024*1024 {
		response.Error(c, http.StatusBadRequest, "文件大小不能超过 5MB")
		return
	}

	// 生成文件名
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixMilli(), uuid.NewString(), ext)
	savePath := filepath.Join(h.uploadDir, "avatars", filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		response.Error(c, http.StatusInternalServerError, "文件保存失败")
		return
	}

	// 构建 URL
	protocol := "http"
	if c.Request.TLS != nil {
		protocol = "https"
	}
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		protocol = strings.Split(proto, ",")[0]
	}

	host := c.Request.Host
	if fwdHost := c.GetHeader("X-Forwarded-Host"); fwdHost != "" {
		host = strings.Split(fwdHost, ",")[0]
	}

	avatarURL := fmt.Sprintf("%s://%s/uploads/avatars/%s", protocol, host, filename)

	// 更新用户头像
	if err := h.db.Model(&database.User{}).Where("id = ?", id).Update("avatar", avatarURL).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "更新头像失败")
		return
	}

	var user database.User
	h.db.Where("id = ?", id).First(&user)

	response.OK(c, gin.H{
		"avatar": avatarURL,
		"user":   toSafeUser(user, false),
	})
}

func (h *UploadHandler) UploadFile(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextUserID)

	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "请上传文件")
		return
	}

	// 验证文件大小 (10MB)
	if file.Size > 10*1024*1024 {
		response.Error(c, http.StatusBadRequest, "文件大小不能超过 10MB")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedAttachmentExts[ext] {
		response.Error(c, http.StatusBadRequest, "不支持的文件类型")
		return
	}

	detectedMime, err := detectMimeType(file)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无法识别文件类型")
		return
	}
	if !allowedAttachmentMIMEs[detectedMime] {
		response.Error(c, http.StatusBadRequest, "文件内容类型不允许")
		return
	}

	// 生成文件名
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixMilli(), uuid.NewString(), ext)
	savePath := filepath.Join(h.uploadDir, "attachments", filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		response.Error(c, http.StatusInternalServerError, "文件保存失败")
		return
	}

	// 构建 URL
	protocol := "http"
	if c.Request.TLS != nil {
		protocol = "https"
	}
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		protocol = strings.Split(proto, ",")[0]
	}

	host := c.Request.Host
	if fwdHost := c.GetHeader("X-Forwarded-Host"); fwdHost != "" {
		host = strings.Split(fwdHost, ",")[0]
	}

	fileURL := fmt.Sprintf("%s://%s/uploads/attachments/%s", protocol, host, filename)

	// 记录上传文件
	uploadedFile := database.UploadedFile{
		ID:           uuid.NewString(),
		FileName:     filename,
		OriginalName: file.Filename,
		FilePath:     savePath,
		FileSize:     int(file.Size),
		MimeType:     detectedMime,
		UploaderID:   strPtr(userID.(string)),
		CreatedAt:    time.Now(),
	}

	if err := h.db.Create(&uploadedFile).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "记录文件信息失败")
		return
	}

	response.OK(c, gin.H{
		"id":           uploadedFile.ID,
		"url":          fileURL,
		"filename":     filename,
		"originalName": file.Filename,
		"size":         file.Size,
		"mimeType":     uploadedFile.MimeType,
	})
}

func (h *UploadHandler) ServeAttachment(c *gin.Context) {
	filename := filepath.Base(c.Param("filename"))
	if filename == "." || filename == "" || strings.Contains(filename, "..") {
		response.Error(c, http.StatusBadRequest, "invalid filename")
		return
	}

	filePath := filepath.Join(h.uploadDir, "attachments", filename)
	if _, err := os.Stat(filePath); err != nil {
		response.Error(c, http.StatusNotFound, "file not found")
		return
	}

	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; sandbox")
	c.FileAttachment(filePath, filename)
}

func detectMimeType(file *multipart.FileHeader) (string, error) {
	opened, err := file.Open()
	if err != nil {
		return "", err
	}
	defer opened.Close()

	buffer := make([]byte, 512)
	n, err := opened.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	return http.DetectContentType(buffer[:n]), nil
}

func strPtr(s string) *string {
	return &s
}
