package stats

import (
	"fmt"
	"testing"
	"time"

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
	if err := db.AutoMigrate(
		&model.User{}, &model.CallLog{},
		&model.Order{}, &model.Payment{}, &model.Refund{},
		&model.TokenProduct{}, &model.UserToken{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func createTestUser(t *testing.T, db *gorm.DB, id uint) {
	t.Helper()
	uid := uniqueID()
	user := model.User{
		Username:     "testuser_" + uid,
		Email:        "user_" + uid + "@test.com",
		Phone:        fmt.Sprintf("138%010d", idCounter),
		PasswordHash: "hash",
		Role:         "user",
		Level:        "normal",
		Status:       1,
	}
	user.ID = id
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
}

func createCallLog(t *testing.T, db *gorm.DB, userID uint, tokensTotal, tokensPrompt, tokensComp int, createdAt time.Time) {
	t.Helper()
	log := model.CallLog{
		TraceID:          "trace-" + uniqueID(),
		UserID:           userID,
		RequestPath:      "/v1/chat/completions",
		RequestModel:     "gpt-4",
		TokensPrompt:     tokensPrompt,
		TokensCompletion: tokensComp,
		TokensTotal:      tokensTotal,
		Status:           "success",
		CreatedAt:        createdAt,
	}
	if err := db.Create(&log).Error; err != nil {
		t.Fatalf("create call_log: %v", err)
	}
}

func createPaidOrder(t *testing.T, db *gorm.DB, userID uint, amount string, createdAt time.Time) model.Order {
	t.Helper()
	uid := uniqueID()
	order := model.Order{
		OrderNo:     "ORD" + uid,
		UserID:      userID,
		ProductID:   1,
		Quantity:    "1",
		UnitPrice:   amount,
		Amount:      amount,
		FinalAmount: amount,
		Currency:    "CNY",
		Status:      "paid",
		CreatedAt:   createdAt,
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	payment := model.Payment{
		OrderID:   order.ID,
		PaymentNo: "PAY" + uid,
		Amount:    amount,
		Method:    "wechat",
		Gateway:   "wechat",
		Status:    "success",
		CreatedAt: createdAt,
	}
	if err := db.Create(&payment).Error; err != nil {
		t.Fatalf("create payment: %v", err)
	}

	return order
}

// ---------- GetUsage tests ----------

func TestGetUsage_Month_Default(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createTestUser(t, db, 1)

	now := time.Now()
	createCallLog(t, db, 1, 1000, 600, 400, now)
	createCallLog(t, db, 1, 2000, 1200, 800, now.Add(-1*time.Hour))

	// Different user, should not be counted
	createCallLog(t, db, 2, 500, 300, 200, now)

	resp, err := svc.GetUsage(1, "month")
	if err != nil {
		t.Fatalf("GetUsage() error = %v", err)
	}
	if resp.TotalTokens != 3000 {
		t.Errorf("total_tokens = %v, want 3000", resp.TotalTokens)
	}
	if resp.PromptTokens != 1800 {
		t.Errorf("prompt_tokens = %v, want 1800", resp.PromptTokens)
	}
	if resp.CompletionTokens != 1200 {
		t.Errorf("completion_tokens = %v, want 1200", resp.CompletionTokens)
	}
	if resp.RequestCount != 2 {
		t.Errorf("request_count = %v, want 2", resp.RequestCount)
	}
	if resp.Period != "month" {
		t.Errorf("period = %v, want month", resp.Period)
	}
}

func TestGetUsage_Day(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createTestUser(t, db, 1)

	now := time.Now()
	createCallLog(t, db, 1, 500, 300, 200, now)

	// Old record from yesterday, should not be counted
	yesterday := now.AddDate(0, 0, -1)
	createCallLog(t, db, 1, 1000, 600, 400, yesterday)

	resp, err := svc.GetUsage(1, "day")
	if err != nil {
		t.Fatalf("GetUsage() error = %v", err)
	}
	if resp.TotalTokens != 500 {
		t.Errorf("total_tokens = %v, want 500", resp.TotalTokens)
	}
	if resp.RequestCount != 1 {
		t.Errorf("request_count = %v, want 1", resp.RequestCount)
	}
}

func TestGetUsage_InvalidPeriod(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	_, err := svc.GetUsage(1, "invalid")
	if err == nil {
		t.Fatal("expected error for invalid period")
	}
}

func TestGetUsage_NoData(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createTestUser(t, db, 1)

	resp, err := svc.GetUsage(1, "month")
	if err != nil {
		t.Fatalf("GetUsage() error = %v", err)
	}
	if resp.TotalTokens != 0 {
		t.Errorf("total_tokens = %v, want 0", resp.TotalTokens)
	}
	if resp.RequestCount != 0 {
		t.Errorf("request_count = %v, want 0", resp.RequestCount)
	}
}

// ---------- GetConsumption tests ----------

func TestGetConsumption_Month(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createTestUser(t, db, 1)

	now := time.Now()
	createPaidOrder(t, db, 1, "100.00", now)
	createPaidOrder(t, db, 1, "200.50", now.Add(-1*time.Hour))

	// Different user
	createPaidOrder(t, db, 2, "50.00", now)

	resp, err := svc.GetConsumption(1, "month")
	if err != nil {
		t.Fatalf("GetConsumption() error = %v", err)
	}
	if resp.OrderCount != 2 {
		t.Errorf("order_count = %v, want 2", resp.OrderCount)
	}
	if resp.PaymentCount != 2 {
		t.Errorf("payment_count = %v, want 2", resp.PaymentCount)
	}
	if resp.TotalAmount != "300.50" {
		t.Errorf("total_amount = %v, want 300.50", resp.TotalAmount)
	}
}

func TestGetConsumption_InvalidPeriod(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	_, err := svc.GetConsumption(1, "bad")
	if err == nil {
		t.Fatal("expected error for invalid period")
	}
}

func TestGetConsumption_NoData(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createTestUser(t, db, 1)

	resp, err := svc.GetConsumption(1, "month")
	if err != nil {
		t.Fatalf("GetConsumption() error = %v", err)
	}
	if resp.TotalAmount != "0.00" {
		t.Errorf("total_amount = %v, want 0.00", resp.TotalAmount)
	}
	if resp.OrderCount != 0 {
		t.Errorf("order_count = %v, want 0", resp.OrderCount)
	}
}

// ---------- GetBills tests ----------

func TestGetBills_Pagination(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createTestUser(t, db, 1)

	now := time.Now()
	for i := 0; i < 5; i++ {
		createPaidOrder(t, db, 1, "10.00", now.Add(time.Duration(-i)*time.Hour))
	}

	bills, total, err := svc.GetBills(1, 1, 2)
	if err != nil {
		t.Fatalf("GetBills() error = %v", err)
	}
	if total != 5 {
		t.Errorf("total = %v, want 5", total)
	}
	if len(bills) != 2 {
		t.Errorf("len = %v, want 2", len(bills))
	}
}

func TestGetBills_SecondPage(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createTestUser(t, db, 1)

	now := time.Now()
	for i := 0; i < 5; i++ {
		createPaidOrder(t, db, 1, "10.00", now.Add(time.Duration(-i)*time.Hour))
	}

	bills, total, err := svc.GetBills(1, 3, 2)
	if err != nil {
		t.Fatalf("GetBills() error = %v", err)
	}
	if total != 5 {
		t.Errorf("total = %v, want 5", total)
	}
	if len(bills) != 1 {
		t.Errorf("len = %v, want 1", len(bills))
	}
}

func TestGetBills_DefaultPageSize(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createTestUser(t, db, 1)

	bills, total, err := svc.GetBills(1, 0, 0) // page=0 -> 1, pageSize=0 -> 20
	if err != nil {
		t.Fatalf("GetBills() error = %v", err)
	}
	if total != 0 {
		t.Errorf("total = %v, want 0", total)
	}
	if len(bills) != 0 {
		t.Errorf("len = %v, want 0", len(bills))
	}
}

func TestGetBills_DifferentUser(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createTestUser(t, db, 1)
	createTestUser(t, db, 2)

	now := time.Now()
	createPaidOrder(t, db, 1, "10.00", now)

	bills, total, err := svc.GetBills(2, 1, 20)
	if err != nil {
		t.Fatalf("GetBills() error = %v", err)
	}
	if total != 0 {
		t.Errorf("total = %v, want 0", total)
	}
	if len(bills) != 0 {
		t.Errorf("len = %v, want 0", len(bills))
	}
}

// ---------- GetSummary tests ----------

func TestGetSummary_WithData(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createTestUser(t, db, 1)

	now := time.Now()
	// Current month data
	createCallLog(t, db, 1, 500, 300, 200, now)
	createCallLog(t, db, 1, 300, 200, 100, now.Add(-1*time.Hour))
	createPaidOrder(t, db, 1, "100.00", now)
	createPaidOrder(t, db, 1, "50.00", now)

	resp, err := svc.GetSummary(1)
	if err != nil {
		t.Fatalf("GetSummary() error = %v", err)
	}
	if resp.TotalTokens != 800 {
		t.Errorf("total_tokens = %v, want 800", resp.TotalTokens)
	}
	if resp.MonthTokens != 800 {
		t.Errorf("month_tokens = %v, want 800", resp.MonthTokens)
	}
	if resp.TotalAmount != "150.00" {
		t.Errorf("total_amount = %v, want 150.00", resp.TotalAmount)
	}
	if resp.MonthAmount != "150.00" {
		t.Errorf("month_amount = %v, want 150.00", resp.MonthAmount)
	}
	if resp.TotalRequests != 2 {
		t.Errorf("total_requests = %v, want 2", resp.TotalRequests)
	}
	if resp.TotalOrders != 2 {
		t.Errorf("total_orders = %v, want 2", resp.TotalOrders)
	}
}

func TestGetSummary_NoData(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createTestUser(t, db, 1)

	resp, err := svc.GetSummary(1)
	if err != nil {
		t.Fatalf("GetSummary() error = %v", err)
	}
	if resp.TotalTokens != 0 {
		t.Errorf("total_tokens = %v, want 0", resp.TotalTokens)
	}
	if resp.TotalAmount != "0.00" {
		t.Errorf("total_amount = %v, want 0.00", resp.TotalAmount)
	}
}

func TestGetSummary_OnlyOldMonthData(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createTestUser(t, db, 1)

	// Data from 60 days ago should appear in total but not in month
	lastMonth := time.Now().AddDate(0, 0, -60)
	createCallLog(t, db, 1, 1000, 600, 400, lastMonth)
	createPaidOrder(t, db, 1, "200.00", lastMonth)

	resp, err := svc.GetSummary(1)
	if err != nil {
		t.Fatalf("GetSummary() error = %v", err)
	}
	if resp.TotalTokens != 1000 {
		t.Errorf("total_tokens = %v, want 1000", resp.TotalTokens)
	}
	if resp.MonthTokens != 0 {
		t.Errorf("month_tokens = %v, want 0", resp.MonthTokens)
	}
	if resp.TotalAmount != "200.00" {
		t.Errorf("total_amount = %v, want 200.00", resp.TotalAmount)
	}
	if resp.MonthAmount != "0.00" {
		t.Errorf("month_amount = %v, want 0.00", resp.MonthAmount)
	}
}

// ---------- helper tests ----------

func TestPeriodStartTime(t *testing.T) {
	tests := []struct {
		period  string
		wantErr bool
	}{
		{"day", false},
		{"week", false},
		{"month", false},
		{"year", false},
		{"", false}, // default to month
		{"invalid", true},
	}

	for _, tt := range tests {
		_, err := periodStartTime(tt.period)
		if (err != nil) != tt.wantErr {
			t.Errorf("periodStartTime(%q) error = %v, wantErr %v", tt.period, err, tt.wantErr)
		}
	}
}

func TestParseFloat(t *testing.T) {
	if v := parseFloat("123.45"); v != 123.45 {
		t.Errorf("parseFloat(123.45) = %v, want 123.45", v)
	}
	if v := parseFloat("0"); v != 0 {
		t.Errorf("parseFloat(0) = %v, want 0", v)
	}
	if v := parseFloat("bad"); v != 0 {
		t.Errorf("parseFloat(bad) = %v, want 0", v)
	}
}

// ---------- GetDashboardSummary tests ----------

func TestGetDashboardSummary_WithData(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	// Create users
	createTestUser(t, db, 1)
	createTestUser(t, db, 2)
	createTestUser(t, db, 3)

	// Create today's user
	uid := uniqueID()
	todayUser := model.User{
		Username:     "today_" + uid,
		Email:        "today_" + uid + "@test.com",
		PasswordHash: "hash",
		Role:         "user",
		Status:       1,
		CreatedAt:    time.Now(),
	}
	db.Create(&todayUser)

	// Create orders
	now := time.Now()
	createPaidOrder(t, db, 1, "100.00", now)
	createPaidOrder(t, db, 2, "200.00", now.Add(-1*time.Hour))

	// Create pending order
	pendingOrder := model.Order{
		OrderNo:     "ORD-PENDING-" + uniqueID(),
		UserID:      1,
		ProductID:   1,
		Quantity:    "1",
		UnitPrice:   "50.00",
		Amount:      "50.00",
		FinalAmount: "50.00",
		Currency:    "CNY",
		Status:      "pending",
		CreatedAt:   now,
	}
	db.Create(&pendingOrder)

	// Create user tokens
	for i := 0; i < 5; i++ {
		token := model.UserToken{
			UserID:      1,
			ProductID:   1,
			TotalAmount: "1000",
			UsedAmount:  "0",
			Status:      1,
			CreatedAt:   now,
		}
		db.Create(&token)
	}

	resp, err := svc.GetDashboardSummary()
	if err != nil {
		t.Fatalf("GetDashboardSummary() error = %v", err)
	}

	// Verify total users (3 + 1 today user = 4)
	if resp.TotalUsers < 3 {
		t.Errorf("total_users = %v, want >= 3", resp.TotalUsers)
	}

	// Verify today new users
	if resp.TodayNewUsers < 1 {
		t.Errorf("today_new_users = %v, want >= 1", resp.TodayNewUsers)
	}

	// Verify total orders (2 paid + 1 pending = 3)
	if resp.TotalOrders < 3 {
		t.Errorf("total_orders = %v, want >= 3", resp.TotalOrders)
	}

	// Verify today new orders
	if resp.TodayNewOrders < 3 {
		t.Errorf("today_new_orders = %v, want >= 3", resp.TodayNewOrders)
	}

	// Verify total revenue (100 + 200 = 300)
	if resp.TotalRevenue != "300.00" {
		t.Errorf("total_revenue = %v, want 300.00", resp.TotalRevenue)
	}

	// Verify active tokens
	if resp.ActiveTokens < 5 {
		t.Errorf("active_tokens = %v, want >= 5", resp.ActiveTokens)
	}
}

func TestGetDashboardSummary_EmptyDB(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	resp, err := svc.GetDashboardSummary()
	if err != nil {
		t.Fatalf("GetDashboardSummary() error = %v", err)
	}

	if resp.TotalUsers != 0 {
		t.Errorf("total_users = %v, want 0", resp.TotalUsers)
	}
	if resp.TodayNewUsers != 0 {
		t.Errorf("today_new_users = %v, want 0", resp.TodayNewUsers)
	}
	if resp.TotalOrders != 0 {
		t.Errorf("total_orders = %v, want 0", resp.TotalOrders)
	}
	if resp.TodayNewOrders != 0 {
		t.Errorf("today_new_orders = %v, want 0", resp.TodayNewOrders)
	}
	if resp.TotalRevenue != "0.00" {
		t.Errorf("total_revenue = %v, want 0.00", resp.TotalRevenue)
	}
	if resp.TodayRevenue != "0.00" {
		t.Errorf("today_revenue = %v, want 0.00", resp.TodayRevenue)
	}
	if resp.ActiveTokens != 0 {
		t.Errorf("active_tokens = %v, want 0", resp.ActiveTokens)
	}
}
