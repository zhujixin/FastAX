package risk

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
	db.AutoMigrate(&model.RiskRule{}, &model.RiskEvent{}, &model.CallLog{})
	return db
}

func TestService_CreateRule(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	rule, err := svc.CreateRule(&RuleRequest{
		Name: "Large Trade", Category: "trade",
		Conditions: "amount > 10000", Action: "alert", RiskLevel: "high",
	})
	if err != nil {
		t.Fatalf("CreateRule() error = %v", err)
	}
	if rule.Name != "Large Trade" {
		t.Errorf("name = %v", rule.Name)
	}
	if rule.Enabled != 1 {
		t.Errorf("enabled = %v, want 1", rule.Enabled)
	}
}

func TestService_ListRules(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	svc.CreateRule(&RuleRequest{Name: "R1", Category: "trade", Conditions: "x", Action: "alert", RiskLevel: "low"})
	svc.CreateRule(&RuleRequest{Name: "R2", Category: "api", Conditions: "x", Action: "freeze", RiskLevel: "high"})

	all, _ := svc.ListRules("")
	if len(all) != 2 {
		t.Errorf("len = %v, want 2", len(all))
	}

	trade, _ := svc.ListRules("trade")
	if len(trade) != 1 {
		t.Errorf("len = %v, want 1", len(trade))
	}
}

func TestService_SetRuleEnabled(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	rule, _ := svc.CreateRule(&RuleRequest{Name: "R1", Category: "trade", Conditions: "x", Action: "alert", RiskLevel: "low"})
	svc.SetRuleEnabled(rule.ID, false)

	rules, _ := svc.ListRules("")
	if rules[0].Enabled != 0 {
		t.Errorf("enabled = %v, want 0", rules[0].Enabled)
	}
}

func TestService_CreateEvent(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	event, err := svc.CreateEvent(&EventRequest{
		UserID: 1, EventType: "abnormal_login", RiskLevel: "high", Description: "test",
	})
	if err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}
	if event.Status != "pending" {
		t.Errorf("status = %v, want pending", event.Status)
	}
}

func TestService_ListEvents(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	svc.CreateEvent(&EventRequest{UserID: 1, EventType: "abnormal_login", RiskLevel: "high"})
	svc.CreateEvent(&EventRequest{UserID: 1, EventType: "rapid_api", RiskLevel: "medium"})
	svc.CreateEvent(&EventRequest{UserID: 2, EventType: "large_trade", RiskLevel: "high"})

	events, total, _ := svc.ListEvents(&EventQuery{UserID: 1})
	if total != 2 {
		t.Errorf("total = %v, want 2", total)
	}
	if len(events) != 2 {
		t.Errorf("len = %v, want 2", len(events))
	}
}

func TestService_HandleEvent(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	event, _ := svc.CreateEvent(&EventRequest{UserID: 1, EventType: "abnormal_login", RiskLevel: "high"})
	err := svc.HandleEvent(event.ID, 2)
	if err != nil {
		t.Fatalf("HandleEvent() error = %v", err)
	}

	events, _, _ := svc.ListEvents(&EventQuery{Status: "handled"})
	if len(events) != 1 {
		t.Errorf("len = %v, want 1", len(events))
	}
}
