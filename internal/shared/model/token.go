package model

import "time"

type TokenProduct struct {
	ID              uint      `gorm:"primaryKey"`
	SupplierID      uint      `gorm:"index;not null"`
	Name            string    `gorm:"size:128;not null"`
	NameI18n        string    `gorm:"type:text"`
	Type            string    `gorm:"size:64;not null"`
	Model           string    `gorm:"size:64;index"`
	Unit            string    `gorm:"size:32;not null"`
	Price           string    `gorm:"size:32;not null"`
	OriginalPrice   string    `gorm:"size:32"`
	Currency        string    `gorm:"default:CNY;size:8"`
	Description     string    `gorm:"type:text"`
	DescriptionI18n string    `gorm:"type:text"`
	ValidityDays    *int      `gorm:"default:null"`
	UsageNotes      string    `gorm:"type:text"`
	SortOrder       int       `gorm:"default:0"`
	Status          int       `gorm:"default:1;index"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type TokenInventory struct {
	ID               uint      `gorm:"primaryKey"`
	SupplierID       uint      `gorm:"index;not null"`
	ProductID        uint      `gorm:"index;not null"`
	TotalAmount      string    `gorm:"size:32;not null"`
	RemainingAmount  string    `gorm:"size:32;not null"`
	AlertThreshold   float64   `gorm:"default:10"`
	LastSyncedAt     *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type UserToken struct {
	ID           uint      `gorm:"primaryKey"`
	UserID       uint      `gorm:"index:idx_user_product,not null"`
	ProductID    uint      `gorm:"index:idx_user_product"`
	OrderID      *uint     `gorm:"default:null"`
	TotalAmount  string    `gorm:"size:32;not null"`
	UsedAmount   string    `gorm:"size:32;default:0"`
	FrozenAmount string    `gorm:"size:32;default:0"`
	ExpiresAt    *time.Time
	Status       int       `gorm:"default:1;index:idx_user_status"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type TokenTransfer struct {
	ID         uint      `gorm:"primaryKey"`
	FromUserID uint      `gorm:"index;not null"`
	ToUserID   uint      `gorm:"index;not null"`
	ProductID  uint      `gorm:"index"`
	Amount     string    `gorm:"size:32;not null"`
	Status     string    `gorm:"default:pending;size:32"`
	CreatedAt  time.Time
	HandledAt  *time.Time
}
