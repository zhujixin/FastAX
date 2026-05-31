package token

import (
	"crypto/rand"
	"encoding/hex"
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

type TransferRequest struct {
	ProductID uint   `json:"product_id" binding:"required"`
	ToUserID  uint   `json:"to_user_id" binding:"required"`
	Amount    string `json:"amount" binding:"required"`
}

type TransferResponse struct {
	TransferID uint   `json:"transfer_id"`
	Status     string `json:"status"`
}

type ExtractRequest struct {
	ProductID uint   `json:"product_id" binding:"required"`
	Amount    string `json:"amount" binding:"required"`
	Address   string `json:"address" binding:"required"`
}

type ExtractResponse struct {
	TransferID uint   `json:"transfer_id"`
	Status     string `json:"status"`
}

func (s *Service) Transfer(fromUserID uint, req *TransferRequest) (*TransferResponse, error) {
	if req.ToUserID == fromUserID {
		return nil, errors.New("cannot transfer to yourself")
	}

	// Parse transfer amount
	transferAmount, err := parseFloat(req.Amount)
	if err != nil || transferAmount <= 0 {
		return nil, errors.New("invalid transfer amount")
	}

	// Find sender's active user token for the product
	var senderToken model.UserToken
	if err := s.db.Where("user_id = ? AND product_id = ? AND status = ?", fromUserID, req.ProductID, 1).
		First(&senderToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no active token found for this product")
		}
		return nil, fmt.Errorf("query sender token: %w", err)
	}

	// Check expiry
	if senderToken.ExpiresAt != nil && senderToken.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token has expired")
	}

	// Calculate remaining balance
	total, _ := parseFloat(senderToken.TotalAmount)
	used, _ := parseFloat(senderToken.UsedAmount)
	frozen, _ := parseFloat(senderToken.FrozenAmount)
	remaining := total - used - frozen

	if transferAmount > remaining {
		return nil, fmt.Errorf("insufficient balance: remaining %.2f, requested %.2f", remaining, transferAmount)
	}

	// Verify receiver exists (by checking user table via a simple query)
	var receiverExists bool
	if err := s.db.Raw("SELECT COUNT(*) > 0 FROM users WHERE id = ?", req.ToUserID).Scan(&receiverExists).Error; err != nil {
		return nil, fmt.Errorf("verify receiver: %w", err)
	}
	if !receiverExists {
		return nil, errors.New("receiver user not found")
	}

	// Update sender's token in a transaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Deduct from sender
		newUsed := fmt.Sprintf("%.2f", used+transferAmount)
		if err := tx.Model(&senderToken).Update("used_amount", newUsed).Error; err != nil {
			return fmt.Errorf("update sender token: %w", err)
		}

		// Find or create receiver's user token
		var receiverToken model.UserToken
		err := tx.Where("user_id = ? AND product_id = ? AND status = ?", req.ToUserID, req.ProductID, 1).
			First(&receiverToken).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("query receiver token: %w", err)
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new token for receiver
			receiverToken = model.UserToken{
				UserID:       req.ToUserID,
				ProductID:    req.ProductID,
				TotalAmount:  req.Amount,
				UsedAmount:   "0",
				FrozenAmount: "0",
				Status:       1,
			}
			// Copy expiry from sender if applicable
			if senderToken.ExpiresAt != nil {
				receiverToken.ExpiresAt = senderToken.ExpiresAt
			}
			if err := tx.Create(&receiverToken).Error; err != nil {
				return fmt.Errorf("create receiver token: %w", err)
			}
		} else {
			// Add to existing receiver token
			recvTotal, _ := parseFloat(receiverToken.TotalAmount)
			newTotal := fmt.Sprintf("%.2f", recvTotal+transferAmount)
			if err := tx.Model(&receiverToken).Update("total_amount", newTotal).Error; err != nil {
				return fmt.Errorf("update receiver token: %w", err)
			}
		}

		// Create transfer record
		transfer := model.TokenTransfer{
			FromUserID: fromUserID,
			ToUserID:   req.ToUserID,
			ProductID:  req.ProductID,
			Amount:     req.Amount,
			Status:     "completed",
			HandledAt:  timePtr(time.Now()),
		}
		if err := tx.Create(&transfer).Error; err != nil {
			return fmt.Errorf("create transfer record: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &TransferResponse{
		Status: "completed",
	}, nil
}

func (s *Service) Extract(userID uint, req *ExtractRequest) (*ExtractResponse, error) {
	// Parse extract amount
	extractAmount, err := parseFloat(req.Amount)
	if err != nil || extractAmount <= 0 {
		return nil, errors.New("invalid extract amount")
	}

	// Find user's active token for the product
	var userToken model.UserToken
	if err := s.db.Where("user_id = ? AND product_id = ? AND status = ?", userID, req.ProductID, 1).
		First(&userToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no active token found for this product")
		}
		return nil, fmt.Errorf("query user token: %w", err)
	}

	// Check expiry
	if userToken.ExpiresAt != nil && userToken.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token has expired")
	}

	// Calculate remaining balance
	total, _ := parseFloat(userToken.TotalAmount)
	used, _ := parseFloat(userToken.UsedAmount)
	frozen, _ := parseFloat(userToken.FrozenAmount)
	remaining := total - used - frozen

	if extractAmount > remaining {
		return nil, fmt.Errorf("insufficient balance: remaining %.2f, requested %.2f", remaining, extractAmount)
	}

	// Update token and create extract record in a transaction
	var transferID uint
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Deduct from user
		newUsed := fmt.Sprintf("%.2f", used+extractAmount)
		if err := tx.Model(&userToken).Update("used_amount", newUsed).Error; err != nil {
			return fmt.Errorf("update user token: %w", err)
		}

		// Create transfer record (ToUserID = 0 for external extract)
		transfer := model.TokenTransfer{
			FromUserID: userID,
			ToUserID:   0,
			ProductID:  req.ProductID,
			Amount:     req.Amount,
			Status:     "completed",
			HandledAt:  timePtr(time.Now()),
		}
		if err := tx.Create(&transfer).Error; err != nil {
			return fmt.Errorf("create extract record: %w", err)
		}
		transferID = transfer.ID

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &ExtractResponse{
		TransferID: transferID,
		Status:     "completed",
	}, nil
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func generateOrderNo() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("ORD%s", hex.EncodeToString(b))
}
