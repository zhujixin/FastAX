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
	ID              uint   `json:"id"`
	SupplierID      uint   `json:"supplier_id"`
	SupplierName    string `json:"supplier_name,omitempty"`
	Name            string `json:"name"`
	NameI18n        string `json:"name_i18n,omitempty"`
	Model           string `json:"model"`
	Type            string `json:"type,omitempty"`
	Unit            string `json:"unit"`
	Price           string `json:"price"`
	OriginalPrice   string `json:"original_price,omitempty"`
	Currency        string `json:"currency"`
	Description     string `json:"description,omitempty"`
	DescriptionI18n string `json:"description_i18n,omitempty"`
	ValidityDays    *int   `json:"validity_days,omitempty"`
	UsageNotes      string `json:"usage_notes,omitempty"`
	SortOrder       int    `json:"sort_order,omitempty"`
	Status          int    `json:"status"`
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
			ID:              p.ID,
			SupplierID:      p.SupplierID,
			Name:            p.Name,
			NameI18n:        p.NameI18n,
			Model:           p.Model,
			Type:            p.Type,
			Unit:            p.Unit,
			Price:           p.Price,
			OriginalPrice:   p.OriginalPrice,
			Currency:        p.Currency,
			Description:     p.Description,
			DescriptionI18n: p.DescriptionI18n,
			ValidityDays:    p.ValidityDays,
			UsageNotes:      p.UsageNotes,
			SortOrder:       p.SortOrder,
			Status:          p.Status,
		}
	}
	return result, nil
}

// GetProduct returns a single product by ID with supplier info.
func (s *Service) GetProduct(id uint) (*ProductResponse, error) {
	var product model.TokenProduct
	if err := s.db.First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("query product: %w", err)
	}

	// Get supplier name
	supplierName := ""
	var supplier model.Supplier
	if err := s.db.First(&supplier, product.SupplierID).Error; err == nil {
		supplierName = supplier.Name
	}

	return &ProductResponse{
		ID:              product.ID,
		SupplierID:      product.SupplierID,
		SupplierName:    supplierName,
		Name:            product.Name,
		NameI18n:        product.NameI18n,
		Model:           product.Model,
		Type:            product.Type,
		Unit:            product.Unit,
		Price:           product.Price,
		OriginalPrice:   product.OriginalPrice,
		Currency:        product.Currency,
		Description:     product.Description,
		DescriptionI18n: product.DescriptionI18n,
		ValidityDays:    product.ValidityDays,
		UsageNotes:      product.UsageNotes,
		SortOrder:       product.SortOrder,
		Status:          product.Status,
	}, nil
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

// ---------- Usage History ----------

type UsageRecord struct {
	ID          uint   `json:"id"`
	TraceID     string `json:"trace_id"`
	Model       string `json:"model"`
	TokensTotal int    `json:"tokens_total"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

// GetUsageHistory returns paginated call logs for the user.
func (s *Service) GetUsageHistory(userID uint, page, pageSize int) ([]UsageRecord, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := s.db.Model(&model.CallLog{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count usage: %w", err)
	}

	var logs []model.CallLog
	if err := query.Order("created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("query usage: %w", err)
	}

	records := make([]UsageRecord, len(logs))
	for i, l := range logs {
		records[i] = UsageRecord{
			ID:          l.ID,
			TraceID:     l.TraceID,
			Model:       l.RequestModel,
			TokensTotal: l.TokensTotal,
			Status:      l.Status,
			CreatedAt:   l.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return records, total, nil
}

// ---------- Admin Product Management ----------

type CreateProductRequest struct {
	SupplierID    uint   `json:"supplier_id" binding:"required"`
	Name          string `json:"name" binding:"required"`
	NameI18n      string `json:"name_i18n,omitempty"`
	Model         string `json:"model" binding:"required"`
	Type          string `json:"type" binding:"required"`
	Unit          string `json:"unit" binding:"required"`
	Price         string `json:"price" binding:"required"`
	OriginalPrice string `json:"original_price,omitempty"`
	Currency      string `json:"currency"`
	Description   string `json:"description,omitempty"`
	DescriptionI18n string `json:"description_i18n,omitempty"`
	ValidityDays  *int   `json:"validity_days,omitempty"`
	UsageNotes    string `json:"usage_notes,omitempty"`
	SortOrder     int    `json:"sort_order,omitempty"`
}

type UpdateProductRequest struct {
	Name          string `json:"name,omitempty"`
	NameI18n      string `json:"name_i18n,omitempty"`
	Price         string `json:"price,omitempty"`
	OriginalPrice string `json:"original_price,omitempty"`
	Currency      string `json:"currency,omitempty"`
	Description   string `json:"description,omitempty"`
	DescriptionI18n string `json:"description_i18n,omitempty"`
	ValidityDays  *int   `json:"validity_days,omitempty"`
	UsageNotes    string `json:"usage_notes,omitempty"`
	SortOrder     *int   `json:"sort_order,omitempty"`
	Status        *int   `json:"status,omitempty"`
}

// CreateProduct creates a new token product (admin).
func (s *Service) CreateProduct(req *CreateProductRequest) (*ProductResponse, error) {
	currency := req.Currency
	if currency == "" {
		currency = "CNY"
	}

	product := model.TokenProduct{
		SupplierID:      req.SupplierID,
		Name:            req.Name,
		NameI18n:        req.NameI18n,
		Model:           req.Model,
		Type:            req.Type,
		Unit:            req.Unit,
		Price:           req.Price,
		OriginalPrice:   req.OriginalPrice,
		Currency:        currency,
		Description:     req.Description,
		DescriptionI18n: req.DescriptionI18n,
		ValidityDays:    req.ValidityDays,
		UsageNotes:      req.UsageNotes,
		SortOrder:       req.SortOrder,
		Status:          1,
	}

	if err := s.db.Create(&product).Error; err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}

	// Get supplier name
	supplierName := ""
	var supplier model.Supplier
	if err := s.db.First(&supplier, product.SupplierID).Error; err == nil {
		supplierName = supplier.Name
	}

	return &ProductResponse{
		ID:              product.ID,
		SupplierID:      product.SupplierID,
		SupplierName:    supplierName,
		Name:            product.Name,
		NameI18n:        product.NameI18n,
		Model:           product.Model,
		Type:            product.Type,
		Unit:            product.Unit,
		Price:           product.Price,
		OriginalPrice:   product.OriginalPrice,
		Currency:        product.Currency,
		Description:     product.Description,
		DescriptionI18n: product.DescriptionI18n,
		ValidityDays:    product.ValidityDays,
		UsageNotes:      product.UsageNotes,
		SortOrder:       product.SortOrder,
		Status:          product.Status,
	}, nil
}

// UpdateProduct updates a token product (admin).
func (s *Service) UpdateProduct(id uint, req *UpdateProductRequest) error {
	var product model.TokenProduct
	if err := s.db.First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("product not found")
		}
		return fmt.Errorf("query product: %w", err)
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.NameI18n != "" {
		updates["name_i18n"] = req.NameI18n
	}
	if req.Price != "" {
		updates["price"] = req.Price
	}
	if req.OriginalPrice != "" {
		updates["original_price"] = req.OriginalPrice
	}
	if req.Currency != "" {
		updates["currency"] = req.Currency
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.DescriptionI18n != "" {
		updates["description_i18n"] = req.DescriptionI18n
	}
	if req.ValidityDays != nil {
		updates["validity_days"] = *req.ValidityDays
	}
	if req.UsageNotes != "" {
		updates["usage_notes"] = req.UsageNotes
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) == 0 {
		return errors.New("no fields to update")
	}

	if err := s.db.Model(&product).Updates(updates).Error; err != nil {
		return fmt.Errorf("update product: %w", err)
	}

	return nil
}
