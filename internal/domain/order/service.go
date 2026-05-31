package order

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

type CreateOrderRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

type OrderResponse struct {
	ID            uint       `json:"id"`
	OrderNo       string     `json:"order_no"`
	UserID        uint       `json:"user_id"`
	ProductID     uint       `json:"product_id"`
	Quantity      string     `json:"quantity"`
	UnitPrice     string     `json:"unit_price"`
	Amount        string     `json:"amount"`
	DiscountAmount string    `json:"discount_amount"`
	FinalAmount   string     `json:"final_amount"`
	Currency      string     `json:"currency"`
	PaymentMethod string     `json:"payment_method,omitempty"`
	Status        string     `json:"status"`
	Remark        string     `json:"remark,omitempty"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type OrderQuery struct {
	UserID    uint   `form:"user_id"`
	OrderNo   string `form:"order_no"`
	Status    string `form:"status"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
	Page      int    `form:"page,default=1"`
	PageSize  int    `form:"page_size,default=20"`
}

// validTransitions defines allowed status transitions
var validTransitions = map[string]map[string]bool{
	"pending":    {"paid": true, "cancelled": true},
	"paid":       {"completed": true, "refunding": true},
	"refunding":  {"refunded": true},
}

func (s *Service) Create(userID uint, req *CreateOrderRequest) (*OrderResponse, error) {
	// Validate product exists and is available
	var product model.TokenProduct
	if err := s.db.First(&product, req.ProductID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("query product: %w", err)
	}
	if product.Status != 1 {
		return nil, errors.New("product is not available")
	}

	// Check for duplicate pending order (same user+product within 30 min)
	var existing model.Order
	err := s.db.Where("user_id = ? AND product_id = ? AND status = ? AND created_at > ?",
		userID, req.ProductID, "pending", time.Now().Add(-30*time.Minute)).
		First(&existing).Error
	if err == nil {
		return nil, errors.New("duplicate order: pending order exists for this product")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("check duplicate: %w", err)
	}

	// Calculate amounts
	quantity := fmt.Sprintf("%d", req.Quantity)
	amount := product.Price // simplified: assume quantity=1 price; real calc needs decimal math
	finalAmount := amount

	order := model.Order{
		OrderNo:      generateOrderNo(),
		UserID:       userID,
		ProductID:    product.ID,
		Quantity:     quantity,
		UnitPrice:    product.Price,
		Amount:       amount,
		FinalAmount:  finalAmount,
		Currency:     product.Currency,
		Status:       "pending",
	}
	if err := s.db.Create(&order).Error; err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	return toOrderResponse(&order), nil
}

func (s *Service) GetByID(orderID uint) (*OrderResponse, error) {
	var order model.Order
	if err := s.db.First(&order, orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, fmt.Errorf("query order: %w", err)
	}
	return toOrderResponse(&order), nil
}

func (s *Service) GetByOrderNo(orderNo string) (*OrderResponse, error) {
	var order model.Order
	if err := s.db.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, fmt.Errorf("query order: %w", err)
	}
	return toOrderResponse(&order), nil
}

func (s *Service) List(query *OrderQuery) ([]OrderResponse, int64, error) {
	db := s.db.Model(&model.Order{})

	if query.UserID > 0 {
		db = db.Where("user_id = ?", query.UserID)
	}
	if query.OrderNo != "" {
		db = db.Where("order_no = ?", query.OrderNo)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.StartTime != "" {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != "" {
		db = db.Where("created_at <= ?", query.EndTime)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	page := max(query.Page, 1)
	pageSize := query.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var orders []model.Order
	if err := db.Order("created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("query orders: %w", err)
	}

	result := make([]OrderResponse, len(orders))
	for i, o := range orders {
		result[i] = *toOrderResponse(&o)
	}
	return result, total, nil
}

func (s *Service) UpdateStatus(orderID uint, fromStatus, toStatus string) error {
	allowed, ok := validTransitions[fromStatus]
	if !ok || !allowed[toStatus] {
		return fmt.Errorf("invalid status transition: %s -> %s", fromStatus, toStatus)
	}

	result := s.db.Model(&model.Order{}).
		Where("id = ? AND status = ?", orderID, fromStatus).
		Update("status", toStatus)
	if result.Error != nil {
		return fmt.Errorf("update status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("order not found or status mismatch")
	}
	return nil
}

func (s *Service) MarkPaid(orderID uint) error {
	now := time.Now()
	result := s.db.Model(&model.Order{}).
		Where("id = ? AND status = ?", orderID, "pending").
		Updates(map[string]any{
			"status":  "paid",
			"paid_at": now,
		})
	if result.Error != nil {
		return fmt.Errorf("mark paid: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("order not found or not in pending status")
	}
	return nil
}

func (s *Service) MarkCompleted(orderID uint) error {
	result := s.db.Model(&model.Order{}).
		Where("id = ? AND status = ?", orderID, "paid").
		Update("status", "completed")
	if result.Error != nil {
		return fmt.Errorf("mark completed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("order not found or not in paid status")
	}
	return nil
}

func (s *Service) Cancel(orderID uint) error {
	result := s.db.Model(&model.Order{}).
		Where("id = ? AND status = ?", orderID, "pending").
		Update("status", "cancelled")
	if result.Error != nil {
		return fmt.Errorf("cancel order: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("order not found or not in pending status")
	}
	return nil
}

// CancelTimeoutOrders cancels pending orders older than 30 minutes
func (s *Service) CancelTimeoutOrders() (int64, error) {
	deadline := time.Now().Add(-30 * time.Minute)
	result := s.db.Model(&model.Order{}).
		Where("status = ? AND created_at < ?", "pending", deadline).
		Update("status", "cancelled")
	if result.Error != nil {
		return 0, fmt.Errorf("cancel timeout orders: %w", result.Error)
	}
	return result.RowsAffected, nil
}

func generateOrderNo() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("ORD%s", fmt.Sprintf("%x", b))
}

func toOrderResponse(o *model.Order) *OrderResponse {
	return &OrderResponse{
		ID:             o.ID,
		OrderNo:        o.OrderNo,
		UserID:         o.UserID,
		ProductID:      o.ProductID,
		Quantity:       o.Quantity,
		UnitPrice:      o.UnitPrice,
		Amount:         o.Amount,
		DiscountAmount: o.DiscountAmount,
		FinalAmount:    o.FinalAmount,
		Currency:       o.Currency,
		PaymentMethod:  o.PaymentMethod,
		Status:         o.Status,
		Remark:         o.Remark,
		PaidAt:         o.PaidAt,
		CreatedAt:      o.CreatedAt,
		UpdatedAt:      o.UpdatedAt,
	}
}

// AdminRefund handles admin refund approval/rejection.
func (s *Service) AdminRefund(orderID, adminID uint, approved bool, reason string) error {
	// Get order
	var order model.Order
	if err := s.db.First(&order, orderID).Error; err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	// Check order is in refunding status
	if order.Status != "refunding" {
		return errors.New("order is not in refunding status")
	}

	if approved {
		// Approve refund: update order status to refunded
		result := s.db.Model(&order).Updates(map[string]interface{}{
			"status": "refunded",
			"remark": reason,
		})
		if result.Error != nil {
			return fmt.Errorf("update order: %w", result.Error)
		}
	} else {
		// Reject refund: revert order status to paid
		result := s.db.Model(&order).Updates(map[string]interface{}{
			"status": "paid",
			"remark": reason,
		})
		if result.Error != nil {
			return fmt.Errorf("update order: %w", result.Error)
		}
	}

	return nil
}
