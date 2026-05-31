package cost

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestRouter(svc *Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(svc)

	// Simulate auth middleware that sets user_id
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	})

	r.GET("/api/user/budget", h.GetBudget)
	r.PUT("/api/user/budget", h.SetBudget)
	r.GET("/api/user/cost-alerts", h.GetAlerts)
	r.PUT("/api/user/cost-alerts", h.SetAlert)
	return r
}

func TestHandler_SetBudget(t *testing.T) {
	svc := NewService(nil) // DB not needed for budget (in-memory)
	r := setupTestRouter(svc)

	body := SetBudgetRequest{Period: "monthly", Limit: 500.0}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/user/budget", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"].(float64) != 0 {
		t.Errorf("code = %v, want 0", resp["code"])
	}
}

func TestHandler_SetBudget_InvalidPeriod(t *testing.T) {
	svc := NewService(nil)
	r := setupTestRouter(svc)

	body := SetBudgetRequest{Period: "yearly", Limit: 500.0}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/user/budget", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_SetBudget_NegativeLimit(t *testing.T) {
	svc := NewService(nil)
	r := setupTestRouter(svc)

	body := SetBudgetRequest{Period: "daily", Limit: -100}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/user/budget", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_GetBudget(t *testing.T) {
	svc := NewService(nil)
	r := setupTestRouter(svc)

	// Set budget first
	svc.SetBudget(1, "monthly", 1000.0)

	req := httptest.NewRequest(http.MethodGet, "/api/user/budget", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["limit"].(float64) != 1000.0 {
		t.Errorf("limit = %v, want 1000", data["limit"])
	}
}

func TestHandler_GetBudget_NotSet(t *testing.T) {
	svc := NewService(nil)
	r := setupTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/user/budget", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandler_SetAlert(t *testing.T) {
	svc := NewService(nil)
	r := setupTestRouter(svc)

	body := SetAlertRequest{Thresholds: []float64{50, 80, 100}}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/user/cost-alerts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, body = %s", w.Code, w.Body.String())
	}
}

func TestHandler_SetAlert_EmptyThresholds(t *testing.T) {
	svc := NewService(nil)
	r := setupTestRouter(svc)

	body := SetAlertRequest{Thresholds: []float64{}}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/user/cost-alerts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_GetAlerts(t *testing.T) {
	svc := NewService(nil)
	r := setupTestRouter(svc)

	svc.SetAlert(1, []float64{50, 80, 100})

	req := httptest.NewRequest(http.MethodGet, "/api/user/cost-alerts", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	thresholds := data["thresholds"].([]interface{})
	if len(thresholds) != 3 {
		t.Errorf("thresholds count = %v, want 3", len(thresholds))
	}
}

func TestHandler_GetAlerts_NotSet(t *testing.T) {
	svc := NewService(nil)
	r := setupTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/user/cost-alerts", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}
