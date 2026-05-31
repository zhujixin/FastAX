package notify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fastax/fastax-server/internal/shared/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestHandler(t *testing.T) (*Handler, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.AutoMigrate(&model.Notification{}, &model.NotificationTemplate{})
	svc := NewService(db)
	return NewHandler(svc), db
}

func setupRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	})
	return r
}

func parseResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return resp
}

func TestHandler_List(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.GET("/notifications", h.List)

	// Seed data
	db.Create(&model.Notification{UserID: 1, Type: "order", Channel: "in_app", Title: "T1", Content: "C1"})
	db.Create(&model.Notification{UserID: 1, Type: "security", Channel: "in_app", Title: "T2", Content: "C2"})
	db.Create(&model.Notification{UserID: 2, Type: "order", Channel: "in_app", Title: "T3", Content: "C3"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notifications", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", w.Code, http.StatusOK)
	}
	resp := parseResponse(t, w)
	if resp["code"].(float64) != 0 {
		t.Errorf("code = %v, want 0", resp["code"])
	}
	data := resp["data"].(map[string]interface{})
	if data["total"].(float64) != 2 {
		t.Errorf("total = %v, want 2", data["total"])
	}
}

func TestHandler_List_FilterType(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.GET("/notifications", h.List)

	db.Create(&model.Notification{UserID: 1, Type: "order", Channel: "in_app", Title: "T1", Content: "C1"})
	db.Create(&model.Notification{UserID: 1, Type: "security", Channel: "in_app", Title: "T2", Content: "C2"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notifications?type=order", nil)
	r.ServeHTTP(w, req)

	resp := parseResponse(t, w)
	data := resp["data"].(map[string]interface{})
	if data["total"].(float64) != 1 {
		t.Errorf("total = %v, want 1", data["total"])
	}
}

func TestHandler_List_FilterIsRead(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.GET("/notifications", h.List)

	n := model.Notification{UserID: 1, Type: "order", Channel: "in_app", Title: "T1", Content: "C1"}
	db.Create(&n)
	db.Model(&model.Notification{}).Where("id = ?", n.ID).Update("is_read", 1)
	db.Create(&model.Notification{UserID: 1, Type: "order", Channel: "in_app", Title: "T2", Content: "C2"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notifications?is_read=true", nil)
	r.ServeHTTP(w, req)

	resp := parseResponse(t, w)
	data := resp["data"].(map[string]interface{})
	if data["total"].(float64) != 1 {
		t.Errorf("total = %v, want 1", data["total"])
	}
}

func TestHandler_UnreadCount(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.GET("/notifications/unread-count", h.UnreadCount)

	db.Create(&model.Notification{UserID: 1, Type: "order", Channel: "in_app", Title: "T1", Content: "C1"})
	db.Create(&model.Notification{UserID: 1, Type: "order", Channel: "in_app", Title: "T2", Content: "C2"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notifications/unread-count", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", w.Code, http.StatusOK)
	}
	resp := parseResponse(t, w)
	data := resp["data"].(map[string]interface{})
	if data["count"].(float64) != 2 {
		t.Errorf("count = %v, want 2", data["count"])
	}
}

func TestHandler_MarkRead(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.PUT("/notifications/:id/read", h.MarkRead)

	n := model.Notification{UserID: 1, Type: "order", Channel: "in_app", Title: "T", Content: "C"}
	db.Create(&n)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/notifications/1/read", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", w.Code, http.StatusOK)
	}

	// Verify it's marked as read
	var notif model.Notification
	db.First(&notif, n.ID)
	if notif.IsRead != 1 {
		t.Errorf("is_read = %v, want 1", notif.IsRead)
	}
}

func TestHandler_MarkRead_InvalidID(t *testing.T) {
	h, _ := setupTestHandler(t)
	r := setupRouter(h)
	r.PUT("/notifications/:id/read", h.MarkRead)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/notifications/abc/read", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_MarkRead_NotFound(t *testing.T) {
	h, _ := setupTestHandler(t)
	r := setupRouter(h)
	r.PUT("/notifications/:id/read", h.MarkRead)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/notifications/999/read", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

func TestHandler_MarkRead_WrongUser(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.PUT("/notifications/:id/read", h.MarkRead)

	// Create notification for user 2
	db.Create(&model.Notification{UserID: 2, Type: "order", Channel: "in_app", Title: "T", Content: "C"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/notifications/1/read", nil)
	r.ServeHTTP(w, req)

	// Should not find it (user_id mismatch)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

func TestHandler_MarkAllRead(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.PUT("/notifications/read-all", h.MarkAllRead)

	db.Create(&model.Notification{UserID: 1, Type: "order", Channel: "in_app", Title: "T1", Content: "C1"})
	db.Create(&model.Notification{UserID: 1, Type: "order", Channel: "in_app", Title: "T2", Content: "C2"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/notifications/read-all", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", w.Code, http.StatusOK)
	}

	// Verify all marked as read
	var count int64
	db.Model(&model.Notification{}).Where("user_id = 1 AND is_read = 0").Count(&count)
	if count != 0 {
		t.Errorf("unread count = %v, want 0", count)
	}
}

func TestHandler_ListTemplates(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.GET("/admin/notifications/templates", h.ListTemplates)

	db.Create(&model.NotificationTemplate{Code: "t1", Name: "T1", Channel: "in_app", Content: "C1", Language: "zh-CN", Status: 1})
	db.Create(&model.NotificationTemplate{Code: "t2", Name: "T2", Channel: "email", Content: "C2", Language: "en", Status: 1})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/notifications/templates", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", w.Code, http.StatusOK)
	}
	resp := parseResponse(t, w)
	if resp["code"].(float64) != 0 {
		t.Errorf("code = %v, want 0", resp["code"])
	}
	data := resp["data"].([]interface{})
	if len(data) != 2 {
		t.Errorf("len = %v, want 2", len(data))
	}
}

func TestHandler_ListTemplates_FilterChannel(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.GET("/admin/notifications/templates", h.ListTemplates)

	db.Create(&model.NotificationTemplate{Code: "t1", Channel: "in_app", Content: "C1", Status: 1})
	db.Create(&model.NotificationTemplate{Code: "t2", Channel: "email", Content: "C2", Status: 1})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/notifications/templates?channel=email", nil)
	r.ServeHTTP(w, req)

	resp := parseResponse(t, w)
	data := resp["data"].([]interface{})
	if len(data) != 1 {
		t.Errorf("len = %v, want 1", len(data))
	}
}

func TestHandler_CreateTemplate(t *testing.T) {
	h, _ := setupTestHandler(t)
	r := setupRouter(h)
	r.POST("/admin/notifications/templates", h.CreateTemplate)

	body := `{"code":"welcome","name":"Welcome","channel":"email","content":"Hello {{name}}","language":"en"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/admin/notifications/templates", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", w.Code, http.StatusOK)
	}
	resp := parseResponse(t, w)
	if resp["code"].(float64) != 0 {
		t.Errorf("code = %v, want 0", resp["code"])
	}
	data := resp["data"].(map[string]interface{})
	// GORM model has no json tags, so keys are capitalized
	if data["Code"].(string) != "welcome" {
		t.Errorf("Code = %v, want welcome", data["Code"])
	}
	if data["Language"].(string) != "en" {
		t.Errorf("Language = %v, want en", data["Language"])
	}
}

func TestHandler_CreateTemplate_DefaultLanguage(t *testing.T) {
	h, _ := setupTestHandler(t)
	r := setupRouter(h)
	r.POST("/admin/notifications/templates", h.CreateTemplate)

	body := `{"code":"test","channel":"in_app","content":"test"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/admin/notifications/templates", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	resp := parseResponse(t, w)
	data := resp["data"].(map[string]interface{})
	if data["Language"].(string) != "zh-CN" {
		t.Errorf("Language = %v, want zh-CN", data["Language"])
	}
}

func TestHandler_CreateTemplate_InvalidBody(t *testing.T) {
	h, _ := setupTestHandler(t)
	r := setupRouter(h)
	r.POST("/admin/notifications/templates", h.CreateTemplate)

	// Missing required "code" and "content"
	body := `{"name":"test"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/admin/notifications/templates", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_UpdateTemplate(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.PUT("/admin/notifications/templates/:id", h.UpdateTemplate)

	tmpl := model.NotificationTemplate{Code: "old", Name: "Old", Channel: "in_app", Content: "Old", Language: "zh-CN", Status: 1}
	db.Create(&tmpl)

	body := `{"code":"new","name":"New","channel":"email","content":"New content","language":"en","status":0}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/admin/notifications/templates/%d", tmpl.ID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", w.Code, http.StatusOK)
	}
	resp := parseResponse(t, w)
	data := resp["data"].(map[string]interface{})
	if data["Code"].(string) != "new" {
		t.Errorf("Code = %v, want new", data["Code"])
	}
	if data["Status"].(float64) != 0 {
		t.Errorf("Status = %v, want 0", data["Status"])
	}
}

func TestHandler_UpdateTemplate_InvalidID(t *testing.T) {
	h, _ := setupTestHandler(t)
	r := setupRouter(h)
	r.PUT("/admin/notifications/templates/:id", h.UpdateTemplate)

	body := `{"code":"x","channel":"in_app","content":"x"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/admin/notifications/templates/abc", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_UpdateTemplate_NotFound(t *testing.T) {
	h, _ := setupTestHandler(t)
	r := setupRouter(h)
	r.PUT("/admin/notifications/templates/:id", h.UpdateTemplate)

	body := `{"code":"x","channel":"in_app","content":"x"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/admin/notifications/templates/999", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %v, want %v", w.Code, http.StatusNotFound)
	}
}
