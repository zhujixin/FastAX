package i18n

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
	if err := db.AutoMigrate(&model.SupportedLanguage{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func createDefaultLanguages(t *testing.T, db *gorm.DB) {
	t.Helper()
	languages := []model.SupportedLanguage{
		{Locale: "zh-CN", Name: "Chinese", NativeName: "中文", IsEnabled: 1, IsDefault: 1, SortOrder: 1, FallbackLocale: "en"},
		{Locale: "en", Name: "English", NativeName: "English", IsEnabled: 1, IsDefault: 0, SortOrder: 2, FallbackLocale: "zh-CN"},
	}
	for _, lang := range languages {
		db.Create(&lang)
	}

	// Create disabled language using map to avoid GORM default value override
	ja := model.SupportedLanguage{
		Locale:         "ja",
		Name:           "Japanese",
		NativeName:     "日本語",
		SortOrder:      3,
		FallbackLocale: "en",
	}
	db.Create(&ja)
	// Explicitly set IsEnabled to 0 after creation
	db.Model(&ja).Update("is_enabled", 0)
}

// ---------- ListLanguages tests ----------

func TestListLanguages_All(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createDefaultLanguages(t, db)

	languages, err := svc.ListLanguages()
	if err != nil {
		t.Fatalf("ListLanguages() error = %v", err)
	}

	if len(languages) != 3 {
		t.Errorf("len = %v, want 3", len(languages))
	}

	// Verify sort order
	if languages[0].Locale != "zh-CN" {
		t.Errorf("first locale = %v, want zh-CN", languages[0].Locale)
	}
}

func TestListLanguages_Empty(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	languages, err := svc.ListLanguages()
	if err != nil {
		t.Fatalf("ListLanguages() error = %v", err)
	}

	if len(languages) != 0 {
		t.Errorf("len = %v, want 0", len(languages))
	}
}

// ---------- ListEnabledLanguages tests ----------

func TestListEnabledLanguages_Filtered(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createDefaultLanguages(t, db)

	languages, err := svc.ListEnabledLanguages()
	if err != nil {
		t.Fatalf("ListEnabledLanguages() error = %v", err)
	}

	// Only zh-CN and en are enabled
	if len(languages) != 2 {
		t.Errorf("len = %v, want 2", len(languages))
	}

	for _, lang := range languages {
		if !lang.IsEnabled {
			t.Errorf("language %s should be enabled", lang.Locale)
		}
	}
}

// ---------- GetLanguage tests ----------

func TestGetLanguage_Found(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createDefaultLanguages(t, db)

	lang, err := svc.GetLanguage("zh-CN")
	if err != nil {
		t.Fatalf("GetLanguage() error = %v", err)
	}

	if lang.Locale != "zh-CN" {
		t.Errorf("locale = %v, want zh-CN", lang.Locale)
	}
	if lang.Name != "Chinese" {
		t.Errorf("name = %v, want Chinese", lang.Name)
	}
	if !lang.IsDefault {
		t.Error("expected is_default = true")
	}
}

func TestGetLanguage_NotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	_, err := svc.GetLanguage("fr")
	if err == nil {
		t.Fatal("GetLanguage() expected error for nonexistent locale")
	}
}

// ---------- GetDefaultLanguage tests ----------

func TestGetDefaultLanguage_Exists(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createDefaultLanguages(t, db)

	lang, err := svc.GetDefaultLanguage()
	if err != nil {
		t.Fatalf("GetDefaultLanguage() error = %v", err)
	}

	if lang.Locale != "zh-CN" {
		t.Errorf("locale = %v, want zh-CN", lang.Locale)
	}
	if !lang.IsDefault {
		t.Error("expected is_default = true")
	}
}

func TestGetDefaultLanguage_NoDefault(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	// No languages configured, should return fallback
	lang, err := svc.GetDefaultLanguage()
	if err != nil {
		t.Fatalf("GetDefaultLanguage() error = %v", err)
	}

	if lang.Locale != "zh-CN" {
		t.Errorf("locale = %v, want zh-CN", lang.Locale)
	}
}

// ---------- CreateLanguage tests ----------

func TestCreateLanguage_Success(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	lang, err := svc.CreateLanguage(&CreateLanguageRequest{
		Locale:     "ko",
		Name:       "Korean",
		NativeName: "한국어",
	})
	if err != nil {
		t.Fatalf("CreateLanguage() error = %v", err)
	}

	if lang.Locale != "ko" {
		t.Errorf("locale = %v, want ko", lang.Locale)
	}
	if !lang.IsEnabled {
		t.Error("expected is_enabled = true")
	}
}

func TestCreateLanguage_Duplicate(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createDefaultLanguages(t, db)

	_, err := svc.CreateLanguage(&CreateLanguageRequest{
		Locale:     "zh-CN",
		Name:       "Chinese",
		NativeName: "中文",
	})
	if err == nil {
		t.Fatal("CreateLanguage() expected error for duplicate locale")
	}
}

// ---------- UpdateLanguage tests ----------

func TestUpdateLanguage_EnableDisable(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createDefaultLanguages(t, db)

	// Enable Japanese
	isEnabled := true
	err := svc.UpdateLanguage("ja", &UpdateLanguageRequest{
		IsEnabled: &isEnabled,
	})
	if err != nil {
		t.Fatalf("UpdateLanguage() error = %v", err)
	}

	lang, _ := svc.GetLanguage("ja")
	if !lang.IsEnabled {
		t.Error("expected ja to be enabled")
	}

	// Disable Japanese
	isEnabled = false
	err = svc.UpdateLanguage("ja", &UpdateLanguageRequest{
		IsEnabled: &isEnabled,
	})
	if err != nil {
		t.Fatalf("UpdateLanguage() error = %v", err)
	}

	lang, _ = svc.GetLanguage("ja")
	if lang.IsEnabled {
		t.Error("expected ja to be disabled")
	}
}

func TestUpdateLanguage_SetDefault(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createDefaultLanguages(t, db)

	// Set English as default
	isDefault := true
	err := svc.UpdateLanguage("en", &UpdateLanguageRequest{
		IsDefault: &isDefault,
	})
	if err != nil {
		t.Fatalf("UpdateLanguage() error = %v", err)
	}

	// Verify zh-CN is no longer default
	zhLang, _ := svc.GetLanguage("zh-CN")
	if zhLang.IsDefault {
		t.Error("expected zh-CN to not be default")
	}

	// Verify en is now default
	enLang, _ := svc.GetLanguage("en")
	if !enLang.IsDefault {
		t.Error("expected en to be default")
	}
}

func TestUpdateLanguage_NotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	err := svc.UpdateLanguage("xx", &UpdateLanguageRequest{})
	if err == nil {
		t.Fatal("UpdateLanguage() expected error for nonexistent locale")
	}
}

// ---------- SetDefaultLanguage tests ----------

func TestSetDefaultLanguage_Success(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createDefaultLanguages(t, db)

	err := svc.SetDefaultLanguage("en")
	if err != nil {
		t.Fatalf("SetDefaultLanguage() error = %v", err)
	}

	// Verify zh-CN is no longer default
	zhLang, _ := svc.GetLanguage("zh-CN")
	if zhLang.IsDefault {
		t.Error("expected zh-CN to not be default")
	}

	// Verify en is now default
	enLang, _ := svc.GetLanguage("en")
	if !enLang.IsDefault {
		t.Error("expected en to be default")
	}
}

func TestSetDefaultLanguage_NotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	err := svc.SetDefaultLanguage("xx")
	if err == nil {
		t.Fatal("SetDefaultLanguage() expected error for nonexistent locale")
	}
}

// ---------- IsValidLocale tests ----------

func TestIsValidLocale_Valid(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createDefaultLanguages(t, db)

	if !svc.IsValidLocale("zh-CN") {
		t.Error("expected zh-CN to be valid")
	}
	if !svc.IsValidLocale("en") {
		t.Error("expected en to be valid")
	}
}

func TestIsValidLocale_Disabled(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createDefaultLanguages(t, db)

	// Debug: check ja language in DB
	var ja model.SupportedLanguage
	db.Where("locale = ?", "ja").First(&ja)
	t.Logf("ja IsEnabled = %d", ja.IsEnabled)

	if svc.IsValidLocale("ja") {
		t.Error("expected ja to be invalid (disabled)")
	}
}

func TestIsValidLocale_NotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	if svc.IsValidLocale("xx") {
		t.Error("expected xx to be invalid")
	}
}

// ---------- NormalizeLocale tests ----------

func TestNormalizeLocale_Underscore(t *testing.T) {
	if v := NormalizeLocale("zh_CN"); v != "zh-CN" {
		t.Errorf("NormalizeLocale(zh_CN) = %v, want zh-CN", v)
	}
}

func TestNormalizeLocale_Hyphen(t *testing.T) {
	if v := NormalizeLocale("zh-CN"); v != "zh-CN" {
		t.Errorf("NormalizeLocale(zh-CN) = %v, want zh-CN", v)
	}
}

func TestNormalizeLocale_Short(t *testing.T) {
	if v := NormalizeLocale("z"); v != "zh-CN" {
		t.Errorf("NormalizeLocale(z) = %v, want zh-CN", v)
	}
}

// ---------- GetTranslations tests ----------

func TestGetTranslations_SupportedLocale(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createDefaultLanguages(t, db)

	translations, err := svc.GetTranslations("zh-CN")
	if err != nil {
		t.Fatalf("GetTranslations() error = %v", err)
	}

	if translations.Locale != "zh-CN" {
		t.Errorf("locale = %v, want zh-CN", translations.Locale)
	}
	if translations.Translations == nil {
		t.Error("translations should not be nil")
	}
}

func TestGetTranslations_UnsupportedLocale(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	createDefaultLanguages(t, db)

	// Should fallback to default
	translations, err := svc.GetTranslations("xx")
	if err != nil {
		t.Fatalf("GetTranslations() error = %v", err)
	}

	// Should fallback to zh-CN (default)
	if translations.Locale != "zh-CN" {
		t.Errorf("locale = %v, want zh-CN (fallback)", translations.Locale)
	}
}
