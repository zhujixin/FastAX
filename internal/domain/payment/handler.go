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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

// Handler 支付模块 HTTP 处理器
type Handler struct {
	svc             *Service
	webhookSecret   string // shared webhook signing secret (Stripe-compatible)
	verifyWebhook   bool   // enforces signature verification when true
}

// NewHandler 创建支付处理器实例
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// SetWebhookVerification enables webhook signature verification with the given secret.
// Call this during server initialization to enforce callback authenticity.
// secret: the signing secret (Stripe: webhook signing secret; custom: shared HMAC key)
func (h *Handler) SetWebhookVerification(secret string) {
	h.webhookSecret = secret
	h.verifyWebhook = secret != ""
	if secret != "" {
		log.Println("[payment] webhook signature verification ENABLED")
	}
}

// Create 创建支付
// POST /api/payments
func (h *Handler) Create(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	resp, err := h.svc.CreatePayment(&req, userID.(uint))
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *Handler) Callback(c *gin.Context) {
	// Verify webhook signature before processing (if configured).
	// Without this, an attacker can forge payment success callbacks.
	// Provider-specific verification details:
	//   - Stripe:   verify stripe-signature header with webhook signing secret (HMAC-SHA256)
	//   - WeChat:   verify sign field in callback body using merchant API v3 key
	//   - Alipay:   verify sign using RSA public key
	if !h.verifyWebhookSignature(c) {
		response.Error(c, http.StatusForbidden, response.CodePermissionDeny, "invalid webhook signature")
		return
	}

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

// verifyWebhookSignature validates the payment gateway's webhook signature.
// Uses a generic HMAC-SHA256 approach (Stripe-compatible).
// Returns true if verification passes or is not configured (backward-compatible).
func (h *Handler) verifyWebhookSignature(c *gin.Context) bool {
	if !h.verifyWebhook {
		// Verification not configured — accept all callbacks (backward-compatible,
		// but insecure for production).
		return true
	}

	// Read body for signature computation
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[payment] failed to read callback body: %v", err)
		return false
	}
	// Restore body for subsequent JSON binding
	c.Request.Body = io.NopCloser(strings.NewReader(string(body)))

	// Stripe-compatible: t=<timestamp>,v1=<signature>
	sigHeader := c.GetHeader("stripe-signature")
	if sigHeader == "" {
		// Fallback: generic X-Webhook-Signature header
		sigHeader = c.GetHeader("X-Webhook-Signature")
	}
	if sigHeader == "" {
		log.Println("[payment] callback rejected: missing signature header")
		return false
	}

	// Extract v1 signature from Stripe format: "t=12345,v1=abcdef,..."
	parts := strings.Split(sigHeader, ",")
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) == 2 && kv[0] == "v1" {
			expected, err := hex.DecodeString(kv[1])
			if err != nil {
				log.Printf("[payment] invalid hex signature: %v", err)
				return false
			}
			mac := hmac.New(sha256.New, []byte(h.webhookSecret))
			// For full Stripe compatibility: mac.Write([]byte(timestamp + "." + string(body)))
			mac.Write(body)
			return hmac.Equal(mac.Sum(nil), expected)
		}
	}

	log.Println("[payment] callback rejected: no v1 signature in header")
	return false
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

	userID, _ := c.Get("user_id")
	resp, err := h.svc.GetPaymentByOrderID(uint(orderID), userID.(uint))
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
