package enterprise

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	db.AutoMigrate(&model.User{}, &model.SubAccount{}, &model.CallLog{})
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

func createEnterpriseUser(t *testing.T, db *gorm.DB) {
	t.Helper()
	db.Create(&model.User{
		ID:   1,
		Role: "enterprise",
	})
}

// --- CreateSubAccount ---

func TestHandler_CreateSubAccount_Success(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.POST("/enterprise/sub-accounts", h.CreateSubAccount)

	createEnterpriseUser(t, db)

	body := SubAccountRequest{
		Email:    "sub@example.com",
		Password: "password123",
		TokenQuota: 1000,
		Permissions: []string{"read", "write"},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/enterprise/sub-accounts", bytes.NewReader(jsonBody))
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
	if data["email"].(string) != "sub@example.com" {
		t.Errorf("email = %v, want sub@example.com", data["email"])
	}
	if data["parent_id"].(float64) != 1 {
		t.Errorf("parent_id = %v, want 1", data["parent_id"])
	}
}

func TestHandler_CreateSubAccount_InvalidJSON(t *testing.T) {
	h, _ := setupTestHandler(t)
	r := setupRouter(h)
	r.POST("/enterprise/sub-accounts", h.CreateSubAccount)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/enterprise/sub-accounts", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_CreateSubAccount_MissingRequired(t *testing.T) {
	h, _ := setupTestHandler(t)
	r := setupRouter(h)
	r.POST("/enterprise/sub-accounts", h.CreateSubAccount)

	body := map[string]string{"email": "sub@example.com"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/enterprise/sub-accounts", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_CreateSubAccount_NotEnterprise(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.POST("/enterprise/sub-accounts", h.CreateSubAccount)

	// Create a non-enterprise user
	db.Create(&model.User{
		ID:   1,
		Role: "user",
	})

	body := SubAccountRequest{
		Email:    "sub@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/enterprise/sub-accounts", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %v, want %v", w.Code, http.StatusForbidden)
	}
}

// --- ListSubAccounts ---

func TestHandler_ListSubAccounts_Empty(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.GET("/enterprise/sub-accounts", h.ListSubAccounts)

	createEnterpriseUser(t, db)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/enterprise/sub-accounts", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", w.Code, http.StatusOK)
	}
	resp := parseResponse(t, w)
	if resp["code"].(float64) != 0 {
		t.Errorf("code = %v, want 0", resp["code"])
	}
	accounts := resp["data"].([]interface{})
	if len(accounts) != 0 {
		t.Errorf("len(accounts) = %v, want 0", len(accounts))
	}
}

func TestHandler_ListSubAccounts_WithData(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.GET("/enterprise/sub-accounts", h.ListSubAccounts)

	createEnterpriseUser(t, db)
	db.Create(&model.SubAccount{ParentID: 1, Email: "a@example.com", PasswordHash: "pw", Status: 1})
	db.Create(&model.SubAccount{ParentID: 1, Email: "b@example.com", PasswordHash: "pw", Status: 1})
	db.Create(&model.SubAccount{ParentID: 2, Email: "c@example.com", PasswordHash: "pw", Status: 1})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/enterprise/sub-accounts", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", w.Code, http.StatusOK)
	}
	resp := parseResponse(t, w)
	accounts := resp["data"].([]interface{})
	if len(accounts) != 2 {
		t.Errorf("len(accounts) = %v, want 2", len(accounts))
	}
}

// --- SetSubAccountStatus ---

func TestHandler_SetSubAccountStatus_Success(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.PUT("/enterprise/sub-accounts/:id/status", h.SetSubAccountStatus)

	createEnterpriseUser(t, db)
	db.Create(&model.SubAccount{ParentID: 1, Email: "a@example.com", PasswordHash: "pw", Status: 1})

	body := map[string]int{"status": 0}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/enterprise/sub-accounts/1/status", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", w.Code, http.StatusOK)
	}

	// Verify DB update
	var updated model.SubAccount
	db.First(&updated, 1)
	if updated.Status != 0 {
		t.Errorf("db status = %v, want 0", updated.Status)
	}
}

func TestHandler_SetSubAccountStatus_NotFound(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.PUT("/enterprise/sub-accounts/:id/status", h.SetSubAccountStatus)

	createEnterpriseUser(t, db)

	body := map[string]int{"status": 0}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/enterprise/sub-accounts/999/status", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

func TestHandler_SetSubAccountStatus_InvalidID(t *testing.T) {
	h, _ := setupTestHandler(t)
	r := setupRouter(h)
	r.PUT("/enterprise/sub-accounts/:id/status", h.SetSubAccountStatus)

	body := map[string]int{"status": 0}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/enterprise/sub-accounts/abc/status", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_SetSubAccountStatus_MissingBody(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.PUT("/enterprise/sub-accounts/:id/status", h.SetSubAccountStatus)

	createEnterpriseUser(t, db)
	db.Create(&model.SubAccount{ParentID: 1, Email: "a@example.com", PasswordHash: "pw", Status: 1})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/enterprise/sub-accounts/1/status", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// --- UpdateQuota ---

func TestHandler_UpdateQuota_Success(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.PUT("/enterprise/sub-accounts/:id/quota", h.UpdateQuota)

	createEnterpriseUser(t, db)
	db.Create(&model.SubAccount{ParentID: 1, Email: "a@example.com", PasswordHash: "pw", TokenQuota: 100, Status: 1})

	body := map[string]int64{"token_quota": 5000}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/enterprise/sub-accounts/1/quota", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", w.Code, http.StatusOK)
	}

	// Verify DB update
	var updated model.SubAccount
	db.First(&updated, 1)
	if updated.TokenQuota != 5000 {
		t.Errorf("db token_quota = %v, want 5000", updated.TokenQuota)
	}
}

func TestHandler_UpdateQuota_NotFound(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.PUT("/enterprise/sub-accounts/:id/quota", h.UpdateQuota)

	createEnterpriseUser(t, db)

	body := map[string]int64{"token_quota": 5000}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/enterprise/sub-accounts/999/quota", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

func TestHandler_UpdateQuota_InvalidID(t *testing.T) {
	h, _ := setupTestHandler(t)
	r := setupRouter(h)
	r.PUT("/enterprise/sub-accounts/:id/quota", h.UpdateQuota)

	body := map[string]int64{"token_quota": 5000}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/enterprise/sub-accounts/abc/quota", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_UpdateQuota_MissingBody(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.PUT("/enterprise/sub-accounts/:id/quota", h.UpdateQuota)

	createEnterpriseUser(t, db)
	db.Create(&model.SubAccount{ParentID: 1, Email: "a@example.com", PasswordHash: "pw", Status: 1})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/enterprise/sub-accounts/1/quota", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// --- GetUsage ---

func TestHandler_GetUsage_Success(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.GET("/enterprise/usage", h.GetUsage)

	createEnterpriseUser(t, db)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/enterprise/usage?period=all", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", w.Code, http.StatusOK)
	}
	resp := parseResponse(t, w)
	if resp["code"].(float64) != 0 {
		t.Errorf("code = %v, want 0", resp["code"])
	}
	data := resp["data"].(map[string]interface{})
	if data["period"].(string) != "all" {
		t.Errorf("period = %v, want all", data["period"])
	}
}

func TestHandler_GetUsage_DefaultPeriod(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.GET("/enterprise/usage", h.GetUsage)

	createEnterpriseUser(t, db)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/enterprise/usage", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", w.Code, http.StatusOK)
	}
	resp := parseResponse(t, w)
	data := resp["data"].(map[string]interface{})
	if data["period"].(string) != "all" {
		t.Errorf("period = %v, want 'all' (default)", data["period"])
	}
}

// --- GetSubAccountUsage ---

func TestHandler_GetSubAccountUsage_Success(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.GET("/enterprise/sub-accounts/:id/usage", h.GetSubAccountUsage)

	createEnterpriseUser(t, db)
	db.Create(&model.SubAccount{ParentID: 1, Email: "a@example.com", PasswordHash: "pw", Status: 1})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/enterprise/sub-accounts/1/usage?period=all", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", w.Code, http.StatusOK)
	}
	resp := parseResponse(t, w)
	if resp["code"].(float64) != 0 {
		t.Errorf("code = %v, want 0", resp["code"])
	}
	data := resp["data"].(map[string]interface{})
	if data["period"].(string) != "all" {
		t.Errorf("period = %v, want all", data["period"])
	}
}

func TestHandler_GetSubAccountUsage_InvalidID(t *testing.T) {
	h, _ := setupTestHandler(t)
	r := setupRouter(h)
	r.GET("/enterprise/sub-accounts/:id/usage", h.GetSubAccountUsage)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/enterprise/sub-accounts/abc/usage", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_GetSubAccountUsage_DefaultPeriod(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.GET("/enterprise/sub-accounts/:id/usage", h.GetSubAccountUsage)

	createEnterpriseUser(t, db)
	db.Create(&model.SubAccount{ParentID: 1, Email: "a@example.com", PasswordHash: "pw", Status: 1})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/enterprise/sub-accounts/1/usage", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", w.Code, http.StatusOK)
	}
	resp := parseResponse(t, w)
	data := resp["data"].(map[string]interface{})
	if data["period"].(string) != "all" {
		t.Errorf("period = %v, want 'all' (default)", data["period"])
	}
}
