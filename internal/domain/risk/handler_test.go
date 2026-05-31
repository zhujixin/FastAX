package risk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fastax/fastax-server/internal/shared/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupHandlerTest(t *testing.T) (*gorm.DB, *Handler, *gin.Engine) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.RiskRule{}, &model.RiskEvent{}, &model.CallLog{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	svc := NewService(db)
	handler := NewHandler(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()

	return db, handler, r
}

func createTestRule(t *testing.T, svc *Service, name, category string) *model.RiskRule {
	t.Helper()
	rule, err := svc.CreateRule(&RuleRequest{
		Name: name, Category: category,
		Conditions: "test", Action: "alert", RiskLevel: "low",
	})
	if err != nil {
		t.Fatalf("create rule: %v", err)
	}
	return rule
}

func createTestEvent(t *testing.T, svc *Service, userID uint, eventType string) *model.RiskEvent {
	t.Helper()
	event, err := svc.CreateEvent(&EventRequest{
		UserID: userID, EventType: eventType,
		RiskLevel: "high", Description: "test event",
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	return event
}

func uintToStr(n uint) string {
	return fmt.Sprintf("%d", n)
}

// --- ListRules ---

func TestHandler_ListRules_All(t *testing.T) {
	_, handler, r := setupHandlerTest(t)

	r.GET("/api/risk/rules", func(c *gin.Context) {
		c.Set("role", "admin")
		handler.ListRules(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/risk/rules", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"].(float64) != 0 {
		t.Errorf("code = %v, want 0", resp["code"])
	}
}

func TestHandler_ListRules_ByCategory(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	svc := NewService(db)
	createTestRule(t, svc, "R1", "trade")
	createTestRule(t, svc, "R2", "api")

	r.GET("/api/risk/rules", func(c *gin.Context) {
		c.Set("role", "admin")
		handler.ListRules(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/risk/rules?category=trade", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	if len(data) != 1 {
		t.Errorf("len = %v, want 1", len(data))
	}
}

// --- CreateRule ---

func TestHandler_CreateRule_Success(t *testing.T) {
	_, handler, r := setupHandlerTest(t)

	r.POST("/api/risk/rules", func(c *gin.Context) {
		c.Set("role", "admin")
		handler.CreateRule(c)
	})

	body, _ := json.Marshal(RuleRequest{
		Name: "Test Rule", Category: "trade",
		Conditions: "amount > 1000", Action: "alert", RiskLevel: "high",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/risk/rules", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestHandler_CreateRule_InvalidBody(t *testing.T) {
	_, handler, r := setupHandlerTest(t)

	r.POST("/api/risk/rules", func(c *gin.Context) {
		c.Set("role", "admin")
		handler.CreateRule(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/risk/rules", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestHandler_CreateRule_InvalidCategory(t *testing.T) {
	_, handler, r := setupHandlerTest(t)

	r.POST("/api/risk/rules", func(c *gin.Context) {
		c.Set("role", "admin")
		handler.CreateRule(c)
	})

	body, _ := json.Marshal(RuleRequest{
		Name: "Bad Rule", Category: "invalid",
		Conditions: "x", Action: "alert", RiskLevel: "low",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/risk/rules", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

// --- SetRuleEnabled ---

func TestHandler_SetRuleEnabled_Success(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	svc := NewService(db)
	rule := createTestRule(t, svc, "R1", "trade")

	r.PUT("/api/risk/rules/:id/enabled", func(c *gin.Context) {
		c.Set("role", "admin")
		handler.SetRuleEnabled(c)
	})

	body, _ := json.Marshal(setEnabledRequest{Enabled: false})
	req := httptest.NewRequest(http.MethodPut, "/api/risk/rules/"+uintToStr(rule.ID)+"/enabled", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestHandler_SetRuleEnabled_InvalidID(t *testing.T) {
	_, handler, r := setupHandlerTest(t)

	r.PUT("/api/risk/rules/:id/enabled", func(c *gin.Context) {
		c.Set("role", "admin")
		handler.SetRuleEnabled(c)
	})

	body, _ := json.Marshal(setEnabledRequest{Enabled: true})
	req := httptest.NewRequest(http.MethodPut, "/api/risk/rules/abc/enabled", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestHandler_SetRuleEnabled_NotFound(t *testing.T) {
	_, handler, r := setupHandlerTest(t)

	r.PUT("/api/risk/rules/:id/enabled", func(c *gin.Context) {
		c.Set("role", "admin")
		handler.SetRuleEnabled(c)
	})

	body, _ := json.Marshal(setEnabledRequest{Enabled: true})
	req := httptest.NewRequest(http.MethodPut, "/api/risk/rules/99999/enabled", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusNotFound, w.Body.String())
	}
}

func TestHandler_SetRuleEnabled_InvalidBody(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	svc := NewService(db)
	rule := createTestRule(t, svc, "R1", "trade")

	r.PUT("/api/risk/rules/:id/enabled", func(c *gin.Context) {
		c.Set("role", "admin")
		handler.SetRuleEnabled(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/api/risk/rules/"+uintToStr(rule.ID)+"/enabled", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

// --- ListEvents ---

func TestHandler_ListEvents_All(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	svc := NewService(db)
	createTestEvent(t, svc, 1, "abnormal_login")

	r.GET("/api/risk/events", func(c *gin.Context) {
		c.Set("role", "admin")
		handler.ListEvents(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/risk/events", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"].(float64) != 0 {
		t.Errorf("code = %v, want 0", resp["code"])
	}
	data := resp["data"].(map[string]interface{})
	if data["total"].(float64) != 1 {
		t.Errorf("total = %v, want 1", data["total"])
	}
}

func TestHandler_ListEvents_FilterByStatus(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	svc := NewService(db)
	createTestEvent(t, svc, 1, "abnormal_login")

	r.GET("/api/risk/events", func(c *gin.Context) {
		c.Set("role", "admin")
		handler.ListEvents(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/risk/events?status=pending", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["total"].(float64) != 1 {
		t.Errorf("total = %v, want 1", data["total"])
	}
}

func TestHandler_ListEvents_Pagination(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	svc := NewService(db)
	for i := 0; i < 5; i++ {
		createTestEvent(t, svc, uint(i+1), "abnormal_login")
	}

	r.GET("/api/risk/events", func(c *gin.Context) {
		c.Set("role", "admin")
		handler.ListEvents(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/risk/events?page=1&page_size=2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["total"].(float64) != 5 {
		t.Errorf("total = %v, want 5", data["total"])
	}
	items := data["items"].([]interface{})
	if len(items) != 2 {
		t.Errorf("items len = %v, want 2", len(items))
	}
}

// --- HandleEvent ---

func TestHandler_HandleEvent_Success(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	svc := NewService(db)
	event := createTestEvent(t, svc, 1, "abnormal_login")

	r.PUT("/api/risk/events/:id/handle", func(c *gin.Context) {
		c.Set("user_id", uint(99))
		c.Set("role", "admin")
		handler.HandleEvent(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/api/risk/events/"+uintToStr(event.ID)+"/handle", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"].(float64) != 0 {
		t.Errorf("code = %v, want 0", resp["code"])
	}
}

func TestHandler_HandleEvent_InvalidID(t *testing.T) {
	_, handler, r := setupHandlerTest(t)

	r.PUT("/api/risk/events/:id/handle", func(c *gin.Context) {
		c.Set("user_id", uint(99))
		c.Set("role", "admin")
		handler.HandleEvent(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/api/risk/events/abc/handle", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestHandler_HandleEvent_NotFound(t *testing.T) {
	_, handler, r := setupHandlerTest(t)

	r.PUT("/api/risk/events/:id/handle", func(c *gin.Context) {
		c.Set("user_id", uint(99))
		c.Set("role", "admin")
		handler.HandleEvent(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/api/risk/events/99999/handle", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusNotFound, w.Body.String())
	}
}

func TestHandler_HandleEvent_AlreadyHandled(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	svc := NewService(db)
	event := createTestEvent(t, svc, 1, "abnormal_login")

	// Handle the event first
	svc.HandleEvent(event.ID, 10)

	r.PUT("/api/risk/events/:id/handle", func(c *gin.Context) {
		c.Set("user_id", uint(99))
		c.Set("role", "admin")
		handler.HandleEvent(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/api/risk/events/"+uintToStr(event.ID)+"/handle", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusNotFound, w.Body.String())
	}
}
