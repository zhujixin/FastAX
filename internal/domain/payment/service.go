package payment

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/fastax/fastax-server/internal/shared/model"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

type CreatePaymentRequest struct {
	OrderID uint   `json:"order_id" binding:"required"`
	Method  string `json:"method" binding:"required,oneof=wechat alipay stripe"`
	Gateway string `json:"gateway" binding:"required,oneof=wechat alipay stripe"`
}

type PaymentCallback struct {
	OrderID        uint   `json:"order_id"`
	PaymentNo      string `json:"payment_no"`
	GatewayTradeNo string `json:"gateway_trade_no"`
	GatewayStatus  string `json:"gateway_status"`
	Success        bool   `json:"success"`
	RawResponse    string `json:"raw_response"`
}

type RefundRequest struct {
	OrderID uint   `json:"order_id" binding:"required"`
	Amount  string `json:"amount" binding:"required"`
	Reason  string `json:"reason"`
}

type RefundReview struct {
	RefundID uint   `json:"refund_id" binding:"required"`
	Approved bool   `json:"approved"`
	Remark   string `json:"remark"`
}

type PaymentResponse struct {
	ID             uint       `json:"id"`
	OrderID        uint       `json:"order_id"`
	PaymentNo      string     `json:"payment_no"`
	Amount         string     `json:"amount"`
	Method         string     `json:"method"`
	Gateway        string     `json:"gateway"`
	GatewayTradeNo string     `json:"gateway_trade_no,omitempty"`
	Status         string     `json:"status"`
	PaidAt         *time.Time `json:"paid_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type RefundResponse struct {
	ID         uint       `json:"id"`
	OrderID    uint       `json:"order_id"`
	PaymentID  uint       `json:"payment_id"`
	RefundNo   string     `json:"refund_no"`
	Amount     string     `json:"amount"`
	Reason     string     `json:"reason,omitempty"`
	Status     string     `json:"status"`
	OperatorID *uint      `json:"operator_id,omitempty"`
	Remark     string     `json:"remark,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	HandledAt  *time.Time `json:"handled_at,omitempty"`
}

func (s *Service) CreatePayment(req *CreatePaymentRequest) (*PaymentResponse, error) {
	// Validate order exists and is pending
	var order model.Order
	if err := s.db.First(&order, req.OrderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, fmt.Errorf("query order: %w", err)
	}
	if order.Status != "pending" {
		return nil, fmt.Errorf("order is not in pending status, current: %s", order.Status)
	}

	payment := model.Payment{
		OrderID:   req.OrderID,
		PaymentNo: generatePaymentNo(),
		Amount:    order.FinalAmount,
		Method:    req.Method,
		Gateway:   req.Gateway,
		Status:    "pending",
	}
	if err := s.db.Create(&payment).Error; err != nil {
		return nil, fmt.Errorf("create payment: %w", err)
	}

	return toPaymentResponse(&payment), nil
}

func (s *Service) HandleCallback(cb *PaymentCallback) error {
	// Find payment
	var payment model.Payment
	if err := s.db.Where("order_id = ? AND status = ?", cb.OrderID, "pending").First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Idempotent: if already processed, skip
			var existing model.Payment
			if err := s.db.Where("order_id = ?", cb.OrderID).First(&existing).Error; err == nil {
				return nil
			}
			return errors.New("pending payment not found")
		}
		return fmt.Errorf("query payment: %w", err)
	}

	now := time.Now()

	// Update payment
	updates := map[string]any{
		"gateway_trade_no": cb.GatewayTradeNo,
		"gateway_status":   cb.GatewayStatus,
		"raw_response":     cb.RawResponse,
	}

	if cb.Success {
		updates["status"] = "success"
		updates["paid_at"] = now
	} else {
		updates["status"] = "failed"
	}

	if err := s.db.Model(&payment).Updates(updates).Error; err != nil {
		return fmt.Errorf("update payment: %w", err)
	}

	// Update order status if payment succeeded
	if cb.Success {
		result := s.db.Model(&model.Order{}).
			Where("id = ? AND status = ?", cb.OrderID, "pending").
			Updates(map[string]any{
				"status":  "paid",
				"paid_at": now,
			})
		if result.Error != nil {
			return fmt.Errorf("update order: %w", result.Error)
		}
	}

	return nil
}

func (s *Service) CreateRefund(req *RefundRequest, operatorID uint) (*RefundResponse, error) {
	// Validate order
	var order model.Order
	if err := s.db.First(&order, req.OrderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, fmt.Errorf("query order: %w", err)
	}
	if order.Status != "paid" {
		return nil, fmt.Errorf("only paid orders can be refunded, current status: %s", order.Status)
	}

	// Find the successful payment
	var payment model.Payment
	if err := s.db.Where("order_id = ? AND status = ?", req.OrderID, "success").First(&payment).Error; err != nil {
		return nil, errors.New("no successful payment found for this order")
	}

	// Create refund record
	refund := model.Refund{
		OrderID:   req.OrderID,
		PaymentID: payment.ID,
		RefundNo:  generateRefundNo(),
		Amount:    req.Amount,
		Reason:    req.Reason,
		Status:    "pending",
	}
	if err := s.db.Create(&refund).Error; err != nil {
		return nil, fmt.Errorf("create refund: %w", err)
	}

	// Update order status
	if err := s.db.Model(&order).Update("status", "refunding").Error; err != nil {
		return nil, fmt.Errorf("update order status: %w", err)
	}

	return toRefundResponse(&refund), nil
}

func (s *Service) ReviewRefund(review *RefundReview, operatorID uint) error {
	var refund model.Refund
	if err := s.db.First(&refund, review.RefundID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("refund not found")
		}
		return fmt.Errorf("query refund: %w", err)
	}
	if refund.Status != "pending" {
		return fmt.Errorf("refund is not in pending status, current: %s", refund.Status)
	}

	now := time.Now()

	if review.Approved {
		// Approve refund
		updates := map[string]any{
			"status":      "refunded",
			"operator_id": operatorID,
			"remark":      review.Remark,
			"handled_at":  now,
		}
		if err := s.db.Model(&refund).Updates(updates).Error; err != nil {
			return fmt.Errorf("update refund: %w", err)
		}

		// Update order status
		if err := s.db.Model(&model.Order{}).Where("id = ?", refund.OrderID).
			Update("status", "refunded").Error; err != nil {
			return fmt.Errorf("update order: %w", err)
		}
	} else {
		// Reject refund: revert order to paid
		updates := map[string]any{
			"status":      "rejected",
			"operator_id": operatorID,
			"remark":      review.Remark,
			"handled_at":  now,
		}
		if err := s.db.Model(&refund).Updates(updates).Error; err != nil {
			return fmt.Errorf("update refund: %w", err)
		}

		if err := s.db.Model(&model.Order{}).Where("id = ?", refund.OrderID).
			Update("status", "paid").Error; err != nil {
			return fmt.Errorf("update order: %w", err)
		}
	}

	return nil
}

func (s *Service) GetPaymentByOrderID(orderID uint) (*PaymentResponse, error) {
	var payment model.Payment
	if err := s.db.Where("order_id = ?", orderID).Order("created_at desc").First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment not found")
		}
		return nil, fmt.Errorf("query payment: %w", err)
	}
	return toPaymentResponse(&payment), nil
}

func (s *Service) ListRefunds(orderID uint) ([]RefundResponse, error) {
	var refunds []model.Refund
	query := s.db.Order("created_at desc")
	if orderID > 0 {
		query = query.Where("order_id = ?", orderID)
	}
	if err := query.Find(&refunds).Error; err != nil {
		return nil, fmt.Errorf("query refunds: %w", err)
	}

	result := make([]RefundResponse, len(refunds))
	for i, r := range refunds {
		result[i] = *toRefundResponse(&r)
	}
	return result, nil
}

func generatePaymentNo() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("PAY%s", fmt.Sprintf("%x", b))
}

func generateRefundNo() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("REF%s", fmt.Sprintf("%x", b))
}

func toPaymentResponse(p *model.Payment) *PaymentResponse {
	return &PaymentResponse{
		ID:             p.ID,
		OrderID:        p.OrderID,
		PaymentNo:      p.PaymentNo,
		Amount:         p.Amount,
		Method:         p.Method,
		Gateway:        p.Gateway,
		GatewayTradeNo: p.GatewayTradeNo,
		Status:         p.Status,
		PaidAt:         p.PaidAt,
		CreatedAt:      p.CreatedAt,
	}
}

func toRefundResponse(r *model.Refund) *RefundResponse {
	return &RefundResponse{
		ID:         r.ID,
		OrderID:    r.OrderID,
		PaymentID:  r.PaymentID,
		RefundNo:   r.RefundNo,
		Amount:     r.Amount,
		Reason:     r.Reason,
		Status:     r.Status,
		OperatorID: r.OperatorID,
		Remark:     r.Remark,
		CreatedAt:  r.CreatedAt,
		HandledAt:  r.HandledAt,
	}
}
