package model

import "time"

type Notification struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index:idx_user_read;not null"`
	Type      string    `gorm:"size:32;not null"`
	Channel   string    `gorm:"size:16;not null"`
	Title     string    `gorm:"size:256;not null"`
	Content   string    `gorm:"type:text;not null"`
	Language  string    `gorm:"size:16"`
	IsRead    int       `gorm:"default:0;index:idx_user_read"`
	ReadAt    *time.Time
	CreatedAt time.Time `gorm:"index:idx_user_created"`
}

type NotificationTemplate struct {
	ID        uint      `gorm:"primaryKey"`
	Code      string    `gorm:"uniqueIndex;size:64;not null"`
	Name      string    `gorm:"size:128"`
	Channel   string    `gorm:"size:16;not null"`
	Content   string    `gorm:"type:text;not null"`
	Language  string    `gorm:"default:zh-CN;size:16"`
	Status    int       `gorm:"default:1"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
