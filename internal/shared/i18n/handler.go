package i18n

import (
	"net/http"

	"github.com/fastax/fastax-server/internal/shared/middleware"
	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GET /api/i18n/languages - List all enabled languages (public)
func (h *Handler) ListLanguages(c *gin.Context) {
	languages, err := h.svc.ListEnabledLanguages()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	defaultLang, _ := h.svc.GetDefaultLanguage()
	defaultLocale := "zh-CN"
	if defaultLang != nil {
		defaultLocale = defaultLang.Locale
	}

	response.Success(c, gin.H{
		"languages":      languages,
		"default_locale": defaultLocale,
	})
}

// GET /api/i18n/translations/:locale - Get translations for a locale (public)
func (h *Handler) GetTranslations(c *gin.Context) {
	locale := c.Param("locale")
	if locale == "" {
		locale = middleware.GetLanguage(c)
	}

	translations, err := h.svc.GetTranslations(locale)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	response.Success(c, translations)
}

// ---------- Admin APIs ----------

// GET /api/admin/i18n/languages - List all languages (admin)
func (h *Handler) ListAllLanguages(c *gin.Context) {
	languages, err := h.svc.ListLanguages()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, languages)
}

// POST /api/admin/i18n/languages - Create a new language (admin)
func (h *Handler) CreateLanguage(c *gin.Context) {
	var req CreateLanguageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	lang, err := h.svc.CreateLanguage(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, lang)
}

// PUT /api/admin/i18n/languages/:locale - Update a language (admin)
func (h *Handler) UpdateLanguage(c *gin.Context) {
	locale := c.Param("locale")

	var req UpdateLanguageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	if err := h.svc.UpdateLanguage(locale, &req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "language updated"})
}

// PUT /api/admin/i18n/default - Set default language (admin)
func (h *Handler) SetDefaultLanguage(c *gin.Context) {
	var req struct {
		Locale string `json:"locale" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	if err := h.svc.SetDefaultLanguage(req.Locale); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "default language set"})
}
