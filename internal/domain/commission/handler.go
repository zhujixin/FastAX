package commission

import (
	"net/http"
	"strconv"

	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListCommissions returns commissions for the authenticated agent.
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
