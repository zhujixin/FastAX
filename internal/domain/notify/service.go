// 通知模块业务逻辑层
//
// 本文件实现了通知相关的业务逻辑：
//
// 通知发送：
//   - Send: 发送通知（站内信/短信/邮件）
//   - SendBatch: 批量发送通知
//   - SendByTemplate: 使用模板发送通知
//
// 用户通知：
//   - List: 获取用户通知列表（分页、筛选）
//   - UnreadCount: 获取未读通知数量
//   - MarkRead: 标记通知为已读
//   - MarkAllRead: 标记所有通知为已读
//
// 模板管理：
//   - ListTemplates: 获取模板列表
//   - CreateTemplate: 创建模板
//   - UpdateTemplate: 更新模板
package notify

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fastax/fastax-server/internal/shared/model"
	"gorm.io/gorm"
)

// Service 通知服务结构体
type Service struct {
	db *gorm.DB
}

// NewService 创建通知服务实例
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// SendRequest 通知发送请求
type SendRequest struct {
	UserID  uint   `json:"user_id" binding:"required"`
	Type    string `json:"type" binding:"required,oneof=order security expiry system"`
	Channel string `json:"channel" binding:"required,oneof=in_app sms email"`
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Language string `json:"language"`
}

// TemplateRequest 模板创建/更新请求
type TemplateRequest struct {
	Code     string `json:"code" binding:"required"`
	Name     string `json:"name"`
	Channel  string `json:"channel" binding:"required,oneof=in_app sms email"`
	Content  string `json:"content" binding:"required"`
	Language string `json:"language"`
	Status   *int   `json:"status"`
}

func (s *Service) Send(req *SendRequest) (*model.Notification, error) {
	notif := model.Notification{
		UserID:   req.UserID,
		Type:     req.Type,
		Channel:  req.Channel,
		Title:    req.Title,
		Content:  req.Content,
		Language: req.Language,
		IsRead:   0,
	}
	if err := s.db.Create(&notif).Error; err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}
	return &notif, nil
}

func (s *Service) SendFromTemplate(userID uint, templateCode string, params map[string]string) (*model.Notification, error) {
	var tmpl model.NotificationTemplate
	if err := s.db.Where("code = ? AND status = 1", templateCode).First(&tmpl).Error; err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// Simple template rendering (replace placeholders)
	title := tmpl.Name
	content := tmpl.Content
	for k, v := range params {
		placeholder := fmt.Sprintf("{{%s}}", k)
		title = replaceAll(title, placeholder, v)
		content = replaceAll(content, placeholder, v)
	}

	return s.Send(&SendRequest{
		UserID:  userID,
		Type:    "system",
		Channel: tmpl.Channel,
		Title:   title,
		Content: content,
		Language: tmpl.Language,
	})
}

func (s *Service) ListByUser(userID uint, notifType string, isRead *bool, page, pageSize int) ([]model.Notification, int64, error) {
	db := s.db.Model(&model.Notification{}).Where("user_id = ?", userID)

	if notifType != "" {
		db = db.Where("type = ?", notifType)
	}
	if isRead != nil {
		v := 0
		if *isRead {
			v = 1
		}
		db = db.Where("is_read = ?", v)
	}

	var total int64
	db.Count(&total)

	page = max(page, 1)
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var notifs []model.Notification
	if err := db.Order("created_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&notifs).Error; err != nil {
		return nil, 0, err
	}
	return notifs, total, nil
}

func (s *Service) MarkRead(notifID uint, userID uint) error {
	result := s.db.Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", notifID, userID).
		Updates(map[string]any{"is_read": 1})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("notification not found")
	}
	return nil
}

func (s *Service) MarkAllRead(userID uint) error {
	return s.db.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = 0", userID).
		Updates(map[string]any{"is_read": 1}).Error
}

func (s *Service) GetUnreadCount(userID uint) (int64, error) {
	var count int64
	err := s.db.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = 0", userID).
		Count(&count).Error
	return count, err
}

// ListTemplates queries notification templates with optional channel and language filters.
func (s *Service) ListTemplates(channel, language string) ([]model.NotificationTemplate, error) {
	db := s.db.Model(&model.NotificationTemplate{})
	if channel != "" {
		db = db.Where("channel = ?", channel)
	}
	if language != "" {
		db = db.Where("language = ?", language)
	}
	var templates []model.NotificationTemplate
	if err := db.Order("id asc").Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	return templates, nil
}

// CreateTemplate creates a new notification template.
func (s *Service) CreateTemplate(req *TemplateRequest) (*model.NotificationTemplate, error) {
	tmpl := model.NotificationTemplate{
		Code:     req.Code,
		Name:     req.Name,
		Channel:  req.Channel,
		Content:  req.Content,
		Language: req.Language,
		Status:   1, // default enabled
	}
	if tmpl.Language == "" {
		tmpl.Language = "zh-CN"
	}
	if req.Status != nil {
		tmpl.Status = *req.Status
	}
	if err := s.db.Create(&tmpl).Error; err != nil {
		return nil, fmt.Errorf("create template: %w", err)
	}
	return &tmpl, nil
}

// UpdateTemplate updates an existing notification template by ID.
func (s *Service) UpdateTemplate(id uint, req *TemplateRequest) (*model.NotificationTemplate, error) {
	var tmpl model.NotificationTemplate
	if err := s.db.First(&tmpl, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("template not found")
		}
		return nil, fmt.Errorf("find template: %w", err)
	}

	updates := map[string]any{
		"code":     req.Code,
		"name":     req.Name,
		"channel":  req.Channel,
		"content":  req.Content,
		"language": req.Language,
	}
	if req.Language == "" {
		updates["language"] = "zh-CN"
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := s.db.Model(&tmpl).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("update template: %w", err)
	}

	// Reload to get updated fields
	s.db.First(&tmpl, id)
	return &tmpl, nil
}

func replaceAll(s, old, new string) string {
	var result strings.Builder
	for {
		idx := findIndex(s, old)
		if idx < 0 {
			result.WriteString(s)
			break
		}
		result.WriteString(s[:idx])
		result.WriteString(new)
		s = s[idx+len(old):]
	}
	return result.String()
}

func findIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
