// 配置管理
//
// 本文件负责应用程序配置的加载和管理：
//   - Config: 主配置结构体，包含所有配置项
//   - ServerConfig: 服务器配置（端口、模式、超时等）
//   - DatabaseConfig: 数据库配置（SQLite 路径、日志级别等）
//   - RedisConfig: Redis 配置（地址、密码、数据库等）
//   - JWTConfig: JWT 认证配置（密钥、过期时间等）
//   - RateLimitConfig: 限流配置（IP 限流、认证限流等）
//   - GuardrailConfig: 安全护栏配置（模式、规则等）
//   - Load: 从配置文件加载配置（支持 YAML/JSON/TOML）
package config

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Config 主配置结构体，包含所有配置项
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	Guardrail  GuardrailConfig  `mapstructure:"guardrail"`
	OAuth    OAuthConfig    `mapstructure:"oauth"`
	Security SecurityConfig `mapstructure:"security"`
}

type ServerConfig struct {
	Port           int           `mapstructure:"port"`
	Mode           string        `mapstructure:"mode"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	MaxBodySize     int           `mapstructure:"max_body_size"` // MB
	AllowedOrigins  []string      `mapstructure:"allowed_origins"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
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
	RefreshSecret string        `mapstructure:"refresh_secret"`
	AccessExpiry  time.Duration `mapstructure:"access_expiry"`
	RefreshExpiry time.Duration `mapstructure:"refresh_expiry"`
}

type GuardrailConfig struct {
	Mode string `mapstructure:"mode"` // "enforce" or "monitor"
}

type OAuthConfig struct {
	Google GoogleOAuthConfig `mapstructure:"google"`
	GitHub GitHubOAuthConfig `mapstructure:"github"`
}

type GoogleOAuthConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

type GitHubOAuthConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

type SecurityConfig struct {
	EncryptionKey string `mapstructure:"encryption_key"` // AES-256 master key (base64-encoded or hex)
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
	v.SetDefault("server.mode", "release")
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "60s")
	v.SetDefault("server.idle_timeout", "120s")
	v.SetDefault("server.max_body_size", 8) // MB
	v.SetDefault("server.shutdown_timeout", "10s")
	v.SetDefault("database.path", "data/fastax.db")
	v.SetDefault("database.wal_mode", true)
	v.SetDefault("database.log_level", "warn")
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("jwt.access_expiry", "24h")
	v.SetDefault("jwt.refresh_expiry", "168h")
	v.SetDefault("jwt.secret", "")
	v.SetDefault("jwt.refresh_secret", "")
	v.SetDefault("rate_limit.ip", 60)
	v.SetDefault("rate_limit.auth", 5)
	v.SetDefault("rate_limit.user_default", 60)
	v.SetDefault("rate_limit.user_enterprise", 300)
	v.SetDefault("guardrail.mode", "monitor")
	v.SetDefault("oauth.google.client_id", "")
	v.SetDefault("oauth.google.client_secret", "")
	v.SetDefault("oauth.github.client_id", "")
	v.SetDefault("oauth.github.client_secret", "")
	v.SetDefault("security.encryption_key", "")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg, viper.DecodeHook(mapstructure.StringToTimeDurationHookFunc())); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.JWT.Secret == "" {
		return nil, errors.New("jwt.secret is required: set it in config.yaml or via FASTAX_JWT_SECRET env var")
	}

	return &cfg, nil
}

// GetEncryptionKey decodes the configured encryption key from hex or base64.
// Returns nil if not configured (API keys stored in plaintext for backward compat).
func (c *Config) GetEncryptionKey() []byte {
	if c.Security.EncryptionKey == "" {
		return nil
	}
	// Try hex first
	key, err := hex.DecodeString(c.Security.EncryptionKey)
	if err == nil && len(key) == 32 {
		return key
	}
	// Fall back to standard base64
	key, err = base64.StdEncoding.DecodeString(c.Security.EncryptionKey)
	if err == nil && len(key) == 32 {
		return key
	}
	// Try URL-safe base64
	key, err = base64.URLEncoding.DecodeString(c.Security.EncryptionKey)
	if err == nil && len(key) == 32 {
		return key
	}
	return nil
}
