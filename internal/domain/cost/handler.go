package cost

import (
	"net/http"

	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// --- Budget endpoints ---

type SetBudgetRequest struct {
	Period string  `json:"period" binding:"required"` // daily, weekly, monthly
	Limit  float64 `json:"limit" binding:"required,gt=0"`
}

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
