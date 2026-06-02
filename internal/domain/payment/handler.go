// 支付模块 HTTP 处理器
//
// 本文件实现了支付相关的 HTTP 接口：
//   - Create: 创建支付（生成支付链接）
//   - Callback: 支付回调处理（微信、支付宝、Stripe）
//   - GetPayment: 获取支付信息
//   - CreateRefund: 创建退款申请
//   - ListRefunds: 获取退款列表
package payment

import (
	"net/http"
	"strconv"

	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

// Handler 支付模块 HTTP 处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建支付处理器实例
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Create 创建支付
// POST /api/payments
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
