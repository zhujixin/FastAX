// Package model 定义了 FastAX 平台的所有数据模型（GORM 模型）
//
// 本文件包含用户相关的数据模型：
//   - User: 用户主表，存储用户基本信息、认证信息、状态等
//   - UserProfile: 用户资料表，存储头像、昵称等扩展信息
//   - SubAccount: 企业子账户，支持团队/企业多账户管理
//   - UserSession: 用户会话表，管理 JWT Token 和设备信息
package model

import "time"

type User struct {
	ID               uint      `gorm:"primaryKey"`
	Username         string    `gorm:"uniqueIndex;size:64;not null"`
	PasswordHash     string    `gorm:"not null"`
	Email            string    `gorm:"uniqueIndex;size:128"`
	Phone            string    `gorm:"uniqueIndex;size:32"`
	Role             string    `gorm:"not null;default:user;size:32"`
	Level            string    `gorm:"not null;default:normal;size:32"`
	Status           int       `gorm:"not null;default:1"`
	EmailVerified    int       `gorm:"default:0"`
	PhoneVerified    int       `gorm:"default:0"`
	PreferredLanguage string   `gorm:"default:zh-CN;size:16"`
	LastLoginIP      string    `gorm:"size:45"`
	LastLoginAt      *time.Time
	LoginFailCount   int       `gorm:"default:0"`
	LockedUntil      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time

	Profile     *UserProfile `gorm:"foreignKey:UserID"`
	SubAccounts []SubAccount `gorm:"foreignKey:ParentID"`
}

type UserProfile struct {
	ID              uint   `gorm:"primaryKey"`
	UserID          uint   `gorm:"uniqueIndex;not null"`
	Avatar          string `gorm:"size:256"`
	RealName        string `gorm:"size:64"`
	IDNumber        string `gorm:"size:256"` // AES-256 encrypted
	CompanyName     string `gorm:"size:128"`
	BusinessLicense string `gorm:"size:256"`
	CompanyAddress  string `gorm:"size:256"`
	InviteCode      string `gorm:"uniqueIndex;size:32"`
	InvitedBy       *uint  `gorm:"index"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type SubAccount struct {
	ID           uint      `gorm:"primaryKey"`
	ParentID     uint      `gorm:"index;not null"`
	Email        string    `gorm:"size:128;not null"`
	PasswordHash string    `gorm:"not null"`
	TokenQuota   int64     `gorm:"default:0"`
	Permissions  string    `gorm:"type:text"`
	Status       int       `gorm:"default:1"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
