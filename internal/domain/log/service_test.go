package log

import (
	"encoding/csv"
	"strings"
	"testing"

	"github.com/fastax/fastax-server/internal/shared/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.AutoMigrate(&model.CallLog{}, &model.AuditLog{})
	return db
}

func TestService_RecordCall(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	err := svc.RecordCall(&CallLogRequest{
		TraceID: "trace-001", UserID: 1, RequestPath: "/v1/chat/completions",
		RequestModel: "gpt-4", TokensPrompt: 100, TokensCompletion: 50,
		TokensTotal: 150, Status: "success", StatusCode: 200,
	})
	if err != nil {
		t.Fatalf("RecordCall() error = %v", err)
	}
}

func TestService_ListCallLogs(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	svc.RecordCall(&CallLogRequest{TraceID: "t1", UserID: 1, RequestPath: "/v1/chat/completions", Status: "success"})
	svc.RecordCall(&CallLogRequest{TraceID: "t2", UserID: 1, RequestPath: "/v1/chat/completions", Status: "error"})
	svc.RecordCall(&CallLogRequest{TraceID: "t3", UserID: 2, RequestPath: "/v1/chat/completions", Status: "success"})

	logs, total, _ := svc.ListCallLogs(&CallLogQuery{UserID: 1})
	if total != 2 {
		t.Errorf("total = %v, want 2", total)
	}
	if len(logs) != 2 {
		t.Errorf("len = %v, want 2", len(logs))
	}
}

func TestService_RecordAudit(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	err := svc.RecordAudit(&AuditLogRequest{
		TraceID: "trace-001", Action: "user.create", ResourceType: "user",
		ResourceID: "1", Result: "success",
	})
	if err != nil {
		t.Fatalf("RecordAudit() error = %v", err)
	}
}

func TestService_ListAuditLogs(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	svc.RecordAudit(&AuditLogRequest{TraceID: "t1", Action: "user.create", Result: "success"})
	svc.RecordAudit(&AuditLogRequest{TraceID: "t2", Action: "order.refund", Result: "success"})

	logs, total, _ := svc.ListAuditLogs(&AuditLogQuery{Action: "user.create"})
	if total != 1 {
		t.Errorf("total = %v, want 1", total)
	}
	if len(logs) != 1 {
		t.Errorf("len = %v, want 1", len(logs))
	}
}

func TestService_Export(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	svc.RecordAudit(&AuditLogRequest{TraceID: "exp-1", Action: "user.create", ResourceType: "user", Result: "success"})
	svc.RecordAudit(&AuditLogRequest{TraceID: "exp-2", Action: "order.refund", ResourceType: "order", Result: "fail", FailReason: "not found"})

	data, err := svc.Export(&AuditLogQuery{})
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	// header + 2 rows
	if len(records) != 3 {
		t.Fatalf("csv rows = %d, want 3", len(records))
	}
	// Verify header
	expectedHeader := []string{"ID", "TraceID", "OperatorID", "OperatorName", "OperatorIP",
		"Action", "ResourceType", "ResourceID", "Detail", "Result", "FailReason", "CreatedAt"}
	for i, h := range expectedHeader {
		if records[0][i] != h {
			t.Errorf("header[%d] = %q, want %q", i, records[0][i], h)
		}
	}
}

func TestService_Export_Filter(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	svc.RecordAudit(&AuditLogRequest{Action: "user.create", Result: "success"})
	svc.RecordAudit(&AuditLogRequest{Action: "order.refund", Result: "success"})

	data, err := svc.Export(&AuditLogQuery{Action: "order.refund"})
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	// header + 1 row
	if len(records) != 2 {
		t.Fatalf("csv rows = %d, want 2", len(records))
	}
	// Verify action column (index 5)
	if records[1][5] != "order.refund" {
		t.Errorf("action = %q, want order.refund", records[1][5])
	}
}

func TestService_Export_Empty(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	data, err := svc.Export(&AuditLogQuery{})
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	// Only header
	if len(records) != 1 {
		t.Fatalf("csv rows = %d, want 1", len(records))
	}
}
