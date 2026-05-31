package payment

import (
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
	if err := db.AutoMigrate(
		&model.Order{}, &model.Payment{}, &model.Refund{},
		&model.TokenProduct{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func createPendingOrder(t *testing.T, db *gorm.DB) model.Order {
	t.Helper()
	order := model.Order{
		OrderNo:     "ORD_TEST_001",
		UserID:      1,
		ProductID:   1,
		Quantity:    "1",
		UnitPrice:   "100.00",
		Amount:      "100.00",
		FinalAmount: "100.00",
		Currency:    "CNY",
		Status:      "pending",
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}
	return order
}

func createPaidOrder(t *testing.T, db *gorm.DB) (model.Order, model.Payment) {
	t.Helper()
	order := model.Order{
		OrderNo:     "ORD_TEST_002",
		UserID:      1,
		ProductID:   1,
		Quantity:    "1",
		UnitPrice:   "200.00",
		Amount:      "200.00",
		FinalAmount: "200.00",
		Currency:    "CNY",
		Status:      "paid",
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	payment := model.Payment{
		OrderID:   order.ID,
		PaymentNo: "PAY_TEST_001",
		Amount:    "200.00",
		Method:    "wechat",
		Gateway:   "wechat",
		Status:    "success",
	}
	if err := db.Create(&payment).Error; err != nil {
		t.Fatalf("create payment: %v", err)
	}
	return order, payment
}

func TestService_CreatePayment_Success(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	order := createPendingOrder(t, db)

	resp, err := svc.CreatePayment(&CreatePaymentRequest{
		OrderID: order.ID,
		Method:  "wechat",
		Gateway: "wechat",
	})
	if err != nil {
		t.Fatalf("CreatePayment() error = %v", err)
	}

	if !strings.HasPrefix(resp.PaymentNo, "PAY") {
		t.Errorf("payment_no = %v, want PAY prefix", resp.PaymentNo)
	}
	if resp.Status != "pending" {
		t.Errorf("status = %v, want pending", resp.Status)
	}
	if resp.Amount != "100.00" {
		t.Errorf("amount = %v, want 100.00", resp.Amount)
	}
	if resp.Method != "wechat" {
		t.Errorf("method = %v, want wechat", resp.Method)
	}
}

func TestCreatePayment_OrderNotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	_, err := svc.CreatePayment(&CreatePaymentRequest{
		OrderID: 99999,
		Method:  "wechat",
		Gateway: "wechat",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent order")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %v, want 'not found'", err)
	}
}

func TestCreatePayment_OrderNotPending(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	order, _ := createPaidOrder(t, db)

	_, err := svc.CreatePayment(&CreatePaymentRequest{
		OrderID: order.ID,
		Method:  "wechat",
		Gateway: "wechat",
	})
	if err == nil {
		t.Fatal("expected error for non-pending order")
	}
	if !strings.Contains(err.Error(), "not in pending") {
		t.Errorf("error = %v, want 'not in pending'", err)
	}
}

func TestService_HandleCallback_Success(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	order := createPendingOrder(t, db)
	svc.CreatePayment(&CreatePaymentRequest{OrderID: order.ID, Method: "wechat", Gateway: "wechat"})

	err := svc.HandleCallback(&PaymentCallback{
		OrderID:        order.ID,
		GatewayTradeNo: "WX123456",
		GatewayStatus:  "SUCCESS",
		Success:        true,
		RawResponse:    `<xml>success</xml>`,
	})
	if err != nil {
		t.Fatalf("HandleCallback() error = %v", err)
	}

	// Verify payment updated
	payment, _ := svc.GetPaymentByOrderID(order.ID)
	if payment.Status != "success" {
		t.Errorf("payment status = %v, want success", payment.Status)
	}
	if payment.PaidAt == nil {
		t.Error("paid_at should be set")
	}

	// Verify order updated
	var updatedOrder model.Order
	db.First(&updatedOrder, order.ID)
	if updatedOrder.Status != "paid" {
		t.Errorf("order status = %v, want paid", updatedOrder.Status)
	}
}

func TestHandleCallback_Failure(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	order := createPendingOrder(t, db)
	svc.CreatePayment(&CreatePaymentRequest{OrderID: order.ID, Method: "wechat", Gateway: "wechat"})

	err := svc.HandleCallback(&PaymentCallback{
		OrderID:        order.ID,
		GatewayTradeNo: "WX789",
		GatewayStatus:  "FAIL",
		Success:        false,
	})
	if err != nil {
		t.Fatalf("HandleCallback() error = %v", err)
	}

	payment, _ := svc.GetPaymentByOrderID(order.ID)
	if payment.Status != "failed" {
		t.Errorf("payment status = %v, want failed", payment.Status)
	}

	// Order should still be pending
	var updatedOrder model.Order
	db.First(&updatedOrder, order.ID)
	if updatedOrder.Status != "pending" {
		t.Errorf("order status = %v, want pending", updatedOrder.Status)
	}
}

func TestHandleCallback_Idempotent(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	order := createPendingOrder(t, db)
	svc.CreatePayment(&CreatePaymentRequest{OrderID: order.ID, Method: "wechat", Gateway: "wechat"})

	cb := &PaymentCallback{
		OrderID:        order.ID,
		GatewayTradeNo: "WX123",
		GatewayStatus:  "SUCCESS",
		Success:        true,
	}

	svc.HandleCallback(cb)
	err := svc.HandleCallback(cb)
	if err != nil {
		t.Fatalf("second HandleCallback() error = %v (should be idempotent)", err)
	}
}

func TestHandleCallback_NoPendingPayment(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	err := svc.HandleCallback(&PaymentCallback{OrderID: 999, Success: true})
	if err == nil {
		t.Fatal("expected error for no pending payment")
	}
}

func TestService_CreateRefund_Success(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	order, _ := createPaidOrder(t, db)

	resp, err := svc.CreateRefund(&RefundRequest{
		OrderID: order.ID,
		Amount:  "200.00",
		Reason:  "not satisfied",
	}, 1)
	if err != nil {
		t.Fatalf("CreateRefund() error = %v", err)
	}

	if !strings.HasPrefix(resp.RefundNo, "REF") {
		t.Errorf("refund_no = %v, want REF prefix", resp.RefundNo)
	}
	if resp.Status != "pending" {
		t.Errorf("status = %v, want pending", resp.Status)
	}

	// Order should be refunding
	var updatedOrder model.Order
	db.First(&updatedOrder, order.ID)
	if updatedOrder.Status != "refunding" {
		t.Errorf("order status = %v, want refunding", updatedOrder.Status)
	}
}

func TestCreateRefund_OrderNotPaid(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	order := createPendingOrder(t, db)

	_, err := svc.CreateRefund(&RefundRequest{
		OrderID: order.ID,
		Amount:  "100.00",
	}, 1)
	if err == nil {
		t.Fatal("expected error for non-paid order")
	}
	if !strings.Contains(err.Error(), "paid") {
		t.Errorf("error = %v, want 'paid'", err)
	}
}

func TestCreateRefund_NoSuccessfulPayment(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	// Create paid order without payment record
	order := model.Order{
		OrderNo: "ORD_NO_PAY", UserID: 1, ProductID: 1,
		Quantity: "1", UnitPrice: "100", Amount: "100",
		FinalAmount: "100", Currency: "CNY", Status: "paid",
	}
	db.Create(&order)

	_, err := svc.CreateRefund(&RefundRequest{
		OrderID: order.ID,
		Amount:  "100.00",
	}, 1)
	if err == nil {
		t.Fatal("expected error for no successful payment")
	}
}

func TestService_ReviewRefund_Approve(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	order, _ := createPaidOrder(t, db)

	refund, _ := svc.CreateRefund(&RefundRequest{
		OrderID: order.ID, Amount: "200.00", Reason: "test",
	}, 1)

	err := svc.ReviewRefund(&RefundReview{
		RefundID: refund.ID,
		Approved: true,
		Remark:   "approved by admin",
	}, 2)
	if err != nil {
		t.Fatalf("ReviewRefund() error = %v", err)
	}

	// Order should be refunded
	var updatedOrder model.Order
	db.First(&updatedOrder, order.ID)
	if updatedOrder.Status != "refunded" {
		t.Errorf("order status = %v, want refunded", updatedOrder.Status)
	}
}

func TestService_ReviewRefund_Reject(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	order, _ := createPaidOrder(t, db)

	refund, _ := svc.CreateRefund(&RefundRequest{
		OrderID: order.ID, Amount: "200.00", Reason: "test",
	}, 1)

	err := svc.ReviewRefund(&RefundReview{
		RefundID: refund.ID,
		Approved: false,
		Remark:   "not eligible",
	}, 2)
	if err != nil {
		t.Fatalf("ReviewRefund() error = %v", err)
	}

	// Order should revert to paid
	var updatedOrder model.Order
	db.First(&updatedOrder, order.ID)
	if updatedOrder.Status != "paid" {
		t.Errorf("order status = %v, want paid", updatedOrder.Status)
	}
}

func TestReviewRefund_AlreadyHandled(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	order, _ := createPaidOrder(t, db)

	refund, _ := svc.CreateRefund(&RefundRequest{
		OrderID: order.ID, Amount: "200.00",
	}, 1)

	svc.ReviewRefund(&RefundReview{RefundID: refund.ID, Approved: true}, 2)

	err := svc.ReviewRefund(&RefundReview{RefundID: refund.ID, Approved: true}, 2)
	if err == nil {
		t.Fatal("expected error for already handled refund")
	}
}

func TestService_ListRefunds(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	// Create two separate paid orders (order becomes "refunding" after first refund)
	order1, _ := createPaidOrder(t, db)
	order2 := model.Order{
		OrderNo: "ORD_TEST_003", UserID: 1, ProductID: 1,
		Quantity: "1", UnitPrice: "300.00", Amount: "300.00",
		FinalAmount: "300.00", Currency: "CNY", Status: "paid",
	}
	db.Create(&order2)
	payment2 := model.Payment{
		OrderID: order2.ID, PaymentNo: "PAY_TEST_002",
		Amount: "300.00", Method: "alipay", Gateway: "alipay", Status: "success",
	}
	db.Create(&payment2)

	svc.CreateRefund(&RefundRequest{OrderID: order1.ID, Amount: "100.00", Reason: "r1"}, 1)
	svc.CreateRefund(&RefundRequest{OrderID: order2.ID, Amount: "50.00", Reason: "r2"}, 1)

	// List all refunds (no order filter)
	refunds, err := svc.ListRefunds(0)
	if err != nil {
		t.Fatalf("ListRefunds() error = %v", err)
	}
	if len(refunds) != 2 {
		t.Errorf("len = %v, want 2", len(refunds))
	}

	// List refunds for specific order
	refunds1, err := svc.ListRefunds(order1.ID)
	if err != nil {
		t.Fatalf("ListRefunds(order1) error = %v", err)
	}
	if len(refunds1) != 1 {
		t.Errorf("len = %v, want 1", len(refunds1))
	}
}

func TestService_GetPaymentByOrderID(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	order := createPendingOrder(t, db)
	svc.CreatePayment(&CreatePaymentRequest{OrderID: order.ID, Method: "alipay", Gateway: "alipay"})

	payment, err := svc.GetPaymentByOrderID(order.ID)
	if err != nil {
		t.Fatalf("GetPaymentByOrderID() error = %v", err)
	}
	if payment.Gateway != "alipay" {
		t.Errorf("gateway = %v, want alipay", payment.Gateway)
	}
}

func TestGetPaymentByOrderID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	_, err := svc.GetPaymentByOrderID(99999)
	if err == nil {
		t.Fatal("expected error for nonexistent payment")
	}
}

func TestGenerateNumbers(t *testing.T) {
	pay := generatePaymentNo()
	if !strings.HasPrefix(pay, "PAY") {
		t.Errorf("payment_no = %v, want PAY prefix", pay)
	}

	ref := generateRefundNo()
	if !strings.HasPrefix(ref, "REF") {
		t.Errorf("refund_no = %v, want REF prefix", ref)
	}

	if pay == ref {
		t.Errorf("payment_no and refund_no should differ: %v", pay)
	}
}
