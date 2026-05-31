package log

import (
	"bytes"
	"encoding/csv"
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

type CallLogRequest struct {
	TraceID          string `json:"trace_id"`
	UserID           uint   `json:"user_id"`
	SubAccountID     uint   `json:"sub_account_id"`
	ProductID        uint   `json:"product_id"`
	SupplierID       uint   `json:"supplier_id"`
	RequestPath      string `json:"request_path"`
	RequestModel     string `json:"request_model"`
	TokensPrompt     int    `json:"tokens_prompt"`
	TokensCompletion int    `json:"tokens_completion"`
	TokensTotal      int    `json:"tokens_total"`
	ResponseTimeMs   int    `json:"response_time_ms"`
	IsStream         bool   `json:"is_stream"`
	StatusCode       int    `json:"status_code"`
	Status           string `json:"status"`
	ErrorMessage     string `json:"error_message"`
	ClientIP         string `json:"client_ip"`
	UserAgent        string `json:"user_agent"`
}

func (s *Service) RecordCall(req *CallLogRequest) error {
	isStream := 0
	if req.IsStream {
		isStream = 1
	}
	log := model.CallLog{
		TraceID:          req.TraceID,
		UserID:           req.UserID,
		SubAccountID:     req.SubAccountID,
		ProductID:        req.ProductID,
		SupplierID:       req.SupplierID,
		RequestPath:      req.RequestPath,
		RequestModel:     req.RequestModel,
		TokensPrompt:     req.TokensPrompt,
		TokensCompletion: req.TokensCompletion,
		TokensTotal:      req.TokensTotal,
		ResponseTimeMs:   req.ResponseTimeMs,
		IsStream:         isStream,
		StatusCode:       req.StatusCode,
		Status:           req.Status,
		ErrorMessage:     req.ErrorMessage,
		ClientIP:         req.ClientIP,
		UserAgent:        req.UserAgent,
	}
	if err := s.db.Create(&log).Error; err != nil {
		return fmt.Errorf("record call log: %w", err)
	}
	return nil
}

type CallLogQuery struct {
	UserID    uint   `form:"user_id"`
	Status    string `form:"status"`
	Model     string `form:"model"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
	Page      int    `form:"page,default=1"`
	PageSize  int    `form:"page_size,default=20"`
}

func (s *Service) ListCallLogs(query *CallLogQuery) ([]model.CallLog, int64, error) {
	db := s.db.Model(&model.CallLog{})

	if query.UserID > 0 {
		db = db.Where("user_id = ?", query.UserID)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Model != "" {
		db = db.Where("request_model = ?", query.Model)
	}
	if query.StartTime != "" {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != "" {
		db = db.Where("created_at <= ?", query.EndTime)
	}

	var total int64
	db.Count(&total)

	page := max(query.Page, 1)
	pageSize := query.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var logs []model.CallLog
	if err := db.Order("created_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

// --- Audit Log ---

type AuditLogRequest struct {
	TraceID      string `json:"trace_id"`
	OperatorID   *uint  `json:"operator_id"`
	OperatorName string `json:"operator_name"`
	OperatorIP   string `json:"operator_ip"`
	Action       string `json:"action" binding:"required"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Detail       string `json:"detail"`
	Result       string `json:"result" binding:"required"`
	FailReason   string `json:"fail_reason"`
}

func (s *Service) RecordAudit(req *AuditLogRequest) error {
	audit := model.AuditLog{
		TraceID:      req.TraceID,
		OperatorID:   req.OperatorID,
		OperatorName: req.OperatorName,
		OperatorIP:   req.OperatorIP,
		Action:       req.Action,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Detail:       req.Detail,
		Result:       req.Result,
		FailReason:   req.FailReason,
	}
	if err := s.db.Create(&audit).Error; err != nil {
		return fmt.Errorf("record audit log: %w", err)
	}
	return nil
}

type AuditLogQuery struct {
	OperatorID   uint   `form:"operator_id"`
	Action       string `form:"action"`
	ResourceType string `form:"resource_type"`
	StartTime    string `form:"start_time"`
	EndTime      string `form:"end_time"`
	Page         int    `form:"page,default=1"`
	PageSize     int    `form:"page_size,default=20"`
}

func (s *Service) ListAuditLogs(query *AuditLogQuery) ([]model.AuditLog, int64, error) {
	db := s.db.Model(&model.AuditLog{})

	if query.OperatorID > 0 {
		db = db.Where("operator_id = ?", query.OperatorID)
	}
	if query.Action != "" {
		db = db.Where("action = ?", query.Action)
	}
	if query.ResourceType != "" {
		db = db.Where("resource_type = ?", query.ResourceType)
	}
	if query.StartTime != "" {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != "" {
		db = db.Where("created_at <= ?", query.EndTime)
	}

	var total int64
	db.Count(&total)

	page := max(query.Page, 1)
	pageSize := query.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var logs []model.AuditLog
	if err := db.Order("created_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

// Export returns audit logs matching the query as CSV bytes (no pagination limit).
func (s *Service) Export(query *AuditLogQuery) ([]byte, error) {
	db := s.db.Model(&model.AuditLog{})

	if query.OperatorID > 0 {
		db = db.Where("operator_id = ?", query.OperatorID)
	}
	if query.Action != "" {
		db = db.Where("action = ?", query.Action)
	}
	if query.ResourceType != "" {
		db = db.Where("resource_type = ?", query.ResourceType)
	}
	if query.StartTime != "" {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != "" {
		db = db.Where("created_at <= ?", query.EndTime)
	}

	var logs []model.AuditLog
	if err := db.Order("created_at desc").Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("export audit logs: %w", err)
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Header row
	_ = w.Write([]string{
		"ID", "TraceID", "OperatorID", "OperatorName", "OperatorIP",
		"Action", "ResourceType", "ResourceID", "Detail", "Result",
		"FailReason", "CreatedAt",
	})

	for _, l := range logs {
		operatorID := ""
		if l.OperatorID != nil {
			operatorID = strconv.FormatUint(uint64(*l.OperatorID), 10)
		}
		_ = w.Write([]string{
			strconv.FormatUint(uint64(l.ID), 10),
			l.TraceID,
			operatorID,
			l.OperatorName,
			l.OperatorIP,
			l.Action,
			l.ResourceType,
			l.ResourceID,
			l.Detail,
			l.Result,
			l.FailReason,
			l.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("flush csv: %w", err)
	}
	return buf.Bytes(), nil
}
