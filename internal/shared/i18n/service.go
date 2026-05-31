package i18n

import (
	"errors"
	"fmt"
	"time"

	"github.com/fastax/fastax-server/internal/shared/model"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ---------- Language Management ----------

type LanguageResponse struct {
	ID             uint   `json:"id"`
	Locale         string `json:"locale"`
	Name           string `json:"name"`
	NativeName     string `json:"native_name"`
	IsEnabled      bool   `json:"is_enabled"`
	IsDefault      bool   `json:"is_default"`
	SortOrder      int    `json:"sort_order"`
	FallbackLocale string `json:"fallback_locale"`
}

type CreateLanguageRequest struct {
	Locale         string `json:"locale" binding:"required"`
	Name           string `json:"name" binding:"required"`
	NativeName     string `json:"native_name" binding:"required"`
	FallbackLocale string `json:"fallback_locale"`
	SortOrder      int    `json:"sort_order"`
}

type UpdateLanguageRequest struct {
	Name           *string `json:"name"`
	NativeName     *string `json:"native_name"`
	IsEnabled      *bool   `json:"is_enabled"`
	IsDefault      *bool   `json:"is_default"`
	SortOrder      *int    `json:"sort_order"`
	FallbackLocale *string `json:"fallback_locale"`
}

// ListLanguages returns all configured languages.
func (s *Service) ListLanguages() ([]LanguageResponse, error) {
	var languages []model.SupportedLanguage
	if err := s.db.Order("sort_order asc, locale asc").Find(&languages).Error; err != nil {
		return nil, fmt.Errorf("list languages: %w", err)
	}

	result := make([]LanguageResponse, len(languages))
	for i, lang := range languages {
		result[i] = toLanguageResponse(&lang)
	}
	return result, nil
}

// ListEnabledLanguages returns only enabled languages.
func (s *Service) ListEnabledLanguages() ([]LanguageResponse, error) {
	var languages []model.SupportedLanguage
	if err := s.db.Where("is_enabled = ?", 1).
		Order("sort_order asc, locale asc").
		Find(&languages).Error; err != nil {
		return nil, fmt.Errorf("list enabled languages: %w", err)
	}

	result := make([]LanguageResponse, len(languages))
	for i, lang := range languages {
		result[i] = toLanguageResponse(&lang)
	}
	return result, nil
}

// GetLanguage returns a language by locale.
func (s *Service) GetLanguage(locale string) (*LanguageResponse, error) {
	var lang model.SupportedLanguage
	if err := s.db.Where("locale = ?", locale).First(&lang).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("language not found: %s", locale)
		}
		return nil, fmt.Errorf("get language: %w", err)
	}
	resp := toLanguageResponse(&lang)
	return &resp, nil
}

// GetDefaultLanguage returns the default language.
func (s *Service) GetDefaultLanguage() (*LanguageResponse, error) {
	var lang model.SupportedLanguage
	if err := s.db.Where("is_default = ? AND is_enabled = ?", 1, 1).First(&lang).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Fallback to zh-CN
			return &LanguageResponse{
				Locale:     "zh-CN",
				Name:       "Chinese",
				NativeName: "中文",
				IsEnabled:  true,
				IsDefault:  true,
			}, nil
		}
		return nil, fmt.Errorf("get default language: %w", err)
	}
	resp := toLanguageResponse(&lang)
	return &resp, nil
}

// CreateLanguage adds a new language configuration.
func (s *Service) CreateLanguage(req *CreateLanguageRequest) (*LanguageResponse, error) {
	// Check if locale already exists
	var count int64
	s.db.Model(&model.SupportedLanguage{}).Where("locale = ?", req.Locale).Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("language already exists: %s", req.Locale)
	}

	fallback := req.FallbackLocale
	if fallback == "" {
		fallback = "en"
	}

	lang := model.SupportedLanguage{
		Locale:         req.Locale,
		Name:           req.Name,
		NativeName:     req.NativeName,
		IsEnabled:      1,
		SortOrder:      req.SortOrder,
		FallbackLocale: fallback,
	}

	if err := s.db.Create(&lang).Error; err != nil {
		return nil, fmt.Errorf("create language: %w", err)
	}

	resp := toLanguageResponse(&lang)
	return &resp, nil
}

// UpdateLanguage updates a language configuration.
func (s *Service) UpdateLanguage(locale string, req *UpdateLanguageRequest) error {
	var lang model.SupportedLanguage
	if err := s.db.Where("locale = ?", locale).First(&lang).Error; err != nil {
		return fmt.Errorf("language not found: %s", locale)
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.NativeName != nil {
		updates["native_name"] = *req.NativeName
	}
	if req.IsEnabled != nil {
		v := 0
		if *req.IsEnabled {
			v = 1
		}
		updates["is_enabled"] = v
	}
	if req.IsDefault != nil {
		v := 0
		if *req.IsDefault {
			v = 1
			// Unset other defaults
			s.db.Model(&model.SupportedLanguage{}).Where("locale != ? AND is_default = ?", locale, 1).Update("is_default", 0)
		}
		updates["is_default"] = v
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.FallbackLocale != nil {
		updates["fallback_locale"] = *req.FallbackLocale
	}

	if len(updates) > 0 {
		if err := s.db.Model(&lang).Updates(updates).Error; err != nil {
			return fmt.Errorf("update language: %w", err)
		}
	}

	return nil
}

// SetDefaultLanguage sets a language as the default.
func (s *Service) SetDefaultLanguage(locale string) error {
	// Verify language exists
	var lang model.SupportedLanguage
	if err := s.db.Where("locale = ?", locale).First(&lang).Error; err != nil {
		return fmt.Errorf("language not found: %s", locale)
	}

	// Unset all defaults
	if err := s.db.Model(&model.SupportedLanguage{}).Where("1 = 1").Update("is_default", 0).Error; err != nil {
		return fmt.Errorf("unset defaults: %w", err)
	}

	// Set new default
	if err := s.db.Model(&lang).Update("is_default", 1).Error; err != nil {
		return fmt.Errorf("set default: %w", err)
	}

	return nil
}

// ---------- Translation Management ----------

type TranslationEntry struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
}

type TranslationFile struct {
	Locale       string            `json:"locale"`
	Translations map[string]string `json:"translations"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// GetTranslations returns translations for a locale.
// In MVP, returns empty map; translations can be loaded from file system or DB.
func (s *Service) GetTranslations(locale string) (*TranslationFile, error) {
	// Verify locale is supported
	var lang model.SupportedLanguage
	if err := s.db.Where("locale = ? AND is_enabled = ?", locale, 1).First(&lang).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Try fallback
			var fallback model.SupportedLanguage
			if err := s.db.Where("is_default = ?", 1).First(&fallback).Error; err != nil {
				locale = "zh-CN"
			} else {
				locale = fallback.Locale
			}
		}
	}

	// In MVP, translations are loaded from static files
	// This is a placeholder for dynamic translation management
	return &TranslationFile{
		Locale:       locale,
		Translations: make(map[string]string),
		UpdatedAt:    time.Now(),
	}, nil
}

// ---------- Validation ----------

// IsValidLocale checks if a locale is enabled.
func (s *Service) IsValidLocale(locale string) bool {
	var count int64
	s.db.Model(&model.SupportedLanguage{}).
		Where("locale = ? AND is_enabled = ?", locale, 1).
		Count(&count)
	return count > 0
}

// NormalizeLocale normalizes a locale string (e.g., "zh_CN" -> "zh-CN").
func NormalizeLocale(locale string) string {
	if len(locale) < 2 {
		return "zh-CN"
	}

	// Replace underscore with hyphen
	result := make([]byte, len(locale))
	for i, c := range locale {
		if c == '_' {
			result[i] = '-'
		} else {
			result[i] = byte(c)
		}
	}
	return string(result)
}

// ---------- helpers ----------

func toLanguageResponse(lang *model.SupportedLanguage) LanguageResponse {
	return LanguageResponse{
		ID:             lang.ID,
		Locale:         lang.Locale,
		Name:           lang.Name,
		NativeName:     lang.NativeName,
		IsEnabled:      lang.IsEnabled == 1,
		IsDefault:      lang.IsDefault == 1,
		SortOrder:      lang.SortOrder,
		FallbackLocale: lang.FallbackLocale,
	}
}
