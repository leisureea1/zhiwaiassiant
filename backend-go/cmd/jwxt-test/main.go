package main

import (
	"net/http"
	"sync"

	"xisu/backend-go/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type sessionStore struct {
	mu   sync.RWMutex
	data map[string]*service.CachedJWXTSession
}

func newSessionStore() *sessionStore {
	return &sessionStore{data: map[string]*service.CachedJWXTSession{}}
}

func (s *sessionStore) set(id string, sess *service.CachedJWXTSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = sess
}

func (s *sessionStore) get(id string) (*service.CachedJWXTSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.data[id]
	return sess, ok
}

func main() {
	r := gin.Default()
	svc := service.NewJwxtDirectService(nil)
	store := newSessionStore()

	r.StaticFile("/", "./static/jwxt_direct_test.html")
	r.StaticFile("/jwxt-direct-test", "./static/jwxt_direct_test.html")

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/api/jwxt/login", func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "username/password required"})
			return
		}

		sess, err := svc.Login(req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		sid := uuid.NewString()
		store.set(sid, sess)

		user, _ := svc.GetUser(sess)
		c.JSON(http.StatusOK, gin.H{
			"sid":  sid,
			"user": user,
		})
	})

	r.GET("/api/jwxt/semester", func(c *gin.Context) {
		sid := c.Query("sid")
		sess, ok := store.get(sid)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid sid, please login first"})
			return
		}

		data, err := svc.GetSemester(sess)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, data)
	})

	r.GET("/api/jwxt/course", func(c *gin.Context) {
		sid := c.Query("sid")
		sess, ok := store.get(sid)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid sid, please login first"})
			return
		}

		semesterID := c.Query("semester_id")
		data, err := svc.GetCourse(sess, semesterID, "")
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, data)
	})

	r.GET("/api/jwxt/user", func(c *gin.Context) {
		sid := c.Query("sid")
		sess, ok := store.get(sid)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid sid, please login first"})
			return
		}

		data, err := svc.GetUser(sess)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, data)
	})

	r.GET("/api/jwxt/grade", func(c *gin.Context) {
		sid := c.Query("sid")
		sess, ok := store.get(sid)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid sid, please login first"})
			return
		}

		data, err := svc.GetGrade(sess, c.Query("semester_id"))
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, data)
	})

	r.GET("/api/jwxt/exam", func(c *gin.Context) {
		sid := c.Query("sid")
		sess, ok := store.get(sid)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid sid, please login first"})
			return
		}

		data, err := svc.GetExam(sess, c.Query("semester_id"))
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, data)
	})

	r.POST("/api/jwxt/quick", func(c *gin.Context) {
		var req struct {
			Username   string `json:"username"`
			Password   string `json:"password"`
			SemesterID string `json:"semester_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "username/password required"})
			return
		}

		sess, err := svc.Login(req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		semester, semErr := svc.GetSemester(sess)
		course, courseErr := svc.GetCourse(sess, req.SemesterID, "")

		resp := gin.H{
			"semester": semester,
			"course":   course,
		}
		if semErr != nil {
			resp["semester_error"] = semErr.Error()
		}
		if courseErr != nil {
			resp["course_error"] = courseErr.Error()
		}
		c.JSON(http.StatusOK, resp)
	})

	if err := r.Run(":3001"); err != nil {
		panic(err)
	}
}
