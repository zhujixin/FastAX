package token

import (
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

type ProductResponse struct {
	ID            uint   `json:"id"`
	SupplierID    uint   `json:"supplier_id"`
	Name          string `json:"name"`
	Model         string `json:"model"`
	Unit          string `json:"unit"`
	Price         string `json:"price"`
	OriginalPrice string `json:"original_price,omitempty"`
	Currency      string `json:"currency"`
	Description   string `json:"description,omitempty"`
	ValidityDays  *int   `json:"validity_days,omitempty"`
	Status        int    `json:"status"`
}

type UserTokenResponse struct {
	ID        uint       `json:"id"`
	ProductID uint       `json:"product_id"`
	Name      string     `json:"product_name"`
	Model     string     `json:"model"`
	Total     string     `json:"total"`
	Used      string     `json:"used"`
	Remaining string     `json:"remaining"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Status    int        `json:"status"`
}

type BuyRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

type BuyResponse struct {
	OrderNo string `json:"order_no"`
	Amount  string `json:"amount"`
	Status  string `json:"status"`
}

func (s *Service) GetProducts() ([]ProductResponse, error) {
	var products []model.TokenProduct
	if err := s.db.Where("status = ?", 1).
		Order("sort_order asc, created_at desc").
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}
	result := make([]ProductResponse, len(products))
	for i, p := range products {
		result[i] = ProductResponse{
			ID:            p.ID,
			SupplierID:    p.SupplierID,
			Name:          p.Name,
			Model:         p.Model,
			Unit:          p.Unit,
			Price:         p.Price,
			OriginalPrice: p.OriginalPrice,
			Currency:      p.Currency,
			Description:   p.Description,
			ValidityDays:  p.ValidityDays,
			Status:        p.Status,
		}
	}
	return result, nil
}

func (s *Service) GetUserTokens(userID uint) ([]UserTokenResponse, error) {
	var tokens []model.UserToken
	if err := s.db.Where("user_id = ? AND status = ?", userID, 1).
		Preload("ProductID").
		Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("query user tokens: %w", err)
	}

	result := make([]UserTokenResponse, 0, len(tokens))
	for _, t := range tokens {
		var product model.TokenProduct
		if err := s.db.First(&product, t.ProductID).Error; err != nil {
			continue
		}
		remaining := "0"
		// remaining = total - used - frozen
		result = append(result, UserTokenResponse{
			ID:        t.ID,
			ProductID: t.ProductID,
			Name:      product.Name,
			Model:     product.Model,
			Total:     t.TotalAmount,
			Used:      t.UsedAmount,
			Remaining: remaining,
			ExpiresAt: t.ExpiresAt,
			Status:    t.Status,
		})
	}
	return result, nil
}

func (s *Service) Buy(userID uint, req *BuyRequest) (*BuyResponse, error) {
	// Validate product exists
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

	// Calculate amount
	quantity := fmt.Sprintf("%d", req.Quantity)

	// Create order
	order := model.Order{
		OrderNo:    generateOrderNo(),
		UserID:     userID,
		ProductID:  product.ID,
		Quantity:   quantity,
		UnitPrice:  product.Price,
		Amount:     product.Price, // simplified: quantity * unit_price
		FinalAmount: product.Price,
		Currency:   product.Currency,
		Status:     "pending",
	}
	if err := s.db.Create(&order).Error; err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	// For S1: auto-complete the order (simplified, no real payment)
	// This will be replaced by real payment flow in S2
	now := time.Now()
	order.Status = "paid"
	order.PaidAt = &now
	s.db.Save(&order)

	// Credit user token
	userToken := model.UserToken{
		UserID:       userID,
		ProductID:    product.ID,
		OrderID:      &order.ID,
		TotalAmount:  product.Price,
		UsedAmount:   "0",
		FrozenAmount: "0",
		Status:       1,
	}
	if product.ValidityDays != nil {
		expiresAt := now.Add(time.Duration(*product.ValidityDays) * 24 * time.Hour)
		userToken.ExpiresAt = &expiresAt
	}
	if err := s.db.Create(&userToken).Error; err != nil {
		return nil, fmt.Errorf("credit user token: %w", err)
	}

	return &BuyResponse{
		OrderNo: order.OrderNo,
		Amount:  order.FinalAmount,
		Status:  "paid",
	}, nil
}

func generateOrderNo() string {
	return fmt.Sprintf("ORD%d", time.Now().UnixNano())
}
