// 请求体大小限制中间件
//
// 防止大负载攻击（DoS），限制请求体最大大小。
// 默认 8MB，可通过配置调整。
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// maxBodyLimitMB is the absolute maximum body size to prevent int64 overflow.
const maxBodyLimitMB = 2047 // ~2 GB, below int64 overflow threshold

// BodyLimit 限制请求体大小（单位：MB）
// 0 或负数表示不限制，最大值 2047MB 防止整数溢出
func BodyLimit(maxMB int) gin.HandlerFunc {
	// Prevent int64 overflow: maxMB * 1024 * 1024 must fit in int64
	if maxMB > maxBodyLimitMB {
		maxMB = maxBodyLimitMB
	}
	maxBytes := int64(maxMB) * 1024 * 1024
	return func(c *gin.Context) {
		if maxBytes > 0 && c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}
