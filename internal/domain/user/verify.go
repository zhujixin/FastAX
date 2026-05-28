package user

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/fastax/fastax-server/internal/shared/cache"
)

const (
	verifyCodeTTL    = 5 * time.Minute
	verifyCodeLength = 6
)

type VerifyService struct {
	cache *cache.RedisClient
}

func NewVerifyService(redis *cache.RedisClient) *VerifyService {
	return &VerifyService{cache: redis}
}

func (s *VerifyService) GenerateCode(identifier string) (string, error) {
	code, err := generateRandomCode(verifyCodeLength)
	if err != nil {
		return "", fmt.Errorf("generate code: %w", err)
	}
	if s.cache == nil {
		return code, nil // dev mode: no Redis, skip storing
	}
	key := cache.VerifyCodeKey(identifier)
	if err := s.cache.Set(key, code, verifyCodeTTL); err != nil {
		return "", fmt.Errorf("store code: %w", err)
	}
	return code, nil
}

func (s *VerifyService) VerifyCode(identifier, code string) (bool, error) {
	if s.cache == nil {
		return true, nil // dev mode: no Redis, accept any code
	}
	key := cache.VerifyCodeKey(identifier)
	stored, err := s.cache.Get(key)
	if err != nil || stored == "" {
		return false, nil // code not found or expired
	}
	if stored != code {
		return false, nil
	}
	// Delete after successful verification (one-time use)
	_ = s.cache.Delete(key)
	return true, nil
}

func generateRandomCode(length int) (string, error) {
	code := make([]byte, length)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code[i] = byte('0') + byte(n.Int64())
	}
	return string(code), nil
}
