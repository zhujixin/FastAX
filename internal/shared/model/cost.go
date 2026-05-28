package model

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
