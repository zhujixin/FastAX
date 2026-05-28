package user

import (
	"errors"
	"fmt"
	"time"

	"github.com/fastax/fastax-server/internal/shared/cache"
	"github.com/fastax/fastax-server/internal/shared/config"
	"github.com/fastax/fastax-server/internal/shared/middleware"
	"github.com/fastax/fastax-server/internal/shared/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	db     *gorm.DB
	cache  *cache.RedisClient
	cfg    *config.JWTConfig
	verify *VerifyService
}

func NewService(db *gorm.DB, redis *cache.RedisClient, cfg *config.JWTConfig) *Service {
	return &Service{
		db:     db,
		cache:  redis,
		cfg:    cfg,
		verify: NewVerifyService(redis),
	}
}

type RegisterRequest struct {
	Username   string `json:"username" binding:"required,min=3,max=32"`
	Password   string `json:"password" binding:"required,min=6,max=64"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	VerifyCode string `json:"verify_code" binding:"required"`
	Language   string `json:"language"`
}

type LoginRequest struct {
	Account  string `json:"account" binding:"required"` // email or phone
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    int64         `json:"expires_in"`
	User         *UserResponse `json:"user"`
}

type UserResponse struct {
	ID                uint   `json:"id"`
	Username          string `json:"username"`
	Email             string `json:"email,omitempty"`
	Phone             string `json:"phone,omitempty"`
	Role              string `json:"role"`
	Level             string `json:"level"`
	PreferredLanguage string `json:"preferred_language"`
}

func (s *Service) Register(req *RegisterRequest) (*LoginResponse, error) {
	// Validate verification code
	identifier := req.Email
	if identifier == "" {
		identifier = req.Phone
	}
	ok, err := s.verify.VerifyCode(identifier, req.VerifyCode)
	if err != nil {
		return nil, fmt.Errorf("verify code error: %w", err)
	}
	if !ok {
		return nil, errors.New("invalid or expired verification code")
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Set language preference
	lang := req.Language
	if lang == "" {
		lang = "zh-CN"
	}

	// Create user
	user := model.User{
		Username:          req.Username,
		PasswordHash:      string(hashed),
		Email:             req.Email,
		Phone:             req.Phone,
		Role:              "user",
		Level:             "normal",
		Status:            1,
		PreferredLanguage: lang,
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return s.generateTokens(&user)
}

func (s *Service) Login(req *LoginRequest) (*LoginResponse, error) {
	// Find user by email or phone
	var user model.User
	err := s.db.Where("email = ? OR phone = ?", req.Account, req.Account).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid account or password")
		}
		return nil, fmt.Errorf("query user: %w", err)
	}

	// Check if account is frozen
	if user.Status == 0 {
		return nil, errors.New("account is frozen")
	}

	// Check if account is locked
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return nil, fmt.Errorf("account locked until %s", user.LockedUntil.Format(time.RFC3339))
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		// Increment fail count
		user.LoginFailCount++
		if user.LoginFailCount >= 5 {
			until := time.Now().Add(15 * time.Minute)
			user.LockedUntil = &until
		}
		s.db.Save(&user)
		return nil, errors.New("invalid account or password")
	}

	// Reset fail count on success
	if user.LoginFailCount > 0 {
		user.LoginFailCount = 0
		user.LockedUntil = nil
	}
	s.db.Save(&user)

	return s.generateTokens(&user)
}

func (s *Service) RefreshToken(refreshToken string) (*LoginResponse, error) {
	// Verify refresh token in Redis
	key := cache.RefreshTokenKey(refreshToken)
	userIDStr, err := s.cache.Get(key)
	if err != nil || userIDStr == "" {
		return nil, errors.New("invalid or expired refresh token")
	}

	// Delete old refresh token (rotation)
	s.cache.Delete(key)

	// Get user
	var user model.User
	if err := s.db.First(&user, userIDStr).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if user.Status == 0 {
		return nil, errors.New("account is frozen")
	}

	return s.generateTokens(&user)
}

func (s *Service) GetUser(userID uint) (*UserResponse, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return toUserResponse(&user), nil
}

func (s *Service) UpdateLanguage(userID uint, language string) error {
	return s.db.Model(&model.User{}).Where("id = ?", userID).
		Update("preferred_language", language).Error
}

func (s *Service) generateTokens(user *model.User) (*LoginResponse, error) {
	accessToken, err := middleware.GenerateAccessToken(user.ID, user.Role, s.cfg.Secret, s.cfg.AccessExpiry)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := middleware.GenerateRefreshToken(user.ID, s.cfg.Secret, s.cfg.RefreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	// Store refresh token in Redis
	key := cache.RefreshTokenKey(refreshToken)
	s.cache.Set(key, fmt.Sprintf("%d", user.ID), s.cfg.RefreshExpiry)

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.cfg.AccessExpiry.Seconds()),
		User:         toUserResponse(user),
	}, nil
}

func toUserResponse(u *model.User) *UserResponse {
	return &UserResponse{
		ID:                u.ID,
		Username:          u.Username,
		Email:             u.Email,
		Phone:             u.Phone,
		Role:              u.Role,
		Level:             u.Level,
		PreferredLanguage: u.PreferredLanguage,
	}
}
