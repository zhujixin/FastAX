// 审计日志模型
//
// 本文件包含审计日志相关的数据模型：
//   - AuditLog: 审计日志表，记录所有管理操作（用户管理、配置变更、数据操作等）
//   - 用于安全审计、合规检查和问题追溯
package model

import "time"

// AuditLog 审计日志表，记录所有管理操作
type AuditLog struct {
	ID           uint      `gorm:"primaryKey"`
	TraceID      string    `gorm:"index;size:64;not null"`
	OperatorID   *uint     `gorm:"index"`
	OperatorName string    `gorm:"size:64"`
	OperatorIP   string    `gorm:"size:45"`
	Action       string    `gorm:"size:64;index;not null"`
	ResourceType string    `gorm:"size:64"`
	ResourceID   string    `gorm:"size:64"`
	Detail       string    `gorm:"type:text"`
	Result       string    `gorm:"size:16;not null"`
	FailReason   string    `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"index"`
}

func (AuditLog) TableName() string { return "audit_log" }
