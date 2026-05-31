package enterprise

import (
	"net/http"
	"strconv"

	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// CreateSubAccount creates a new sub-account under the authenticated enterprise user.
func (h *Handler) CreateSubAccount(c *gin.Context) {
	parentID, _ := c.Get("user_id")

	var req SubAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	account, err := h.svc.CreateSubAccount(parentID.(uint), &req)
	if err != nil {
		if err.Error() == "only enterprise users can create sub-accounts" || err.Error() == "parent user not found" {
			response.Error(c, http.StatusForbidden, response.CodePermissionDeny, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, account)
}

// ListSubAccounts returns all sub-accounts for the authenticated enterprise user.
func (h *Handler) ListSubAccounts(c *gin.Context) {
	parentID, _ := c.Get("user_id")

	accounts, err := h.svc.ListSubAccounts(parentID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, accounts)
}

// SetSubAccountStatus updates the status of a sub-account.
func (h *Handler) SetSubAccountStatus(c *gin.Context) {
	parentID, _ := c.Get("user_id")

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid sub-account id")
		return
	}

	var body struct {
		Status *int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	if body.Status == nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "status is required")
		return
	}

	if err := h.svc.SetSubAccountStatus(uint(id), parentID.(uint), *body.Status); err != nil {
		if err.Error() == "sub-account not found" {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "status updated"})
}

// UpdateQuota updates the token quota of a sub-account.
func (h *Handler) UpdateQuota(c *gin.Context) {
	parentID, _ := c.Get("user_id")

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid sub-account id")
		return
	}

	var body struct {
		TokenQuota *int64 `json:"token_quota" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	if body.TokenQuota == nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "token_quota is required")
		return
	}

	if err := h.svc.UpdateQuota(uint(id), parentID.(uint), *body.TokenQuota); err != nil {
		if err.Error() == "sub-account not found" {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "quota updated"})
}

// GetUsage returns enterprise-wide usage statistics.
func (h *Handler) GetUsage(c *gin.Context) {
	parentID, _ := c.Get("user_id")
	period := c.DefaultQuery("period", "all")

	stats, err := h.svc.GetEnterpriseUsage(parentID.(uint), period)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, stats)
}

// GetSubAccountUsage returns usage statistics for a specific sub-account.
func (h *Handler) GetSubAccountUsage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid sub-account id")
		return
	}

	period := c.DefaultQuery("period", "all")

	stats, err := h.svc.GetSubAccountUsage(uint(id), period)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, stats)
}

// --- SSO Configuration ---

// GET /api/admin/sso/config
func (h *Handler) GetSSOConfig(c *gin.Context) {
	config, err := h.svc.GetSSOConfig()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, config)
}

// PUT /api/admin/sso/config
func (h *Handler) UpdateSSOConfig(c *gin.Context) {
	var req SSOConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	if err := h.svc.UpdateSSOConfig(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "SSO config updated"})
}

// --- Team Management ---

// GET /api/admin/teams
func (h *Handler) ListTeams(c *gin.Context) {
	teams, err := h.svc.ListTeams(0)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, teams)
}

// POST /api/admin/teams
func (h *Handler) CreateTeam(c *gin.Context) {
	parentID, _ := c.Get("user_id")

	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	team, err := h.svc.CreateTeam(parentID.(uint), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, team)
}

// PUT /api/admin/teams/:id
func (h *Handler) UpdateTeam(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid team id")
		return
	}

	var req UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	if err := h.svc.UpdateTeam(uint(id), &req); err != nil {
		if err.Error() == "team not found" {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "team updated"})
}

// DELETE /api/admin/teams/:id
func (h *Handler) DeleteTeam(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid team id")
		return
	}

	if err := h.svc.DeleteTeam(uint(id)); err != nil {
		if err.Error() == "team not found" {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "team deleted"})
}
