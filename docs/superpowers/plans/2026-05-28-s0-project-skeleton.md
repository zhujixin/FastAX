# S0: 项目骨架 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Scaffold the FastAX Go project with shared infrastructure: config, all GORM models, Redis cache, Gin middleware, and entry point — so `go run ./cmd/fastax` starts a working server with auto-migrated DB.

**Architecture:** Go 1.22+ monolithic with domain packages under `internal/`. S0 establishes `internal/shared/{config,model,cache,middleware,response}` and `internal/router/`. No business logic yet — only plumbing.

**Tech Stack:** Go 1.22+ / Gin / GORM / SQLite (WAL mode, via `go-sqlite3` or `gorm.io/driver/sqlite`) / Redis (go-redis) / Viper / golang-jwt

---

### Task 1: Go module + project scaffolding

**Files:**
- Create: `go.mod`
- Create: `cmd/fastax/main.go`
- Create: `config.example.yaml`

- [ ] **Step 1: Initialize Go module**

Run:
```bash
cd d:\Token\FastAX
go mod init github.com/fastax/fastax-server
```

- [ ] **Step 2: Create config.example.yaml**

```yaml
server:
  port: 8080
  mode: debug        # debug | release | test
  read_timeout: 30s
  write_timeout: 30s

database:
  path: data/fastax.db
  wal_mode: true
  log_level: warn    # silent | error | warn | info

redis:
  addr: localhost:6379
  password: ""
  db: 0
  pool_size: 10

jwt:
  secret: "change-me-in-production"
  access_expiry: 24h
  refresh_expiry: 168h  # 7 days

rate_limit:
  ip: 60
  auth: 5
  user_default: 60
  user_enterprise: 300
```

- [ ] **Step 3: Create cmd/fastax/main.go**

```go
package main

import (
	"log"

	"github.com/fastax/fastax-server/internal/shared/config"
	"github.com/fastax/fastax-server/internal/shared/model"
	"github.com/fastax/fastax-server/internal/router"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load config
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := model.InitDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// Create Gin engine
	gin.SetMode(cfg.Server.Mode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Register routes
	router.RegisterRoutes(r, db, cfg)

	// Start server
	log.Printf("FastAX server starting on :%d", cfg.Server.Port)
	if err := r.Run(cfg.Server.Addr()); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
```

- [ ] **Step 4: Verify it compiles**

Run: `go mod tidy`
Expected: `go.mod` and `go.sum` created, no errors.

---

### Task 2: Config management (Viper)

**Files:**
- Create: `internal/shared/config/config.go`

- [ ] **Step 1: Write config.go**

```go
package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

func (s ServerConfig) Addr() string {
	return fmt.Sprintf(":%d", s.Port)
}

type DatabaseConfig struct {
	Path     string `mapstructure:"path"`
	WALMode  bool   `mapstructure:"wal_mode"`
	LogLevel string `mapstructure:"log_level"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type JWTConfig struct {
	Secret        string        `mapstructure:"secret"`
	AccessExpiry  time.Duration `mapstructure:"access_expiry"`
	RefreshExpiry time.Duration `mapstructure:"refresh_expiry"`
}

type RateLimitConfig struct {
	IP              int `mapstructure:"ip"`
	Auth            int `mapstructure:"auth"`
	UserDefault     int `mapstructure:"user_default"`
	UserEnterprise  int `mapstructure:"user_enterprise"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.AutomaticEnv()
	v.SetEnvPrefix("FASTAX")

	// Defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("database.path", "data/fastax.db")
	v.SetDefault("database.wal_mode", true)
	v.SetDefault("database.log_level", "warn")
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("jwt.access_expiry", "24h")
	v.SetDefault("jwt.refresh_expiry", "168h")
	v.SetDefault("rate_limit.ip", 60)
	v.SetDefault("rate_limit.auth", 5)
	v.SetDefault("rate_limit.user_default", 60)
	v.SetDefault("rate_limit.user_enterprise", 300)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./internal/shared/config/`
Expected: no errors.

---

### Task 3: Unified response helpers + error codes

**Files:**
- Create: `internal/shared/response/response.go`

- [ ] **Step 1: Write response.go**

```go
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
}

type PaginatedData struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

const (
	CodeSuccess = 0

	CodeParamInvalid    = 1001
	CodeVerifyFailed    = 1002
	CodeTokenExpired    = 2001
	CodeAccountFrozen   = 2002
	CodePermissionDeny  = 3001
	CodeRateLimited     = 3002
	CodeNotFound        = 4001
	CodeDuplicateOp     = 5001
	CodeBalanceInsufficient = 6001
	CodeTokenExpiredOp  = 6002
	CodeOverLimit       = 6003
	CodeTooFrequent     = 7001
	CodeInternalError   = 9001
	CodeServiceUnavail  = 9002
)

var codeMessages = map[int]string{
	CodeSuccess:         "success",
	CodeParamInvalid:    "invalid parameters",
	CodeVerifyFailed:    "verification code expired or invalid",
	CodeTokenExpired:    "token expired or invalid",
	CodeAccountFrozen:   "account frozen",
	CodePermissionDeny:  "permission denied",
	CodeRateLimited:     "rate limit exceeded",
	CodeNotFound:        "resource not found",
	CodeDuplicateOp:     "duplicate operation",
	CodeBalanceInsufficient: "insufficient balance",
	CodeTokenExpiredOp:  "token expired",
	CodeOverLimit:       "purchase limit exceeded",
	CodeTooFrequent:     "too frequent",
	CodeInternalError:   "internal error",
	CodeServiceUnavail:  "service unavailable",
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

func SuccessPaginated(c *gin.Context, items interface{}, total int64, page, size int) {
	Success(c, PaginatedData{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

func Error(c *gin.Context, httpStatus, code int, msg ...string) {
	message := codeMessages[code]
	if len(msg) > 0 && msg[0] != "" {
		message = msg[0]
	}
	c.AbortWithStatusJSON(httpStatus, APIResponse{
		Code:    code,
		Message: message,
	})
}

func InternalError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, CodeInternalError)
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./internal/shared/response/`
Expected: no errors.

---

### Task 4: Core user/token/supplier GORM models

**Files:**
- Create: `internal/shared/model/user.go`
- Create: `internal/shared/model/token.go`
- Create: `internal/shared/model/channel.go`

- [ ] **Step 1: Write user.go** (User, UserProfile, SubAccount)

```go
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
	LastLoginAt      *time.Time`
	LoginFailCount   int       `gorm:"default:0"`
	LockedUntil      *time.Time`
	CreatedAt        time.Time`
	UpdatedAt        time.Time`

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
	CreatedAt       time.Time`
	UpdatedAt       time.Time`
}

type SubAccount struct {
	ID           uint      `gorm:"primaryKey"`
	ParentID     uint      `gorm:"index;not null"`
	Email        string    `gorm:"size:128;not null"`
	PasswordHash string    `gorm:"not null"`
	TokenQuota   int64     `gorm:"default:0"`
	Permissions  string    `gorm:"type:text"`
	Status       int       `gorm:"default:1"`
	CreatedAt    time.Time`
	UpdatedAt    time.Time`
}
```

- [ ] **Step 2: Write token.go** (TokenProduct, TokenInventory, UserToken, TokenTransfer)

```go
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
	ValidityDays    *int`
	UsageNotes      string    `gorm:"type:text"`
	SortOrder       int       `gorm:"default:0"`
	Status          int       `gorm:"default:1;index"`
	CreatedAt       time.Time`
	UpdatedAt       time.Time`
}

type TokenInventory struct {
	ID               uint      `gorm:"primaryKey"`
	SupplierID       uint      `gorm:"index;not null"`
	ProductID        uint      `gorm:"index;not null"`
	TotalAmount      string    `gorm:"size:32;not null"`
	RemainingAmount  string    `gorm:"size:32;not null"`
	AlertThreshold   float64   `gorm:"default:10"`
	LastSyncedAt     *time.Time`
	CreatedAt        time.Time`
	UpdatedAt        time.Time`
}

type UserToken struct {
	ID           uint      `gorm:"primaryKey"`
	UserID       uint      `gorm:"index:idx_user_product,not null"`
	ProductID    uint      `gorm:"index:idx_user_product"`
	OrderID      *uint`
	TotalAmount  string    `gorm:"size:32;not null"`
	UsedAmount   string    `gorm:"size:32;default:0"`
	FrozenAmount string    `gorm:"size:32;default:0"`
	ExpiresAt    *time.Time`
	Status       int       `gorm:"default:1;index:idx_user_status"`
	CreatedAt    time.Time`
	UpdatedAt    time.Time`
}

type TokenTransfer struct {
	ID         uint      `gorm:"primaryKey"`
	FromUserID uint      `gorm:"index;not null"`
	ToUserID   uint      `gorm:"index;not null"`
	ProductID  uint      `gorm:"index"`
	Amount     string    `gorm:"size:32;not null"`
	Status     string    `gorm:"default:pending;size:32"`
	CreatedAt  time.Time`
	HandledAt  *time.Time`
}
```

- [ ] **Step 3: Write channel.go** (Channel/Supplier, Ability)

```go
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
	CreatedAt       time.Time`
	UpdatedAt       time.Time`
}

// Ability maps model+group+channel for O(1) routing
type Ability struct {
	ID        uint   `gorm:"primaryKey"`
	Group     string `gorm:"uniqueIndex:idx_ability;size:32;not null"`
	Model     string `gorm:"uniqueIndex:idx_ability;size:64;not null"`
	ChannelID uint   `gorm:"uniqueIndex:idx_ability;not null"`
	Enabled   bool   `gorm:"default:true"`
}
```

- [ ] **Step 4: Verify compilation**

Run: `go build ./internal/shared/model/`
Expected: no errors.

---

### Task 5: Order/payment + call_log models

**Files:**
- Create: `internal/shared/model/order.go`
- Create: `internal/shared/model/call_log.go`

- [ ] **Step 1: Write order.go** (Order, Payment, Refund)

```go
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
	PaidAt         *time.Time`
	CreatedAt      time.Time`
	UpdatedAt      time.Time`

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
	PaidAt          *time.Time`
	CreatedAt       time.Time`
	UpdatedAt       time.Time`
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
	CreatedAt  time.Time`
	HandledAt  *time.Time`
}
```

- [ ] **Step 2: Write call_log.go** (CallLog)

```go
package model

import "time"

type CallLog struct {
	ID               uint      `gorm:"primaryKey"`
	TraceID          string    `gorm:"index;size:64;not null"`
	UserID           uint      `gorm:"index:idx_user_created;not null"`
	SubAccountID     uint      `gorm:"default:0"`
	ProductID        uint      `gorm:"index"`
	SupplierID       uint      `gorm:"index:idx_supplier_created"`
	RequestPath      string    `gorm:"size:128;not null"`
	RequestModel     string    `gorm:"size:64"`
	TokensPrompt     int       `gorm:"default:0"`
	TokensCompletion int       `gorm:"default:0"`
	TokensTotal      int       `gorm:"default:0"`
	ResponseTimeMs   int       `gorm:"default:0"`
	IsStream         int       `gorm:"default:0"`
	StatusCode       int       `gorm:"type:smallint"`
	Status           string    `gorm:"size:32;not null"`
	ErrorMessage     string    `gorm:"type:text"`
	ClientIP         string    `gorm:"size:45"`
	UserAgent        string    `gorm:"size:256"`
	CreatedAt        time.Time `gorm:"index"`
}

func (CallLog) TableName() string { return "call_log" }
```

- [ ] **Step 3: Verify compilation**

Run: `go build ./internal/shared/model/`
Expected: no errors.

---

### Task 6: Risk, notify, commission, audit_log models

**Files:**
- Create: `internal/shared/model/risk.go`
- Create: `internal/shared/model/notify.go`
- Create: `internal/shared/model/commission.go`
- Create: `internal/shared/model/audit_log.go`

- [ ] **Step 1: Write risk.go** (RiskEvent, RiskRule)

```go
package model

import "time"

type RiskEvent struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"index;not null"`
	EventType   string    `gorm:"size:64;not null"`
	RiskLevel   string    `gorm:"size:16;not null"`
	Description string    `gorm:"type:text"`
	RuleID      uint      `gorm:"default:0"`
	RelatedInfo string    `gorm:"type:text"`
	ActionTaken string    `gorm:"size:64"`
	Status      string    `gorm:"default:pending;size:32"`
	HandlerID   *uint     `gorm:"index"`
	HandledAt   *time.Time`
	CreatedAt   time.Time`
}

type RiskRule struct {
	ID         uint      `gorm:"primaryKey"`
	Name       string    `gorm:"size:128;not null"`
	Category   string    `gorm:"size:32;not null"`
	Conditions string    `gorm:"type:text;not null"`
	Action     string    `gorm:"size:32;not null"`
	RiskLevel  string    `gorm:"size:16;not null"`
	Priority   int       `gorm:"default:0"`
	Enabled    int       `gorm:"default:1"`
	CreatedAt  time.Time`
	UpdatedAt  time.Time`
}
```

- [ ] **Step 2: Write notify.go** (Notification, NotificationTemplate)

```go
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
	ReadAt    *time.Time`
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
	CreatedAt time.Time`
	UpdatedAt time.Time`
}
```

- [ ] **Step 3: Write commission.go**

```go
package model

import "time"

type Commission struct {
	ID               uint      `gorm:"primaryKey"`
	AgentID          uint      `gorm:"index;not null"`
	CustomerID       uint      `gorm:"index;not null"`
	OrderID          uint      `gorm:"index;not null"`
	OrderAmount      string    `gorm:"size:32;not null"`
	CommissionRate   string    `gorm:"size:16;not null"`
	CommissionAmount string    `gorm:"size:32;not null"`
	Status           string    `gorm:"default:pending;size:32"`
	SettledAt        *time.Time`
	CreatedAt        time.Time`
}

func (Commission) TableName() string { return "commission" }
```

- [ ] **Step 4: Write audit_log.go**

```go
package model

import "time"

type AuditLog struct {
	ID           uint      `gorm:"primaryKey"`
	TraceID      string    `gorm:"index;size:64;not null"`
	OperatorID   *uint     `gorm:"index"`
	OperatorName string    `gorm:"size:64"`
	OperatorIP   string    `gorm:"size:45"`
	Action       string    `gorm:"size:64;index;not null"`
	ResourceType string    `gorm:"size:64"`
	ResourceID   string    `gorm:"size:64"`
	Detail       string    `gorm:"type:text"`
	Result       string    `gorm:"size:16;not null"`
	FailReason   string    `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"index"`
}

func (AuditLog) TableName() string { return "audit_log" }
```

- [ ] **Step 5: Verify compilation**

Run: `go build ./internal/shared/model/`
Expected: no errors.

---

### Task 7: Vendor/i18n/extended models

**Files:**
- Create: `internal/shared/model/vendor.go`
- Create: `internal/shared/model/guardrail.go`
- Create: `internal/shared/model/byok.go`
- Create: `internal/shared/model/i18n.go`
- Create: `internal/shared/model/cost.go`
- Create: `internal/shared/model/health.go`

- [ ] **Step 1: Write vendor.go** (SupplierVendor, SupplierProduct, Settlement)

```go
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
	ApprovedAt       *time.Time`
	CreatedAt        time.Time`
	UpdatedAt        time.Time`
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
	CreatedAt      time.Time`
	UpdatedAt      time.Time`
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
	PaidAt          *time.Time`
	Remark          string    `gorm:"type:text"`
	CreatedAt       time.Time`
	UpdatedAt       time.Time`
}
```

- [ ] **Step 2: Write guardrail.go**

```go
package model

type GuardrailRule struct {
	ID         uint   `gorm:"primaryKey"`
	Name       string `gorm:"size:128;not null"`
	Stage      string `gorm:"size:16;not null"`   // before | after
	Type       string `gorm:"size:32;not null"`    // pii | injection | secret | content
	Action     string `gorm:"size:16;not null"`    // enforce | monitor | log
	Conditions string `gorm:"type:text"`
	Priority   int    `gorm:"default:0"`
	Enabled    int    `gorm:"default:1"`
	CreatedAt  int64  `gorm:"autoCreateTime"`
}

type GuardrailLog struct {
	ID               uint   `gorm:"primaryKey"`
	TraceID          string `gorm:"index;size:64;not null"`
	UserID           uint   `gorm:"index"`
	RuleID           uint   `gorm:"index"`
	Stage            string `gorm:"size:16;not null"`
	DetectedEntities string `gorm:"type:text"`
	ActionTaken      string `gorm:"size:16;not null"`
	CreatedAt        int64  `gorm:"autoCreateTime;index"`
}
```

- [ ] **Step 3: Write byok.go**

```go
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
```

- [ ] **Step 4: Write i18n.go** (SupportedLanguage, ModelVariant)

```go
package model

type SupportedLanguage struct {
	ID             uint   `gorm:"primaryKey"`
	Locale         string `gorm:"uniqueIndex;size:16;not null"`
	Name           string `gorm:"size:64;not null"`
	NativeName     string `gorm:"size:64;not null"`
	IsEnabled      int    `gorm:"default:1"`
	SortOrder      int    `gorm:"default:0"`
	IsDefault      int    `gorm:"default:0"`
	FallbackLocale string `gorm:"default:en;size:16"`
}

type ModelVariant struct {
	ID               uint    `gorm:"primaryKey"`
	BaseModel        string  `gorm:"size:64;not null"`
	Suffix           string  `gorm:"size:32;not null"`
	ProviderID       uint    `gorm:"index"`
	PriceCoefficient float64 `gorm:"default:1.0"`
	Priority         int     `gorm:"default:0"`
}
```

- [ ] **Step 5: Write cost.go** (SemanticCache)

```go
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
```

- [ ] **Step 6: Write health.go** (ProviderHealth)

```go
package model

type ProviderHealth struct {
	ID           uint   `gorm:"primaryKey"`
	ProviderID   uint   `gorm:"index:idx_health_provider;not null"`
	Status       int    `gorm:"not null"`
	AvgLatencyMs int    `gorm:"default:0"`
	ErrorRate    float64`gorm:"default:0"`
	CheckCount   int    `gorm:"default:0"`
	PeriodStart  int64  `gorm:"not null"`
	PeriodEnd    int64  `gorm:"not null"`
}
```

- [ ] **Step 7: Verify compilation**

Run: `go build ./internal/shared/model/`
Expected: no errors.

---

### Task 8: DB initialization + AutoMigrate

**Files:**
- Modify: `internal/shared/model/init.go` (Create)

- [ ] **Step 1: Write model/init.go**

```go
package model

import (
	"fmt"
	"time"

	"github.com/fastax/fastax-server/internal/shared/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./internal/shared/model/`
Expected: no errors.

---

### Task 9: Redis cache wrapper

**Files:**
- Create: `internal/shared/cache/redis.go`

- [ ] **Step 1: Write redis.go**

```go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fastax/fastax-server/internal/shared/config"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(cfg config.RedisConfig) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &RedisClient{client: client}, nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Get returns string value or empty string if key not found
func (r *RedisClient) Get(key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Set stores a string value with TTL
func (r *RedisClient) Set(key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// GetJSON unmarshals JSON from cache into target
func (r *RedisClient) GetJSON(key string, target interface{}) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil // cache miss, not an error
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// SetJSON marshals value to JSON and stores with TTL
func (r *RedisClient) SetJSON(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

// Exists checks if key exists
func (r *RedisClient) Exists(key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	return n > 0, err
}

// Delete removes one or more keys
func (r *RedisClient) Delete(keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Incr increments a counter and returns the new value
func (r *RedisClient) Incr(key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// Expire sets TTL on a key
func (r *RedisClient) Expire(key string, ttl time.Duration) error {
	return r.client.Expire(ctx, key, ttl).Err()
}

// HGetAll returns all fields of a hash
func (r *RedisClient) HGetAll(key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// HSet sets fields in a hash
func (r *RedisClient) HSet(key string, fields map[string]interface{}) error {
	return r.client.HSet(ctx, key, fields).Err()
}

// CacheKey helpers for the key naming convention from PDD
func UserSessionKey(userID uint) string        { return fmt.Sprintf("user:session:%d", userID) }
func UserQuotaKey(userID uint) string          { return fmt.Sprintf("user:quota:%d", userID) }
func TokenProductKey(id uint) string           { return fmt.Sprintf("token:product:%d", id) }
func SupplierKey(id uint) string               { return fmt.Sprintf("token:supplier:%d", id) }
func RouteHealthKey(id uint) string            { return fmt.Sprintf("route:health:%d", id) }
func RateLimitKey(key string) string           { return fmt.Sprintf("rate:limit:%s", key) }
func VerifyCodeKey(identifier string) string   { return fmt.Sprintf("verify:code:%s", identifier) }
func RefreshTokenKey(token string) string      { return fmt.Sprintf("refresh:token:%s", token) }
func I18nKey(locale, ns string) string         { return fmt.Sprintf("i18n:translations:%s:%s", locale, ns) }
func SystemConfigKey(key string) string        { return fmt.Sprintf("config:system:%s", key) }
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./internal/shared/cache/`
Expected: no errors.

---

### Task 10: Auth + CORS middleware

**Files:**
- Create: `internal/shared/middleware/auth.go`
- Create: `internal/shared/middleware/cors.go`

- [ ] **Step 1: Write auth.go** (JWT middleware + RBAC)

```go
package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// AuthRequired validates JWT and sets user info in context
func AuthRequired(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			response.Error(c, http.StatusUnauthorized, response.CodeTokenExpired, "missing authorization token")
			return
		}

		claims, err := parseJWT(tokenStr, secret)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, response.CodeTokenExpired, "invalid or expired token")
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// AdminRequired ensures the user has admin role
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" && role != "super_admin" {
			response.Error(c, http.StatusForbidden, response.CodePermissionDeny)
			return
		}
		c.Next()
	}
}

// RoleRequired checks for specific role(s)
func RoleRequired(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}
		response.Error(c, http.StatusForbidden, response.CodePermissionDeny)
	}
}

func extractToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}
	return ""
}

func GenerateAccessToken(userID uint, role, secret string, expiry time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "fastax",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(userID uint, secret string, expiry time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "fastax-refresh",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(fmt.Sprintf("%s-refresh", secret)))
}

func parseJWT(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}
```

- [ ] **Step 2: Write cors.go**

```go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept-Language, X-Trace-Id")
		c.Header("Access-Control-Expose-Headers", "X-Trace-Id")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
```

- [ ] **Step 3: Verify compilation**

Run: `go build ./internal/shared/middleware/`
Expected: no errors.

---

### Task 11: Rate limit + Language middleware

**Files:**
- Create: `internal/shared/middleware/rate_limit.go`
- Create: `internal/shared/middleware/language.go`

- [ ] **Step 1: Write rate_limit.go**

```go
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type rateEntry struct {
	count    int
	resetAt  time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	entries  map[string]*rateEntry
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		entries: make(map[string]*rateEntry),
		limit:   limit,
		window:  window,
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[key]

	if !exists || now.After(entry.resetAt) {
		rl.entries[key] = &rateEntry{
			count:   1,
			resetAt: now.Add(rl.window),
		}
		return true
	}

	if entry.count >= rl.limit {
		return false
	}

	entry.count++
	return true
}

// RateLimitIP limits by client IP (60 req/min default)
func RateLimitIP(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.Allow(ip) {
			response.Error(c, http.StatusTooManyRequests, response.CodeRateLimited)
			return
		}
		c.Next()
	}
}

// RateLimitAuth limits auth endpoints (5 req/min)
func RateLimitAuth(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "auth:" + c.ClientIP()
		if !limiter.Allow(key) {
			response.Error(c, http.StatusTooManyRequests, response.CodeRateLimited)
			return
		}
		c.Next()
	}
}
```

- [ ] **Step 2: Write language.go**

```go
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// DetectLanguage reads Accept-Language header and sets it in context
// Falls back to zh-CN if no acceptable language found
func DetectLanguage() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("Accept-Language")
		lang = normalizeLanguage(lang)

		if lang == "" {
			lang = "zh-CN"
		}
		c.Set("lang", lang)
		c.Request.Header.Set("Accept-Language", lang)
		c.Next()
	}
}

func GetLanguage(c *gin.Context) string {
	lang, _ := c.Get("lang")
	return lang.(string)
}

// normalizeLanguage extracts the primary language tag from Accept-Language header
// and maps it to a supported locale code
func normalizeLanguage(header string) string {
	if header == "" {
		return ""
	}

	// Accept-Language can be: zh-CN,zh;q=0.9,en;q=0.8
	// Take the first tag, strip quality value
	parts := strings.Split(header, ",")
	lang := strings.TrimSpace(parts[0])
	if idx := strings.Index(lang, ";"); idx > 0 {
		lang = lang[:idx]
	}

	// Supported languages
	supported := map[string]string{
		"zh":   "zh-CN",
		"zh-CN": "zh-CN",
		"zh-TW": "zh-TW",
		"en":   "en",
		"ja":   "ja",
		"ko":   "ko",
	}

	if mapped, ok := supported[lang]; ok {
		return mapped
	}
	// Fallback chain: zh-TW→zh-CN→en
	if strings.HasPrefix(lang, "zh") {
		return "zh-CN"
	}
	return "en"
}
```

- [ ] **Step 3: Verify compilation**

Run: `go build ./internal/shared/middleware/`
Expected: no errors.

---

### Task 12: Router setup + main.go entry point

**Files:**
- Create: `internal/router/router.go`
- Modify: `cmd/fastax/main.go`

- [ ] **Step 1: Write router.go**

```go
package router

import (
	"time"

	"github.com/fastax/fastax-server/internal/shared/config"
	"github.com/fastax/fastax-server/internal/shared/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// Global middleware
	r.Use(middleware.CORS())
	r.Use(middleware.DetectLanguage())

	// Rate limiters
	ipLimiter := middleware.NewRateLimiter(cfg.RateLimit.IP, time.Minute)
	authLimiter := middleware.NewRateLimiter(cfg.RateLimit.Auth, time.Minute)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	api := r.Group("/api")
	{
		// Auth endpoints (no JWT required)
		auth := api.Group("/auth")
		auth.Use(middleware.RateLimitAuth(authLimiter))
		{
			auth.POST("/register", placeholder)
			auth.POST("/login", placeholder)
			auth.POST("/refresh", placeholder)
			auth.POST("/send-code", placeholder)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthRequired(cfg.JWT.Secret))
		{
			tokens := protected.Group("/tokens")
			{
				tokens.GET("/products", placeholder)
				tokens.GET("/my", placeholder)
				tokens.POST("/buy", placeholder)
			}

			orders := protected.Group("/orders")
			{
				orders.GET("", placeholder)
			}
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(middleware.AuthRequired(cfg.JWT.Secret), middleware.AdminRequired())
		{
			admin.GET("/dashboard/summary", placeholder)
			admin.GET("/users", placeholder)
			admin.GET("/suppliers", placeholder)
			admin.GET("/orders", placeholder)
		}
	}

	// OpenAI-compatible relay routes (S3+)
	v1 := r.Group("/v1")
	v1.Use(middleware.RateLimitIP(ipLimiter))
	{
		v1.POST("/chat/completions", placeholder)
		v1.POST("/embeddings", placeholder)
	}
}

func placeholder(c *gin.Context) {
	c.JSON(200, gin.H{"message": "not implemented yet"})
}
```

- [ ] **Step 2: Update cmd/fastax/main.go** (with Redis initialization)

```go
package main

import (
	"log"

	"github.com/fastax/fastax-server/internal/router"
	"github.com/fastax/fastax-server/internal/shared/cache"
	"github.com/fastax/fastax-server/internal/shared/config"
	"github.com/fastax/fastax-server/internal/shared/model"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load config
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := model.InitDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// Initialize Redis (optional, warn if unavailable)
	redisClient, err := cache.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Printf("WARNING: Redis not available (running without cache): %v", err)
		redisClient = nil
	}
	_ = redisClient // available for injection into domain services

	// Create Gin engine
	gin.SetMode(cfg.Server.Mode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Register routes (pass db and cfg for handler injection)
	router.RegisterRoutes(r, db, cfg)

	// Start server
	log.Printf("FastAX server starting on :%d", cfg.Server.Port)
	if err := r.Run(cfg.Server.Addr()); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
```

- [ ] **Step 3: Verify full build**

Run: `go build ./cmd/fastax/`
Expected: binary `fastax.exe` or `fastax` created in current directory.

- [ ] **Step 4: Verify server starts**

Run: `go run ./cmd/fastax/`
Expected: Server starts, visits `http://localhost:8080/health` returns `{"status":"ok"}`.

---

### Task 13: Add .gitignore + verify end-to-end

**Files:**
- Create: `.gitignore`

- [ ] **Step 1: Write .gitignore**

```
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/

# Go
*.test
*.out
vendor/

# Data
data/
*.db
*.db-wal
*.db-shm

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Config (keep example, ignore actual)
config.yaml
!config.example.yaml

# Environment
.env
```

- [ ] **Step 2: Run full build + test**

```bash
go build ./...
go vet ./...
```

Expected: no errors, binary builds.

---

## Self-Review

**Spec coverage check** (against PDD 02-database, PDD 03-api, PDD 04-deploy):

| PDD Requirement | Covered By |
|-----------------|-----------|
| All 25+ database tables | Tasks 4-7 (model/*.go) |
| Redis key naming convention | Task 9 (cache/redis.go Key helpers) |
| Database init + WAL mode | Task 8 (model/init.go) |
| Viper config | Task 2 (config/config.go) |
| Unified API response format | Task 3 (response/response.go) |
| JWT auth + RBAC middleware | Task 10 (middleware/auth.go) |
| Rate limiting (IP/Auth) | Task 11 (middleware/rate_limit.go) |
| Language detection | Task 11 (middleware/language.go) |
| Route scaffolding (all API groups) | Task 12 (router/router.go) |
| CORS | Task 10 (middleware/cors.go) |
| Server entry point | Task 1 + Task 12 |

**Placeholder scan:** No TBD/TODO/fill-in-later in code files. All model fields from PDD are explicitly defined. Placeholder handlers in router.go are intentional — they mark endpoints for future S1+ implementation.

**Type consistency:** Verify: `config.Load` returns `*Config`, `model.InitDB` returns `*gorm.DB`, `cache.NewRedisClient` returns `*RedisClient`, all middleware signatures match Gin's `gin.HandlerFunc` pattern. Consistent.

**Scope check:** This plan covers ONLY S0 (infrastructure). No business logic, no domain services, no handlers with real implementations. Properly scoped for a single plan.
