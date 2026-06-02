// 风控与安全模型
//
// 本文件包含风控相关的数据模型：
//   - RiskEvent: 风控事件表，记录检测到的异常行为（频繁调用、异常消费等）
//   - RiskRule: 风控规则表，定义触发风控的条件和处理方式
//   - Blacklist: 黑名单表，存储被封禁的用户、IP、设备等
package model

import "time"

// RiskEvent 风控事件表，记录检测到的异常行为
type RiskEvent struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"index;not null"`
	EventType   string    `gorm:"size:64;not null"`
	RiskLevel   string    `gorm:"size:16;not null"`
	Description string    `gorm:"type:text"`
	RuleID      uint      `gorm:"default:0"`
	RelatedInfo string    `gorm:"type:text"`
	ActionTaken string    `gorm:"size:64"`
	Status      string    `gorm:"default:pending;size:32"`
	HandlerID   *uint     `gorm:"index"`
	HandledAt   *time.Time
	CreatedAt   time.Time
}

type RiskRule struct {
	ID         uint      `gorm:"primaryKey"`
	Name       string    `gorm:"size:128;not null"`
	Category   string    `gorm:"size:32;not null"`
	Conditions string    `gorm:"type:text;not null"`
	Action     string    `gorm:"size:32;not null"`
	RiskLevel  string    `gorm:"size:16;not null"`
	Priority   int       `gorm:"default:0"`
	Enabled    int       `gorm:"default:1"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
