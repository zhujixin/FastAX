// 供应商健康状态模型
//
// 本文件包含供应商健康监控相关的数据模型：
//   - ProviderHealth: 供应商健康状态表，记录供应商的可用性、延迟、错误率等指标
//   - 用于路由决策和熔断判断
package model

// ProviderHealth 供应商健康状态表，记录供应商的可用性、延迟、错误率等指标
type ProviderHealth struct {
	ID           uint    `gorm:"primaryKey"`
	ProviderID   uint    `gorm:"index:idx_health_provider;not null"`
	Status       int     `gorm:"not null"`
	AvgLatencyMs int     `gorm:"default:0"`
	ErrorRate    float64 `gorm:"default:0"`
	CheckCount   int     `gorm:"default:0"`
	PeriodStart  int64   `gorm:"not null"`
	PeriodEnd    int64   `gorm:"not null"`
}
