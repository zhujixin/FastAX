package user

import (
	"fmt"
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

func registerUserWithPhone(t *testing.T, svc *Service, username, email, phone, password string) *LoginResponse {
	t.Helper()
	resp, err := svc.Register(&RegisterRequest{
		Username:   username,
		Password:   password,
		Email:      email,
		Phone:      phone,
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

func TestService_Logout_NoRedis(t *testing.T) {
	svc := setupTestService(t)
	resp := registerUser(t, svc, "logoutuser", "logout@test.com", "pass123")

	// With nil cache, Logout should succeed silently
	err := svc.Logout(resp.User.ID, "")
	if err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
}

func TestService_Logout_WithRefreshToken_NoRedis(t *testing.T) {
	svc := setupTestService(t)
	resp := registerUser(t, svc, "logoutrt", "logoutrt@test.com", "pass123")

	// With nil cache, even providing a refresh token should succeed
	err := svc.Logout(resp.User.ID, resp.RefreshToken)
	if err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
}

func TestService_ResetPassword_Success(t *testing.T) {
	svc := setupTestService(t)
	registerUser(t, svc, "resetpw", "resetpw@test.com", "oldpass")

	// In dev mode (nil cache), VerifyCode always returns true
	err := svc.ResetPassword(&ResetPasswordRequest{
		Email:       "resetpw@test.com",
		Code:        "123456",
		NewPassword: "newpass123",
	})
	if err != nil {
		t.Fatalf("ResetPassword() error = %v", err)
	}

	// Old password should no longer work
	_, err = svc.Login(&LoginRequest{
		Account:  "resetpw@test.com",
		Password: "oldpass",
	})
	if err == nil {
		t.Fatal("Login() with old password should fail after reset")
	}

	// New password should work
	resp, err := svc.Login(&LoginRequest{
		Account:  "resetpw@test.com",
		Password: "newpass123",
	})
	if err != nil {
		t.Fatalf("Login() with new password error = %v", err)
	}
	if resp.User.Email != "resetpw@test.com" {
		t.Errorf("email = %v, want resetpw@test.com", resp.User.Email)
	}
}

func TestService_ResetPassword_UserNotFound(t *testing.T) {
	svc := setupTestService(t)

	err := svc.ResetPassword(&ResetPasswordRequest{
		Email:       "nonexistent@test.com",
		Code:        "123456",
		NewPassword: "newpass123",
	})
	if err == nil {
		t.Fatal("ResetPassword() expected error for nonexistent user")
	}
	if err.Error() != "user not found" {
		t.Errorf("error = %v, want 'user not found'", err)
	}
}

func TestService_ResetPassword_LoginAfterReset(t *testing.T) {
	svc := setupTestService(t)
	registerUser(t, svc, "resetlogin", "resetlogin@test.com", "original")

	// Reset password
	err := svc.ResetPassword(&ResetPasswordRequest{
		Email:       "resetlogin@test.com",
		Code:        "000000",
		NewPassword: "brandnew456",
	})
	if err != nil {
		t.Fatalf("ResetPassword() error = %v", err)
	}

	// Login with new password should succeed
	resp, err := svc.Login(&LoginRequest{
		Account:  "resetlogin@test.com",
		Password: "brandnew456",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if resp.User.Username != "resetlogin" {
		t.Errorf("username = %v, want resetlogin", resp.User.Username)
	}
}

// ---------- ListUsers tests ----------

func TestService_ListUsers_All(t *testing.T) {
	svc := setupTestService(t)

	// Register multiple users with unique phones
	registerUserWithPhone(t, svc, "user1", "user1@test.com", "13800000001", "pass123")
	registerUserWithPhone(t, svc, "user2", "user2@test.com", "13800000002", "pass123")
	registerUserWithPhone(t, svc, "user3", "user3@test.com", "13800000003", "pass123")

	resp, err := svc.ListUsers(1, 20, "")
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}

	if resp.Total < 3 {
		t.Errorf("total = %v, want >= 3", resp.Total)
	}
	if len(resp.Items) < 3 {
		t.Errorf("items len = %v, want >= 3", len(resp.Items))
	}
	if resp.Page != 1 {
		t.Errorf("page = %v, want 1", resp.Page)
	}
	if resp.PageSize != 20 {
		t.Errorf("page_size = %v, want 20", resp.PageSize)
	}
}

func TestService_ListUsers_Pagination(t *testing.T) {
	svc := setupTestService(t)

	// Register 5 users with unique phones
	for i := 0; i < 5; i++ {
		phone := fmt.Sprintf("1380000%04d", i+1)
		registerUserWithPhone(t, svc, "paguser"+string(rune('a'+i)), "pag"+string(rune('a'+i))+"@test.com", phone, "pass123")
	}

	// Get page 1 with page_size=2
	resp, err := svc.ListUsers(1, 2, "")
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}

	if resp.Total < 5 {
		t.Errorf("total = %v, want >= 5", resp.Total)
	}
	if len(resp.Items) != 2 {
		t.Errorf("items len = %v, want 2", len(resp.Items))
	}
	if resp.Page != 1 {
		t.Errorf("page = %v, want 1", resp.Page)
	}
}

func TestService_ListUsers_Search(t *testing.T) {
	svc := setupTestService(t)

	registerUserWithPhone(t, svc, "searchable", "searchable@test.com", "13800001001", "pass123")
	registerUserWithPhone(t, svc, "other", "other@test.com", "13800001002", "pass123")

	resp, err := svc.ListUsers(1, 20, "searchable")
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}

	found := false
	for _, item := range resp.Items {
		if item.Username == "searchable" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find user 'searchable' in search results")
	}
}

func TestService_ListUsers_SearchByEmail(t *testing.T) {
	svc := setupTestService(t)

	registerUser(t, svc, "emailuser", "unique.email@test.com", "pass123")

	resp, err := svc.ListUsers(1, 20, "unique.email")
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}

	found := false
	for _, item := range resp.Items {
		if item.Email == "unique.email@test.com" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find user by email search")
	}
}

// ---------- SetUserStatus tests ----------

func TestService_SetUserStatus_Freeze(t *testing.T) {
	svc := setupTestService(t)
	resp := registerUser(t, svc, "freezeme", "freezeme@test.com", "pass123")

	err := svc.SetUserStatus(resp.User.ID, 0)
	if err != nil {
		t.Fatalf("SetUserStatus() error = %v", err)
	}

	// Try to login - should fail
	_, err = svc.Login(&LoginRequest{
		Account:  "freezeme@test.com",
		Password: "pass123",
	})
	if err == nil {
		t.Fatal("Login() expected error for frozen account")
	}
	if err.Error() != "account is frozen" {
		t.Errorf("error = %v, want 'account is frozen'", err)
	}
}

func TestService_SetUserStatus_Unfreeze(t *testing.T) {
	svc := setupTestService(t)
	resp := registerUser(t, svc, "unfreezeme", "unfreezeme@test.com", "pass123")

	// Freeze
	err := svc.SetUserStatus(resp.User.ID, 0)
	if err != nil {
		t.Fatalf("SetUserStatus(0) error = %v", err)
	}

	// Unfreeze
	err = svc.SetUserStatus(resp.User.ID, 1)
	if err != nil {
		t.Fatalf("SetUserStatus(1) error = %v", err)
	}

	// Login should succeed
	_, err = svc.Login(&LoginRequest{
		Account:  "unfreezeme@test.com",
		Password: "pass123",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
}

func TestService_SetUserStatus_PreventFreezeSuperAdmin(t *testing.T) {
	svc := setupTestService(t)

	// Create a super admin user directly in DB
	db := svc.db
	superAdmin := model.User{
		Username:     "superadmin",
		Email:        "superadmin@test.com",
		PasswordHash: "hash",
		Role:         "super_admin",
		Status:       1,
	}
	db.Create(&superAdmin)

	err := svc.SetUserStatus(superAdmin.ID, 0)
	if err == nil {
		t.Fatal("SetUserStatus() expected error for freezing super admin")
	}
	if err.Error() != "cannot freeze super admin account" {
		t.Errorf("error = %v, want 'cannot freeze super admin account'", err)
	}
}

func TestService_SetUserStatus_UserNotFound(t *testing.T) {
	svc := setupTestService(t)

	err := svc.SetUserStatus(99999, 0)
	if err == nil {
		t.Fatal("SetUserStatus() expected error for nonexistent user")
	}
}
