// 安全护栏模块 HTTP 处理器
//
// 本文件实现了安全护栏相关的 HTTP 接口：
//
// 护栏规则管理：
//   - ListRules: 规则列表（支持按阶段筛选）
//   - CreateRule: 创建规则
//   - SetRuleEnabled: 启用/禁用规则
//
// 检测日志：
//   - ListLogs: 检测日志列表（支持分页、筛选）
//
// 实时检测：
//   - Detect: 执行实时内容检测
package guardrail

import (
	"net/http"
	"strconv"

	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

// Handler 安全护栏模块 HTTP 处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建安全护栏处理器实例
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListRules 获取护栏规则列表
// GET /api/admin/guardrails/rules
func (h *Handler) ListRules(c *gin.Context) {
	stage := c.Query("stage")
	rules, err := h.svc.ListRules(stage)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, rules)
}

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

func (h *Handler) SetRuleEnabled(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid rule id")
		return
	}
	var req struct {
		Enabled bool `json:"enabled"`
	}
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
	response.Success(c, gin.H{"message": "rule enabled status updated"})
}

func (h *Handler) ListLogs(c *gin.Context) {
	traceID := c.Query("trace_id")
	stage := c.Query("stage")
	var userID uint
	if v := c.Query("user_id"); v != "" {
		uid, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid user_id")
			return
		}
		userID = uint(uid)
	}
	logs, err := h.svc.ListLogs(traceID, userID, stage)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, logs)
}

// UpdateRule 更新护栏规则
// PUT /api/admin/guardrails/rules/:id
func (h *Handler) UpdateRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid rule id")
		return
	}
	var req RuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	rule, err := h.svc.UpdateRule(uint(id), &req)
	if err != nil {
		if err.Error() == "rule not found" {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, rule)
}

// DeleteRule 删除护栏规则
// DELETE /api/admin/guardrails/rules/:id
func (h *Handler) DeleteRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid rule id")
		return
	}
	if err := h.svc.DeleteRule(uint(id)); err != nil {
		if err.Error() == "rule not found" {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "rule deleted"})
}

func (h *Handler) Detect(c *gin.Context) {
	var req DetectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	result, err := h.svc.Detect(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}

// UpdateConfig 更新护栏全局配置
// PUT /api/admin/guardrails/config
func (h *Handler) UpdateConfig(c *gin.Context) {
	var req struct {
		Mode    string `json:"mode" binding:"required,oneof=enforce monitor log"`
		Enabled *bool  `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	if err := h.svc.UpdateGlobalConfig(req.Mode, enabled); err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"mode": req.Mode, "enabled": enabled})
}
