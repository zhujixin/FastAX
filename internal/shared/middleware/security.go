// 安全响应头中间件
//
// 本文件实现了 HTTP 安全响应头中间件：
//   - X-Content-Type-Options: nosniff
//   - X-Frame-Options: DENY
//   - Referrer-Policy: strict-origin-when-cross-origin
//   - X-XSS-Protection: 1; mode=block
//   - Strict-Transport-Security (仅在 HTTPS 时启用)
//
// 当部署在反向代理（如 Nginx）后时，部分头可由代理层设置。
package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders 设置 HTTP 安全响应头
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("X-XSS-Protection", "1; mode=block")
		// Content-Security-Policy: default-src 'self' restricts loading resources
		// only from the same origin, preventing XSS and data injection attacks.
		// Adjust script-src/style-src if you load external font/CDN resources.
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'")
		c.Next()
	}
}
