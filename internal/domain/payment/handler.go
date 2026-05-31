package payment

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

func (h *Handler) Create(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	resp, err := h.svc.CreatePayment(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *Handler) Callback(c *gin.Context) {
	var cb PaymentCallback
	if err := c.ShouldBindJSON(&cb); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	if err := h.svc.HandleCallback(&cb); err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "callback processed"})
}

func (h *Handler) CreateRefund(c *gin.Context) {
	var req RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	resp, err := h.svc.CreateRefund(&req, userID.(uint))
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *Handler) ReviewRefund(c *gin.Context) {
	var review RefundReview
	if err := c.ShouldBindJSON(&review); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.svc.ReviewRefund(&review, userID.(uint)); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "refund reviewed"})
}

func (h *Handler) GetPayment(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid order id")
		return
	}

	resp, err := h.svc.GetPaymentByOrderID(uint(orderID))
	if err != nil {
		response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *Handler) ListRefunds(c *gin.Context) {
	orderIDStr := c.Query("order_id")
	var orderID uint
	if orderIDStr != "" {
		id, err := strconv.ParseUint(orderIDStr, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid order id")
			return
		}
		orderID = uint(id)
	}

	resp, err := h.svc.ListRefunds(orderID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, resp)
}
