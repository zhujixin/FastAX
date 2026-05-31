package stats

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
		&model.User{}, &model.CallLog{},
		&model.Order{}, &model.Payment{}, &model.Refund{},
		&model.TokenProduct{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	svc := NewService(db)
	handler := NewHandler(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()

	return db, handler, r
}

func addTestUser(t *testing.T, db *gorm.DB, id uint) {
	t.Helper()
	uid := uniqueID()
	user := model.User{Username: "testuser_" + uid, Email: "user_" + uid + "@test.com", Phone: fmt.Sprintf("139%010d", idCounter), PasswordHash: "hash", Role: "user", Level: "normal", Status: 1}
	user.ID = id
	db.Create(&user)
}

func addCallLog(t *testing.T, db *gorm.DB, userID uint, tokens int, createdAt time.Time) {
	t.Helper()
	cl := model.CallLog{
		TraceID:      "tr-" + uniqueID(),
		UserID:       userID,
		RequestPath:  "/v1/chat/completions",
		TokensTotal:  tokens,
		Status:       "success",
		CreatedAt:    createdAt,
	}
	db.Create(&cl)
}

func addPaidOrder(t *testing.T, db *gorm.DB, userID uint, amount string, createdAt time.Time) {
	t.Helper()
	uid := uniqueID()
	o := model.Order{
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
	db.Create(&o)
	p := model.Payment{
		OrderID:   o.ID,
		PaymentNo: "PAY" + uid,
		Amount:    amount,
		Method:    "wechat",
		Gateway:   "wechat",
		Status:    "success",
		CreatedAt: createdAt,
	}
	db.Create(&p)
}

func parseResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return resp
}

// ---------- Handler GetUsage ----------

func TestHandler_GetUsage_Success(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	addTestUser(t, db, 1)
	addCallLog(t, db, 1, 1000, time.Now())

	r.GET("/api/stats/usage", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		handler.GetUsage(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/stats/usage?period=month", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	resp := parseResponse(t, w)
	if resp["code"].(float64) != 0 {
		t.Errorf("code = %v, want 0", resp["code"])
	}
	data := resp["data"].(map[string]interface{})
	if data["total_tokens"].(float64) != 1000 {
		t.Errorf("total_tokens = %v, want 1000", data["total_tokens"])
	}
}

func TestHandler_GetUsage_InvalidPeriod(t *testing.T) {
	_, handler, r := setupHandlerTest(t)

	r.GET("/api/stats/usage", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		handler.GetUsage(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/stats/usage?period=bad", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ---------- Handler GetConsumption ----------

func TestHandler_GetConsumption_Success(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	addTestUser(t, db, 1)
	addPaidOrder(t, db, 1, "100.00", time.Now())

	r.GET("/api/stats/consumption", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		handler.GetConsumption(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/stats/consumption?period=month", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	resp := parseResponse(t, w)
	if resp["code"].(float64) != 0 {
		t.Errorf("code = %v, want 0", resp["code"])
	}
	data := resp["data"].(map[string]interface{})
	if data["order_count"].(float64) != 1 {
		t.Errorf("order_count = %v, want 1", data["order_count"])
	}
}

func TestHandler_GetConsumption_InvalidPeriod(t *testing.T) {
	_, handler, r := setupHandlerTest(t)

	r.GET("/api/stats/consumption", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		handler.GetConsumption(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/stats/consumption?period=bad", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ---------- Handler GetBills ----------

func TestHandler_GetBills_Success(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	addTestUser(t, db, 1)
	addPaidOrder(t, db, 1, "100.00", time.Now())
	addPaidOrder(t, db, 1, "200.00", time.Now().Add(-1*time.Hour))

	r.GET("/api/stats/bills", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		handler.GetBills(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/stats/bills?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	resp := parseResponse(t, w)
	if resp["code"].(float64) != 0 {
		t.Errorf("code = %v, want 0", resp["code"])
	}
}

func TestHandler_GetBills_DefaultParams(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	addTestUser(t, db, 1)

	r.GET("/api/stats/bills", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		handler.GetBills(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/stats/bills", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

// ---------- Handler GetSummary ----------

func TestHandler_GetSummary_Success(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	addTestUser(t, db, 1)
	addCallLog(t, db, 1, 500, time.Now())
	addPaidOrder(t, db, 1, "100.00", time.Now())

	r.GET("/api/stats/summary", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		handler.GetSummary(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/stats/summary", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	resp := parseResponse(t, w)
	if resp["code"].(float64) != 0 {
		t.Errorf("code = %v, want 0", resp["code"])
	}
	data := resp["data"].(map[string]interface{})
	if data["total_tokens"].(float64) != 500 {
		t.Errorf("total_tokens = %v, want 500", data["total_tokens"])
	}
	if data["total_amount"] != "100.00" {
		t.Errorf("total_amount = %v, want 100.00", data["total_amount"])
	}
}

func TestHandler_GetSummary_EmptyData(t *testing.T) {
	db, handler, r := setupHandlerTest(t)
	addTestUser(t, db, 1)

	r.GET("/api/stats/summary", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		handler.GetSummary(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/stats/summary", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	resp := parseResponse(t, w)
	data := resp["data"].(map[string]interface{})
	if data["total_tokens"].(float64) != 0 {
		t.Errorf("total_tokens = %v, want 0", data["total_tokens"])
	}
	if data["total_amount"] != "0.00" {
		t.Errorf("total_amount = %v, want 0.00", data["total_amount"])
	}
}
