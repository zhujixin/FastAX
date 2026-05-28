package model

import "time"

type SupplierVendor struct {
	ID               uint      `gorm:"primaryKey"`
	UserID           uint      `gorm:"uniqueIndex;not null"`
	CompanyName      string    `gorm:"size:128;not null"`
	ContactName      string    `gorm:"size:64"`
	ContactEmail     string    `gorm:"size:128;not null"`
	ContactPhone     string    `gorm:"size:32"`
	BusinessLicense  string    `gorm:"size:256"`
	APIBaseURL       string    `gorm:"size:256;not null"`
	APIAuthType      string    `gorm:"default:api_key;size:32"`
	APIKeyEncrypted  string    `gorm:"size:512"`
	CommissionRate   string    `gorm:"size:16;not null"`
	SettlementCycle  string    `gorm:"default:t+7;size:16"`
	Status           string    `gorm:"not null;default:pending;size:32;index"`
	RejectReason     string    `gorm:"type:text"`
	ApprovedAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type SupplierProduct struct {
	ID             uint      `gorm:"primaryKey"`
	VendorID       uint      `gorm:"index;not null"`
	Name           string    `gorm:"size:128;not null"`
	NameI18n       string    `gorm:"type:text"`
	Type           string    `gorm:"size:64;not null"`
	Model          string    `gorm:"size:64;not null"`
	APIEndpoint    string    `gorm:"size:256;not null"`
	AuthType       string    `gorm:"default:api_key;size:32"`
	Unit           string    `gorm:"size:32;not null"`
	Price          string    `gorm:"size:32;not null"`
	Currency       string    `gorm:"default:USD;size:8"`
	StockTotal     string    `gorm:"size:32"`
	StockRemaining string    `gorm:"size:32"`
	Status         string    `gorm:"default:pending_review;size:32;index"`
	HealthStatus   string    `gorm:"default:unknown;size:32;index"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Settlement struct {
	ID              uint      `gorm:"primaryKey"`
	VendorID        uint      `gorm:"index;not null"`
	SettlementNo    string    `gorm:"uniqueIndex;size:64;not null"`
	PeriodStart     time.Time `gorm:"not null"`
	PeriodEnd       time.Time `gorm:"not null"`
	TotalSales      string    `gorm:"size:32;not null"`
	CommissionAmount string   `gorm:"size:32;not null"`
	NetAmount       string    `gorm:"size:32;not null"`
	Currency        string    `gorm:"default:USD;size:8"`
	Status          string    `gorm:"default:pending;size:32;index"`
	PaymentMethod   string    `gorm:"size:32"`
	PaidAt          *time.Time
	Remark          string    `gorm:"type:text"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
