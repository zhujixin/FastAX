// BYOK（自带 Key）模块 HTTP 处理器
//
// 本文件实现了 BYOK 相关的 HTTP 接口：
//
// 用户 Key 管理：
//   - ListKeys: 获取用户 Key 列表
//   - AddKey: 添加新 Key
//   - DeleteKey: 删除 Key
//   - SetKeyStatus: 启用/禁用 Key
package byok

import (
	"net/http"
	"strconv"

	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

// Handler BYOK 模块 HTTP 处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建 BYOK 处理器实例
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListKeys 获取用户 Key 列表
// GET /api/byok/keys
func (h *Handler) ListKeys(c *gin.Context) {
	userID, _ := c.Get("user_id")

	keys, err := h.svc.ListKeys(userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, keys)
}

// AddKey creates a new BYOK key for the authenticated user.
func (h *Handler) AddKey(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req AddKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	key, err := h.svc.AddKey(userID.(uint), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, key)
}

// DeleteKey deletes a BYOK key by ID for the authenticated user.
func (h *Handler) DeleteKey(c *gin.Context) {
	userID, _ := c.Get("user_id")

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid key id")
		return
	}

	if err := h.svc.DeleteKey(uint(id), userID.(uint)); err != nil {
		if err.Error() == "key not found" {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "key deleted"})
}

// SetKeyStatus updates the status of a BYOK key.
func (h *Handler) SetKeyStatus(c *gin.Context) {
	userID, _ := c.Get("user_id")

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid key id")
		return
	}

	var body struct {
		Status *int `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	if body.Status == nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "status is required")
		return
	}

	if err := h.svc.SetKeyStatus(uint(id), userID.(uint), *body.Status); err != nil {
		if err.Error() == "key not found" {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "status updated"})
}

// GetUsage 获取 BYOK 用量统计
// GET /api/byok/usage
func (h *Handler) GetUsage(c *gin.Context) {
	userID, _ := c.Get("user_id")

	stats, err := h.svc.GetUsageStats(userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, stats)
}

// SetPreference 设置 BYOK 路由偏好
// PUT /api/byok/preference
func (h *Handler) SetPreference(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req struct {
		Mode              string `json:"mode" binding:"required,oneof=byok_first platform_only byok_only"`
		FallbackEnabled   *bool  `json:"fallback_enabled"`
		MaxPlatformFeePct int    `json:"max_platform_fee_pct"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	pref, err := h.svc.SetPreference(userID.(uint), req.Mode, req.FallbackEnabled, req.MaxPlatformFeePct)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, pref)
}
