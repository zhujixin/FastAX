package model

type BYOKKey struct {
	ID             uint   `gorm:"primaryKey"`
	UserID         uint   `gorm:"index:idx_byok_user;not null"`
	Provider       string `gorm:"size:32;not null"`
	KeyEncrypted   string `gorm:"size:512;not null"`
	KeyIV          string `gorm:"size:64;not null"`
	Alias          string `gorm:"size:64"`
	ModelWhitelist string `gorm:"type:text"`
	Status         int    `gorm:"default:1"`
	LastUsedAt     int64  `gorm:"default:0"`
	ExpiresAt      int64  `gorm:"default:0"`
	CreatedAt      int64  `gorm:"autoCreateTime"`
}
