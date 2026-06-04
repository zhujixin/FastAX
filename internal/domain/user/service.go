// 用户模块业务逻辑层
//
// 本文件实现了用户相关的业务逻辑：
//   - Service: 用户服务结构体
//   - Register: 用户注册（密码加密存储）
//   - Login: 用户登录（密码验证、JWT 生成）
//   - RefreshToken: 刷新 Access Token
//   - GetUser: 获取用户信息
//   - UpdateLanguage: 更新用户语言偏好
//   - OAuthLogin: OAuth 第三方登录
//   - 管理员功能: ListUsers、GetUserDetail、SetUserStatus、SetUserLevel
package user

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/fastax/fastax-server/internal/shared/cache"
	"github.com/fastax/fastax-server/internal/shared/config"
	"github.com/fastax/fastax-server/internal/shared/constants"
	"github.com/fastax/fastax-server/internal/shared/middleware"
	"github.com/fastax/fastax-server/internal/shared/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ─── 安全常量 ───

const (
	BcryptCost        = 12
	MaxLoginAttempts  = 5
	LockoutDuration   = 15 * time.Minute
)

// Service 用户服务结构体
type Service struct {
	db       *gorm.DB
	cache    *cache.RedisClient
	cfg      *config.JWTConfig
	verify   *VerifyService
	oauthCfg *OAuthConfig
}

// NewService 创建用户服务实例
func NewService(db *gorm.DB, redis *cache.RedisClient, cfg *config.Config) *Service {
	return &Service{
		db:    db,
		cache: redis,
		cfg:   &cfg.JWT,
		verify: NewVerifyService(redis),
		oauthCfg: &OAuthConfig{
			GoogleClientID:     cfg.OAuth.Google.ClientID,
			GoogleClientSecret: cfg.OAuth.Google.ClientSecret,
			GitHubClientID:     cfg.OAuth.GitHub.ClientID,
			GitHubClientSecret: cfg.OAuth.GitHub.ClientSecret,
		},
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

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), BcryptCost)
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
		Role:              constants.RoleUser,
		Level:             constants.LevelNormal,
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
		// Atomic update to prevent race condition on login failure counter
		s.db.Model(&model.User{}).Where("id = ?", user.ID).
			UpdateColumn("login_fail_count", gorm.Expr("login_fail_count + 1"))
		user.LoginFailCount++
		if user.LoginFailCount >= MaxLoginAttempts {
			until := time.Now().Add(LockoutDuration)
			s.db.Model(&user).Update("locked_until", until)
		}
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

	var user model.User
	if err := s.db.First(&user, userIDStr).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	if user.Status == 0 {
		return nil, errors.New("account is frozen")
	}

	// Generate new tokens first, then delete old refresh token.
	// This prevents permanent lockout if token generation fails mid-way.
	resp, err := s.generateTokens(&user)
	if err != nil {
		return nil, err
	}
	s.cache.Delete(key)

	return resp, nil
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
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
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

// effectiveRefreshSecret returns the refresh token signing key.
// If jwt.refresh_secret is configured, it is used directly for cryptographic independence.
// Otherwise, falls back to secret+"-refresh" for backward compatibility.
func (s *Service) effectiveRefreshSecret() string {
	if s.cfg.RefreshSecret != "" {
		return s.cfg.RefreshSecret
	}
	return s.cfg.Secret + "-refresh"
}

func (s *Service) generateTokens(user *model.User) (*LoginResponse, error) {
	accessToken, err := middleware.GenerateAccessToken(user.ID, user.Role, s.cfg.Secret, s.cfg.AccessExpiry)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := middleware.GenerateRefreshToken(user.ID, s.effectiveRefreshSecret(), s.cfg.RefreshExpiry)
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
	if user.Role == constants.RoleSuperAdmin && status == 0 {
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
	validLevels := map[string]bool{constants.LevelNormal: true, constants.LevelVIP: true, constants.LevelEnterprise: true}
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

// --- Admin Account Management ---

// ListAdmins returns all users with admin or super_admin roles.
func (s *Service) ListAdmins() ([]UserResponse, error) {
	var users []model.User
	if err := s.db.Where("role IN ?", []string{constants.RoleAdmin, constants.RoleSuperAdmin}).Order("created_at desc").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list admins: %w", err)
	}
	result := make([]UserResponse, len(users))
	for i, u := range users {
		result[i] = *toUserResponse(&u)
	}
	return result, nil
}

// CreateAdmin creates a new admin account.
func (s *Service) CreateAdmin(username, password, email, role string) (*UserResponse, error) {
	if role != constants.RoleAdmin && role != constants.RoleSuperAdmin {
		return nil, fmt.Errorf("invalid role: %s", role)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	user := model.User{
		Username:     username,
		PasswordHash: string(hash),
		Email:        email,
		Role:         role,
		Level:        constants.LevelNormal,
		Status:       1,
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("create admin: %w", err)
	}
	return toUserResponse(&user), nil
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

// IsOAuthConfigured returns true if the OAuth provider has client credentials configured.
func (s *Service) IsOAuthConfigured(provider string) bool {
	switch provider {
	case "google":
		return s.oauthCfg.GoogleClientID != "" && s.oauthCfg.GoogleClientSecret != ""
	case "github":
		return s.oauthCfg.GitHubClientID != "" && s.oauthCfg.GitHubClientSecret != ""
	default:
		return false
	}
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
			Role:              constants.RoleUser,
			Level:             constants.LevelNormal,
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
// callbackURL is validated against allowed origins to prevent open redirect attacks.
func (s *Service) GetOAuthRedirectURL(provider, callbackURL, requestHost string) (string, error) {
	// Validate callback URL against whitelist to prevent open redirect.
	// Allowed hosts: localhost/127.0.0.1 (dev) + the actual request host (production).
	allowedHosts := map[string]bool{
		"localhost": true, "127.0.0.1": true,
		requestHost: true,
	}
	if !isAllowedCallback(callbackURL, allowedHosts) {
		return "", fmt.Errorf("invalid callback URL: %s", callbackURL)
	}
	switch provider {
	case "google":
		clientID := s.oauthCfg.GoogleClientID
		return fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=email%%20profile", clientID, callbackURL), nil
	case "github":
		clientID := s.oauthCfg.GitHubClientID
		return fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email", clientID, callbackURL), nil
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}
}

// isAllowedCallback validates the callback URL against a whitelist of hosts.
// Uses url.Parse for strict host comparison instead of string prefix matching,
// which prevents bypasses like https://localhost.evil.com matching "localhost".
func isAllowedCallback(urlStr string, allowed map[string]bool) bool {
	if urlStr == "" {
		return false
	}
	parsed, err := url.Parse(urlStr)
	if err != nil || parsed.Host == "" {
		return false
	}
	// Only allow http/https schemes
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	// Strip port for host comparison (url.Parse includes port in Host)
	host := parsed.Hostname()
	return allowed[host]
}
