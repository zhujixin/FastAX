package market

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
	db.AutoMigrate(&model.Supplier{}, &model.TokenProduct{}, &model.SupplierVendor{}, &model.SupplierProduct{}, &model.ModelVariant{}, &model.ProviderHealth{})
	return db
}

func TestService_ListModels(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	db.Create(&model.Supplier{Name: "OpenAI", Code: "openai", APIBaseURL: "https://api.openai.com", APIKeyEncrypted: "k", Status: 1})
	db.Create(&model.TokenProduct{SupplierID: 1, Name: "GPT-4", Model: "gpt-4", Unit: "tokens", Price: "0.03", Currency: "USD", Status: 1})

	models, err := svc.ListModels("", "")
	if err != nil {
		t.Fatalf("ListModels() error = %v", err)
	}
	if len(models) < 1 {
		t.Errorf("len = %v, want >= 1", len(models))
	}
}

func TestService_ListModels_FilterProvider(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	db.Create(&model.Supplier{Name: "OpenAI", Code: "openai", APIBaseURL: "https://api.openai.com", APIKeyEncrypted: "k", Status: 1})
	db.Create(&model.Supplier{Name: "Anthropic", Code: "anthropic", APIBaseURL: "https://api.anthropic.com", APIKeyEncrypted: "k", Status: 1})
	db.Create(&model.TokenProduct{SupplierID: 1, Name: "GPT-4", Model: "gpt-4", Unit: "tokens", Price: "0.03", Currency: "USD", Status: 1})
	db.Create(&model.TokenProduct{SupplierID: 2, Name: "Claude", Model: "claude-3", Unit: "tokens", Price: "0.015", Currency: "USD", Status: 1})

	models, _ := svc.ListModels("openai", "")
	if len(models) != 1 {
		t.Errorf("len = %v, want 1", len(models))
	}
}

func TestService_CreateVariant(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	db.Create(&model.Supplier{Name: "OpenAI", Code: "openai", APIBaseURL: "https://api.openai.com", APIKeyEncrypted: "k", Status: 1})

	variant, err := svc.CreateVariant(&VariantRequest{
		BaseModel: "gpt-4", Suffix: "-turbo", ProviderID: 1, PriceCoefficient: 0.5,
	})
	if err != nil {
		t.Fatalf("CreateVariant() error = %v", err)
	}
	if variant.PriceCoefficient != 0.5 {
		t.Errorf("coefficient = %v, want 0.5", variant.PriceCoefficient)
	}
}

func TestService_ListVariants(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	db.Create(&model.Supplier{Name: "OpenAI", Code: "openai", APIBaseURL: "https://api.openai.com", APIKeyEncrypted: "k", Status: 1})
	svc.CreateVariant(&VariantRequest{BaseModel: "gpt-4", Suffix: "-turbo", ProviderID: 1})
	svc.CreateVariant(&VariantRequest{BaseModel: "gpt-4", Suffix: "-32k", ProviderID: 1})

	variants, _ := svc.ListVariants("gpt-4")
	if len(variants) != 2 {
		t.Errorf("len = %v, want 2", len(variants))
	}
}

func TestService_FindVariant(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	db.Create(&model.Supplier{Name: "OpenAI", Code: "openai", APIBaseURL: "https://api.openai.com", APIKeyEncrypted: "k", Status: 1})
	svc.CreateVariant(&VariantRequest{BaseModel: "gpt-4", Suffix: "-turbo", ProviderID: 1})

	v, err := svc.FindVariant("gpt-4-turbo")
	if err != nil {
		t.Fatalf("FindVariant() error = %v", err)
	}
	if v.BaseModel != "gpt-4" {
		t.Errorf("base_model = %v", v.BaseModel)
	}
}

func TestService_FindVariant_NotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	_, err := svc.FindVariant("nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestService_CompareModels(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	db.Create(&model.Supplier{Name: "OpenAI", Code: "openai", APIBaseURL: "https://api.openai.com", APIKeyEncrypted: "k", Status: 1})
	db.Create(&model.Supplier{Name: "Anthropic", Code: "anthropic", APIBaseURL: "https://api.anthropic.com", APIKeyEncrypted: "k", Status: 1})
	db.Create(&model.TokenProduct{SupplierID: 1, Name: "GPT-4", Model: "gpt-4", Unit: "tokens", Price: "0.03", Currency: "USD", Status: 1})
	db.Create(&model.TokenProduct{SupplierID: 2, Name: "Claude", Model: "claude-3", Unit: "tokens", Price: "0.015", Currency: "USD", Status: 1})

	result, err := svc.CompareModels([]string{"gpt-4", "claude-3"})
	if err != nil {
		t.Fatalf("CompareModels() error = %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("len = %v, want 2", len(result))
	}
	if result[0].Model != "gpt-4" {
		t.Errorf("model[0] = %v, want gpt-4", result[0].Model)
	}
	if len(result[0].Providers) != 1 {
		t.Errorf("providers[0] len = %v, want 1", len(result[0].Providers))
	}
	if result[0].Providers[0].ProviderName != "OpenAI" {
		t.Errorf("provider = %v, want OpenAI", result[0].Providers[0].ProviderName)
	}
}

func TestService_CompareModels_Empty(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	result, err := svc.CompareModels([]string{})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(result) != 0 {
		t.Errorf("len = %v, want 0", len(result))
	}
}

func TestService_CompareModels_WithHealth(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	db.Create(&model.Supplier{Name: "OpenAI", Code: "openai", APIBaseURL: "https://api.openai.com", APIKeyEncrypted: "k", Status: 1})
	db.Create(&model.TokenProduct{SupplierID: 1, Name: "GPT-4", Model: "gpt-4", Unit: "tokens", Price: "0.03", Currency: "USD", Status: 1})
	db.Create(&model.ProviderHealth{ProviderID: 1, Status: 1, AvgLatencyMs: 200, ErrorRate: 0.01, CheckCount: 100, PeriodStart: 1000, PeriodEnd: 2000})

	result, err := svc.CompareModels([]string{"gpt-4"})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if result[0].Providers[0].AvgLatencyMs != 200 {
		t.Errorf("latency = %v, want 200", result[0].Providers[0].AvgLatencyMs)
	}
	if result[0].Providers[0].ErrorRate != 0.01 {
		t.Errorf("error_rate = %v, want 0.01", result[0].Providers[0].ErrorRate)
	}
}

func TestService_GetProviderHealth(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	db.Create(&model.Supplier{Name: "OpenAI", Code: "openai", APIBaseURL: "https://api.openai.com", APIKeyEncrypted: "k", Status: 1})
	db.Create(&model.ProviderHealth{ProviderID: 1, Status: 1, AvgLatencyMs: 150, ErrorRate: 0.02, CheckCount: 50, PeriodStart: 1000, PeriodEnd: 2000})
	db.Create(&model.ProviderHealth{ProviderID: 1, Status: 1, AvgLatencyMs: 180, ErrorRate: 0.03, CheckCount: 60, PeriodStart: 2000, PeriodEnd: 3000})

	records, err := svc.GetProviderHealth(1)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("len = %v, want 2", len(records))
	}
	// Ordered by period_end desc, so first record should be the newer one
	if records[0].PeriodEnd != 3000 {
		t.Errorf("first period_end = %v, want 3000", records[0].PeriodEnd)
	}
	if records[0].ProviderName != "OpenAI" {
		t.Errorf("provider_name = %v, want OpenAI", records[0].ProviderName)
	}
}

func TestService_GetProviderHealth_Empty(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	records, err := svc.GetProviderHealth(999)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(records) != 0 {
		t.Errorf("len = %v, want 0", len(records))
	}
}

func TestService_ListProviders(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	db.Create(&model.Supplier{Name: "OpenAI", Code: "openai", APIBaseURL: "https://api.openai.com", APIKeyEncrypted: "k", Status: 1})
	db.Create(&model.Supplier{Name: "Anthropic", Code: "anthropic", APIBaseURL: "https://api.anthropic.com", APIKeyEncrypted: "k", Status: 1})
	db.Create(&model.ProviderHealth{ProviderID: 1, Status: 1, AvgLatencyMs: 200, ErrorRate: 0.01, CheckCount: 100, PeriodStart: 1000, PeriodEnd: 2000})

	providers, err := svc.ListProviders()
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(providers) != 2 {
		t.Fatalf("len = %v, want 2", len(providers))
	}
	// First supplier has health data
	if providers[0].ProviderName != "OpenAI" {
		t.Errorf("name = %v, want OpenAI", providers[0].ProviderName)
	}
	if providers[0].Status != 1 {
		t.Errorf("status = %v, want 1", providers[0].Status)
	}
	if providers[0].AvgLatencyMs != 200 {
		t.Errorf("latency = %v, want 200", providers[0].AvgLatencyMs)
	}
	// Second supplier has no health data
	if providers[1].ProviderName != "Anthropic" {
		t.Errorf("name = %v, want Anthropic", providers[1].ProviderName)
	}
	if providers[1].Status != 0 {
		t.Errorf("status = %v, want 0", providers[1].Status)
	}
}
