package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	Guardrail  GuardrailConfig  `mapstructure:"guardrail"`
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

type GuardrailConfig struct {
	Mode string `mapstructure:"mode"` // "enforce" or "monitor"
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
	v.SetDefault("guardrail.mode", "monitor")

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
