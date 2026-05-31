package enterprise

import (
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
	db.AutoMigrate(&model.User{}, &model.SubAccount{}, &model.CallLog{})
	return db
}

func TestService_CreateSubAccount(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	db.Create(&model.User{Username: "enterprise", PasswordHash: "x", Role: "enterprise", Status: 1})

	resp, err := svc.CreateSubAccount(1, &SubAccountRequest{
		Email: "sub@test.com", Password: "password123", TokenQuota: 10000,
		Permissions: []string{"api:chat", "api:embedding"},
	})
	if err != nil {
		t.Fatalf("CreateSubAccount() error = %v", err)
	}
	if resp.Email != "sub@test.com" {
		t.Errorf("email = %v", resp.Email)
	}
	if resp.TokenQuota != 10000 {
		t.Errorf("token_quota = %v, want 10000", resp.TokenQuota)
	}
}

func TestCreateSubAccount_NotEnterprise(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	db.Create(&model.User{Username: "normal", PasswordHash: "x", Role: "user", Status: 1})

	_, err := svc.CreateSubAccount(1, &SubAccountRequest{
		Email: "sub@test.com", Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error for non-enterprise user")
	}
}

func TestService_ListSubAccounts(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	db.Create(&model.User{Username: "enterprise", PasswordHash: "x", Role: "enterprise", Status: 1})

	svc.CreateSubAccount(1, &SubAccountRequest{Email: "s1@test.com", Password: "p123456"})
	svc.CreateSubAccount(1, &SubAccountRequest{Email: "s2@test.com", Password: "p123456"})

	accounts, _ := svc.ListSubAccounts(1)
	if len(accounts) != 2 {
		t.Errorf("len = %v, want 2", len(accounts))
	}
}

func TestService_SetSubAccountStatus(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	db.Create(&model.User{Username: "enterprise", PasswordHash: "x", Role: "enterprise", Status: 1})
	svc.CreateSubAccount(1, &SubAccountRequest{Email: "s@test.com", Password: "p123456"})

	err := svc.SetSubAccountStatus(1, 1, 0)
	if err != nil {
		t.Fatalf("SetSubAccountStatus() error = %v", err)
	}

	accounts, _ := svc.ListSubAccounts(1)
	if accounts[0].Status != 0 {
		t.Errorf("status = %v, want 0", accounts[0].Status)
	}
}

func TestService_UpdateQuota(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	db.Create(&model.User{Username: "enterprise", PasswordHash: "x", Role: "enterprise", Status: 1})
	svc.CreateSubAccount(1, &SubAccountRequest{Email: "s@test.com", Password: "p123456"})

	err := svc.UpdateQuota(1, 1, 50000)
	if err != nil {
		t.Fatalf("UpdateQuota() error = %v", err)
	}

	accounts, _ := svc.ListSubAccounts(1)
	if accounts[0].TokenQuota != 50000 {
		t.Errorf("token_quota = %v, want 50000", accounts[0].TokenQuota)
	}
}
