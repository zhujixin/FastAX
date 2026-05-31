package order

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fastax/fastax-server/internal/domain/payment"
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
	if err := db.AutoMigrate(
		&model.Order{}, &model.Payment{}, &model.Refund{},
		&model.TokenProduct{}, &model.User{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	orderSvc := NewService(db)
	paymentSvc := payment.NewService(db)
	handler := NewHandler(orderSvc, paymentSvc)

	gin.SetMode(gin.TestMode)
	r := gin.New()

	return db, handler, r
}

func createPaidOrder(t *testing.T, db *gorm.DB, orderSvc *Service, userID uint) *OrderResponse {
	t.Helper()
	product := model.TokenProduct{
		SupplierID: 1,
		Name:       "Test Product",
		Model:      "gpt-4",
		Unit:       "tokens",
		Price:      "100.00",
		Currency:   "CNY",
		Status:     1,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}

	order, err := orderSvc.Create(userID, &CreateOrderRequest{
		ProductID: product.ID,
		Quantity:  1,
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	// Mark as paid
	if err := orderSvc.MarkPaid(order.ID); err != nil {
		t.Fatalf("mark paid: %v", err)
	}

	// Create a successful payment record
	paymentRecord := model.Payment{
		OrderID:   order.ID,
		PaymentNo: "PAY_TEST_001",
		Amount:    order.FinalAmount,
		Method:    "wechat",
		Gateway:   "wechat",
		Status:    "success",
	}
	if err := db.Create(&paymentRecord).Error; err != nil {
		t.Fatalf("create payment: %v", err)
	}

	// Re-fetch to get updated status
	updated, _ := orderSvc.GetByID(order.ID)
	return updated
}

func TestHandler_RequestRefund_Success(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	orderSvc := NewService(db)
	paidOrder := createPaidOrder(t, db, orderSvc, 1)

	r.POST("/api/orders/:id/refund", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "user")
		handler.RequestRefund(c)
	})

	body, _ := json.Marshal(RefundRequestBody{Reason: "not satisfied"})
	req := httptest.NewRequest(http.MethodPost, "/api/orders/1/refund", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
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

	// Verify order status changed to refunding
	updated, _ := orderSvc.GetByID(paidOrder.ID)
	if updated.Status != "refunding" {
		t.Errorf("order status = %v, want refunding", updated.Status)
	}
}

func TestHandler_RequestRefund_InvalidID(t *testing.T) {
	_, handler, r := setupHandlerTest(t)

	r.POST("/api/orders/:id/refund", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "user")
		handler.RequestRefund(c)
	})

	body, _ := json.Marshal(RefundRequestBody{Reason: "test"})
	req := httptest.NewRequest(http.MethodPost, "/api/orders/abc/refund", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_RequestRefund_OrderNotFound(t *testing.T) {
	_, handler, r := setupHandlerTest(t)

	r.POST("/api/orders/:id/refund", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "user")
		handler.RequestRefund(c)
	})

	body, _ := json.Marshal(RefundRequestBody{Reason: "test"})
	req := httptest.NewRequest(http.MethodPost, "/api/orders/99999/refund", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

func TestHandler_RequestRefund_Forbidden(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	orderSvc := NewService(db)
	createPaidOrder(t, db, orderSvc, 1)

	r.POST("/api/orders/:id/refund", func(c *gin.Context) {
		c.Set("user_id", uint(2)) // Different user
		c.Set("role", "user")
		handler.RequestRefund(c)
	})

	body, _ := json.Marshal(RefundRequestBody{Reason: "test"})
	req := httptest.NewRequest(http.MethodPost, "/api/orders/1/refund", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %v, want %v", w.Code, http.StatusForbidden)
	}
}

func TestHandler_RequestRefund_AdminCanRefundAnyOrder(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	orderSvc := NewService(db)
	paidOrder := createPaidOrder(t, db, orderSvc, 1)

	r.POST("/api/orders/:id/refund", func(c *gin.Context) {
		c.Set("user_id", uint(99)) // Admin user
		c.Set("role", "admin")
		handler.RequestRefund(c)
	})

	body, _ := json.Marshal(RefundRequestBody{Reason: "admin refund"})
	req := httptest.NewRequest(http.MethodPost, "/api/orders/1/refund", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	updated, _ := orderSvc.GetByID(paidOrder.ID)
	if updated.Status != "refunding" {
		t.Errorf("order status = %v, want refunding", updated.Status)
	}
}

func TestHandler_RequestRefund_NotPaidOrder(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	orderSvc := NewService(db)
	product := model.TokenProduct{
		SupplierID: 1, Name: "P", Model: "gpt-4", Unit: "tokens",
		Price: "50.00", Currency: "CNY", Status: 1,
	}
	db.Create(&product)
	order, _ := orderSvc.Create(1, &CreateOrderRequest{ProductID: product.ID, Quantity: 1})

	r.POST("/api/orders/:id/refund", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "user")
		handler.RequestRefund(c)
	})

	body, _ := json.Marshal(RefundRequestBody{Reason: "test"})
	req := httptest.NewRequest(http.MethodPost, "/api/orders/"+uintToStr(order.ID)+"/refund", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should fail because order is pending, not paid
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestHandler_RequestRefund_EmptyBody(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	orderSvc := NewService(db)
	createPaidOrder(t, db, orderSvc, 1)

	r.POST("/api/orders/:id/refund", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "user")
		handler.RequestRefund(c)
	})

	// Empty JSON body is valid (reason is optional)
	req := httptest.NewRequest(http.MethodPost, "/api/orders/1/refund", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func uintToStr(n uint) string {
	return fmt.Sprintf("%d", n)
}
