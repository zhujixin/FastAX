package byok

import (
	"errors"
	"fmt"
	"time"

	"github.com/fastax/fastax-server/internal/shared/model"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

type AddKeyRequest struct {
	Provider       string `json:"provider" binding:"required"`
	KeyEncrypted   string `json:"key_encrypted" binding:"required"`
	KeyIV          string `json:"key_iv" binding:"required"`
	Alias          string `json:"alias"`
	ModelWhitelist string `json:"model_whitelist"`
}

type KeyResponse struct {
	ID             uint   `json:"id"`
	Provider       string `json:"provider"`
	Alias          string `json:"alias"`
	ModelWhitelist string `json:"model_whitelist"`
	Status         int    `json:"status"`
	LastUsedAt     int64  `json:"last_used_at"`
	CreatedAt      int64  `json:"created_at"`
}

func (s *Service) AddKey(userID uint, req *AddKeyRequest) (*KeyResponse, error) {
	key := model.BYOKKey{
		UserID:         userID,
		Provider:       req.Provider,
		KeyEncrypted:   req.KeyEncrypted,
		KeyIV:          req.KeyIV,
		Alias:          req.Alias,
		ModelWhitelist: req.ModelWhitelist,
		Status:         1,
		CreatedAt:      time.Now().Unix(),
	}
	if err := s.db.Create(&key).Error; err != nil {
		return nil, fmt.Errorf("add key: %w", err)
	}
	return toKeyResponse(&key), nil
}

func (s *Service) GetKey(id uint, userID uint) (*KeyResponse, error) {
	var key model.BYOKKey
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("key not found")
		}
		return nil, err
	}
	return toKeyResponse(&key), nil
}

func (s *Service) ListKeys(userID uint) ([]KeyResponse, error) {
	var keys []model.BYOKKey
	if err := s.db.Where("user_id = ?", userID).Order("created_at desc").Find(&keys).Error; err != nil {
		return nil, err
	}
	result := make([]KeyResponse, len(keys))
	for i, k := range keys {
		result[i] = *toKeyResponse(&k)
	}
	return result, nil
}

func (s *Service) SetKeyStatus(id, userID uint, status int) error {
	result := s.db.Model(&model.BYOKKey{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("key not found")
	}
	return nil
}

func (s *Service) DeleteKey(id, userID uint) error {
	result := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.BYOKKey{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("key not found")
	}
	return nil
}

// FindKeyForModel finds a BYOK key that matches the provider and model
func (s *Service) FindKeyForModel(userID uint, provider, modelName string) (*model.BYOKKey, error) {
	var keys []model.BYOKKey
	if err := s.db.Where("user_id = ? AND provider = ? AND status = 1", userID, provider).Find(&keys).Error; err != nil {
		return nil, err
	}

	for _, key := range keys {
		if key.ModelWhitelist == "" {
			return &key, nil
		}
		if containsModel(key.ModelWhitelist, modelName) {
			return &key, nil
		}
	}
	return nil, errors.New("no matching BYOK key")
}

func (s *Service) TouchKey(id uint) {
	s.db.Model(&model.BYOKKey{}).Where("id = ?", id).
		Update("last_used_at", time.Now().Unix())
}

func containsModel(whitelist, model string) bool {
	// Simple comma-separated check
	start := 0
	for i := 0; i <= len(whitelist); i++ {
		if i == len(whitelist) || whitelist[i] == ',' {
			if start < i && whitelist[start:i] == model {
				return true
			}
			start = i + 1
		}
	}
	return false
}

func toKeyResponse(k *model.BYOKKey) *KeyResponse {
	return &KeyResponse{
		ID:             k.ID,
		Provider:       k.Provider,
		Alias:          k.Alias,
		ModelWhitelist: k.ModelWhitelist,
		Status:         k.Status,
		LastUsedAt:     k.LastUsedAt,
		CreatedAt:      k.CreatedAt,
	}
}
