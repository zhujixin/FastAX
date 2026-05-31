package order

import (
	"net/http"
	"strconv"

	"github.com/fastax/fastax-server/internal/domain/payment"
	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc       *Service
	paymentSvc *payment.Service
}

func NewHandler(svc *Service, paymentSvc *payment.Service) *Handler {
	return &Handler{svc: svc, paymentSvc: paymentSvc}
}

func (h *Handler) Create(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	resp, err := h.svc.Create(userID.(uint), &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid order id")
		return
	}

	resp, err := h.svc.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *Handler) GetByOrderNo(c *gin.Context) {
	orderNo := c.Param("order_no")
	resp, err := h.svc.GetByOrderNo(orderNo)
	if err != nil {
		response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *Handler) List(c *gin.Context) {
	var query OrderQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	// Non-admin users can only see their own orders
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	if role != "admin" {
		query.UserID = userID.(uint)
	}

	items, total, err := h.svc.List(&query)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.SuccessPaginated(c, items, total, query.Page, query.PageSize)
}

func (h *Handler) Cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid order id")
		return
	}

	// Verify ownership
	resp, err := h.svc.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	if role != "admin" && resp.UserID != userID.(uint) {
		response.Error(c, http.StatusForbidden, response.CodePermissionDeny, "not your order")
		return
	}

	if err := h.svc.Cancel(uint(id)); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "order cancelled"})
}

type RefundRequestBody struct {
	Reason string `json:"reason"`
}

func (h *Handler) RequestRefund(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid order id")
		return
	}

	// Verify ownership
	orderResp, err := h.svc.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	if role != "admin" && orderResp.UserID != userID.(uint) {
		response.Error(c, http.StatusForbidden, response.CodePermissionDeny, "not your order")
		return
	}

	var body RefundRequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	// Delegate to payment service
	refundReq := &payment.RefundRequest{
		OrderID: uint(id),
		Amount:  orderResp.FinalAmount,
		Reason:  body.Reason,
	}

	resp, err := h.paymentSvc.CreateRefund(refundReq, userID.(uint))
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, resp)
}

// GET /api/admin/orders - Admin list all orders with filters
func (h *Handler) ListAdmin(c *gin.Context) {
	var query OrderQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	// Admin can see all orders (no user_id filter unless explicitly provided)
	items, total, err := h.svc.List(&query)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.SuccessPaginated(c, items, total, query.Page, query.PageSize)
}

// POST /api/admin/orders/:id/refund - Admin approve/reject refund
func (h *Handler) AdminRefund(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid order id")
		return
	}

	adminID, _ := c.Get("user_id")

	var req struct {
		Approved bool   `json:"approved"`
		Reason   string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	if err := h.svc.AdminRefund(uint(id), adminID.(uint), req.Approved, req.Reason); err != nil {
		if err.Error() == "order is not in refunding status" {
			response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	status := "refunded"
	if !req.Approved {
		status = "refund_rejected"
	}
	response.Success(c, gin.H{"message": "refund " + status})
}
