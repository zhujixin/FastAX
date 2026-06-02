// 成本优化模型
//
// 本文件包含成本优化相关的数据模型：
//   - SemanticCache: 语义缓存表，存储相似查询的缓存结果，降低重复调用成本
//   - UserBudget: 用户预算表，设置用户的消费预算和告警阈值
//   - CostAlert: 成本告警表，记录预算告警触发历史
package model

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
