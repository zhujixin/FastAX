// 供应商渠道与能力模型
//
// 本文件包含渠道相关的数据模型：
//   - Supplier: 供应商表，存储 Token 供应商信息（OpenAI、Claude 等）
//   - Channel: 渠道表，定义供应商的 API 接入点、认证方式、状态等
//   - AbilityIndex: 能力索引表，记录渠道支持的模型列表，用于路由决策
//   - ModelVariant: 模型变体表，支持同一模型的多种配置（如不同参数、价格）
package model

import "time"

// Supplier 供应商表，存储 Token 供应商信息（如 OpenAI、Claude、Gemini 等）
type Supplier struct {
	ID              uint      `gorm:"primaryKey"`
	Name            string    `gorm:"size:128;not null"`
	Code            string    `gorm:"uniqueIndex;size:32;not null"`
	Description     string    `gorm:"type:text"`
	APIBaseURL      string    `gorm:"size:256;not null"`
	APIKeyEncrypted string    `gorm:"size:512;not null"`
	Models          string    `gorm:"type:text"`
	Region          string    `gorm:"default:overseas;size:32;index"`
	Status          int       `gorm:"default:1;index"`
	Priority        int       `gorm:"default:0"`
	Weight          int       `gorm:"default:10"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Ability maps model+group+channel for O(1) routing
type Ability struct {
	ID        uint   `gorm:"primaryKey"`
	Group     string `gorm:"uniqueIndex:idx_ability;size:32;not null"`
	Model     string `gorm:"uniqueIndex:idx_ability;size:64;not null"`
	ChannelID uint   `gorm:"uniqueIndex:idx_ability;not null"`
	Enabled   bool   `gorm:"default:true"`
}
