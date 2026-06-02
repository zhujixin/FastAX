// API 调用日志模型
//
// 本文件包含调用日志相关的数据模型：
//   - CallLog: API 调用日志表，记录每次代理转发的详细信息（请求、响应、耗时、Token 消耗等）
//   - 用于计费、审计、问题排查和性能分析
package model

import "time"

// CallLog API 调用日志表，记录每次代理转发的详细信息
type CallLog struct {
	ID               uint      `gorm:"primaryKey"`
	TraceID          string    `gorm:"index;size:64;not null"`
	UserID           uint      `gorm:"index:idx_user_created;not null"`
	SubAccountID     uint      `gorm:"default:0"`
	ProductID        uint      `gorm:"index"`
	SupplierID       uint      `gorm:"index:idx_supplier_created"`
	RequestPath      string    `gorm:"size:128;not null"`
	RequestModel     string    `gorm:"size:64"`
	TokensPrompt     int       `gorm:"default:0"`
	TokensCompletion int       `gorm:"default:0"`
	TokensTotal      int       `gorm:"default:0"`
	ResponseTimeMs   int       `gorm:"default:0"`
	IsStream         int       `gorm:"default:0"`
	StatusCode       int       `gorm:"type:smallint"`
	Status           string    `gorm:"size:32;not null"`
	ErrorMessage     string    `gorm:"type:text"`
	ClientIP         string    `gorm:"size:45"`
	UserAgent        string    `gorm:"size:256"`
	CreatedAt        time.Time `gorm:"index"`
}

func (CallLog) TableName() string { return "call_log" }
