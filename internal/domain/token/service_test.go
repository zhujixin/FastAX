package token

import (
	"fmt"
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
		&model.User{},
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

func createTestUser(t *testing.T, db *gorm.DB, id uint, username string) {
	t.Helper()
	user := model.User{
		ID:           id,
		Username:     username,
		PasswordHash: "hash",
		Email:        username + "@test.com",
		Phone:        fmt.Sprintf("1380000%04d", id),
		Role:         "user",
		Status:       1,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
}

func createUserToken(t *testing.T, db *gorm.DB, userID uint, productID uint, total, used, frozen string) model.UserToken {
	t.Helper()
	ut := model.UserToken{
		UserID:       userID,
		ProductID:    productID,
		TotalAmount:  total,
		UsedAmount:   used,
		FrozenAmount: frozen,
		Status:       1,
	}
	if err := db.Create(&ut).Error; err != nil {
		t.Fatalf("create user token: %v", err)
	}
	return ut
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

// ===== Transfer Tests =====

func TestService_Transfer_Success(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	createTestUser(t, db, 1, "sender")
	createTestUser(t, db, 2, "receiver")
	product := createProduct(t, db, "Transfer Product", 1, 1, nil)
	createUserToken(t, db, 1, product.ID, "1000.00", "0", "0")

	resp, err := svc.Transfer(1, &TransferRequest{
		ProductID: product.ID,
		ToUserID:  2,
		Amount:    "300.00",
	})
	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}
	if resp.Status != "completed" {
		t.Errorf("status = %v, want completed", resp.Status)
	}

	// Verify sender's used amount increased
	var senderToken model.UserToken
	db.Where("user_id = ? AND product_id = ?", 1, product.ID).First(&senderToken)
	if senderToken.UsedAmount != "300.00" {
		t.Errorf("sender used_amount = %v, want 300.00", senderToken.UsedAmount)
	}

	// Verify receiver got a new token
	var receiverToken model.UserToken
	db.Where("user_id = ? AND product_id = ?", 2, product.ID).First(&receiverToken)
	if receiverToken.TotalAmount != "300.00" {
		t.Errorf("receiver total_amount = %v, want 300.00", receiverToken.TotalAmount)
	}

	// Verify transfer record
	var transfer model.TokenTransfer
	db.Where("from_user_id = ? AND to_user_id = ?", 1, 2).First(&transfer)
	if transfer.Amount != "300.00" {
		t.Errorf("transfer amount = %v, want 300.00", transfer.Amount)
	}
	if transfer.Status != "completed" {
		t.Errorf("transfer status = %v, want completed", transfer.Status)
	}
	if transfer.HandledAt == nil {
		t.Error("handled_at should be set")
	}
}

func TestService_Transfer_ToExistingReceiverToken(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	createTestUser(t, db, 1, "sender")
	createTestUser(t, db, 2, "receiver")
	product := createProduct(t, db, "Product", 1, 1, nil)
	createUserToken(t, db, 1, product.ID, "1000.00", "0", "0")
	createUserToken(t, db, 2, product.ID, "500.00", "0", "0")

	_, err := svc.Transfer(1, &TransferRequest{
		ProductID: product.ID,
		ToUserID:  2,
		Amount:    "200.00",
	})
	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}

	// Receiver's total should be 500 + 200 = 700
	var receiverToken model.UserToken
	db.Where("user_id = ? AND product_id = ?", 2, product.ID).First(&receiverToken)
	if receiverToken.TotalAmount != "700.00" {
		t.Errorf("receiver total_amount = %v, want 700.00", receiverToken.TotalAmount)
	}
}

func TestService_Transfer_InsufficientBalance(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	createTestUser(t, db, 1, "sender")
	createTestUser(t, db, 2, "receiver")
	product := createProduct(t, db, "Product", 1, 1, nil)
	createUserToken(t, db, 1, product.ID, "100.00", "0", "0")

	_, err := svc.Transfer(1, &TransferRequest{
		ProductID: product.ID,
		ToUserID:  2,
		Amount:    "200.00",
	})
	if err == nil {
		t.Fatal("Transfer() expected error for insufficient balance")
	}
	if !strings.Contains(err.Error(), "insufficient balance") {
		t.Errorf("error = %v, want 'insufficient balance'", err)
	}
}

func TestService_Transfer_InsufficientBalance_AfterUse(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	createTestUser(t, db, 1, "sender")
	createTestUser(t, db, 2, "receiver")
	product := createProduct(t, db, "Product", 1, 1, nil)
	// Total 1000, already used 800, remaining = 200
	createUserToken(t, db, 1, product.ID, "1000.00", "800.00", "0")

	_, err := svc.Transfer(1, &TransferRequest{
		ProductID: product.ID,
		ToUserID:  2,
		Amount:    "300.00",
	})
	if err == nil {
		t.Fatal("Transfer() expected error for insufficient balance")
	}
	if !strings.Contains(err.Error(), "insufficient balance") {
		t.Errorf("error = %v, want 'insufficient balance'", err)
	}
}

func TestService_Transfer_InsufficientBalance_AfterFrozen(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	createTestUser(t, db, 1, "sender")
	createTestUser(t, db, 2, "receiver")
	product := createProduct(t, db, "Product", 1, 1, nil)
	// Total 1000, used 500, frozen 300, remaining = 200
	createUserToken(t, db, 1, product.ID, "1000.00", "500.00", "300.00")

	_, err := svc.Transfer(1, &TransferRequest{
		ProductID: product.ID,
		ToUserID:  2,
		Amount:    "300.00",
	})
	if err == nil {
		t.Fatal("Transfer() expected error for insufficient balance")
	}
}

func TestService_Transfer_SelfTransfer(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	createTestUser(t, db, 1, "user")
	product := createProduct(t, db, "Product", 1, 1, nil)
	createUserToken(t, db, 1, product.ID, "1000.00", "0", "0")

	_, err := svc.Transfer(1, &TransferRequest{
		ProductID: product.ID,
		ToUserID:  1,
		Amount:    "100.00",
	})
	if err == nil {
		t.Fatal("Transfer() expected error for self-transfer")
	}
	if !strings.Contains(err.Error(), "cannot transfer to yourself") {
		t.Errorf("error = %v, want 'cannot transfer to yourself'", err)
	}
}

func TestService_Transfer_NoTokenFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	createTestUser(t, db, 1, "sender")
	createTestUser(t, db, 2, "receiver")
	product := createProduct(t, db, "Product", 1, 1, nil)

	_, err := svc.Transfer(1, &TransferRequest{
		ProductID: product.ID,
		ToUserID:  2,
		Amount:    "100.00",
	})
	if err == nil {
		t.Fatal("Transfer() expected error when no token found")
	}
	if !strings.Contains(err.Error(), "no active token found") {
		t.Errorf("error = %v, want 'no active token found'", err)
	}
}

func TestService_Transfer_ReceiverNotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	createTestUser(t, db, 1, "sender")
	product := createProduct(t, db, "Product", 1, 1, nil)
	createUserToken(t, db, 1, product.ID, "1000.00", "0", "0")

	_, err := svc.Transfer(1, &TransferRequest{
		ProductID: product.ID,
		ToUserID:  999,
		Amount:    "100.00",
	})
	if err == nil {
		t.Fatal("Transfer() expected error for nonexistent receiver")
	}
	if !strings.Contains(err.Error(), "receiver user not found") {
		t.Errorf("error = %v, want 'receiver user not found'", err)
	}
}

func TestService_Transfer_InvalidAmount(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	createTestUser(t, db, 1, "sender")
	createTestUser(t, db, 2, "receiver")
	product := createProduct(t, db, "Product", 1, 1, nil)
	createUserToken(t, db, 1, product.ID, "1000.00", "0", "0")

	_, err := svc.Transfer(1, &TransferRequest{
		ProductID: product.ID,
		ToUserID:  2,
		Amount:    "0",
	})
	if err == nil {
		t.Fatal("Transfer() expected error for zero amount")
	}

	_, err = svc.Transfer(1, &TransferRequest{
		ProductID: product.ID,
		ToUserID:  2,
		Amount:    "-50",
	})
	if err == nil {
		t.Fatal("Transfer() expected error for negative amount")
	}
}

func TestService_Transfer_ExpiredToken(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	createTestUser(t, db, 1, "sender")
	createTestUser(t, db, 2, "receiver")
	product := createProduct(t, db, "Product", 1, 1, nil)

	pastTime := time.Now().Add(-24 * time.Hour)
	ut := model.UserToken{
		UserID:      1,
		ProductID:   product.ID,
		TotalAmount: "1000.00",
		UsedAmount:  "0",
		Status:      1,
		ExpiresAt:   &pastTime,
	}
	db.Create(&ut)

	_, err := svc.Transfer(1, &TransferRequest{
		ProductID: product.ID,
		ToUserID:  2,
		Amount:    "100.00",
	})
	if err == nil {
		t.Fatal("Transfer() expected error for expired token")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("error = %v, want 'expired'", err)
	}
}

// ===== Extract Tests =====

func TestService_Extract_Success(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	product := createProduct(t, db, "Extract Product", 1, 1, nil)
	createUserToken(t, db, 1, product.ID, "1000.00", "0", "0")

	resp, err := svc.Extract(1, &ExtractRequest{
		ProductID: product.ID,
		Amount:    "500.00",
		Address:   "0x1234567890abcdef",
	})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	if resp.Status != "completed" {
		t.Errorf("status = %v, want completed", resp.Status)
	}
	if resp.TransferID == 0 {
		t.Error("transfer_id should be nonzero")
	}

	// Verify user's used amount increased
	var userToken model.UserToken
	db.Where("user_id = ? AND product_id = ?", 1, product.ID).First(&userToken)
	if userToken.UsedAmount != "500.00" {
		t.Errorf("used_amount = %v, want 500.00", userToken.UsedAmount)
	}

	// Verify transfer record
	var transfer model.TokenTransfer
	db.Where("from_user_id = ? AND to_user_id = ?", 1, 0).First(&transfer)
	if transfer.Amount != "500.00" {
		t.Errorf("transfer amount = %v, want 500.00", transfer.Amount)
	}
	if transfer.Status != "completed" {
		t.Errorf("transfer status = %v, want completed", transfer.Status)
	}
}

func TestService_Extract_InsufficientBalance(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	product := createProduct(t, db, "Product", 1, 1, nil)
	createUserToken(t, db, 1, product.ID, "100.00", "0", "0")

	_, err := svc.Extract(1, &ExtractRequest{
		ProductID: product.ID,
		Amount:    "200.00",
		Address:   "0xabc",
	})
	if err == nil {
		t.Fatal("Extract() expected error for insufficient balance")
	}
	if !strings.Contains(err.Error(), "insufficient balance") {
		t.Errorf("error = %v, want 'insufficient balance'", err)
	}
}

func TestService_Extract_InsufficientBalance_AfterUse(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	product := createProduct(t, db, "Product", 1, 1, nil)
	// Total 1000, used 800, remaining = 200
	createUserToken(t, db, 1, product.ID, "1000.00", "800.00", "0")

	_, err := svc.Extract(1, &ExtractRequest{
		ProductID: product.ID,
		Amount:    "300.00",
		Address:   "0xabc",
	})
	if err == nil {
		t.Fatal("Extract() expected error for insufficient balance")
	}
}

func TestService_Extract_NoTokenFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	product := createProduct(t, db, "Product", 1, 1, nil)

	_, err := svc.Extract(1, &ExtractRequest{
		ProductID: product.ID,
		Amount:    "100.00",
		Address:   "0xabc",
	})
	if err == nil {
		t.Fatal("Extract() expected error when no token found")
	}
	if !strings.Contains(err.Error(), "no active token found") {
		t.Errorf("error = %v, want 'no active token found'", err)
	}
}

func TestService_Extract_InvalidAmount(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	product := createProduct(t, db, "Product", 1, 1, nil)
	createUserToken(t, db, 1, product.ID, "1000.00", "0", "0")

	_, err := svc.Extract(1, &ExtractRequest{
		ProductID: product.ID,
		Amount:    "0",
		Address:   "0xabc",
	})
	if err == nil {
		t.Fatal("Extract() expected error for zero amount")
	}

	_, err = svc.Extract(1, &ExtractRequest{
		ProductID: product.ID,
		Amount:    "-100",
		Address:   "0xabc",
	})
	if err == nil {
		t.Fatal("Extract() expected error for negative amount")
	}
}

func TestService_Extract_ExpiredToken(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	product := createProduct(t, db, "Product", 1, 1, nil)
	pastTime := time.Now().Add(-24 * time.Hour)
	ut := model.UserToken{
		UserID:      1,
		ProductID:   product.ID,
		TotalAmount: "1000.00",
		UsedAmount:  "0",
		Status:      1,
		ExpiresAt:   &pastTime,
	}
	db.Create(&ut)

	_, err := svc.Extract(1, &ExtractRequest{
		ProductID: product.ID,
		Amount:    "100.00",
		Address:   "0xabc",
	})
	if err == nil {
		t.Fatal("Extract() expected error for expired token")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("error = %v, want 'expired'", err)
	}
}

func TestService_Extract_AccountsForFrozen(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	product := createProduct(t, db, "Product", 1, 1, nil)
	// Total 1000, used 500, frozen 300, remaining = 200
	createUserToken(t, db, 1, product.ID, "1000.00", "500.00", "300.00")

	_, err := svc.Extract(1, &ExtractRequest{
		ProductID: product.ID,
		Amount:    "300.00",
		Address:   "0xabc",
	})
	if err == nil {
		t.Fatal("Extract() expected error when remaining (after frozen) is insufficient")
	}

	// Should succeed with 200
	resp, err := svc.Extract(1, &ExtractRequest{
		ProductID: product.ID,
		Amount:    "200.00",
		Address:   "0xabc",
	})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	if resp.Status != "completed" {
		t.Errorf("status = %v, want completed", resp.Status)
	}
}
