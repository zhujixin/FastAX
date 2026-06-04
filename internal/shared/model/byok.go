// BYOK（自带 Key）模型
//
// 本文件包含 BYOK 相关的数据模型：
//   - BYOKKey: 用户自带 Key 表，存储用户提供的供应商 API Key（加密存储）
//   - BYOKPreference: BYOK 路由偏好表，持久化用户的路由策略配置
package model

import "time"

// BYOKKey 用户自带 Key 表，存储用户提供的供应商 API Key（加密存储）
type BYOKKey struct {
	ID             uint   `gorm:"primaryKey"`
	UserID         uint   `gorm:"index:idx_byok_user;not null"`
	Provider       string `gorm:"size:32;not null"`
	KeyEncrypted   string `gorm:"size:512;not null"`
	KeyIV          string `gorm:"size:64;not null"`
	Alias          string `gorm:"size:64"`
	ModelWhitelist string `gorm:"type:text"`
	Status         int    `gorm:"default:1"`
	LastUsedAt     int64  `gorm:"default:0"`
	ExpiresAt      int64  `gorm:"default:0"`
	CreatedAt      int64  `gorm:"autoCreateTime"`
}

// BYOKPreference BYOK 路由偏好表，存储用户的路由策略配置
type BYOKPreference struct {
	ID                uint      `gorm:"primaryKey"`
	UserID            uint      `gorm:"uniqueIndex;not null"`
	Mode              string    `gorm:"size:32;not null;default:byok_first"` // byok_first, platform_only, byok_only
	FallbackEnabled   bool      `gorm:"default:true"`
	MaxPlatformFeePct int       `gorm:"default:5"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime"`
	CreatedAt         time.Time
}
