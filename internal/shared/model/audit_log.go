package model

import "time"

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
