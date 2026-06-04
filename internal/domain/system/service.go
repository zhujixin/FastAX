// 系统配置模块业务逻辑层
//
// 本文件实现了系统配置相关的业务逻辑：
//   - GetConfig: 获取系统配置（所有 key-value）
//   - UpdateConfig: 批量更新系统配置
//   - 配置项存储在 system_config 数据库表中，同时缓存到内存
package system

import (
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
)

// Service 系统配置服务结构体
type Service struct {
	db     *gorm.DB
	mu     sync.RWMutex
	cache  map[string]string
}

// ConfigEntry 配置条目
type ConfigEntry struct {
	ID          uint   `json:"id"`
	ConfigKey   string `json:"config_key"`
	ConfigValue string `json:"config_value"`
	Description string `json:"description,omitempty"`
}

// NewService 创建系统配置服务实例
func NewService(db *gorm.DB) *Service {
	svc := &Service{
		db:    db,
		cache: make(map[string]string),
	}
	// Auto-migrate system_config table
	db.Exec(`CREATE TABLE IF NOT EXISTS system_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		config_key TEXT UNIQUE NOT NULL,
		config_value TEXT NOT NULL,
		description TEXT,
		created_at INTEGER NOT NULL DEFAULT 0,
		updated_at INTEGER NOT NULL DEFAULT 0
	)`)
	// Seed defaults
	svc.seedDefaults()
	return svc
}

func (s *Service) seedDefaults() {
	defaults := map[string]string{
		"site_name":        "FastAX",
		"max_retries":      "3",
		"log_level":        "info",
		"enable_signup":    "true",
		"auto_disable_rate": "0.5",
	}
	for k, v := range defaults {
		s.db.Exec(`INSERT OR IGNORE INTO system_configs (config_key, config_value, description, created_at, updated_at)
			VALUES (?, ?, 'auto-seeded', ?, ?)`, k, v, now(), now())
		s.cache[k] = v
	}
}

// TableName returns the table name for GORM
type SystemConfig struct {
	ID          uint   `gorm:"primaryKey"`
	ConfigKey   string `gorm:"uniqueIndex;size:128;not null"`
	ConfigValue string `gorm:"not null"`
	Description string `gorm:"size:256"`
	CreatedAt   int64  `gorm:"not null;default:0"`
	UpdatedAt   int64  `gorm:"not null;default:0"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}

// GetAll returns all config entries
func (s *Service) GetAll() ([]ConfigEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var rows []SystemConfig
	if err := s.db.Order("config_key").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query config: %w", err)
	}
	entries := make([]ConfigEntry, len(rows))
	for i, r := range rows {
		entries[i] = ConfigEntry{
			ID:          r.ID,
			ConfigKey:   r.ConfigKey,
			ConfigValue: r.ConfigValue,
			Description: r.Description,
		}
	}
	return entries, nil
}

// allowedConfigKeys defines the whitelist of config keys that can be updated via API.
var allowedConfigKeys = map[string]bool{
	"site_name":         true,
	"max_retries":       true,
	"log_level":         true,
	"enable_signup":     true,
	"auto_disable_rate": true,
}

// UpdateAll updates multiple config values at once.
// Only keys in allowedConfigKeys are accepted; unknown keys are rejected.
func (s *Service) UpdateAll(updates map[string]string) error {
	// Validate all keys before applying any changes
	for key := range updates {
		if !allowedConfigKeys[key] {
			return fmt.Errorf("unknown config key: %s", key)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for key, value := range updates {
		result := s.db.Exec(`INSERT INTO system_configs (config_key, config_value, description, created_at, updated_at)
			VALUES (?, ?, 'manual', ?, ?)
			ON CONFLICT(config_key) DO UPDATE SET config_value = ?, updated_at = ?`,
			key, value, now(), now(), value, now())
		if result.Error != nil {
			return fmt.Errorf("update config %s: %w", key, result.Error)
		}
		s.cache[key] = value
	}
	return nil
}

// Get returns a single config value.
// Uses a write lock for simplicity — Get is not on a hot path and the extra
// contention from briefly holding a write lock is negligible compared to the
// complexity of lock escalation (RLock → Lock → RLock).
func (s *Service) Get(key string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if v, ok := s.cache[key]; ok {
		return v
	}

	// Lazy-load from DB
	var row SystemConfig
	if err := s.db.Where("config_key = ?", key).First(&row).Error; err == nil {
		s.cache[key] = row.ConfigValue
		return row.ConfigValue
	}
	return ""
}

func now() int64 {
	return time.Now().Unix()
}
