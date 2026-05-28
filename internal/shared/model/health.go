package model

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
