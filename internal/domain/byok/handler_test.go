package byok

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
	db.AutoMigrate(&model.BYOKKey{})
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

// --- ListKeys ---

func TestHandler_ListKeys_Empty(t *testing.T) {
	h, _ := setupTestHandler(t)
	r := setupRouter(h)
	r.GET("/byok/keys", h.ListKeys)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/byok/keys", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", w.Code, http.StatusOK)
	}
	resp := parseResponse(t, w)
	if resp["code"].(float64) != 0 {
		t.Errorf("code = %v, want 0", resp["code"])
	}
	keys := resp["data"].([]interface{})
	if len(keys) != 0 {
		t.Errorf("len(keys) = %v, want 0", len(keys))
	}
}

func TestHandler_ListKeys_WithData(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.GET("/byok/keys", h.ListKeys)

	db.Create(&model.BYOKKey{UserID: 1, Provider: "openai", KeyEncrypted: "e1", KeyIV: "iv1", Status: 1})
	db.Create(&model.BYOKKey{UserID: 1, Provider: "anthropic", KeyEncrypted: "e2", KeyIV: "iv2", Status: 1})
	db.Create(&model.BYOKKey{UserID: 2, Provider: "openai", KeyEncrypted: "e3", KeyIV: "iv3", Status: 1})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/byok/keys", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", w.Code, http.StatusOK)
	}
	resp := parseResponse(t, w)
	keys := resp["data"].([]interface{})
	if len(keys) != 2 {
		t.Errorf("len(keys) = %v, want 2", len(keys))
	}
}

// --- AddKey ---

func TestHandler_AddKey_Success(t *testing.T) {
	h, _ := setupTestHandler(t)
	r := setupRouter(h)
	r.POST("/byok/keys", h.AddKey)

	body := AddKeyRequest{
		Provider:     "openai",
		KeyEncrypted: "encrypted-data",
		KeyIV:        "iv-data",
		Alias:        "my key",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/byok/keys", bytes.NewReader(jsonBody))
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
	if data["provider"].(string) != "openai" {
		t.Errorf("provider = %v, want openai", data["provider"])
	}
	if data["alias"].(string) != "my key" {
		t.Errorf("alias = %v, want 'my key'", data["alias"])
	}
}

func TestHandler_AddKey_MissingRequired(t *testing.T) {
	h, _ := setupTestHandler(t)
	r := setupRouter(h)
	r.POST("/byok/keys", h.AddKey)

	body := map[string]string{"alias": "no provider"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/byok/keys", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_AddKey_InvalidJSON(t *testing.T) {
	h, _ := setupTestHandler(t)
	r := setupRouter(h)
	r.POST("/byok/keys", h.AddKey)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/byok/keys", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// --- DeleteKey ---

func TestHandler_DeleteKey_Success(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.DELETE("/byok/keys/:id", h.DeleteKey)

	key := model.BYOKKey{UserID: 1, Provider: "openai", KeyEncrypted: "e", KeyIV: "iv", Status: 1}
	db.Create(&key)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/byok/keys/1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", w.Code, http.StatusOK)
	}
	resp := parseResponse(t, w)
	if resp["code"].(float64) != 0 {
		t.Errorf("code = %v, want 0", resp["code"])
	}
}

func TestHandler_DeleteKey_NotFound(t *testing.T) {
	h, _ := setupTestHandler(t)
	r := setupRouter(h)
	r.DELETE("/byok/keys/:id", h.DeleteKey)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/byok/keys/999", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

func TestHandler_DeleteKey_WrongUser(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.DELETE("/byok/keys/:id", h.DeleteKey)

	db.Create(&model.BYOKKey{UserID: 2, Provider: "openai", KeyEncrypted: "e", KeyIV: "iv", Status: 1})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/byok/keys/1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

func TestHandler_DeleteKey_InvalidID(t *testing.T) {
	h, _ := setupTestHandler(t)
	r := setupRouter(h)
	r.DELETE("/byok/keys/:id", h.DeleteKey)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/byok/keys/abc", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// --- SetKeyStatus ---

func TestHandler_SetKeyStatus_Success(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.PUT("/byok/keys/:id/status", h.SetKeyStatus)

	key := model.BYOKKey{UserID: 1, Provider: "openai", KeyEncrypted: "e", KeyIV: "iv", Status: 1}
	db.Create(&key)

	body := map[string]int{"status": 0}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/byok/keys/1/status", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", w.Code, http.StatusOK)
	}

	// Verify DB update
	var updated model.BYOKKey
	db.First(&updated, key.ID)
	if updated.Status != 0 {
		t.Errorf("db status = %v, want 0", updated.Status)
	}
}

func TestHandler_SetKeyStatus_NotFound(t *testing.T) {
	h, _ := setupTestHandler(t)
	r := setupRouter(h)
	r.PUT("/byok/keys/:id/status", h.SetKeyStatus)

	body := map[string]int{"status": 0}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/byok/keys/999/status", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

func TestHandler_SetKeyStatus_InvalidID(t *testing.T) {
	h, _ := setupTestHandler(t)
	r := setupRouter(h)
	r.PUT("/byok/keys/:id/status", h.SetKeyStatus)

	body := map[string]int{"status": 0}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/byok/keys/abc/status", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_SetKeyStatus_MissingBody(t *testing.T) {
	h, db := setupTestHandler(t)
	r := setupRouter(h)
	r.PUT("/byok/keys/:id/status", h.SetKeyStatus)

	db.Create(&model.BYOKKey{UserID: 1, Provider: "openai", KeyEncrypted: "e", KeyIV: "iv", Status: 1})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/byok/keys/1/status", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}
