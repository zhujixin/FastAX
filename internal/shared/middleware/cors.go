// CORS 跨域中间件
//
// 本文件实现了跨域资源共享（CORS）中间件：
//   - CORS: 处理跨域请求，设置允许的源、方法、头部等
//   - 支持预检请求（OPTIONS）的快速响应
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件，处理跨域请求
// allowedOrigins: 允许的来源列表，为空时使用 "*"（仅开发环境推荐）
func CORS(allowedOrigins ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := "*"
		hasWhitelist := len(allowedOrigins) > 0 && allowedOrigins[0] != "" && allowedOrigins[0] != "*"
		if hasWhitelist {
			reqOrigin := c.Request.Header.Get("Origin")
			matched := false
			for _, allowed := range allowedOrigins {
				if reqOrigin == allowed {
					origin = reqOrigin
					matched = true
					break
				}
			}
			if !matched {
				// Origin不在白名单中：使用第一个允许的来源作为默认值。
				// 对于跨域请求，浏览器会检查 Origin 是否匹配，不匹配则阻止。
				// 对于同源请求（无 Origin 头），此设置不影响正常访问。
				origin = allowedOrigins[0]
			}
		}
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept-Language, X-Trace-Id")
		c.Header("Access-Control-Expose-Headers", "X-Trace-Id")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
