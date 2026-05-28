package model

import "time"

// Supplier represents a model provider (OpenAI, Claude, etc.)
type Supplier struct {
	ID              uint      `gorm:"primaryKey"`
	Name            string    `gorm:"size:128;not null"`
	Code            string    `gorm:"uniqueIndex;size:32;not null"`
	Description     string    `gorm:"type:text"`
	APIBaseURL      string    `gorm:"size:256;not null"`
	APIKeyEncrypted string    `gorm:"size:512;not null"`
	Models          string    `gorm:"type:text"`
	Region          string    `gorm:"default:overseas;size:32;index"`
	Status          int       `gorm:"default:1;index"`
	Priority        int       `gorm:"default:0"`
	Weight          int       `gorm:"default:10"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Ability maps model+group+channel for O(1) routing
type Ability struct {
	ID        uint   `gorm:"primaryKey"`
	Group     string `gorm:"uniqueIndex:idx_ability;size:32;not null"`
	Model     string `gorm:"uniqueIndex:idx_ability;size:64;not null"`
	ChannelID uint   `gorm:"uniqueIndex:idx_ability;not null"`
	Enabled   bool   `gorm:"default:true"`
}
