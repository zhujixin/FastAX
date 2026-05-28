package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDetectLanguage_ZhCN(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	DetectLanguage()(c)

	lang, exists := c.Get("lang")
	if !exists {
		t.Fatal("lang not set in context")
	}
	if lang.(string) != "zh-CN" {
		t.Errorf("lang = %v, want zh-CN", lang)
	}
}

func TestDetectLanguage_Zh(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Accept-Language", "zh")

	DetectLanguage()(c)

	lang, _ := c.Get("lang")
	if lang.(string) != "zh-CN" {
		t.Errorf("lang = %v, want zh-CN", lang)
	}
}

func TestDetectLanguage_En(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Accept-Language", "en-US,en;q=0.9")

	DetectLanguage()(c)

	lang, _ := c.Get("lang")
	if lang.(string) != "en" {
		t.Errorf("lang = %v, want en", lang)
	}
}

func TestDetectLanguage_Ja(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Accept-Language", "ja")

	DetectLanguage()(c)

	lang, _ := c.Get("lang")
	if lang.(string) != "ja" {
		t.Errorf("lang = %v, want ja", lang)
	}
}

func TestDetectLanguage_Ko(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Accept-Language", "ko")

	DetectLanguage()(c)

	lang, _ := c.Get("lang")
	if lang.(string) != "ko" {
		t.Errorf("lang = %v, want ko", lang)
	}
}

func TestDetectLanguage_Unknown_DefaultsToEn(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Accept-Language", "fr")

	DetectLanguage()(c)

	lang, _ := c.Get("lang")
	if lang.(string) != "en" {
		t.Errorf("lang = %v, want en", lang)
	}
}

func TestDetectLanguage_Empty_DefaultsToZhCN(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	DetectLanguage()(c)

	lang, _ := c.Get("lang")
	if lang.(string) != "zh-CN" {
		t.Errorf("lang = %v, want zh-CN", lang)
	}
}

func TestDetectLanguage_ZhTW(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Accept-Language", "zh-TW")

	DetectLanguage()(c)

	lang, _ := c.Get("lang")
	if lang.(string) != "zh-TW" {
		t.Errorf("lang = %v, want zh-TW", lang)
	}
}

func TestGetLanguage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Accept-Language", "en")

	DetectLanguage()(c)

	lang := GetLanguage(c)
	if lang != "en" {
		t.Errorf("GetLanguage() = %v, want en", lang)
	}
}
