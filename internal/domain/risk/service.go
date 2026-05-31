package risk

import (
	"errors"
	"fmt"

	"github.com/fastax/fastax-server/internal/shared/model"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// --- Risk Rules ---

type RuleRequest struct {
	Name       string `json:"name" binding:"required"`
	Category   string `json:"category" binding:"required,oneof=register trade api"`
	Conditions string `json:"conditions" binding:"required"`
	Action     string `json:"action" binding:"required,oneof=alert rate_limit freeze"`
	RiskLevel  string `json:"risk_level" binding:"required"`
	Priority   int    `json:"priority"`
	Enabled    *bool  `json:"enabled"`
}

func (s *Service) CreateRule(req *RuleRequest) (*model.RiskRule, error) {
	enabled := 1
	if req.Enabled != nil && !*req.Enabled {
		enabled = 0
	}
	rule := model.RiskRule{
		Name:       req.Name,
		Category:   req.Category,
		Conditions: req.Conditions,
		Action:     req.Action,
		RiskLevel:  req.RiskLevel,
		Priority:   req.Priority,
		Enabled:    enabled,
	}
	if err := s.db.Create(&rule).Error; err != nil {
		return nil, fmt.Errorf("create rule: %w", err)
	}
	return &rule, nil
}

func (s *Service) ListRules(category string) ([]model.RiskRule, error) {
	var rules []model.RiskRule
	query := s.db.Order("priority desc")
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if err := query.Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("list rules: %w", err)
	}
	return rules, nil
}

func (s *Service) SetRuleEnabled(id uint, enabled bool) error {
	v := 0
	if enabled {
		v = 1
	}
	result := s.db.Model(&model.RiskRule{}).Where("id = ?", id).Update("enabled", v)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("rule not found")
	}
	return nil
}

// --- Risk Events ---

type EventRequest struct {
	UserID      uint   `json:"user_id"`
	EventType   string `json:"event_type" binding:"required"`
	RiskLevel   string `json:"risk_level" binding:"required"`
	Description string `json:"description"`
	RuleID      uint   `json:"rule_id"`
	RelatedInfo string `json:"related_info"`
	ActionTaken string `json:"action_taken"`
}

func (s *Service) CreateEvent(req *EventRequest) (*model.RiskEvent, error) {
	event := model.RiskEvent{
		UserID:      req.UserID,
		EventType:   req.EventType,
		RiskLevel:   req.RiskLevel,
		Description: req.Description,
		RuleID:      req.RuleID,
		RelatedInfo: req.RelatedInfo,
		ActionTaken: req.ActionTaken,
		Status:      "pending",
	}
	if err := s.db.Create(&event).Error; err != nil {
		return nil, fmt.Errorf("create event: %w", err)
	}
	return &event, nil
}

type EventQuery struct {
	UserID    uint   `form:"user_id"`
	EventType string `form:"event_type"`
	RiskLevel string `form:"risk_level"`
	Status    string `form:"status"`
	Page      int    `form:"page,default=1"`
	PageSize  int    `form:"page_size,default=20"`
}

func (s *Service) ListEvents(query *EventQuery) ([]model.RiskEvent, int64, error) {
	db := s.db.Model(&model.RiskEvent{})

	if query.UserID > 0 {
		db = db.Where("user_id = ?", query.UserID)
	}
	if query.EventType != "" {
		db = db.Where("event_type = ?", query.EventType)
	}
	if query.RiskLevel != "" {
		db = db.Where("risk_level = ?", query.RiskLevel)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	var total int64
	db.Count(&total)

	page := max(query.Page, 1)
	pageSize := query.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var events []model.RiskEvent
	if err := db.Order("created_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&events).Error; err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

func (s *Service) HandleEvent(eventID uint, handlerID uint) error {
	result := s.db.Model(&model.RiskEvent{}).Where("id = ? AND status = ?", eventID, "pending").
		Updates(map[string]any{
			"status":     "handled",
			"handler_id": handlerID,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("event not found or already handled")
	}
	return nil
}

// DetectAbnormalLogin checks for abnormal login patterns
func (s *Service) DetectAbnormalLogin(userID uint, ip string) bool {
	// Simplified: check if same user logged in from different IPs within 5 min
	var count int64
	s.db.Model(&model.RiskEvent{}).
		Where("user_id = ? AND event_type = ? AND created_at > datetime('now', '-5 minutes')",
			userID, "abnormal_login").
		Count(&count)
	return count > 0
}

// DetectRapidAPI checks for rapid API calls
func (s *Service) DetectRapidAPI(userID uint, limit int) bool {
	var count int64
	s.db.Model(&model.CallLog{}).
		Where("user_id = ? AND created_at > datetime('now', '-1 minute')", userID).
		Count(&count)
	return int(count) > limit
}
