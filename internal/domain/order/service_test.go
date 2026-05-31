package order

import (
	"strings"
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
		&model.Order{}, &model.Payment{}, &model.Refund{},
		&model.TokenProduct{}, &model.User{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func createTestProduct(t *testing.T, db *gorm.DB) model.TokenProduct {
	t.Helper()
	p := model.TokenProduct{
		SupplierID: 1,
		Name:       "Test Product",
		Model:      "gpt-4",
		Unit:       "tokens",
		Price:      "100.00",
		Currency:   "CNY",
		Status:     1,
	}
	if err := db.Create(&p).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}
	return p
}

func TestService_Create_Success(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	product := createTestProduct(t, db)

	resp, err := svc.Create(1, &CreateOrderRequest{
		ProductID: product.ID,
		Quantity:  1,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if !strings.HasPrefix(resp.OrderNo, "ORD") {
		t.Errorf("order_no = %v, want ORD prefix", resp.OrderNo)
	}
	if resp.Status != "pending" {
		t.Errorf("status = %v, want pending", resp.Status)
	}
	if resp.UserID != 1 {
		t.Errorf("user_id = %v, want 1", resp.UserID)
	}
	if resp.ProductID != product.ID {
		t.Errorf("product_id = %v, want %v", resp.ProductID, product.ID)
	}
}

func TestCreate_ProductNotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	_, err := svc.Create(1, &CreateOrderRequest{
		ProductID: 99999,
		Quantity:  1,
	})
	if err == nil {
		t.Fatal("expected error for nonexistent product")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %v, want 'not found'", err)
	}
}

func TestCreate_InactiveProduct(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	product := createTestProduct(t, db)
	db.Model(&model.TokenProduct{}).Where("id = ?", product.ID).Update("status", 0)

	_, err := svc.Create(1, &CreateOrderRequest{
		ProductID: product.ID,
		Quantity:  1,
	})
	if err == nil {
		t.Fatal("expected error for inactive product")
	}
	if !strings.Contains(err.Error(), "not available") {
		t.Errorf("error = %v, want 'not available'", err)
	}
}

func TestCreate_DuplicatePendingOrder(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	product := createTestProduct(t, db)

	_, err := svc.Create(1, &CreateOrderRequest{ProductID: product.ID, Quantity: 1})
	if err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	_, err = svc.Create(1, &CreateOrderRequest{ProductID: product.ID, Quantity: 1})
	if err == nil {
		t.Fatal("expected error for duplicate pending order")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("error = %v, want 'duplicate'", err)
	}
}

func TestService_GetByID(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	product := createTestProduct(t, db)

	created, _ := svc.Create(1, &CreateOrderRequest{ProductID: product.ID, Quantity: 1})

	found, err := svc.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found.OrderNo != created.OrderNo {
		t.Errorf("order_no = %v, want %v", found.OrderNo, created.OrderNo)
	}
}

func TestService_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	_, err := svc.GetByID(99999)
	if err == nil {
		t.Fatal("expected error for nonexistent order")
	}
}

func TestService_GetByOrderNo(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	product := createTestProduct(t, db)

	created, _ := svc.Create(1, &CreateOrderRequest{ProductID: product.ID, Quantity: 1})

	found, err := svc.GetByOrderNo(created.OrderNo)
	if err != nil {
		t.Fatalf("GetByOrderNo() error = %v", err)
	}
	if found.ID != created.ID {
		t.Errorf("id = %v, want %v", found.ID, created.ID)
	}
}

func TestService_List_FilterByStatus(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	product := createTestProduct(t, db)

	svc.Create(1, &CreateOrderRequest{ProductID: product.ID, Quantity: 1})
	// Create another product to avoid duplicate check
	p2 := model.TokenProduct{
		SupplierID: 1, Name: "P2", Model: "gpt-4", Unit: "tokens",
		Price: "50.00", Currency: "CNY", Status: 1,
	}
	db.Create(&p2)
	svc.Create(1, &CreateOrderRequest{ProductID: p2.ID, Quantity: 1})

	items, total, err := svc.List(&OrderQuery{UserID: 1, Status: "pending"})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 2 {
		t.Errorf("total = %v, want 2", total)
	}
	if len(items) != 2 {
		t.Errorf("len = %v, want 2", len(items))
	}
}

func TestService_List_Pagination(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	// Create 5 orders with different products
	for i := range 5 {
		p := model.TokenProduct{
			SupplierID: 1, Name: "P" + string(rune('A'+i)), Model: "gpt-4",
			Unit: "tokens", Price: "10.00", Currency: "CNY", Status: 1,
		}
		db.Create(&p)
		svc.Create(1, &CreateOrderRequest{ProductID: p.ID, Quantity: 1})
	}

	items, total, err := svc.List(&OrderQuery{UserID: 1, Page: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 5 {
		t.Errorf("total = %v, want 5", total)
	}
	if len(items) != 2 {
		t.Errorf("len = %v, want 2", len(items))
	}
}

func TestService_UpdateStatus_Valid(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	product := createTestProduct(t, db)

	created, _ := svc.Create(1, &CreateOrderRequest{ProductID: product.ID, Quantity: 1})

	err := svc.UpdateStatus(created.ID, "pending", "paid")
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	found, _ := svc.GetByID(created.ID)
	if found.Status != "paid" {
		t.Errorf("status = %v, want paid", found.Status)
	}
}

func TestService_UpdateStatus_Invalid(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	product := createTestProduct(t, db)

	created, _ := svc.Create(1, &CreateOrderRequest{ProductID: product.ID, Quantity: 1})

	err := svc.UpdateStatus(created.ID, "pending", "completed")
	if err == nil {
		t.Fatal("expected error for invalid transition")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("error = %v, want 'invalid'", err)
	}
}

func TestService_MarkPaid(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	product := createTestProduct(t, db)

	created, _ := svc.Create(1, &CreateOrderRequest{ProductID: product.ID, Quantity: 1})

	err := svc.MarkPaid(created.ID)
	if err != nil {
		t.Fatalf("MarkPaid() error = %v", err)
	}

	found, _ := svc.GetByID(created.ID)
	if found.Status != "paid" {
		t.Errorf("status = %v, want paid", found.Status)
	}
	if found.PaidAt == nil {
		t.Error("paid_at should be set")
	}
}

func TestService_MarkPaid_NotPending(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	product := createTestProduct(t, db)

	created, _ := svc.Create(1, &CreateOrderRequest{ProductID: product.ID, Quantity: 1})
	svc.MarkPaid(created.ID)

	err := svc.MarkPaid(created.ID)
	if err == nil {
		t.Fatal("expected error for non-pending order")
	}
}

func TestService_MarkCompleted(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	product := createTestProduct(t, db)

	created, _ := svc.Create(1, &CreateOrderRequest{ProductID: product.ID, Quantity: 1})
	svc.MarkPaid(created.ID)

	err := svc.MarkCompleted(created.ID)
	if err != nil {
		t.Fatalf("MarkCompleted() error = %v", err)
	}

	found, _ := svc.GetByID(created.ID)
	if found.Status != "completed" {
		t.Errorf("status = %v, want completed", found.Status)
	}
}

func TestService_Cancel(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	product := createTestProduct(t, db)

	created, _ := svc.Create(1, &CreateOrderRequest{ProductID: product.ID, Quantity: 1})

	err := svc.Cancel(created.ID)
	if err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}

	found, _ := svc.GetByID(created.ID)
	if found.Status != "cancelled" {
		t.Errorf("status = %v, want cancelled", found.Status)
	}
}

func TestService_Cancel_NotPending(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	product := createTestProduct(t, db)

	created, _ := svc.Create(1, &CreateOrderRequest{ProductID: product.ID, Quantity: 1})
	svc.MarkPaid(created.ID)

	err := svc.Cancel(created.ID)
	if err == nil {
		t.Fatal("expected error for non-pending order")
	}
}

func TestService_CancelTimeoutOrders(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	product := createTestProduct(t, db)

	// Create an order and backdate it
	created, _ := svc.Create(1, &CreateOrderRequest{ProductID: product.ID, Quantity: 1})
	db.Model(&model.Order{}).Where("id = ?", created.ID).
		Update("created_at", time.Now().Add(-31*time.Minute))

	count, err := svc.CancelTimeoutOrders()
	if err != nil {
		t.Fatalf("CancelTimeoutOrders() error = %v", err)
	}
	if count != 1 {
		t.Errorf("cancelled count = %v, want 1", count)
	}

	found, _ := svc.GetByID(created.ID)
	if found.Status != "cancelled" {
		t.Errorf("status = %v, want cancelled", found.Status)
	}
}

func TestService_CancelTimeoutOrders_NoExpired(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	product := createTestProduct(t, db)

	svc.Create(1, &CreateOrderRequest{ProductID: product.ID, Quantity: 1})

	count, err := svc.CancelTimeoutOrders()
	if err != nil {
		t.Fatalf("CancelTimeoutOrders() error = %v", err)
	}
	if count != 0 {
		t.Errorf("cancelled count = %v, want 0", count)
	}
}

func TestService_FullLifecycle(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	product := createTestProduct(t, db)

	// Create → Paid → Completed
	created, _ := svc.Create(1, &CreateOrderRequest{ProductID: product.ID, Quantity: 1})
	if created.Status != "pending" {
		t.Fatalf("initial status = %v, want pending", created.Status)
	}

	svc.MarkPaid(created.ID)
	paid, _ := svc.GetByID(created.ID)
	if paid.Status != "paid" {
		t.Fatalf("after pay status = %v, want paid", paid.Status)
	}

	svc.MarkCompleted(created.ID)
	completed, _ := svc.GetByID(created.ID)
	if completed.Status != "completed" {
		t.Fatalf("after complete status = %v, want completed", completed.Status)
	}
}

func TestGenerateOrderNo(t *testing.T) {
	no := generateOrderNo()
	if !strings.HasPrefix(no, "ORD") {
		t.Errorf("order_no = %v, want ORD prefix", no)
	}
	if len(no) <= 3 {
		t.Errorf("order_no too short: %v", no)
	}
	// Test uniqueness
	no2 := generateOrderNo()
	if no == no2 {
		t.Errorf("duplicate order_no: %v", no)
	}
}
