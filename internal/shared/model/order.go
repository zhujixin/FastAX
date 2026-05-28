package model

import "time"

type Order struct {
	ID             uint      `gorm:"primaryKey"`
	OrderNo        string    `gorm:"uniqueIndex;size:64;not null"`
	UserID         uint      `gorm:"index;not null"`
	ProductID      uint      `gorm:"index"`
	Quantity       string    `gorm:"size:32;not null"`
	UnitPrice      string    `gorm:"size:32;not null"`
	Amount         string    `gorm:"size:32;not null"`
	DiscountAmount string    `gorm:"size:32;default:0"`
	FinalAmount    string    `gorm:"size:32;not null"`
	Currency       string    `gorm:"default:CNY;size:8"`
	PaymentMethod  string    `gorm:"size:32"`
	Status         string    `gorm:"size:32;not null;index"`
	Remark         string    `gorm:"type:text"`
	PaidAt         *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time

	Payments []Payment `gorm:"foreignKey:OrderID"`
	Refunds  []Refund  `gorm:"foreignKey:OrderID"`
}

type Payment struct {
	ID              uint      `gorm:"primaryKey"`
	OrderID         uint      `gorm:"index;not null"`
	PaymentNo       string    `gorm:"uniqueIndex;size:64"`
	Amount          string    `gorm:"size:32;not null"`
	Method          string    `gorm:"size:32;not null"`
	Gateway         string    `gorm:"size:32;not null"`
	GatewayTradeNo  string    `gorm:"size:128"`
	GatewayStatus   string    `gorm:"size:32"`
	Status          string    `gorm:"size:32;not null"`
	RawResponse     string    `gorm:"type:text"`
	PaidAt          *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Refund struct {
	ID         uint      `gorm:"primaryKey"`
	OrderID    uint      `gorm:"index;not null"`
	PaymentID  uint      `gorm:"index"`
	RefundNo   string    `gorm:"uniqueIndex;size:64"`
	Amount     string    `gorm:"size:32;not null"`
	Reason     string    `gorm:"type:text"`
	Status     string    `gorm:"size:32;not null"`
	OperatorID *uint     `gorm:"index"`
	Remark     string    `gorm:"type:text"`
	CreatedAt  time.Time
	HandledAt  *time.Time
}
