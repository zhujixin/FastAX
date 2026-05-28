package user

import (
	"strings"
	"testing"
	"time"

	"github.com/fastax/fastax-server/internal/shared/config"
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
	if err := db.AutoMigrate(&model.User{}, &model.UserProfile{}, &model.SubAccount{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func setupTestService(t *testing.T) *Service {
	t.Helper()
	db := setupTestDB(t)
	cfg := &config.JWTConfig{
		Secret:        "test-secret",
		AccessExpiry:  time.Hour,
		RefreshExpiry: 7 * 24 * time.Hour,
	}
	return NewService(db, nil, cfg)
}

func registerUser(t *testing.T, svc *Service, username, email, password string) *LoginResponse {
	t.Helper()
	// In dev mode (nil cache), VerifyCode always returns true
	resp, err := svc.Register(&RegisterRequest{
		Username:   username,
		Password:   password,
		Email:      email,
		VerifyCode: "123456",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	return resp
}

func TestService_Register_Success(t *testing.T) {
	svc := setupTestService(t)

	resp, err := svc.Register(&RegisterRequest{
		Username:   "testuser",
		Password:   "pass123",
		Email:      "test@test.com",
		VerifyCode: "123456",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if resp.AccessToken == "" {
		t.Error("access_token is empty")
	}
	if resp.RefreshToken == "" {
		t.Error("refresh_token is empty")
	}
	if resp.User.Username != "testuser" {
		t.Errorf("username = %v, want testuser", resp.User.Username)
	}
	if resp.User.Email != "test@test.com" {
		t.Errorf("email = %v, want test@test.com", resp.User.Email)
	}
	if resp.User.Role != "user" {
		t.Errorf("role = %v, want user", resp.User.Role)
	}
	if resp.User.Level != "normal" {
		t.Errorf("level = %v, want normal", resp.User.Level)
	}
	if resp.User.PreferredLanguage != "zh-CN" {
		t.Errorf("language = %v, want zh-CN", resp.User.PreferredLanguage)
	}
}

func TestService_Register_DefaultLanguage(t *testing.T) {
	svc := setupTestService(t)

	resp, err := svc.Register(&RegisterRequest{
		Username:   "testuser",
		Password:   "pass123",
		Email:      "lang@test.com",
		VerifyCode: "123456",
		Language:   "en",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if resp.User.PreferredLanguage != "en" {
		t.Errorf("language = %v, want en", resp.User.PreferredLanguage)
	}
}

func TestService_Register_DuplicateEmail(t *testing.T) {
	svc := setupTestService(t)

	registerUser(t, svc, "user1", "dup@test.com", "pass123")

	_, err := svc.Register(&RegisterRequest{
		Username:   "user2",
		Password:   "pass123",
		Email:      "dup@test.com",
		VerifyCode: "123456",
	})
	if err == nil {
		t.Fatal("Register() expected error for duplicate email")
	}
	if !strings.Contains(err.Error(), "create user") {
		t.Errorf("error = %v, want 'create user'", err)
	}
}

func TestService_Login_Success(t *testing.T) {
	svc := setupTestService(t)
	registerUser(t, svc, "loginuser", "login@test.com", "mypass123")

	resp, err := svc.Login(&LoginRequest{
		Account:  "login@test.com",
		Password: "mypass123",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if resp.User.Email != "login@test.com" {
		t.Errorf("email = %v, want login@test.com", resp.User.Email)
	}
	if resp.AccessToken == "" {
		t.Error("access_token is empty")
	}
}

func TestService_Login_ByPhone(t *testing.T) {
	svc := setupTestService(t)

	_, err := svc.Register(&RegisterRequest{
		Username:   "phoneuser",
		Password:   "pass123",
		Phone:      "13800138000",
		VerifyCode: "123456",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	resp, err := svc.Login(&LoginRequest{
		Account:  "13800138000",
		Password: "pass123",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if resp.User.Phone != "13800138000" {
		t.Errorf("phone = %v, want 13800138000", resp.User.Phone)
	}
}

func TestService_Login_WrongPassword(t *testing.T) {
	svc := setupTestService(t)
	registerUser(t, svc, "wrongpw", "wrongpw@test.com", "correct")

	_, err := svc.Login(&LoginRequest{
		Account:  "wrongpw@test.com",
		Password: "wrong",
	})
	if err == nil {
		t.Fatal("Login() expected error for wrong password")
	}
	if !strings.Contains(err.Error(), "invalid account or password") {
		t.Errorf("error = %v, want 'invalid account or password'", err)
	}
}

func TestService_Login_FrozenAccount(t *testing.T) {
	svc := setupTestService(t)
	registerUser(t, svc, "frozen", "frozen@test.com", "pass123")

	// Freeze the account
	svc.db.Model(&model.User{}).Where("email = ?", "frozen@test.com").Update("status", 0)

	_, err := svc.Login(&LoginRequest{
		Account:  "frozen@test.com",
		Password: "pass123",
	})
	if err == nil {
		t.Fatal("Login() expected error for frozen account")
	}
	if !strings.Contains(err.Error(), "frozen") {
		t.Errorf("error = %v, want 'frozen'", err)
	}
}

func TestService_Login_LockoutAfter5Failures(t *testing.T) {
	svc := setupTestService(t)
	registerUser(t, svc, "locktest", "lock@test.com", "correct")

	for i := 0; i < 5; i++ {
		svc.Login(&LoginRequest{
			Account:  "lock@test.com",
			Password: "wrong",
		})
	}

	// Check lockout
	var user model.User
	svc.db.Where("email = ?", "lock@test.com").First(&user)
	if user.LoginFailCount != 5 {
		t.Errorf("LoginFailCount = %v, want 5", user.LoginFailCount)
	}
	if user.LockedUntil == nil {
		t.Fatal("LockedUntil should be set")
	}

	// Correct password should also fail when locked
	_, err := svc.Login(&LoginRequest{
		Account:  "lock@test.com",
		Password: "correct",
	})
	if err == nil {
		t.Fatal("Login() expected error for locked account")
	}
	if !strings.Contains(err.Error(), "locked") {
		t.Errorf("error = %v, want 'locked'", err)
	}
}

func TestService_Login_ResetsFailCount(t *testing.T) {
	svc := setupTestService(t)
	registerUser(t, svc, "reset", "reset@test.com", "correct")

	// 3 failed attempts
	for i := 0; i < 3; i++ {
		svc.Login(&LoginRequest{
			Account:  "reset@test.com",
			Password: "wrong",
		})
	}

	// Successful login
	_, err := svc.Login(&LoginRequest{
		Account:  "reset@test.com",
		Password: "correct",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	var user model.User
	svc.db.Where("email = ?", "reset@test.com").First(&user)
	if user.LoginFailCount != 0 {
		t.Errorf("LoginFailCount = %v, want 0", user.LoginFailCount)
	}
	if user.LockedUntil != nil {
		t.Error("LockedUntil should be nil after successful login")
	}
}

func TestService_Login_AccountNotFound(t *testing.T) {
	svc := setupTestService(t)

	_, err := svc.Login(&LoginRequest{
		Account:  "nonexistent@test.com",
		Password: "pass",
	})
	if err == nil {
		t.Fatal("Login() expected error for nonexistent account")
	}
	if !strings.Contains(err.Error(), "invalid account or password") {
		t.Errorf("error = %v, want 'invalid account or password'", err)
	}
}

func TestService_GetUser_Success(t *testing.T) {
	svc := setupTestService(t)
	resp := registerUser(t, svc, "getuser", "getuser@test.com", "pass123")

	user, err := svc.GetUser(resp.User.ID)
	if err != nil {
		t.Fatalf("GetUser() error = %v", err)
	}
	if user.Username != "getuser" {
		t.Errorf("username = %v, want getuser", user.Username)
	}
}

func TestService_GetUser_NotFound(t *testing.T) {
	svc := setupTestService(t)

	_, err := svc.GetUser(99999)
	if err == nil {
		t.Fatal("GetUser() expected error for nonexistent user")
	}
}

func TestService_UpdateLanguage(t *testing.T) {
	svc := setupTestService(t)
	resp := registerUser(t, svc, "languser", "languser@test.com", "pass123")

	if err := svc.UpdateLanguage(resp.User.ID, "en"); err != nil {
		t.Fatalf("UpdateLanguage() error = %v", err)
	}

	user, _ := svc.GetUser(resp.User.ID)
	if user.PreferredLanguage != "en" {
		t.Errorf("language = %v, want en", user.PreferredLanguage)
	}
}

func TestService_RefreshToken_NoRedis(t *testing.T) {
	svc := setupTestService(t)

	_, err := svc.RefreshToken("some-token")
	if err == nil {
		t.Fatal("RefreshToken() expected error without Redis")
	}
	if !strings.Contains(err.Error(), "Redis") {
		t.Errorf("error = %v, want 'Redis'", err)
	}
}

func TestService_RefreshToken_InvalidToken(t *testing.T) {
	svc := setupTestService(t)
	// Even with nil cache, RefreshToken checks cache first
	_, err := svc.RefreshToken("invalid")
	if err == nil {
		t.Fatal("RefreshToken() expected error for invalid token")
	}
}
