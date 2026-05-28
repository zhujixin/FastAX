package token

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
		&model.TokenProduct{},
		&model.TokenInventory{},
		&model.UserToken{},
		&model.TokenTransfer{},
		&model.Order{},
		&model.Payment{},
		&model.Refund{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func createProduct(t *testing.T, db *gorm.DB, name string, status int, sortOrder int, validityDays *int) model.TokenProduct {
	t.Helper()
	p := model.TokenProduct{
		SupplierID:   1,
		Name:         name,
		Model:        "gpt-4",
		Unit:         "tokens",
		Price:        "100.00",
		Currency:     "CNY",
		Status:       status,
		SortOrder:    sortOrder,
		ValidityDays: validityDays,
	}
	if err := db.Create(&p).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}
	return p
}

func TestService_GetProducts_OnlyActive(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	createProduct(t, db, "Active 1", 1, 1, nil)
	createProduct(t, db, "Active 2", 1, 2, nil)
	// Create as active then deactivate (GORM default:1 overrides zero value)
	inactive := createProduct(t, db, "Inactive", 1, 3, nil)
	db.Model(&model.TokenProduct{}).Where("id = ?", inactive.ID).Update("status", 0)

	products, err := svc.GetProducts()
	if err != nil {
		t.Fatalf("GetProducts() error = %v", err)
	}
	if len(products) != 2 {
		t.Errorf("len = %v, want 2", len(products))
	}
	for _, p := range products {
		if p.Status != 1 {
			t.Errorf("product %s has status %d, want 1", p.Name, p.Status)
		}
	}
}

func TestService_GetProducts_OrderBySortOrder(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	createProduct(t, db, "Second", 1, 2, nil)
	createProduct(t, db, "First", 1, 1, nil)

	products, err := svc.GetProducts()
	if err != nil {
		t.Fatalf("GetProducts() error = %v", err)
	}
	if len(products) != 2 {
		t.Fatalf("len = %v, want 2", len(products))
	}
	if products[0].Name != "First" {
		t.Errorf("first product = %v, want First", products[0].Name)
	}
	if products[1].Name != "Second" {
		t.Errorf("second product = %v, want Second", products[1].Name)
	}
}

func TestService_GetProducts_Empty(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	products, err := svc.GetProducts()
	if err != nil {
		t.Fatalf("GetProducts() error = %v", err)
	}
	if len(products) != 0 {
		t.Errorf("len = %v, want 0", len(products))
	}
}

func TestService_Buy_Success(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	product := createProduct(t, db, "Test Product", 1, 1, nil)

	resp, err := svc.Buy(1, &BuyRequest{
		ProductID: product.ID,
		Quantity:  2,
	})
	if err != nil {
		t.Fatalf("Buy() error = %v", err)
	}

	if !strings.HasPrefix(resp.OrderNo, "ORD") {
		t.Errorf("order_no = %v, want ORD prefix", resp.OrderNo)
	}
	if resp.Status != "paid" {
		t.Errorf("status = %v, want paid", resp.Status)
	}

	// Verify order in DB
	var order model.Order
	db.Where("order_no = ?", resp.OrderNo).First(&order)
	if order.Status != "paid" {
		t.Errorf("order status = %v, want paid", order.Status)
	}
	if order.PaidAt == nil {
		t.Error("paid_at should be set")
	}

	// Verify user token
	var token model.UserToken
	db.Where("user_id = ? AND product_id = ?", 1, product.ID).First(&token)
	if token.UserID != 1 {
		t.Errorf("user_id = %v, want 1", token.UserID)
	}
	if token.OrderID == nil || *token.OrderID != order.ID {
		t.Error("order_id mismatch")
	}
}

func TestService_Buy_ProductNotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	_, err := svc.Buy(1, &BuyRequest{
		ProductID: 99999,
		Quantity:  1,
	})
	if err == nil {
		t.Fatal("Buy() expected error for nonexistent product")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %v, want 'not found'", err)
	}
}

func TestService_Buy_InactiveProduct(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	// Create as active, then deactivate (GORM default:1 overrides zero value)
	product := createProduct(t, db, "Inactive", 1, 1, nil)
	db.Model(&model.TokenProduct{}).Where("id = ?", product.ID).Update("status", 0)

	_, err := svc.Buy(1, &BuyRequest{
		ProductID: product.ID,
		Quantity:  1,
	})
	if err == nil {
		t.Fatal("Buy() expected error for inactive product")
	}
	if !strings.Contains(err.Error(), "not available") {
		t.Errorf("error = %v, want 'not available'", err)
	}
}

func TestService_Buy_SetsExpiresAt_WithValidityDays(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	days := 30
	product := createProduct(t, db, "Expiring", 1, 1, &days)

	_, err := svc.Buy(1, &BuyRequest{
		ProductID: product.ID,
		Quantity:  1,
	})
	if err != nil {
		t.Fatalf("Buy() error = %v", err)
	}

	var token model.UserToken
	db.Where("user_id = ? AND product_id = ?", 1, product.ID).First(&token)
	if token.ExpiresAt == nil {
		t.Fatal("ExpiresAt should be set")
	}
	// Should be approximately 30 days from now
	diff := time.Until(*token.ExpiresAt)
	if diff < 29*24*time.Hour || diff > 31*24*time.Hour {
		t.Errorf("ExpiresAt is %v, approximately 30 days from now", token.ExpiresAt)
	}
}

func TestService_Buy_NoExpiresAt_WithoutValidityDays(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	product := createProduct(t, db, "No Expiry", 1, 1, nil)

	_, err := svc.Buy(1, &BuyRequest{
		ProductID: product.ID,
		Quantity:  1,
	})
	if err != nil {
		t.Fatalf("Buy() error = %v", err)
	}

	var token model.UserToken
	db.Where("user_id = ? AND product_id = ?", 1, product.ID).First(&token)
	if token.ExpiresAt != nil {
		t.Errorf("ExpiresAt should be nil, got %v", token.ExpiresAt)
	}
}

func TestService_GetUserTokens_ReturnsUserSpecificTokens(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	p1 := createProduct(t, db, "Product A", 1, 1, nil)
	p2 := createProduct(t, db, "Product B", 1, 2, nil)

	// User 1 buys product A
	svc.Buy(1, &BuyRequest{ProductID: p1.ID, Quantity: 1})
	// User 2 buys product B
	svc.Buy(2, &BuyRequest{ProductID: p2.ID, Quantity: 1})

	tokens, err := svc.GetUserTokens(1)
	if err != nil {
		t.Fatalf("GetUserTokens() error = %v", err)
	}
	if len(tokens) != 1 {
		t.Fatalf("len = %v, want 1", len(tokens))
	}
	if tokens[0].ProductID != p1.ID {
		t.Errorf("product_id = %v, want %v", tokens[0].ProductID, p1.ID)
	}
	if tokens[0].Name != "Product A" {
		t.Errorf("name = %v, want Product A", tokens[0].Name)
	}
}

func TestService_GetUserTokens_Empty(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	tokens, err := svc.GetUserTokens(999)
	if err != nil {
		t.Fatalf("GetUserTokens() error = %v", err)
	}
	if len(tokens) != 0 {
		t.Errorf("len = %v, want 0", len(tokens))
	}
}

func TestService_GetUserTokens_OnlyActive(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	product := createProduct(t, db, "Active", 1, 1, nil)

	// Buy and then deactivate the token
	svc.Buy(1, &BuyRequest{ProductID: product.ID, Quantity: 1})
	db.Model(&model.UserToken{}).Where("user_id = ? AND product_id = ?", 1, product.ID).Update("status", 0)

	tokens, err := svc.GetUserTokens(1)
	if err != nil {
		t.Fatalf("GetUserTokens() error = %v", err)
	}
	if len(tokens) != 0 {
		t.Errorf("len = %v, want 0", len(tokens))
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
}
