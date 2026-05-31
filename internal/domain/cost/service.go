package cost

import (
	"fmt"
	"sync"
	"time"

	"github.com/fastax/fastax-server/internal/shared/model"
	"gorm.io/gorm"
)

type Service struct {
	db       *gorm.DB
	mu       sync.RWMutex
	budgets  map[uint]*BudgetSetting   // userID -> budget
	alerts   map[uint]*AlertSetting    // userID -> alert config
	spending map[uint]float64          // userID -> current period spending
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db:       db,
		budgets:  make(map[uint]*BudgetSetting),
		alerts:   make(map[uint]*AlertSetting),
		spending: make(map[uint]float64),
	}
}

// --- Budget & Alert types ---

type BudgetSetting struct {
	UserID    uint    `json:"user_id"`
	Period    string  `json:"period"`    // "daily", "weekly", "monthly"
	Limit     float64 `json:"limit"`     // budget limit in currency units
	Spent     float64 `json:"spent"`     // current period spent
	UpdatedAt int64   `json:"updated_at"`
}

type AlertSetting struct {
	UserID     uint      `json:"user_id"`
	Thresholds []float64 `json:"thresholds"` // percentage thresholds, e.g. [50, 80, 100]
	UpdatedAt  int64     `json:"updated_at"`
}

type BudgetStatus struct {
	*BudgetSetting
	Exceeded    bool      `json:"exceeded"`
	UsagePct    float64   `json:"usage_pct"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
}

// --- Budget Management ---

func (s *Service) SetBudget(userID uint, period string, limit float64) (*BudgetSetting, error) {
	if period != "daily" && period != "weekly" && period != "monthly" {
		return nil, fmt.Errorf("invalid period: %s, must be daily/weekly/monthly", period)
	}
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be positive")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	budget := &BudgetSetting{
		UserID:    userID,
		Period:    period,
		Limit:     limit,
		Spent:     s.spending[userID],
		UpdatedAt: time.Now().Unix(),
	}
	s.budgets[userID] = budget
	return budget, nil
}

func (s *Service) GetBudget(userID uint) (*BudgetStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	budget, ok := s.budgets[userID]
	if !ok {
		return nil, fmt.Errorf("budget not set for user %d", userID)
	}

	// Refresh spent from spending tracker
	budget.Spent = s.spending[userID]

	pct := 0.0
	if budget.Limit > 0 {
		pct = (budget.Spent / budget.Limit) * 100
	}

	now := time.Now()
	start, end := s.periodRange(budget.Period, now)

	return &BudgetStatus{
		BudgetSetting: budget,
		Exceeded:      budget.Spent >= budget.Limit,
		UsagePct:      pct,
		PeriodStart:   start,
		PeriodEnd:     end,
	}, nil
}

func (s *Service) CheckBudget(userID uint) (bool, float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	budget, ok := s.budgets[userID]
	if !ok {
		// No budget set, allow by default
		return true, 0, nil
	}

	spent := s.spending[userID]
	return spent < budget.Limit, spent, nil
}

// RecordSpending adds cost for a user (called by proxy after a successful call)
func (s *Service) RecordSpending(userID uint, amount float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.spending[userID] += amount
}

// --- Alert Management ---

func (s *Service) SetAlert(userID uint, thresholds []float64) (*AlertSetting, error) {
	if len(thresholds) == 0 {
		return nil, fmt.Errorf("thresholds cannot be empty")
	}
	for _, t := range thresholds {
		if t <= 0 || t > 200 {
			return nil, fmt.Errorf("threshold %f is out of range (0, 200]", t)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	alert := &AlertSetting{
		UserID:     userID,
		Thresholds: thresholds,
		UpdatedAt:  time.Now().Unix(),
	}
	s.alerts[userID] = alert
	return alert, nil
}

func (s *Service) GetAlerts(userID uint) (*AlertSetting, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	alert, ok := s.alerts[userID]
	if !ok {
		return nil, fmt.Errorf("alert not configured for user %d", userID)
	}
	return alert, nil
}

// CheckAlerts checks current spending against thresholds, returns triggered thresholds
func (s *Service) CheckAlerts(userID uint) ([]float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	budget, hasBudget := s.budgets[userID]
	alert, hasAlert := s.alerts[userID]
	if !hasBudget || !hasAlert {
		return nil, nil
	}

	spent := s.spending[userID]
	pct := 0.0
	if budget.Limit > 0 {
		pct = (spent / budget.Limit) * 100
	}

	var triggered []float64
	for _, threshold := range alert.Thresholds {
		if pct >= threshold {
			triggered = append(triggered, threshold)
		}
	}
	return triggered, nil
}

func (s *Service) periodRange(period string, now time.Time) (time.Time, time.Time) {
	switch period {
	case "daily":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return start, start.AddDate(0, 0, 1)
	case "weekly":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start := time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
		return start, start.AddDate(0, 0, 7)
	case "monthly":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return start, start.AddDate(0, 1, 0)
	default:
		return now, now
	}
}

// --- Semantic Cache ---

type CacheRequest struct {
	PromptHash        string `json:"prompt_hash" binding:"required"`
	ResponseEncrypted string `json:"response_encrypted" binding:"required"`
	Model             string `json:"model" binding:"required"`
	TTLSeconds        int    `json:"ttl_seconds"`
}

func (s *Service) SetCache(req *CacheRequest) error {
	ttl := req.TTLSeconds
	if ttl <= 0 {
		ttl = 3600 // default 1 hour
	}
	cache := model.SemanticCache{
		PromptHash:        req.PromptHash,
		ResponseEncrypted: req.ResponseEncrypted,
		Model:             req.Model,
		CreatedAt:         time.Now().Unix(),
		ExpiresAt:         time.Now().Add(time.Duration(ttl) * time.Second).Unix(),
	}
	if err := s.db.Create(&cache).Error; err != nil {
		return fmt.Errorf("set cache: %w", err)
	}
	return nil
}

func (s *Service) GetCache(promptHash, modelName string) (*model.SemanticCache, error) {
	var cache model.SemanticCache
	err := s.db.Where("prompt_hash = ? AND model = ? AND expires_at > ?",
		promptHash, modelName, time.Now().Unix()).
		First(&cache).Error
	if err != nil {
		return nil, err
	}
	// Increment hit count
	s.db.Model(&cache).Update("hit_count", gorm.Expr("hit_count + 1"))
	return &cache, nil
}

func (s *Service) CleanExpired() (int64, error) {
	result := s.db.Where("expires_at < ?", time.Now().Unix()).Delete(&model.SemanticCache{})
	return result.RowsAffected, result.Error
}

// --- Cost Tracking ---

type CostRecord struct {
	SupplierID uint   `json:"supplier_id"`
	Model      string `json:"model"`
	Tokens     int    `json:"tokens"`
	Cost       string `json:"cost"`
	Period     string `json:"period"` // "2026-05-28" or "2026-05"
}

func (s *Service) GetCostBySupplier(period string) ([]CostRecord, error) {
	// Simplified: query from call_log joined with supplier
	var results []struct {
		SupplierID uint
		Model      string
		TotalTokens int
	}
	err := s.db.Raw(`
		SELECT supplier_id, request_model as model, SUM(tokens_total) as total_tokens
		FROM call_log
		WHERE created_at >= ? AND status = 'success'
		GROUP BY supplier_id, request_model
	`, period).Scan(&results).Error
	if err != nil {
		return nil, err
	}

	records := make([]CostRecord, len(results))
	for i, r := range results {
		records[i] = CostRecord{
			SupplierID: r.SupplierID,
			Model:      r.Model,
			Tokens:     r.TotalTokens,
		}
	}
	return records, nil
}

func (s *Service) GetCostByModel(period string) ([]CostRecord, error) {
	var results []struct {
		Model       string
		TotalTokens int
	}
	err := s.db.Raw(`
		SELECT request_model as model, SUM(tokens_total) as total_tokens
		FROM call_log
		WHERE created_at >= ? AND status = 'success'
		GROUP BY request_model
	`, period).Scan(&results).Error
	if err != nil {
		return nil, err
	}

	records := make([]CostRecord, len(results))
	for i, r := range results {
		records[i] = CostRecord{
			Model:  r.Model,
			Tokens: r.TotalTokens,
		}
	}
	return records, nil
}
