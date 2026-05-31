package guardrail

import (
	"errors"
	"fmt"
	"time"

	"github.com/fastax/fastax-server/internal/shared/model"
	"gorm.io/gorm"
)

type Service struct {
	db    *gorm.DB
	mode  string // "enforce" or "monitor"
}

func NewService(db *gorm.DB, mode string) *Service {
	if mode == "" {
		mode = "monitor"
	}
	return &Service{db: db, mode: mode}
}

// --- Rules ---

type RuleRequest struct {
	Name       string `json:"name" binding:"required"`
	Stage      string `json:"stage" binding:"required,oneof=before after"`
	Type       string `json:"type" binding:"required,oneof=pii injection secret content"`
	Action     string `json:"action" binding:"required,oneof=enforce monitor log"`
	Conditions string `json:"conditions"`
	Priority   int    `json:"priority"`
}

func (s *Service) CreateRule(req *RuleRequest) (*model.GuardrailRule, error) {
	rule := model.GuardrailRule{
		Name:       req.Name,
		Stage:      req.Stage,
		Type:       req.Type,
		Action:     req.Action,
		Conditions: req.Conditions,
		Priority:   req.Priority,
		Enabled:    1,
		CreatedAt:  time.Now().Unix(),
	}
	if err := s.db.Create(&rule).Error; err != nil {
		return nil, fmt.Errorf("create rule: %w", err)
	}
	return &rule, nil
}

func (s *Service) ListRules(stage string) ([]model.GuardrailRule, error) {
	var rules []model.GuardrailRule
	query := s.db.Where("enabled = 1").Order("priority desc")
	if stage != "" {
		query = query.Where("stage = ?", stage)
	}
	if err := query.Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

func (s *Service) SetRuleEnabled(id uint, enabled bool) error {
	v := 0
	if enabled {
		v = 1
	}
	result := s.db.Model(&model.GuardrailRule{}).Where("id = ?", id).Update("enabled", v)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("rule not found")
	}
	return nil
}

// --- Detection ---

type DetectRequest struct {
	TraceID string `json:"trace_id"`
	UserID  uint   `json:"user_id"`
	Stage   string `json:"stage" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type DetectResult struct {
	Blocked         bool     `json:"blocked"`
	DetectedEntities []string `json:"detected_entities,omitempty"`
	ActionTaken     string   `json:"action_taken"`
	RuleID          uint     `json:"rule_id,omitempty"`
}

func (s *Service) Detect(req *DetectRequest) (*DetectResult, error) {
	rules, err := s.ListRules(req.Stage)
	if err != nil {
		return nil, err
	}

	for _, rule := range rules {
		detected := s.applyRule(rule, req.Content)
		if !detected {
			continue
		}

		// Log the detection
		log := model.GuardrailLog{
			TraceID:          req.TraceID,
			UserID:           req.UserID,
			RuleID:           rule.ID,
			Stage:            req.Stage,
			DetectedEntities: rule.Type,
			ActionTaken:      rule.Action,
			CreatedAt:        time.Now().Unix(),
		}
		s.db.Create(&log)

		result := &DetectResult{
			DetectedEntities: []string{rule.Type},
			ActionTaken:      rule.Action,
			RuleID:           rule.ID,
		}

		if s.mode == "enforce" && rule.Action == "enforce" {
			result.Blocked = true
		}

		return result, nil
	}

	return &DetectResult{ActionTaken: "pass"}, nil
}

func (s *Service) applyRule(rule model.GuardrailRule, content string) bool {
	// Simplified detection logic
	switch rule.Type {
	case "injection":
		// Check for common injection patterns
		return containsInjection(content)
	case "pii":
		// Check for PII patterns
		return containsPII(content)
	default:
		return false
	}
}

func containsInjection(content string) bool {
	patterns := []string{"ignore previous", "system prompt", "you are now", "forget your instructions"}
	for _, p := range patterns {
		if containsIgnoreCase(content, p) {
			return true
		}
	}
	return false
}

func containsPII(content string) bool {
	// Simplified: check for patterns like email, phone
	return false
}

func containsIgnoreCase(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c1 := s[i+j]
			c2 := substr[j]
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 32
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 32
			}
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// --- Logs ---

func (s *Service) ListLogs(traceID string, userID uint, stage string) ([]model.GuardrailLog, error) {
	var logs []model.GuardrailLog
	query := s.db.Order("created_at desc")
	if traceID != "" {
		query = query.Where("trace_id = ?", traceID)
	}
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if stage != "" {
		query = query.Where("stage = ?", stage)
	}
	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
