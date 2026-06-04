// 成本优化模型
//
// 本文件包含成本优化相关的数据模型：
//   - UserBudget: 用户预算表，设置用户的消费预算（持久化到 DB）
//   - CostAlert: 成本告警表，存储用户告警配置（持久化到 DB）
//   - SemanticCache: 语义缓存表，存储相似查询的缓存结果，降低重复调用成本
package model

import "time"

// UserBudget 用户预算表，设置用户的消费预算和周期（持久化到 DB）
type UserBudget struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"uniqueIndex;not null"`
	Period    string    `gorm:"size:16;not null"` // daily, weekly, monthly
	Limit     float64   `gorm:"not null"`          // 预算上限
	Spent     float64   `gorm:"default:0"`         // 当前周期已消费
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CostAlert 成本告警表，存储用户告警配置（持久化到 DB）
type CostAlert struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `gorm:"uniqueIndex;not null"`
	Thresholds string    `gorm:"type:text;not null"` // JSON array e.g. "[50,80,100]"
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// SemanticCache 语义缓存表，存储相似查询的缓存结果
type SemanticCache struct {
	ID                uint   `gorm:"primaryKey"`
	PromptHash        string `gorm:"index;size:64;not null"`
	PromptVector      []byte `gorm:"type:blob"`
	ResponseEncrypted string `gorm:"type:text;not null"`
	Model             string `gorm:"size:64;not null"`
	HitCount          int    `gorm:"default:0"`
	CreatedAt         int64  `gorm:"autoCreateTime"`
	ExpiresAt         int64  `gorm:"index;not null"`
}
