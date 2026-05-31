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
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=64"`
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

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	lang := req.Language
	if lang == "" {
		lang = "zh-CN"
	}

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
	var user model.User
	err := s.db.Where("email = ? OR phone = ?", req.Account, req.Account).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid account or password")
		}
		return nil, fmt.Errorf("query user: %w", err)
	}

	if user.Status == 0 {
		return nil, errors.New("account is frozen")
	}

	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return nil, fmt.Errorf("account locked until %s", user.LockedUntil.Format(time.RFC3339))
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		user.LoginFailCount++
		if user.LoginFailCount >= 5 {
			until := time.Now().Add(15 * time.Minute)
			user.LockedUntil = &until
		}
		s.db.Save(&user)
		return nil, errors.New("invalid account or password")
	}

	if user.LoginFailCount > 0 {
		user.LoginFailCount = 0
		user.LockedUntil = nil
	}
	s.db.Save(&user)

	return s.generateTokens(&user)
}

func (s *Service) RefreshToken(refreshToken string) (*LoginResponse, error) {
	if s.cache == nil {
		return nil, errors.New("refresh not available without Redis")
	}
	key := cache.RefreshTokenKey(refreshToken)
	userIDStr, err := s.cache.Get(key)
	if err != nil || userIDStr == "" {
		return nil, errors.New("invalid or expired refresh token")
	}
	s.cache.Delete(key)

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

func (s *Service) Logout(userID uint, refreshToken string) error {
	if s.cache == nil {
		return nil // dev mode: no Redis, skip
	}

	// Delete refresh token if provided
	if refreshToken != "" {
		key := cache.RefreshTokenKey(refreshToken)
		_ = s.cache.Delete(key)
	}

	// Blacklist the user's access tokens until they expire
	blacklistKey := cache.BlacklistKey(userID)
	return s.cache.Set(blacklistKey, time.Now().Unix(), s.cfg.AccessExpiry)
}

func (s *Service) ResetPassword(req *ResetPasswordRequest) error {
	// Find user by email
	var user model.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("query user: %w", err)
	}

	// Verify the code
	ok, err := s.verify.VerifyCode(req.Email, req.Code)
	if err != nil {
		return fmt.Errorf("verify code error: %w", err)
	}
	if !ok {
		return errors.New("invalid or expired verification code")
	}

	// Hash the new password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// Update password in DB
	if err := s.db.Model(&model.User{}).Where("id = ?", user.ID).
		Update("password_hash", string(hashed)).Error; err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
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

	if s.cache != nil {
		key := cache.RefreshTokenKey(refreshToken)
		s.cache.Set(key, fmt.Sprintf("%d", user.ID), s.cfg.RefreshExpiry)
	}

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

// ---------- Admin methods ----------

type UserListResponse struct {
	Items    []UserResponse `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// ListUsers returns paginated user list with optional keyword search.
func (s *Service) ListUsers(page, pageSize int, keyword string) (*UserListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := s.db.Model(&model.User{})
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("username LIKE ? OR email LIKE ? OR phone LIKE ?", like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	var users []model.User
	if err := query.Order("id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}

	items := make([]UserResponse, len(users))
	for i, u := range users {
		items[i] = *toUserResponse(&u)
	}

	return &UserListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// SetUserStatus updates user status (0=frozen, 1=active).
func (s *Service) SetUserStatus(id uint, status int) error {
	// Prevent freezing super admin
	var user model.User
	if err := s.db.First(&user, id).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	if user.Role == "super_admin" && status == 0 {
		return errors.New("cannot freeze super admin account")
	}

	result := s.db.Model(&model.User{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("update status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// SetUserLevel updates user level.
func (s *Service) SetUserLevel(id uint, level string) error {
	validLevels := map[string]bool{"normal": true, "vip": true, "enterprise": true}
	if !validLevels[level] {
		return fmt.Errorf("invalid level: %s, must be normal/vip/enterprise", level)
	}

	result := s.db.Model(&model.User{}).Where("id = ?", id).Update("level", level)
	if result.Error != nil {
		return fmt.Errorf("update level: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// ---------- User Detail ----------

type UserDetailResponse struct {
	UserResponse
	TokenBalance  string `json:"token_balance"`
	TotalOrders   int64  `json:"total_orders"`
	TotalSpent    string `json:"total_spent"`
	LastLoginAt   string `json:"last_login_at,omitempty"`
	CreatedAt     string `json:"created_at"`
}

// GetUserDetail returns detailed user information for admin.
func (s *Service) GetUserDetail(id uint) (*UserDetailResponse, error) {
	var user model.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("query user: %w", err)
	}

	// Get token balance
	var tokenBalance float64
	s.db.Model(&model.UserToken{}).
		Where("user_id = ? AND status = ?", id, 1).
		Select("COALESCE(SUM(CAST(total_amount AS REAL)), 0) - COALESCE(SUM(CAST(used_amount AS REAL)), 0)").
		Row().Scan(&tokenBalance)

	// Get order stats
	var totalOrders int64
	s.db.Model(&model.Order{}).Where("user_id = ?", id).Count(&totalOrders)

	var totalSpent float64
	s.db.Model(&model.Payment{}).
		Joins("JOIN orders ON orders.id = payments.order_id").
		Where("orders.user_id = ? AND payments.status = ?", id, "success").
		Select("COALESCE(SUM(CAST(payments.amount AS REAL)), 0)").
		Row().Scan(&totalSpent)

	// Format last login
	lastLogin := ""
	if user.LastLoginAt != nil {
		lastLogin = user.LastLoginAt.Format("2006-01-02 15:04:05")
	}

	return &UserDetailResponse{
		UserResponse:  *toUserResponse(&user),
		TokenBalance:  fmt.Sprintf("%.2f", tokenBalance),
		TotalOrders:   totalOrders,
		TotalSpent:    fmt.Sprintf("%.2f", totalSpent),
		LastLoginAt:   lastLogin,
		CreatedAt:     user.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// ---------- OAuth ----------

type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string
}

type OAuthUserInfo struct {
	Provider    string
	ProviderID  string
	Email       string
	Name        string
	Avatar      string
}

type OAuthLoginRequest struct {
	Provider string `json:"provider" binding:"required,oneof=google github"`
	Code     string `json:"code" binding:"required"`
}

// OAuthLogin handles OAuth login flow.
func (s *Service) OAuthLogin(req *OAuthLoginRequest) (*LoginResponse, error) {
	// In production, exchange code for token and get user info from provider
	// For MVP, simulate OAuth user info
	userInfo := &OAuthUserInfo{
		Provider:   req.Provider,
		ProviderID: fmt.Sprintf("%s_%s", req.Provider, req.Code),
		Email:      fmt.Sprintf("%s_user@example.com", req.Provider),
		Name:       fmt.Sprintf("%s User", req.Provider),
	}

	// Find existing user by OAuth provider info
	var user model.User
	err := s.db.Where("email = ?", userInfo.Email).First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new user
		user = model.User{
			Username:          userInfo.Name,
			Email:             userInfo.Email,
			Role:              "user",
			Level:             "normal",
			Status:            1,
			PreferredLanguage: "en",
		}
		if err := s.db.Create(&user).Error; err != nil {
			return nil, fmt.Errorf("create oauth user: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}

	// Check if user is frozen
	if user.Status == 0 {
		return nil, errors.New("account is frozen")
	}

	// Generate tokens
	return s.generateTokens(&user)
}

// GetOAuthRedirectURL returns the OAuth redirect URL for a provider.
func (s *Service) GetOAuthRedirectURL(provider, callbackURL string) (string, error) {
	switch provider {
	case "google":
		return fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?client_id=GOOGLE_CLIENT_ID&redirect_uri=%s&response_type=code&scope=email%%20profile", callbackURL), nil
	case "github":
		return fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=GITHUB_CLIENT_ID&redirect_uri=%s&scope=user:email", callbackURL), nil
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}
}
