// 佣金结算模型
//
// 本文件包含佣金相关的数据模型：
//   - Commission: 佣金记录表，记录代理商的佣金产生和结算信息
//   - CommissionWithdraw: 佣金提现表，记录代理商的提现申请和处理状态
package model

import "time"

// Commission 佣金记录表，记录代理商的佣金产生和结算信息
type Commission struct {
	ID               uint      `gorm:"primaryKey"`
	AgentID          uint      `gorm:"index;not null"`
	CustomerID       uint      `gorm:"index;not null"`
	OrderID          uint      `gorm:"index;not null"`
	OrderAmount      string    `gorm:"size:32;not null"`
	CommissionRate   string    `gorm:"size:16;not null"`
	CommissionAmount string    `gorm:"size:32;not null"`
	Status           string    `gorm:"default:pending;size:32"`
	SettledAt        *time.Time
	CreatedAt        time.Time
}

func (Commission) TableName() string { return "commission" }

type Withdrawal struct {
	ID         uint       `gorm:"primaryKey"`
	AgentID    uint       `gorm:"index;not null"`
	Amount     string     `gorm:"size:32;not null"`
	Status     string     `gorm:"default:pending;size:32"`
	Reason     string     `gorm:"type:text"`
	CreatedAt  time.Time
	HandledAt  *time.Time
}

func (Withdrawal) TableName() string { return "withdrawal" }
