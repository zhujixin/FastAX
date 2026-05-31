package cost

import (
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
	db.AutoMigrate(&model.SemanticCache{})
	return db
}

// --- Budget Tests ---

func TestService_SetBudget(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	budget, err := svc.SetBudget(1, "monthly", 1000.0)
	if err != nil {
		t.Fatalf("SetBudget() error = %v", err)
	}
	if budget.UserID != 1 || budget.Period != "monthly" || budget.Limit != 1000.0 {
		t.Errorf("unexpected budget: %+v", budget)
	}
}

func TestService_SetBudget_InvalidPeriod(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	_, err := svc.SetBudget(1, "yearly", 1000.0)
	if err == nil {
		t.Fatal("expected error for invalid period")
	}
}

func TestService_SetBudget_NegativeLimit(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	_, err := svc.SetBudget(1, "daily", -100)
	if err == nil {
		t.Fatal("expected error for negative limit")
	}
}

func TestService_GetBudget(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	svc.SetBudget(1, "monthly", 1000.0)

	status, err := svc.GetBudget(1)
	if err != nil {
		t.Fatalf("GetBudget() error = %v", err)
	}
	if status.Limit != 1000.0 {
		t.Errorf("limit = %v, want 1000", status.Limit)
	}
	if status.Exceeded {
		t.Error("should not be exceeded with 0 spending")
	}
}

func TestService_GetBudget_NotSet(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	_, err := svc.GetBudget(999)
	if err == nil {
		t.Fatal("expected error for unset budget")
	}
}

func TestService_CheckBudget(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	// No budget set — should allow
	ok, spent, err := svc.CheckBudget(1)
	if err != nil {
		t.Fatalf("CheckBudget() error = %v", err)
	}
	if !ok {
		t.Error("should allow when no budget set")
	}
	if spent != 0 {
		t.Errorf("spent = %v, want 0", spent)
	}

	// Set budget and check
	svc.SetBudget(1, "monthly", 100.0)
	ok, _, _ = svc.CheckBudget(1)
	if !ok {
		t.Error("should allow with 0 spending")
	}

	// Record spending over limit
	svc.RecordSpending(1, 150.0)
	ok, spent, _ = svc.CheckBudget(1)
	if ok {
		t.Error("should deny when over budget")
	}
	if spent != 150.0 {
		t.Errorf("spent = %v, want 150", spent)
	}
}

func TestService_RecordSpending(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	svc.RecordSpending(1, 50.0)
	svc.RecordSpending(1, 30.0)

	svc.SetBudget(1, "daily", 1000.0)
	status, _ := svc.GetBudget(1)
	if status.Spent != 80.0 {
		t.Errorf("spent = %v, want 80", status.Spent)
	}
}

func TestService_GetBudget_Exceeded(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	svc.SetBudget(1, "daily", 100.0)
	svc.RecordSpending(1, 100.0)

	status, _ := svc.GetBudget(1)
	if !status.Exceeded {
		t.Error("should be exceeded")
	}
	if status.UsagePct != 100.0 {
		t.Errorf("usage_pct = %v, want 100", status.UsagePct)
	}
}

// --- Alert Tests ---

func TestService_SetAlert(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	alert, err := svc.SetAlert(1, []float64{50, 80, 100})
	if err != nil {
		t.Fatalf("SetAlert() error = %v", err)
	}
	if len(alert.Thresholds) != 3 {
		t.Errorf("thresholds count = %v, want 3", len(alert.Thresholds))
	}
}

func TestService_SetAlert_EmptyThresholds(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	_, err := svc.SetAlert(1, []float64{})
	if err == nil {
		t.Fatal("expected error for empty thresholds")
	}
}

func TestService_SetAlert_InvalidThreshold(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	_, err := svc.SetAlert(1, []float64{-10})
	if err == nil {
		t.Fatal("expected error for negative threshold")
	}
}

func TestService_GetAlerts(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	svc.SetAlert(1, []float64{50, 80, 100})

	alert, err := svc.GetAlerts(1)
	if err != nil {
		t.Fatalf("GetAlerts() error = %v", err)
	}
	if len(alert.Thresholds) != 3 {
		t.Errorf("thresholds count = %v, want 3", len(alert.Thresholds))
	}
}

func TestService_GetAlerts_NotSet(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	_, err := svc.GetAlerts(999)
	if err == nil {
		t.Fatal("expected error for unset alert")
	}
}

func TestService_CheckAlerts(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	// No budget or alert — no triggers
	triggered, _ := svc.CheckAlerts(1)
	if triggered != nil {
		t.Errorf("expected nil, got %v", triggered)
	}

	// Set budget and alert
	svc.SetBudget(1, "monthly", 1000.0)
	svc.SetAlert(1, []float64{50, 80, 100})

	// At 0% — no triggers
	triggered, _ = svc.CheckAlerts(1)
	if len(triggered) != 0 {
		t.Errorf("expected 0 triggers, got %v", triggered)
	}

	// At 60% — should trigger 50
	svc.RecordSpending(1, 600.0)
	triggered, _ = svc.CheckAlerts(1)
	if len(triggered) != 1 || triggered[0] != 50 {
		t.Errorf("expected [50], got %v", triggered)
	}

	// At 85% — should trigger 50, 80
	svc.RecordSpending(1, 250.0)
	triggered, _ = svc.CheckAlerts(1)
	if len(triggered) != 2 {
		t.Errorf("expected 2 triggers, got %v", triggered)
	}
}

// --- Cache Tests (existing) ---

func TestService_SetCache(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	err := svc.SetCache(&CacheRequest{
		PromptHash: "abc123", ResponseEncrypted: "resp", Model: "gpt-4", TTLSeconds: 3600,
	})
	if err != nil {
		t.Fatalf("SetCache() error = %v", err)
	}
}

func TestService_GetCache(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	svc.SetCache(&CacheRequest{PromptHash: "abc123", ResponseEncrypted: "resp", Model: "gpt-4", TTLSeconds: 3600})

	cache, err := svc.GetCache("abc123", "gpt-4")
	if err != nil {
		t.Fatalf("GetCache() error = %v", err)
	}
	if cache.ResponseEncrypted != "resp" {
		t.Errorf("response = %v", cache.ResponseEncrypted)
	}
}

func TestService_GetCache_Miss(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	_, err := svc.GetCache("nonexistent", "gpt-4")
	if err == nil {
		t.Fatal("expected error for cache miss")
	}
}

func TestService_CleanExpired(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	db.Exec("INSERT INTO semantic_caches (prompt_hash, response_encrypted, model, created_at, expires_at) VALUES (?, ?, ?, ?, ?)",
		"old", "r", "gpt-4", 1000, 1000)
	svc.SetCache(&CacheRequest{PromptHash: "new", ResponseEncrypted: "r", Model: "gpt-4", TTLSeconds: 3600})

	count, _ := svc.CleanExpired()
	if count != 1 {
		t.Errorf("cleaned = %v, want 1", count)
	}
}
