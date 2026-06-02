// 成本优化模块 HTTP 处理器
//
// 本文件实现了成本优化相关的 HTTP 接口：
//
// 预算管理：
//   - GetBudget: 获取用户预算设置
//   - SetBudget: 设置用户预算（按日/周/月）
//
// 成本告警：
//   - GetAlerts: 获取成本告警配置
//   - SetAlert: 设置成本告警
package cost

import (
	"net/http"

	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

// Handler 成本优化模块 HTTP 处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建成本优化处理器实例
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// --- 预算管理 ---

// SetBudgetRequest 预算设置请求
type SetBudgetRequest struct {
	Period string  `json:"period" binding:"required"` // daily, weekly, monthly
	Limit  float64 `json:"limit" binding:"required,gt=0"`
}

// GetBudget 获取用户预算设置
// GET /api/user/budget
func (h *Handler) GetBudget(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uint)

	status, err := h.svc.GetBudget(uid)
	if err != nil {
		response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
		return
	}
	response.Success(c, status)
}

func (h *Handler) SetBudget(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uint)

	var req SetBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	budget, err := h.svc.SetBudget(uid, req.Period, req.Limit)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, budget)
}

// --- Alert endpoints ---

type SetAlertRequest struct {
	Thresholds []float64 `json:"thresholds" binding:"required,min=1"`
}

func (h *Handler) GetAlerts(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uint)

	alert, err := h.svc.GetAlerts(uid)
	if err != nil {
		response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
		return
	}
	response.Success(c, alert)
}

func (h *Handler) SetAlert(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uint)

	var req SetAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	alert, err := h.svc.SetAlert(uid, req.Thresholds)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, alert)
}
