package byok

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
	db.AutoMigrate(&model.BYOKKey{})
	return db
}

func TestService_AddKey(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	key, err := svc.AddKey(1, &AddKeyRequest{
		Provider: "openai", KeyEncrypted: "sk-xxx", KeyIV: "iv-xxx", Alias: "My Key",
	})
	if err != nil {
		t.Fatalf("AddKey() error = %v", err)
	}
	if key.Provider != "openai" {
		t.Errorf("provider = %v", key.Provider)
	}
	if key.Status != 1 {
		t.Errorf("status = %v, want 1", key.Status)
	}
}

func TestService_ListKeys(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	svc.AddKey(1, &AddKeyRequest{Provider: "openai", KeyEncrypted: "k1", KeyIV: "iv1"})
	svc.AddKey(1, &AddKeyRequest{Provider: "anthropic", KeyEncrypted: "k2", KeyIV: "iv2"})
	svc.AddKey(2, &AddKeyRequest{Provider: "openai", KeyEncrypted: "k3", KeyIV: "iv3"})

	keys, _ := svc.ListKeys(1)
	if len(keys) != 2 {
		t.Errorf("len = %v, want 2", len(keys))
	}
}

func TestService_GetKey(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	created, _ := svc.AddKey(1, &AddKeyRequest{Provider: "openai", KeyEncrypted: "k", KeyIV: "iv", Alias: "Test"})

	found, err := svc.GetKey(created.ID, 1)
	if err != nil {
		t.Fatalf("GetKey() error = %v", err)
	}
	if found.Alias != "Test" {
		t.Errorf("alias = %v", found.Alias)
	}
}

func TestService_GetKey_NotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	_, err := svc.GetKey(999, 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestService_SetKeyStatus(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	key, _ := svc.AddKey(1, &AddKeyRequest{Provider: "openai", KeyEncrypted: "k", KeyIV: "iv"})
	svc.SetKeyStatus(key.ID, 1, 0)

	keys, _ := svc.ListKeys(1)
	if keys[0].Status != 0 {
		t.Errorf("status = %v, want 0", keys[0].Status)
	}
}

func TestService_DeleteKey(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	key, _ := svc.AddKey(1, &AddKeyRequest{Provider: "openai", KeyEncrypted: "k", KeyIV: "iv"})
	err := svc.DeleteKey(key.ID, 1)
	if err != nil {
		t.Fatalf("DeleteKey() error = %v", err)
	}

	keys, _ := svc.ListKeys(1)
	if len(keys) != 0 {
		t.Errorf("len = %v, want 0", len(keys))
	}
}

func TestService_FindKeyForModel(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	svc.AddKey(1, &AddKeyRequest{
		Provider: "openai", KeyEncrypted: "k", KeyIV: "iv",
		ModelWhitelist: "gpt-4,gpt-3.5-turbo",
	})

	key, err := svc.FindKeyForModel(1, "openai", "gpt-4")
	if err != nil {
		t.Fatalf("FindKeyForModel() error = %v", err)
	}
	if key.Provider != "openai" {
		t.Errorf("provider = %v", key.Provider)
	}
}

func TestService_FindKeyForModel_NoWhitelist(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	svc.AddKey(1, &AddKeyRequest{
		Provider: "openai", KeyEncrypted: "k", KeyIV: "iv",
	})

	key, err := svc.FindKeyForModel(1, "openai", "any-model")
	if err != nil {
		t.Fatalf("FindKeyForModel() error = %v", err)
	}
	if key == nil {
		t.Fatal("expected key")
	}
}

func TestService_FindKeyForModel_NotInWhitelist(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	svc.AddKey(1, &AddKeyRequest{
		Provider: "openai", KeyEncrypted: "k", KeyIV: "iv",
		ModelWhitelist: "gpt-4",
	})

	_, err := svc.FindKeyForModel(1, "openai", "gpt-3.5-turbo")
	if err == nil {
		t.Fatal("expected error for model not in whitelist")
	}
}

func TestContainsModel(t *testing.T) {
	tests := []struct {
		whitelist, model string
		expected         bool
	}{
		{"gpt-4,gpt-3.5-turbo", "gpt-4", true},
		{"gpt-4,gpt-3.5-turbo", "gpt-3.5-turbo", true},
		{"gpt-4,gpt-3.5-turbo", "claude-3", false},
		{"", "any", false},
	}
	for _, tt := range tests {
		if got := containsModel(tt.whitelist, tt.model); got != tt.expected {
			t.Errorf("containsModel(%q, %q) = %v, want %v", tt.whitelist, tt.model, got, tt.expected)
		}
	}
}
