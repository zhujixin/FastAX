package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	Success(c, map[string]string{"key": "value"})

	if w.Code != http.StatusOK {
		t.Errorf("status = %v, want 200", w.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Code != CodeSuccess {
		t.Errorf("code = %v, want %v", resp.Code, CodeSuccess)
	}
	if resp.Data == nil {
		t.Error("data should not be nil")
	}
}

func TestSuccessPaginated(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	items := []string{"a", "b", "c"}
	SuccessPaginated(c, items, 10, 1, 3)

	var resp APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	paginated, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not paginated: %T", resp.Data)
	}
	if paginated["total"].(float64) != 10 {
		t.Errorf("total = %v, want 10", paginated["total"])
	}
	if paginated["page"].(float64) != 1 {
		t.Errorf("page = %v, want 1", paginated["page"])
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	Error(c, http.StatusBadRequest, CodeParamInvalid, "bad input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %v, want 400", w.Code)
	}

	var resp APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != CodeParamInvalid {
		t.Errorf("code = %v, want %v", resp.Code, CodeParamInvalid)
	}
	if resp.Message != "bad input" {
		t.Errorf("message = %v, want 'bad input'", resp.Message)
	}
}

func TestError_DefaultMessage(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	Error(c, http.StatusInternalServerError, CodeInternalError)

	var resp APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Message == "" {
		t.Error("should have default message for known code")
	}
}

func TestInternalError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	InternalError(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %v, want 500", w.Code)
	}

	var resp APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != CodeInternalError {
		t.Errorf("code = %v, want %v", resp.Code, CodeInternalError)
	}
}
