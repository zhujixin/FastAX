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
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
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
