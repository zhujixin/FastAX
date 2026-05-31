package enterprise

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

// --- Sub Account Management ---

type SubAccountRequest struct {
	Email      string   `json:"email" binding:"required,email"`
	Password   string   `json:"password" binding:"required,min=6"`
	TokenQuota int64    `json:"token_quota"`
	Permissions []string `json:"permissions"`
}

type SubAccountResponse struct {
	ID          uint     `json:"id"`
	ParentID    uint     `json:"parent_id"`
	Email       string   `json:"email"`
	TokenQuota  int64    `json:"token_quota"`
	Permissions []string `json:"permissions"`
	Status      int      `json:"status"`
}

func (s *Service) CreateSubAccount(parentID uint, req *SubAccountRequest) (*SubAccountResponse, error) {
	// Verify parent is enterprise user
	var parent model.User
	if err := s.db.First(&parent, parentID).Error; err != nil {
		return nil, errors.New("parent user not found")
	}
	if parent.Role != "enterprise" && parent.Role != "admin" {
		return nil, errors.New("only enterprise users can create sub-accounts")
	}

	permStr := ""
	for i, p := range req.Permissions {
		if i > 0 {
			permStr += ","
		}
		permStr += p
	}

	account := model.SubAccount{
		ParentID:    parentID,
		Email:       req.Email,
		PasswordHash: req.Password, // In production, hash the password
		TokenQuota:  req.TokenQuota,
		Permissions: permStr,
		Status:      1,
	}
	if err := s.db.Create(&account).Error; err != nil {
		return nil, fmt.Errorf("create sub-account: %w", err)
	}
	return toSubAccountResponse(&account), nil
}

func (s *Service) ListSubAccounts(parentID uint) ([]SubAccountResponse, error) {
	var accounts []model.SubAccount
	if err := s.db.Where("parent_id = ?", parentID).Find(&accounts).Error; err != nil {
		return nil, err
	}
	result := make([]SubAccountResponse, len(accounts))
	for i, a := range accounts {
		result[i] = *toSubAccountResponse(&a)
	}
	return result, nil
}

func (s *Service) SetSubAccountStatus(id, parentID uint, status int) error {
	result := s.db.Model(&model.SubAccount{}).
		Where("id = ? AND parent_id = ?", id, parentID).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("sub-account not found")
	}
	return nil
}

func (s *Service) UpdateQuota(id, parentID uint, quota int64) error {
	result := s.db.Model(&model.SubAccount{}).
		Where("id = ? AND parent_id = ?", id, parentID).
		Update("token_quota", quota)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("sub-account not found")
	}
	return nil
}

// --- Usage Statistics ---

type UsageStats struct {
	TotalTokens    int    `json:"total_tokens"`
	TotalRequests  int    `json:"total_requests"`
	Period         string `json:"period"`
}

func (s *Service) GetEnterpriseUsage(parentID uint, period string) (*UsageStats, error) {
	// Get sub-account IDs
	var subAccounts []model.SubAccount
	s.db.Where("parent_id = ?", parentID).Find(&subAccounts)

	ids := make([]uint, len(subAccounts))
	for i, sa := range subAccounts {
		ids[i] = sa.ID
	}
	ids = append(ids, parentID)

	var stats struct {
		TotalTokens   int
		TotalRequests int
	}
	err := s.db.Raw(`
		SELECT COALESCE(SUM(tokens_total), 0) as total_tokens, COUNT(*) as total_requests
		FROM call_log
		WHERE user_id IN ? AND created_at >= ?
	`, ids, period).Scan(&stats).Error
	if err != nil {
		return nil, err
	}

	return &UsageStats{
		TotalTokens:   stats.TotalTokens,
		TotalRequests: stats.TotalRequests,
		Period:        period,
	}, nil
}

func (s *Service) GetSubAccountUsage(subAccountID uint, period string) (*UsageStats, error) {
	var stats struct {
		TotalTokens   int
		TotalRequests int
	}
	err := s.db.Raw(`
		SELECT COALESCE(SUM(tokens_total), 0) as total_tokens, COUNT(*) as total_requests
		FROM call_log
		WHERE sub_account_id = ? AND created_at >= ?
	`, subAccountID, period).Scan(&stats).Error
	if err != nil {
		return nil, err
	}

	return &UsageStats{
		TotalTokens:   stats.TotalTokens,
		TotalRequests: stats.TotalRequests,
		Period:        period,
	}, nil
}

func toSubAccountResponse(a *model.SubAccount) *SubAccountResponse {
	perms := []string{}
	if a.Permissions != "" {
		for _, p := range splitString(a.Permissions, ",") {
			perms = append(perms, p)
		}
	}
	return &SubAccountResponse{
		ID:          a.ID,
		ParentID:    a.ParentID,
		Email:       a.Email,
		TokenQuota:  a.TokenQuota,
		Permissions: perms,
		Status:      a.Status,
	}
}

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == sep[0] {
			if start < i {
				result = append(result, s[start:i])
			}
			start = i + 1
		}
	}
	return result
}

// ---------- SSO Configuration ----------

type SSOConfig struct {
	Enabled     bool   `json:"enabled"`
	Protocol    string `json:"protocol"` // saml, oidc
	EntityID    string `json:"entity_id,omitempty"`
	MetadataURL string `json:"metadata_url,omitempty"`
	ClientID    string `json:"client_id,omitempty"`
	Issuer      string `json:"issuer,omitempty"`
	CallbackURL string `json:"callback_url,omitempty"`
}

// In-memory SSO config (in production, store in DB)
var ssoConfig = &SSOConfig{}

// GetSSOConfig returns current SSO configuration.
func (s *Service) GetSSOConfig() (*SSOConfig, error) {
	return ssoConfig, nil
}

// UpdateSSOConfig updates SSO configuration.
func (s *Service) UpdateSSOConfig(req *SSOConfig) error {
	if req.Protocol != "" && req.Protocol != "saml" && req.Protocol != "oidc" {
		return fmt.Errorf("invalid protocol: %s, must be saml or oidc", req.Protocol)
	}

	ssoConfig = req
	return nil
}

// ---------- Team Management ----------

type Team struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    uint   `json:"parent_id"`
	MemberCount int    `json:"member_count"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

type CreateTeamRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateTeamRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
}

// In-memory teams (in production, use DB table)
var teams = make(map[uint]*Team)
var teamIDCounter uint = 0

// CreateTeam creates a new team.
func (s *Service) CreateTeam(parentID uint, req *CreateTeamRequest) (*Team, error) {
	teamIDCounter++
	team := &Team{
		ID:          teamIDCounter,
		Name:        req.Name,
		Description: req.Description,
		ParentID:    parentID,
		MemberCount: 0,
		Status:      "active",
		CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
	}
	teams[team.ID] = team
	return team, nil
}

// ListTeams returns all teams for a parent.
func (s *Service) ListTeams(parentID uint) ([]Team, error) {
	result := make([]Team, 0)
	for _, t := range teams {
		if t.ParentID == parentID || parentID == 0 {
			result = append(result, *t)
		}
	}
	return result, nil
}

// GetTeam returns a team by ID.
func (s *Service) GetTeam(id uint) (*Team, error) {
	team, ok := teams[id]
	if !ok {
		return nil, errors.New("team not found")
	}
	return team, nil
}

// UpdateTeam updates a team.
func (s *Service) UpdateTeam(id uint, req *UpdateTeamRequest) error {
	team, ok := teams[id]
	if !ok {
		return errors.New("team not found")
	}

	if req.Name != "" {
		team.Name = req.Name
	}
	if req.Description != "" {
		team.Description = req.Description
	}
	if req.Status != "" {
		team.Status = req.Status
	}
	return nil
}

// DeleteTeam deletes a team.
func (s *Service) DeleteTeam(id uint) error {
	if _, ok := teams[id]; !ok {
		return errors.New("team not found")
	}
	delete(teams, id)
	return nil
}
