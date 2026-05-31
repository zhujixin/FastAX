package commission

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
	db.AutoMigrate(&model.Commission{}, &model.Withdrawal{})
	return db
}

func TestService_Create(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	comm, err := svc.Create(&CreateRequest{
		AgentID: 1, CustomerID: 2, OrderID: 1,
		OrderAmount: "100.00", Rate: "0.10",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if comm.Status != "pending" {
		t.Errorf("status = %v, want pending", comm.Status)
	}
	if comm.CommissionAmount != "10.00" {
		t.Errorf("CommissionAmount = %v, want 10.00", comm.CommissionAmount)
	}
}

func TestService_ListByAgent(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	svc.Create(&CreateRequest{AgentID: 1, CustomerID: 2, OrderID: 1, OrderAmount: "100", Rate: "0.1"})
	svc.Create(&CreateRequest{AgentID: 1, CustomerID: 3, OrderID: 2, OrderAmount: "200", Rate: "0.1"})
	svc.Create(&CreateRequest{AgentID: 2, CustomerID: 4, OrderID: 3, OrderAmount: "300", Rate: "0.1"})

	comms, _ := svc.ListByAgent(1, "")
	if len(comms) != 2 {
		t.Errorf("len = %v, want 2", len(comms))
	}
}

func TestService_ListByAgent_WithStatus(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	comm, _ := svc.Create(&CreateRequest{AgentID: 1, CustomerID: 2, OrderID: 1, OrderAmount: "100", Rate: "0.1"})
	svc.Create(&CreateRequest{AgentID: 1, CustomerID: 3, OrderID: 2, OrderAmount: "200", Rate: "0.1"})
	svc.Settle(comm.ID)

	comms, _ := svc.ListByAgent(1, "settled")
	if len(comms) != 1 {
		t.Errorf("settled len = %v, want 1", len(comms))
	}
	comms, _ = svc.ListByAgent(1, "pending")
	if len(comms) != 1 {
		t.Errorf("pending len = %v, want 1", len(comms))
	}
}

func TestService_Settle(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	comm, _ := svc.Create(&CreateRequest{AgentID: 1, CustomerID: 2, OrderID: 1, OrderAmount: "100", Rate: "0.1"})
	err := svc.Settle(comm.ID)
	if err != nil {
		t.Fatalf("Settle() error = %v", err)
	}

	comms, _ := svc.ListByAgent(1, "settled")
	if len(comms) != 1 {
		t.Errorf("len = %v, want 1", len(comms))
	}
}

func TestService_Settle_AlreadySettled(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	comm, _ := svc.Create(&CreateRequest{AgentID: 1, CustomerID: 2, OrderID: 1, OrderAmount: "100", Rate: "0.1"})
	svc.Settle(comm.ID)

	err := svc.Settle(comm.ID)
	if err == nil {
		t.Fatal("expected error for already settled")
	}
}

func TestService_Settle_NotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	err := svc.Settle(999)
	if err == nil {
		t.Fatal("expected error for non-existent commission")
	}
}

func TestService_GetTotalByAgent(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	// Create two commissions for agent 1
	c1, _ := svc.Create(&CreateRequest{AgentID: 1, CustomerID: 2, OrderID: 1, OrderAmount: "100", Rate: "0.1"})
	c2, _ := svc.Create(&CreateRequest{AgentID: 1, CustomerID: 3, OrderID: 2, OrderAmount: "200", Rate: "0.2"})

	// Before settling, total should be 0
	total, err := svc.GetTotalByAgent(1)
	if err != nil {
		t.Fatalf("GetTotalByAgent() error = %v", err)
	}
	if total != "0.00" {
		t.Errorf("total before settle = %v, want 0.00", total)
	}

	// Settle both
	svc.Settle(c1.ID)
	svc.Settle(c2.ID)

	// After settling: 10 + 40 = 50
	total, err = svc.GetTotalByAgent(1)
	if err != nil {
		t.Fatalf("GetTotalByAgent() error = %v", err)
	}
	if total != "50.00" {
		t.Errorf("total after settle = %v, want 50.00", total)
	}
}

func TestService_GetTotalByAgent_DifferentAgents(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	c1, _ := svc.Create(&CreateRequest{AgentID: 1, CustomerID: 2, OrderID: 1, OrderAmount: "100", Rate: "0.1"})
	c2, _ := svc.Create(&CreateRequest{AgentID: 2, CustomerID: 3, OrderID: 2, OrderAmount: "200", Rate: "0.1"})
	svc.Settle(c1.ID)
	svc.Settle(c2.ID)

	total1, _ := svc.GetTotalByAgent(1)
	total2, _ := svc.GetTotalByAgent(2)
	if total1 != "10.00" {
		t.Errorf("agent 1 total = %v, want 10.00", total1)
	}
	if total2 != "20.00" {
		t.Errorf("agent 2 total = %v, want 20.00", total2)
	}
}

func TestService_GetAvailableBalance(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	// Create and settle a commission: 100 * 0.10 = 10.00
	c, _ := svc.Create(&CreateRequest{AgentID: 1, CustomerID: 2, OrderID: 1, OrderAmount: "100", Rate: "0.1"})
	svc.Settle(c.ID)

	// Available balance should be 10.00
	balance, err := svc.GetAvailableBalance(1)
	if err != nil {
		t.Fatalf("GetAvailableBalance() error = %v", err)
	}
	if balance != "10.00" {
		t.Errorf("balance = %v, want 10.00", balance)
	}

	// Make a withdrawal of 3.00
	svc.Withdraw(1, &WithdrawRequest{Amount: "3.00"})
	// Manually set withdrawal to completed
	db.Model(&model.Withdrawal{}).Where("agent_id = ?", 1).Update("status", "completed")

	// Available balance should be 7.00
	balance, err = svc.GetAvailableBalance(1)
	if err != nil {
		t.Fatalf("GetAvailableBalance() error = %v", err)
	}
	if balance != "7.00" {
		t.Errorf("balance after withdrawal = %v, want 7.00", balance)
	}
}

func TestService_GetAvailableBalance_NoWithdrawals(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	balance, err := svc.GetAvailableBalance(1)
	if err != nil {
		t.Fatalf("GetAvailableBalance() error = %v", err)
	}
	if balance != "0.00" {
		t.Errorf("balance = %v, want 0.00", balance)
	}
}

func TestService_Withdraw(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	// Settle a commission worth 10.00
	c, _ := svc.Create(&CreateRequest{AgentID: 1, CustomerID: 2, OrderID: 1, OrderAmount: "100", Rate: "0.1"})
	svc.Settle(c.ID)

	// Withdraw 5.00
	w, err := svc.Withdraw(1, &WithdrawRequest{Amount: "5.00", Reason: "test"})
	if err != nil {
		t.Fatalf("Withdraw() error = %v", err)
	}
	if w.Status != "pending" {
		t.Errorf("withdrawal status = %v, want pending", w.Status)
	}
	if w.Amount != "5.00" {
		t.Errorf("withdrawal amount = %v, want 5.00", w.Amount)
	}
	if w.AgentID != 1 {
		t.Errorf("withdrawal agentID = %v, want 1", w.AgentID)
	}
}

func TestService_Withdraw_InsufficientBalance(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	// Settle a commission worth 10.00
	c, _ := svc.Create(&CreateRequest{AgentID: 1, CustomerID: 2, OrderID: 1, OrderAmount: "100", Rate: "0.1"})
	svc.Settle(c.ID)

	// Try to withdraw 20.00 (more than available)
	_, err := svc.Withdraw(1, &WithdrawRequest{Amount: "20.00"})
	if err == nil {
		t.Fatal("expected error for insufficient balance")
	}
}

func TestService_Withdraw_ZeroAmount(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	c, _ := svc.Create(&CreateRequest{AgentID: 1, CustomerID: 2, OrderID: 1, OrderAmount: "100", Rate: "0.1"})
	svc.Settle(c.ID)

	_, err := svc.Withdraw(1, &WithdrawRequest{Amount: "0.00"})
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestService_Withdraw_NegativeAmount(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	c, _ := svc.Create(&CreateRequest{AgentID: 1, CustomerID: 2, OrderID: 1, OrderAmount: "100", Rate: "0.1"})
	svc.Settle(c.ID)

	_, err := svc.Withdraw(1, &WithdrawRequest{Amount: "-5.00"})
	if err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestService_Withdraw_MultiplePending(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	// Settle 20.00
	c1, _ := svc.Create(&CreateRequest{AgentID: 1, CustomerID: 2, OrderID: 1, OrderAmount: "100", Rate: "0.1"})
	c2, _ := svc.Create(&CreateRequest{AgentID: 1, CustomerID: 3, OrderID: 2, OrderAmount: "100", Rate: "0.1"})
	svc.Settle(c1.ID)
	svc.Settle(c2.ID)

	// Two pending withdrawals of 5 each = 10 pending, 10 remaining available
	svc.Withdraw(1, &WithdrawRequest{Amount: "5.00"})
	svc.Withdraw(1, &WithdrawRequest{Amount: "5.00"})

	// Pending withdrawals don't reduce available balance (only completed ones do)
	// So we can still withdraw up to 20
	_, err := svc.Withdraw(1, &WithdrawRequest{Amount: "15.00"})
	if err != nil {
		t.Fatalf("Withdraw() should succeed, error = %v", err)
	}
}

func TestService_Withdraw_CompletedReducesBalance(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	c, _ := svc.Create(&CreateRequest{AgentID: 1, CustomerID: 2, OrderID: 1, OrderAmount: "100", Rate: "0.1"})
	svc.Settle(c.ID)

	// Withdraw and complete 8.00
	svc.Withdraw(1, &WithdrawRequest{Amount: "8.00"})
	db.Model(&model.Withdrawal{}).Where("agent_id = ?", 1).Update("status", "completed")

	// Only 2.00 left
	_, err := svc.Withdraw(1, &WithdrawRequest{Amount: "3.00"})
	if err == nil {
		t.Fatal("expected error: only 2.00 available")
	}

	// Exactly 2.00 should work
	_, err = svc.Withdraw(1, &WithdrawRequest{Amount: "2.00"})
	if err != nil {
		t.Fatalf("Withdraw() should succeed for exact balance, error = %v", err)
	}
}

func TestCalculateCommission(t *testing.T) {
	tests := []struct {
		amount, rate, want string
	}{
		{"100.00", "0.10", "10.00"},
		{"200.00", "0.05", "10.00"},
		{"99.99", "0.15", "15.00"},    // 99.99 * 0.15 = 14.9985 -> 15.00
		{"0.00", "0.10", "0.00"},
		{"100", "0", "0.00"},
		{"0", "0.10", "0.00"},
	}
	for _, tt := range tests {
		got := calculateCommission(tt.amount, tt.rate)
		if got != tt.want {
			t.Errorf("calculateCommission(%s, %s) = %s, want %s", tt.amount, tt.rate, got, tt.want)
		}
	}
}
