package commission

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/fastax/fastax-server/internal/shared/model"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

type CreateRequest struct {
	AgentID     uint   `json:"agent_id" binding:"required"`
	CustomerID  uint   `json:"customer_id" binding:"required"`
	OrderID     uint   `json:"order_id" binding:"required"`
	OrderAmount string `json:"order_amount" binding:"required"`
	Rate        string `json:"commission_rate" binding:"required"`
}

type WithdrawRequest struct {
	Amount string `json:"amount" binding:"required"`
	Reason string `json:"reason"`
}

func (s *Service) Create(req *CreateRequest) (*model.Commission, error) {
	commission := model.Commission{
		AgentID:          req.AgentID,
		CustomerID:       req.CustomerID,
		OrderID:          req.OrderID,
		OrderAmount:      req.OrderAmount,
		CommissionRate:   req.Rate,
		CommissionAmount: calculateCommission(req.OrderAmount, req.Rate),
		Status:           "pending",
	}
	if err := s.db.Create(&commission).Error; err != nil {
		return nil, fmt.Errorf("create commission: %w", err)
	}
	return &commission, nil
}

func (s *Service) ListByAgent(agentID uint, status string) ([]model.Commission, error) {
	var commissions []model.Commission
	query := s.db.Where("agent_id = ?", agentID).Order("created_at desc")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Find(&commissions).Error; err != nil {
		return nil, err
	}
	return commissions, nil
}

func (s *Service) Settle(commissionID uint) error {
	result := s.db.Model(&model.Commission{}).
		Where("id = ? AND status = ?", commissionID, "pending").
		Update("status", "settled")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("commission not found or already settled")
	}
	return nil
}

// GetTotalByAgent returns total settled commission amount for an agent.
func (s *Service) GetTotalByAgent(agentID uint) (string, error) {
	var commissions []model.Commission
	if err := s.db.Where("agent_id = ? AND status = ?", agentID, "settled").Find(&commissions).Error; err != nil {
		return "0", err
	}
	total := 0.0
	for _, c := range commissions {
		total += parseFloat(c.CommissionAmount)
	}
	return formatAmount(total), nil
}

// GetAvailableBalance returns withdrawable balance = settled commissions - completed withdrawals.
func (s *Service) GetAvailableBalance(agentID uint) (string, error) {
	commissionTotal, err := s.GetTotalByAgent(agentID)
	if err != nil {
		return "0", err
	}

	var withdrawals []model.Withdrawal
	if err := s.db.Where("agent_id = ? AND status = ?", agentID, "completed").Find(&withdrawals).Error; err != nil {
		return "0", err
	}

	balance := parseFloat(commissionTotal)
	for _, w := range withdrawals {
		balance -= parseFloat(w.Amount)
	}
	if balance < 0 {
		balance = 0
	}
	return formatAmount(balance), nil
}

// Withdraw creates a withdrawal request after verifying sufficient balance.
func (s *Service) Withdraw(agentID uint, req *WithdrawRequest) (*model.Withdrawal, error) {
	amount := parseFloat(req.Amount)
	if amount <= 0 {
		return nil, errors.New("withdrawal amount must be positive")
	}

	balanceStr, err := s.GetAvailableBalance(agentID)
	if err != nil {
		return nil, fmt.Errorf("check balance: %w", err)
	}
	balance := parseFloat(balanceStr)
	if balance < amount {
		return nil, fmt.Errorf("insufficient balance: available %s, requested %s", balanceStr, req.Amount)
	}

	withdrawal := model.Withdrawal{
		AgentID: agentID,
		Amount:  formatAmount(amount),
		Status:  "pending",
		Reason:  req.Reason,
	}
	if err := s.db.Create(&withdrawal).Error; err != nil {
		return nil, fmt.Errorf("create withdrawal: %w", err)
	}
	return &withdrawal, nil
}

// calculateCommission computes commission = amount * rate.
func calculateCommission(amount, rate string) string {
	a := parseFloat(amount)
	r := parseFloat(rate)
	return formatAmount(a * r)
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func formatAmount(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}
