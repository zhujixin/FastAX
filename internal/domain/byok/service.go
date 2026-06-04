// BYOK（自带 Key）模块业务逻辑层
//
// 本文件实现了 BYOK 相关的业务逻辑：
//
// Key 管理：
//   - ListKeys: 获取用户 Key 列表
//   - AddKey: 添加新 Key（加密存储）
//   - DeleteKey: 删除 Key
//   - SetKeyStatus: 启用/禁用 Key
//   - GetKey: 获取单个 Key 详情
//
// 安全特性：
//   - Key 加密存储（AES-256-GCM）
//   - 模型白名单限制
//   - 使用量统计
package byok

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fastax/fastax-server/internal/shared/model"
	"gorm.io/gorm"
)

// Service BYOK 服务结构体
type Service struct {
	db *gorm.DB
}

// NewService 创建 BYOK 服务实例
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// AddKeyRequest Key 添加请求
type AddKeyRequest struct {
	Provider       string `json:"provider" binding:"required"`
	KeyEncrypted   string `json:"key_encrypted" binding:"required"`
	KeyIV          string `json:"key_iv" binding:"required"`
	Alias          string `json:"alias"`
	ModelWhitelist string `json:"model_whitelist"`
}

// KeyResponse Key 响应
type KeyResponse struct {
	ID             uint   `json:"id"`
	Provider       string `json:"provider"`
	Alias          string `json:"alias"`
	ModelWhitelist string `json:"model_whitelist"`
	Status         int    `json:"status"`
	LastUsedAt     int64  `json:"last_used_at"`
	CreatedAt      int64  `json:"created_at"`
}

func (s *Service) AddKey(userID uint, req *AddKeyRequest) (*KeyResponse, error) {
	// Validate that key_encrypted looks like base64-encoded AES-GCM ciphertext
	// (minimum length: 12 bytes nonce + 16 bytes tag + 1 byte data = 29 bytes → ~40 base64 chars)
	if len(req.KeyEncrypted) < 40 {
		return nil, fmt.Errorf("key_encrypted appears to be plaintext; must be AES-GCM encrypted and base64-encoded")
	}

	// Limit users to max 20 BYOK keys to prevent abuse
	var count int64
	s.db.Model(&model.BYOKKey{}).Where("user_id = ?", userID).Count(&count)
	if count >= 20 {
		return nil, fmt.Errorf("maximum 20 BYOK keys per user")
	}

	key := model.BYOKKey{
		UserID:         userID,
		Provider:       req.Provider,
		KeyEncrypted:   req.KeyEncrypted,
		KeyIV:          req.KeyIV,
		Alias:          req.Alias,
		ModelWhitelist: req.ModelWhitelist,
		Status:         1,
		CreatedAt:      time.Now().Unix(),
	}
	if err := s.db.Create(&key).Error; err != nil {
		return nil, fmt.Errorf("add key: %w", err)
	}
	return toKeyResponse(&key), nil
}

func (s *Service) GetKey(id uint, userID uint) (*KeyResponse, error) {
	var key model.BYOKKey
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("key not found")
		}
		return nil, err
	}
	return toKeyResponse(&key), nil
}

func (s *Service) ListKeys(userID uint) ([]KeyResponse, error) {
	var keys []model.BYOKKey
	if err := s.db.Where("user_id = ?", userID).Order("created_at desc").Find(&keys).Error; err != nil {
		return nil, err
	}
	result := make([]KeyResponse, len(keys))
	for i, k := range keys {
		result[i] = *toKeyResponse(&k)
	}
	return result, nil
}

func (s *Service) SetKeyStatus(id, userID uint, status int) error {
	result := s.db.Model(&model.BYOKKey{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("key not found")
	}
	return nil
}

func (s *Service) DeleteKey(id, userID uint) error {
	result := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.BYOKKey{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("key not found")
	}
	return nil
}

// FindKeyForModel finds a BYOK key that matches the provider and model
func (s *Service) FindKeyForModel(userID uint, provider, modelName string) (*model.BYOKKey, error) {
	var keys []model.BYOKKey
	if err := s.db.Where("user_id = ? AND provider = ? AND status = 1", userID, provider).Find(&keys).Error; err != nil {
		return nil, err
	}

	for _, key := range keys {
		if key.ModelWhitelist == "" {
			return &key, nil
		}
		if containsModel(key.ModelWhitelist, modelName) {
			return &key, nil
		}
	}
	return nil, errors.New("no matching BYOK key")
}

func (s *Service) TouchKey(id uint) {
	s.db.Model(&model.BYOKKey{}).Where("id = ?", id).
		Update("last_used_at", time.Now().Unix())
}

func containsModel(whitelist, model string) bool {
	for _, part := range strings.Split(whitelist, ",") {
		if strings.TrimSpace(part) == model {
			return true
		}
	}
	return false
}

// UsageStats BYOK 用量统计响应
type UsageStats struct {
	TotalKeys     int   `json:"total_keys"`
	ActiveKeys    int   `json:"active_keys"`
	TotalCalls    int64 `json:"total_calls"`
	TotalTokens   int64 `json:"total_tokens"`
	RecentCalls   int64 `json:"recent_calls_30d"`
}

// GetUsageStats 聚合查询 BYOK 用量统计
func (s *Service) GetUsageStats(userID uint) (*UsageStats, error) {
	// Count keys
	var totalKeys, activeKeys int64
	s.db.Model(&model.BYOKKey{}).Where("user_id = ?", userID).Count(&totalKeys)
	s.db.Model(&model.BYOKKey{}).Where("user_id = ? AND status = 1", userID).Count(&activeKeys)

	// Count calls via call_log: find BYOK-related calls by matching supplier type
	// BYOK keys use the supplier code that matches the key's provider
	var totalCalls, totalTokens int64
	s.db.Model(&model.CallLog{}).
		Joins("JOIN suppliers ON suppliers.id = call_log.supplier_id").
		Where("call_log.user_id = ? AND suppliers.code IN (SELECT DISTINCT provider FROM byok_keys WHERE user_id = ?)", userID, userID).
		Count(&totalCalls)

	s.db.Model(&model.CallLog{}).
		Joins("JOIN suppliers ON suppliers.id = call_log.supplier_id").
		Where("call_log.user_id = ? AND suppliers.code IN (SELECT DISTINCT provider FROM byok_keys WHERE user_id = ?)", userID, userID).
		Select("COALESCE(SUM(tokens_total), 0)").
		Scan(&totalTokens)

	// Recent calls (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var recentCalls int64
	s.db.Model(&model.CallLog{}).
		Joins("JOIN suppliers ON suppliers.id = call_log.supplier_id").
		Where("call_log.user_id = ? AND suppliers.code IN (SELECT DISTINCT provider FROM byok_keys WHERE user_id = ?) AND call_log.created_at >= ?", userID, userID, thirtyDaysAgo).
		Count(&recentCalls)

	return &UsageStats{
		TotalKeys:   int(totalKeys),
		ActiveKeys:  int(activeKeys),
		TotalCalls:  totalCalls,
		TotalTokens: totalTokens,
		RecentCalls: recentCalls,
	}, nil
}

// RoutingPreference BYOK 路由偏好配置
type RoutingPreference struct {
	UserID            uint   `json:"user_id"`
	Mode              string `json:"mode"`               // byok_first, platform_only, byok_only
	FallbackEnabled   bool   `json:"fallback_enabled"`
	MaxPlatformFeePct int    `json:"max_platform_fee_pct"`
}

// SetPreference 设置用户的 BYOK 路由偏好（持久化到数据库）
func (s *Service) SetPreference(userID uint, mode string, fallbackEnabled *bool, maxPlatformFeePct int) (*RoutingPreference, error) {
	fallback := true
	if fallbackEnabled != nil {
		fallback = *fallbackEnabled
	}
	fee := maxPlatformFeePct
	if fee == 0 {
		fee = 5
	}

	// Validate fee percentage range
	if fee < 0 || fee > 100 {
		return nil, fmt.Errorf("max_platform_fee_pct must be between 0 and 100")
	}

	pref := model.BYOKPreference{
		UserID:            userID,
		Mode:              mode,
		FallbackEnabled:   fallback,
		MaxPlatformFeePct: fee,
	}
	if err := s.db.Where("user_id = ?", userID).Assign(pref).FirstOrCreate(&pref).Error; err != nil {
		return nil, fmt.Errorf("save preference: %w", err)
	}

	return &RoutingPreference{
		UserID:            pref.UserID,
		Mode:              pref.Mode,
		FallbackEnabled:   pref.FallbackEnabled,
		MaxPlatformFeePct: pref.MaxPlatformFeePct,
	}, nil
}

// GetPreference 获取用户的 BYOK 路由偏好（从数据库读取，未配置时返回默认值）
func (s *Service) GetPreference(userID uint) *RoutingPreference {
	var pref model.BYOKPreference
	if err := s.db.Where("user_id = ?", userID).First(&pref).Error; err != nil {
		// Not configured: return sensible defaults
		return &RoutingPreference{
			UserID:            userID,
			Mode:              "byok_first",
			FallbackEnabled:   true,
			MaxPlatformFeePct: 5,
		}
	}
	return &RoutingPreference{
		UserID:            pref.UserID,
		Mode:              pref.Mode,
		FallbackEnabled:   pref.FallbackEnabled,
		MaxPlatformFeePct: pref.MaxPlatformFeePct,
	}
}

func toKeyResponse(k *model.BYOKKey) *KeyResponse {
	return &KeyResponse{
		ID:             k.ID,
		Provider:       k.Provider,
		Alias:          k.Alias,
		ModelWhitelist: k.ModelWhitelist,
		Status:         k.Status,
		LastUsedAt:     k.LastUsedAt,
		CreatedAt:      k.CreatedAt,
	}
}
