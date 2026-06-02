// 风控模块 HTTP 处理器
//
// 本文件实现了风控相关的 HTTP 接口：
//
// 风控规则管理：
//   - ListRules: 规则列表（支持按类别筛选）
//   - CreateRule: 创建规则
//   - SetRuleEnabled: 启用/禁用规则
//
// 风控事件管理：
//   - ListEvents: 事件列表（支持按状态、级别筛选）
//   - HandleEvent: 处理事件
//
// 黑名单管理：
//   - ListBlacklist: 黑名单列表
//   - AddBlacklist: 添加黑名单
//   - RemoveBlacklist: 移除黑名单
package risk

import (
	"net/http"
	"strconv"

	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

// Handler 风控模块 HTTP 处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建风控处理器实例
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListRules 获取风控规则列表
// GET /api/admin/risk/rules
func (h *Handler) ListRules(c *gin.Context) {
	category := c.Query("category")
	rules, err := h.svc.ListRules(category)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, rules)
}

// CreateRule creates a new risk rule.
func (h *Handler) CreateRule(c *gin.Context) {
	var req RuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	rule, err := h.svc.CreateRule(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, rule)
}

type setEnabledRequest struct {
	Enabled bool `json:"enabled"`
}

// SetRuleEnabled enables or disables a risk rule by ID.
func (h *Handler) SetRuleEnabled(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid rule id")
		return
	}

	var req setEnabledRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	if err := h.svc.SetRuleEnabled(uint(id), req.Enabled); err != nil {
		if err.Error() == "rule not found" {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "rule updated"})
}

// ListEvents returns risk events with pagination and filtering.
func (h *Handler) ListEvents(c *gin.Context) {
	var query EventQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	items, total, err := h.svc.ListEvents(&query)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.SuccessPaginated(c, items, total, query.Page, query.PageSize)
}

// HandleEvent marks a risk event as handled, using the current user's ID as handler_id.
func (h *Handler) HandleEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid event id")
		return
	}

	handlerID, _ := c.Get("user_id")
	if err := h.svc.HandleEvent(uint(id), handlerID.(uint)); err != nil {
		if err.Error() == "event not found or already handled" {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "event handled"})
}

// --- Blacklist ---

// GET /api/admin/risk/blacklist
func (h *Handler) ListBlacklist(c *gin.Context) {
	entries, err := h.svc.ListBlacklist()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, entries)
}

// POST /api/admin/risk/blacklist
func (h *Handler) AddBlacklist(c *gin.Context) {
	var req AddBlacklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	if err := h.svc.AddBlacklist(&req); err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "blacklist entry added"})
}

// DELETE /api/admin/risk/blacklist/:id
func (h *Handler) RemoveBlacklist(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid blacklist id")
		return
	}

	if err := h.svc.RemoveBlacklist(uint(id)); err != nil {
		if err.Error() == "blacklist entry not found" {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "blacklist entry removed"})
}
