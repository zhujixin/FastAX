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
func BlacklistKey(userID uint) string          { return fmt.Sprintf("user:blacklist:%d", userID) }
func I18nKey(locale, ns string) string         { return fmt.Sprintf("i18n:translations:%s:%s", locale, ns) }
func SystemConfigKey(key string) string        { return fmt.Sprintf("config:system:%s", key) }
