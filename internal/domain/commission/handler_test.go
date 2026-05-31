package commission

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fastax/fastax-server/internal/shared/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupHandlerTest(t *testing.T) (*Handler, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.AutoMigrate(&model.Commission{}, &model.Withdrawal{})
	svc := NewService(db)
	handler := NewHandler(svc)
	return handler, db
}

func setupRouter(h *Handler, userID uint, role string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	})
	return r
}

func TestHandler_ListCommissions(t *testing.T) {
	h, db := setupHandlerTest(t)

	// Seed data
	svc := NewService(db)
	svc.Create(&CreateRequest{AgentID: 1, CustomerID: 2, OrderID: 1, OrderAmount: "100", Rate: "0.1"})
	svc.Create(&CreateRequest{AgentID: 1, CustomerID: 3, OrderID: 2, OrderAmount: "200", Rate: "0.1"})
	svc.Create(&CreateRequest{AgentID: 2, CustomerID: 4, OrderID: 3, OrderAmount: "300", Rate: "0.1"})

	r := setupRouter(h, 1, "user")
	r.GET("/commissions", h.ListCommissions)

	req := httptest.NewRequest("GET", "/commissions", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	if len(data) != 2 {
		t.Errorf("data len = %d, want 2", len(data))
	}
}

func TestHandler_ListCommissions_WithStatus(t *testing.T) {
	h, db := setupHandlerTest(t)

	svc := NewService(db)
	comm, _ := svc.Create(&CreateRequest{AgentID: 1, CustomerID: 2, OrderID: 1, OrderAmount: "100", Rate: "0.1"})
	svc.Create(&CreateRequest{AgentID: 1, CustomerID: 3, OrderID: 2, OrderAmount: "200", Rate: "0.1"})
	svc.Settle(comm.ID)

	r := setupRouter(h, 1, "user")
	r.GET("/commissions", h.ListCommissions)

	req := httptest.NewRequest("GET", "/commissions?status=settled", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	if len(data) != 1 {
		t.Errorf("settled data len = %d, want 1", len(data))
	}
}

func TestHandler_GetTotal(t *testing.T) {
	h, db := setupHandlerTest(t)

	svc := NewService(db)
	c1, _ := svc.Create(&CreateRequest{AgentID: 1, CustomerID: 2, OrderID: 1, OrderAmount: "100", Rate: "0.1"})
	c2, _ := svc.Create(&CreateRequest{AgentID: 1, CustomerID: 3, OrderID: 2, OrderAmount: "200", Rate: "0.2"})
	svc.Settle(c1.ID)
	svc.Settle(c2.ID)

	r := setupRouter(h, 1, "user")
	r.GET("/commissions/total", h.GetTotal)

	req := httptest.NewRequest("GET", "/commissions/total", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["total_settled"] != "50.00" {
		t.Errorf("total_settled = %v, want 50.00", data["total_settled"])
	}
	if data["available_balance"] != "50.00" {
		t.Errorf("available_balance = %v, want 50.00", data["available_balance"])
	}
}

func TestHandler_Settle(t *testing.T) {
	h, db := setupHandlerTest(t)

	svc := NewService(db)
	comm, _ := svc.Create(&CreateRequest{AgentID: 1, CustomerID: 2, OrderID: 1, OrderAmount: "100", Rate: "0.1"})

	r := setupRouter(h, 1, "admin")
	r.POST("/commissions/:id/settle", h.Settle)

	req := httptest.NewRequest("POST", "/commissions/1/settle", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	// Verify settled
	comms, _ := svc.ListByAgent(1, "settled")
	if len(comms) != 1 {
		t.Errorf("settled len = %d, want 1", len(comms))
	}
	_ = comm
}

func TestHandler_Settle_InvalidID(t *testing.T) {
	h, _ := setupHandlerTest(t)

	r := setupRouter(h, 1, "admin")
	r.POST("/commissions/:id/settle", h.Settle)

	req := httptest.NewRequest("POST", "/commissions/abc/settle", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_Withdraw(t *testing.T) {
	h, db := setupHandlerTest(t)

	svc := NewService(db)
	c, _ := svc.Create(&CreateRequest{AgentID: 1, CustomerID: 2, OrderID: 1, OrderAmount: "100", Rate: "0.1"})
	svc.Settle(c.ID)

	r := setupRouter(h, 1, "user")
	r.POST("/commissions/withdraw", h.Withdraw)

	body, _ := json.Marshal(WithdrawRequest{Amount: "5.00", Reason: "test"})
	req := httptest.NewRequest("POST", "/commissions/withdraw", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d body=%s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestHandler_Withdraw_InsufficientBalance(t *testing.T) {
	h, db := setupHandlerTest(t)

	svc := NewService(db)
	c, _ := svc.Create(&CreateRequest{AgentID: 1, CustomerID: 2, OrderID: 1, OrderAmount: "100", Rate: "0.1"})
	svc.Settle(c.ID)

	r := setupRouter(h, 1, "user")
	r.POST("/commissions/withdraw", h.Withdraw)

	body, _ := json.Marshal(WithdrawRequest{Amount: "20.00"})
	req := httptest.NewRequest("POST", "/commissions/withdraw", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_Withdraw_InvalidBody(t *testing.T) {
	h, _ := setupHandlerTest(t)

	r := setupRouter(h, 1, "user")
	r.POST("/commissions/withdraw", h.Withdraw)

	req := httptest.NewRequest("POST", "/commissions/withdraw", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
