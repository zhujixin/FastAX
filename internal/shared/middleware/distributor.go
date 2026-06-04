// 渠道分发中间件
//
// 本中间件在请求到达 proxy handler 之前，预选渠道并注入到 Gin context 中。
// 参考 one-api middleware/distributor.go 的 Distribute 模式。
//
// 流程：
//   1. 从 JWT 获取 user_id
//   2. 查询用户所属 group（默认 "default"）
//   3. 从请求中提取 model 参数
//   4. 调用路由引擎预选渠道
//   5. 注入 channel_id 到 context
//
// 注意：本中间件执行预选，实际转发时 handler 仍可重新选择渠道（重试逻辑）。
package middleware

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// readCloser wraps a byte slice as io.ReadCloser
type readCloser struct {
	*bytes.Reader
}

func (r *readCloser) Close() error { return nil }

func newReadCloser(data []byte) io.ReadCloser {
	return &readCloser{bytes.NewReader(data)}
}

// ChannelContextKey 渠道信息在 Gin context 中的 key
const ChannelContextKey = "pre_selected_channel"

// Distributor 渠道分发中间件工厂函数
// db 用于查询用户 group 和 model-channel 映射
func Distributor(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户 ID（由 AuthRequired 中间件注入）
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		// 获取请求中的 model 参数
		model := extractModel(c)
		if model == "" {
			c.Next()
			return
		}

		// 预选渠道：查询用户在指定 model 上的首选渠道
		channelID := preselectedChannel(db, userID.(uint), model)
		if channelID > 0 {
			c.Set(ChannelContextKey, channelID)
		}

		c.Next()
	}
}

// extractModel 从请求中提取 model 字段
// 支持 JSON body 中的 "model" 字段
func extractModel(c *gin.Context) string {
	// Try query parameter first
	if m := c.Query("model"); m != "" {
		return m
	}

	// Try to parse JSON body using proper JSON decoding
	if c.Request.Body != nil && c.Request.ContentLength > 0 {
		bodyBytes, err := c.GetRawData()
		if err == nil && len(bodyBytes) > 0 {
			// Restore body for downstream handlers
			c.Request.Body = &readCloser{bytes.NewReader(bodyBytes)}
			return extractModelFromJSON(bodyBytes)
		}
	}
	return ""
}

// extractModelFromJSON uses proper JSON unmarshaling to extract the "model" field.
func extractModelFromJSON(data []byte) string {
	var body struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(data, &body); err == nil && body.Model != "" {
		return body.Model
	}
	return ""
}

// preselectedChannel 查询用户首选渠道
// MVP 实现：返回上次成功使用的渠道
func preselectedChannel(db *gorm.DB, userID uint, model string) uint {
	var result struct {
		SupplierID uint
	}
	db.Raw(`
		SELECT supplier_id FROM call_log
		WHERE user_id = ? AND request_model = ? AND status = 'success'
		ORDER BY created_at DESC LIMIT 1
	`, userID, model).Scan(&result)
	return result.SupplierID
}
