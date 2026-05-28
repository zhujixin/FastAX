package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGenerateAccessToken(t *testing.T) {
	secret := "test-secret"
	token, err := GenerateAccessToken(42, "admin", secret, time.Hour)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	if token == "" {
		t.Fatal("GenerateAccessToken() returned empty token")
	}

	claims, err := parseJWT(token, secret)
	if err != nil {
		t.Fatalf("parseJWT() error = %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("Claims.UserID = %v, want 42", claims.UserID)
	}
	if claims.Role != "admin" {
		t.Errorf("Claims.Role = %v, want admin", claims.Role)
	}
	if claims.Issuer != "fastax" {
		t.Errorf("Claims.Issuer = %v, want fastax", claims.Issuer)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	secret := "test-secret"
	token, err := GenerateRefreshToken(42, secret, time.Hour)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	if token == "" {
		t.Fatal("GenerateRefreshToken() returned empty token")
	}

	// Refresh token uses secret+"-refresh" for signing
	claims, err := parseJWT(token, secret+"-refresh")
	if err != nil {
		t.Fatalf("parseJWT() error = %v", err)
	}
	// Refresh token only encodes RegisteredClaims, not UserID
	if claims.Issuer != "fastax-refresh" {
		t.Errorf("Claims.Issuer = %v, want fastax-refresh", claims.Issuer)
	}

	// Should fail with original secret
	_, err = parseJWT(token, secret)
	if err == nil {
		t.Error("parseJWT() expected error with wrong secret, got nil")
	}
}

func TestParseJWT_ExpiredToken(t *testing.T) {
	secret := "test-secret"
	token, err := GenerateAccessToken(1, "user", secret, -time.Hour)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	_, err = parseJWT(token, secret)
	if err == nil {
		t.Fatal("parseJWT() expected error for expired token, got nil")
	}
}

func TestParseJWT_WrongSecret(t *testing.T) {
	token, _ := GenerateAccessToken(1, "user", "correct", time.Hour)
	_, err := parseJWT(token, "wrong")
	if err == nil {
		t.Fatal("parseJWT() expected error with wrong secret, got nil")
	}
}

func TestAuthRequired_MissingHeader(t *testing.T) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.GET("/", AuthRequired("secret"), func(c *gin.Context) {
		c.Status(200)
	})
	req := httptest.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %v, want 401", w.Code)
	}
}

func TestAuthRequired_InvalidToken(t *testing.T) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.GET("/", AuthRequired("secret"), func(c *gin.Context) {
		c.Status(200)
	})
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %v, want 401", w.Code)
	}
}

func TestAuthRequired_ValidToken(t *testing.T) {
	secret := "test-secret"
	token, _ := GenerateAccessToken(42, "user", secret, time.Hour)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	var gotUserID uint
	var gotRole string
	r.GET("/", AuthRequired(secret), func(c *gin.Context) {
		uid, exists := c.Get("user_id")
		if !exists {
			t.Error("user_id not set in context")
		}
		gotUserID = uid.(uint)
		if r, ok := c.Get("role"); ok {
			gotRole = r.(string)
		}
		c.Status(200)
	})
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %v, want 200", w.Code)
	}
	if gotUserID != 42 {
		t.Errorf("user_id = %v, want 42", gotUserID)
	}
	if gotRole != "user" {
		t.Errorf("role = %v, want user", gotRole)
	}
}

func setRoleMiddleware(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("role", role)
		c.Next()
	}
}

func TestAdminRequired_NonAdmin(t *testing.T) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.GET("/", setRoleMiddleware("user"), AdminRequired(), func(c *gin.Context) {
		c.Status(200)
	})
	req := httptest.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %v, want 403", w.Code)
	}
}

func TestAdminRequired_Admin(t *testing.T) {
	for _, role := range []string{"admin", "super_admin"} {
		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)

		r.GET("/", setRoleMiddleware(role), AdminRequired(), func(c *gin.Context) {
			c.Status(200)
		})
		req := httptest.NewRequest("GET", "/", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("role=%s: status = %v, want 200", role, w.Code)
		}
	}
}

func TestRoleRequired_Allowed(t *testing.T) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.GET("/", setRoleMiddleware("editor"), RoleRequired("admin", "editor"), func(c *gin.Context) {
		c.Status(200)
	})
	req := httptest.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("allowed role: status = %v, want 200", w.Code)
	}
}

func TestRoleRequired_Denied(t *testing.T) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.GET("/", setRoleMiddleware("viewer"), RoleRequired("admin", "editor"), func(c *gin.Context) {
		c.Status(200)
	})
	req := httptest.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("denied role: status = %v, want 403", w.Code)
	}
}
