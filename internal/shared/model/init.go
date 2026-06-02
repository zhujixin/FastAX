// 数据库初始化与迁移
//
// 本文件负责数据库的初始化和表结构迁移：
//   - InitDB: 初始化 SQLite 数据库连接（WAL 模式）
//   - AutoMigrate: 自动迁移所有表结构
//   - 设置数据库连接池参数（最大连接数、空闲连接数、连接生命周期）
package model

import (
	"fmt"
	"time"

	"github.com/fastax/fastax-server/internal/shared/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库实例
var DB *gorm.DB

func InitDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	logLevel := logger.Warn
	switch cfg.LogLevel {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "info":
		logLevel = logger.Info
	}

	db, err := gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if cfg.WALMode {
		db.Exec("PRAGMA journal_mode=WAL")
		db.Exec("PRAGMA busy_timeout=5000")
		db.Exec("PRAGMA synchronous=NORMAL")
		db.Exec("PRAGMA cache_size=-20000") // 20MB cache
	}

	// AutoMigrate all tables
	if err := db.AutoMigrate(
		&User{}, &UserProfile{}, &SubAccount{},
		&TokenProduct{}, &TokenInventory{}, &UserToken{}, &TokenTransfer{},
		&Supplier{}, &Ability{},
		&Order{}, &Payment{}, &Refund{},
		&CallLog{},
		&RiskEvent{}, &RiskRule{},
		&Notification{}, &NotificationTemplate{},
		&Commission{},
		&Withdrawal{},
		&AuditLog{},
		&SupplierVendor{}, &SupplierProduct{}, &Settlement{},
		&GuardrailRule{}, &GuardrailLog{},
		&BYOKKey{},
		&SupportedLanguage{}, &ModelVariant{},
		&SemanticCache{}, &ProviderHealth{},
	); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	// Set global DB for use by domain packages (for cross-domain queries)
	DB = db

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get underlying db: %w", err)
	}
	sqlDB.SetMaxOpenConns(1) // SQLite WAL: single writer
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
