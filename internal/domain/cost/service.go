// 成本优化模块业务逻辑层
//
// 本文件实现了成本优化相关的业务逻辑：
//
// 预算管理：
//   - GetBudget: 获取用户预算设置
//   - SetBudget: 设置用户预算（按日/周/月）
//   - CheckBudget: 检查是否超出预算
//
// 成本告警：
//   - GetAlerts: 获取用户告警配置
//   - SetAlert: 设置成本告警
//   - CheckAlert: 检查是否触发告警
//
// 消费跟踪：
//   - RecordSpending: 记录消费金额
//   - GetSpending: 获取当前周期消费
package cost

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/fastax/fastax-server/internal/shared/model"
	"gorm.io/gorm"
)

// Service 成本优化服务结构体
type Service struct {
	db       *gorm.DB
	mu       sync.RWMutex
	spending map[uint]float64 // userID -> 当前周期消费（瞬时数据，周期切换时重置）
}

// NewService 创建成本优化服务实例
func NewService(db *gorm.DB) *Service {
	return &Service{
		db:       db,
		spending: make(map[uint]float64),
	}
}

// --- 预算与告警类型 ---

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
	spent := s.spending[userID]
	s.mu.Unlock()

	// Upsert budget to database
	budget := model.UserBudget{
		UserID: userID,
		Period: period,
		Limit:  limit,
		Spent:  spent,
	}
	if err := s.db.Where("user_id = ?", userID).Assign(budget).FirstOrCreate(&budget).Error; err != nil {
		return nil, fmt.Errorf("set budget: %w", err)
	}

	return &BudgetSetting{
		UserID: userID,
		Period: period,
		Limit:  limit,
		Spent:  spent,
	}, nil
}

func (s *Service) GetBudget(userID uint) (*BudgetStatus, error) {
	var budget model.UserBudget
	if err := s.db.Where("user_id = ?", userID).First(&budget).Error; err != nil {
		return nil, fmt.Errorf("budget not set for user %d", userID)
	}

	// Refresh spent from spending tracker
	s.mu.RLock()
	spent := s.spending[userID]
	s.mu.RUnlock()

	// Persist updated spent to DB periodically (on read)
	if spent != budget.Spent {
		s.db.Model(&budget).Update("spent", spent)
		budget.Spent = spent
	}

	pct := 0.0
	if budget.Limit > 0 {
		pct = (budget.Spent / budget.Limit) * 100
	}

	now := time.Now()
	start, end := s.periodRange(budget.Period, now)

	return &BudgetStatus{
		BudgetSetting: &BudgetSetting{
			UserID: budget.UserID,
			Period: budget.Period,
			Limit:  budget.Limit,
			Spent:  budget.Spent,
		},
		Exceeded:    budget.Spent >= budget.Limit,
		UsagePct:    pct,
		PeriodStart: start,
		PeriodEnd:   end,
	}, nil
}

func (s *Service) CheckBudget(userID uint) (bool, float64, error) {
	var budget model.UserBudget
	if err := s.db.Where("user_id = ?", userID).First(&budget).Error; err != nil {
		// No budget set, allow by default
		return true, 0, nil
	}

	s.mu.RLock()
	spent := s.spending[userID]
	s.mu.RUnlock()

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

	thresholdsJSON, err := json.Marshal(thresholds)
	if err != nil {
		return nil, fmt.Errorf("marshal thresholds: %w", err)
	}

	alert := model.CostAlert{
		UserID:     userID,
		Thresholds: string(thresholdsJSON),
	}
	if err := s.db.Where("user_id = ?", userID).Assign(alert).FirstOrCreate(&alert).Error; err != nil {
		return nil, fmt.Errorf("set alert: %w", err)
	}

	return &AlertSetting{
		UserID:     userID,
		Thresholds: thresholds,
	}, nil
}

func (s *Service) GetAlerts(userID uint) (*AlertSetting, error) {
	var alert model.CostAlert
	if err := s.db.Where("user_id = ?", userID).First(&alert).Error; err != nil {
		return nil, fmt.Errorf("alert not configured for user %d", userID)
	}

	var thresholds []float64
	if err := json.Unmarshal([]byte(alert.Thresholds), &thresholds); err != nil {
		return nil, fmt.Errorf("unmarshal thresholds: %w", err)
	}

	return &AlertSetting{
		UserID:     userID,
		Thresholds: thresholds,
	}, nil
}

// CheckAlerts checks current spending against thresholds, returns triggered thresholds
func (s *Service) CheckAlerts(userID uint) ([]float64, error) {
	var budget model.UserBudget
	if err := s.db.Where("user_id = ?", userID).First(&budget).Error; err != nil {
		return nil, nil
	}

	var alert model.CostAlert
	if err := s.db.Where("user_id = ?", userID).First(&alert).Error; err != nil {
		return nil, nil
	}

	var thresholds []float64
	if err := json.Unmarshal([]byte(alert.Thresholds), &thresholds); err != nil {
		return nil, nil
	}

	s.mu.RLock()
	spent := s.spending[userID]
	s.mu.RUnlock()

	pct := 0.0
	if budget.Limit > 0 {
		pct = (spent / budget.Limit) * 100
	}

	var triggered []float64
	for _, threshold := range thresholds {
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

// --- Cache Stats & Config ---

type CacheStats struct {
	TotalEntries    int     `json:"total_entries"`
	ActiveEntries   int     `json:"active_entries"`
	TotalHits       int64   `json:"total_hits"`
	HitRate         float64 `json:"hit_rate"`
	EstimateSavings string  `json:"estimate_savings"`
}

type CacheConfig struct {
	Enabled              bool    `json:"enabled"`
	SimilarityThreshold  float64 `json:"similarity_threshold"`
	TTLSeconds           int     `json:"ttl_seconds"`
	MaxEntries           int     `json:"max_entries"`
}

var cacheConfig = CacheConfig{
	Enabled:              true,
	SimilarityThreshold:  0.85,
	TTLSeconds:           3600,
	MaxEntries:           10000,
}

func (s *Service) GetCacheStats() (*CacheStats, error) {
	var totalEntries, activeEntries int64
	s.db.Model(&model.SemanticCache{}).Count(&totalEntries)
	s.db.Model(&model.SemanticCache{}).Where("expires_at > ?", time.Now().Unix()).Count(&activeEntries)

	var totalHits int64
	s.db.Model(&model.SemanticCache{}).Select("COALESCE(SUM(hit_count), 0)").Scan(&totalHits)

	hitRate := 0.0
	if totalEntries > 0 {
		hitRate = float64(totalHits) / float64(totalEntries)
	}

	return &CacheStats{
		TotalEntries:    int(totalEntries),
		ActiveEntries:   int(activeEntries),
		TotalHits:       totalHits,
		HitRate:         hitRate,
		EstimateSavings: fmt.Sprintf("%.2f%%", hitRate*90),
	}, nil
}

func (s *Service) UpdateCacheConfig(req *struct {
	Enabled              bool    `json:"enabled"`
	SimilarityThreshold  float64 `json:"similarity_threshold"`
	TTLSeconds           int     `json:"ttl_seconds"`
	MaxEntries           int     `json:"max_entries"`
}) error {
	if req.TTLSeconds > 0 {
		cacheConfig.TTLSeconds = req.TTLSeconds
	}
	if req.SimilarityThreshold > 0 {
		cacheConfig.SimilarityThreshold = req.SimilarityThreshold
	}
	if req.MaxEntries > 0 {
		cacheConfig.MaxEntries = req.MaxEntries
	}
	cacheConfig.Enabled = req.Enabled
	return nil
}
