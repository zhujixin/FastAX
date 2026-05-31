package log

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
	if err := db.AutoMigrate(&model.CallLog{}, &model.AuditLog{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	svc := NewService(db)
	handler := NewHandler(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()

	return db, handler, r
}

func seedAuditLogs(t *testing.T, svc *Service, count int) {
	t.Helper()
	for i := 0; i < count; i++ {
		err := svc.RecordAudit(&AuditLogRequest{
			TraceID:      "trace-" + string(rune('A'+i)),
			OperatorName: "admin",
			Action:       "user.create",
			ResourceType: "user",
			ResourceID:    "1",
			Result:       "success",
		})
		if err != nil {
			t.Fatalf("seed audit log %d: %v", i, err)
		}
	}
}

func seedCallLogs(t *testing.T, svc *Service) {
	t.Helper()
	svc.RecordCall(&CallLogRequest{TraceID: "c1", UserID: 1, RequestPath: "/v1/chat/completions", RequestModel: "gpt-4", Status: "success"})
	svc.RecordCall(&CallLogRequest{TraceID: "c2", UserID: 2, RequestPath: "/v1/chat/completions", RequestModel: "gpt-3.5-turbo", Status: "error"})
}

// --- ListAuditLogs ---

func TestHandler_ListAuditLogs_All(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	svc := NewService(db)
	seedAuditLogs(t, svc, 3)

	r.GET("/api/admin/audit/logs", handler.ListAuditLogs)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/audit/logs", nil)
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
	if data["total"].(float64) != 3 {
		t.Errorf("total = %v, want 3", data["total"])
	}
}

func TestHandler_ListAuditLogs_Filter(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	svc := NewService(db)
	svc.RecordAudit(&AuditLogRequest{Action: "user.create", Result: "success"})
	svc.RecordAudit(&AuditLogRequest{Action: "order.refund", Result: "success"})

	r.GET("/api/admin/audit/logs", handler.ListAuditLogs)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/audit/logs?action=user.create", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v", w.Code, http.StatusOK)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["total"].(float64) != 1 {
		t.Errorf("total = %v, want 1", data["total"])
	}
}

func TestHandler_ListAuditLogs_Pagination(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	svc := NewService(db)
	seedAuditLogs(t, svc, 5)

	r.GET("/api/admin/audit/logs", handler.ListAuditLogs)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/audit/logs?page=1&page_size=2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v", w.Code, http.StatusOK)
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

// --- ExportAuditLogs ---

func TestHandler_ExportAuditLogs_Success(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	svc := NewService(db)
	svc.RecordAudit(&AuditLogRequest{TraceID: "exp-1", Action: "user.create", ResourceType: "user", Result: "success"})
	svc.RecordAudit(&AuditLogRequest{TraceID: "exp-2", Action: "order.refund", ResourceType: "order", Result: "fail", FailReason: "not found"})

	r.GET("/api/admin/audit/export", handler.ExportAuditLogs)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/audit/export", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/csv") {
		t.Errorf("Content-Type = %q, want text/csv prefix", ct)
	}
	cd := w.Header().Get("Content-Disposition")
	if !strings.Contains(cd, "audit_logs.csv") {
		t.Errorf("Content-Disposition = %q, want audit_logs.csv", cd)
	}

	// Parse CSV and verify content
	reader := csv.NewReader(w.Body)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	// header + 2 data rows
	if len(records) != 3 {
		t.Fatalf("csv rows = %d, want 3", len(records))
	}
	// Check header
	if records[0][0] != "ID" || records[0][5] != "Action" {
		t.Errorf("header = %v", records[0])
	}
}

func TestHandler_ExportAuditLogs_Empty(t *testing.T) {
	_, handler, r := setupHandlerTest(t)

	r.GET("/api/admin/audit/export", handler.ExportAuditLogs)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/audit/export", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v", w.Code, http.StatusOK)
	}

	reader := csv.NewReader(w.Body)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	// Only header row
	if len(records) != 1 {
		t.Fatalf("csv rows = %d, want 1 (header only)", len(records))
	}
}

func TestHandler_ExportAuditLogs_FilterByAction(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	svc := NewService(db)
	svc.RecordAudit(&AuditLogRequest{Action: "user.create", Result: "success"})
	svc.RecordAudit(&AuditLogRequest{Action: "order.refund", Result: "success"})

	r.GET("/api/admin/audit/export", handler.ExportAuditLogs)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/audit/export?action=order.refund", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v", w.Code, http.StatusOK)
	}

	reader := csv.NewReader(w.Body)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	// header + 1 data row
	if len(records) != 2 {
		t.Fatalf("csv rows = %d, want 2", len(records))
	}
	// Verify the action column (index 5)
	if records[1][5] != "order.refund" {
		t.Errorf("action = %q, want order.refund", records[1][5])
	}
}

// --- ListCallLogs ---

func TestHandler_ListCallLogs_All(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	svc := NewService(db)
	seedCallLogs(t, svc)

	r.GET("/api/admin/call-logs", handler.ListCallLogs)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/call-logs", nil)
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
	if data["total"].(float64) != 2 {
		t.Errorf("total = %v, want 2", data["total"])
	}
}

func TestHandler_ListCallLogs_FilterByStatus(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	svc := NewService(db)
	seedCallLogs(t, svc)

	r.GET("/api/admin/call-logs", handler.ListCallLogs)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/call-logs?status=error", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v", w.Code, http.StatusOK)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["total"].(float64) != 1 {
		t.Errorf("total = %v, want 1", data["total"])
	}
}

func TestHandler_ListCallLogs_Pagination(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	svc := NewService(db)
	for i := 0; i < 5; i++ {
		svc.RecordCall(&CallLogRequest{TraceID: "pag-" + string(rune('A'+i)), UserID: 1, RequestPath: "/v1/chat/completions", Status: "success"})
	}

	r.GET("/api/admin/call-logs", handler.ListCallLogs)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/call-logs?page=1&page_size=2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v", w.Code, http.StatusOK)
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
