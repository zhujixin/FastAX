// 佣金模块 HTTP 处理器
//
// 本文件实现了佣金相关的 HTTP 接口：
//
// 代理商佣金：
//   - ListCommissions: 佣金列表（支持按状态筛选）
//   - GetTotal: 佣金总计
//   - Withdraw: 申请提现
//
// 管理员操作：
//   - Settle: 结算佣金
package commission

import (
	"net/http"
	"strconv"

	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

// Handler 佣金模块 HTTP 处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建佣金处理器实例
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListCommissions 获取佣金列表
// GET /api/commissions
func (h *Handler) ListCommissions(c *gin.Context) {
	userID, _ := c.Get("user_id")
	status := c.Query("status")

	commissions, err := h.svc.ListByAgent(userID.(uint), status)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, commissions)
}

// GetTotal returns settled commission total and available balance for the agent.
func (h *Handler) GetTotal(c *gin.Context) {
	userID, _ := c.Get("user_id")

	total, err := h.svc.GetTotalByAgent(userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	balance, err := h.svc.GetAvailableBalance(userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"total_settled":    total,
		"available_balance": balance,
	})
}

// Settle marks a commission as settled (admin only).
func (h *Handler) Settle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid commission id")
		return
	}

	if err := h.svc.Settle(uint(id)); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "commission settled"})
}

// Withdraw creates a withdrawal request for the authenticated agent.
func (h *Handler) Withdraw(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	withdrawal, err := h.svc.Withdraw(userID.(uint), &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBalanceInsufficient, err.Error())
		return
	}
	response.Success(c, withdrawal)
}
