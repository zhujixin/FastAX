package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// DetectLanguage reads Accept-Language header and sets it in context
// Falls back to zh-CN if no acceptable language found
func DetectLanguage() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("Accept-Language")
		lang = normalizeLanguage(lang)

		if lang == "" {
			lang = "zh-CN"
		}
		c.Set("lang", lang)
		c.Request.Header.Set("Accept-Language", lang)
		c.Next()
	}
}

func GetLanguage(c *gin.Context) string {
	lang, _ := c.Get("lang")
	return lang.(string)
}

// normalizeLanguage extracts the primary language tag from Accept-Language header
// and maps it to a supported locale code
func normalizeLanguage(header string) string {
	if header == "" {
		return ""
	}

	// Accept-Language can be: zh-CN,zh;q=0.9,en;q=0.8
	// Take the first tag, strip quality value
	parts := strings.Split(header, ",")
	lang := strings.TrimSpace(parts[0])
	if idx := strings.Index(lang, ";"); idx > 0 {
		lang = lang[:idx]
	}

	// Supported languages
	supported := map[string]string{
		"zh":    "zh-CN",
		"zh-CN": "zh-CN",
		"zh-TW": "zh-TW",
		"en":    "en",
		"ja":    "ja",
		"ko":    "ko",
	}

	if mapped, ok := supported[lang]; ok {
		return mapped
	}
	// Fallback chain: zh-TW→zh-CN→en
	if strings.HasPrefix(lang, "zh") {
		return "zh-CN"
	}
	return "en"
}
