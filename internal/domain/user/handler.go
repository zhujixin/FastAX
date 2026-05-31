package user

import (
	"fmt"
	"net/http"

	"github.com/fastax/fastax-server/internal/shared/middleware"
	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	// Get language from middleware if not provided
	if req.Language == "" {
		req.Language = middleware.GetLanguage(c)
	}
	resp, err := h.svc.Register(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	resp, err := h.svc.Login(&req)
	if err != nil {
		if err.Error() == "account is frozen" {
			response.Error(c, http.StatusUnauthorized, response.CodeAccountFrozen, err.Error())
			return
		}
		response.Error(c, http.StatusUnauthorized, response.CodeTokenExpired, err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *Handler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	resp, err := h.svc.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, response.CodeTokenExpired, err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *Handler) GetUser(c *gin.Context) {
	userID, _ := c.Get("user_id")
	resp, err := h.svc.GetUser(userID.(uint))
	if err != nil {
		response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *Handler) Logout(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	// Body is optional; if present, extract refresh token
	_ = c.ShouldBindJSON(&req)

	if err := h.svc.Logout(userID.(uint), req.RefreshToken); err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "logged out"})
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	if err := h.svc.ResetPassword(&req); err != nil {
		if err.Error() == "user not found" {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
			return
		}
		if err.Error() == "invalid or expired verification code" {
			response.Error(c, http.StatusBadRequest, response.CodeVerifyFailed, err.Error())
			return
		}
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "password reset successfully"})
}

func (h *Handler) SendCode(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
		Phone string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	identifier := req.Email
	if identifier == "" {
		identifier = req.Phone
	}
	if identifier == "" {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "email or phone required")
		return
	}
	code, err := h.svc.verify.GenerateCode(identifier)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	// TODO: Send code via email/SMS service (S4+)
	_ = code
	response.Success(c, gin.H{"message": "verification code sent"})
}

// GET /api/admin/users?page=1&page_size=20&keyword=xxx
func (h *Handler) ListUsers(c *gin.Context) {
	page, pageSize := 1, 20
	if p := c.Query("page"); p != "" {
		if v, err := fmt.Sscanf(p, "%d", &page); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, err := fmt.Sscanf(ps, "%d", &pageSize); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}
	keyword := c.Query("keyword")

	resp, err := h.svc.ListUsers(page, pageSize, keyword)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.SuccessPaginated(c, resp.Items, resp.Total, resp.Page, resp.PageSize)
}

// PUT /api/admin/users/:id/status
func (h *Handler) SetUserStatus(c *gin.Context) {
	var req struct {
		Status int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	var id uint
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid user id")
		return
	}

	if err := h.svc.SetUserStatus(id, req.Status); err != nil {
		if err.Error() == "cannot freeze super admin account" {
			response.Error(c, http.StatusForbidden, response.CodePermissionDeny, err.Error())
			return
		}
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "status updated"})
}

// PUT /api/user/language - Update user's preferred language
func (h *Handler) UpdateLanguage(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req struct {
		Language string `json:"language" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	if err := h.svc.UpdateLanguage(userID.(uint), req.Language); err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "language updated"})
}

// GET /api/admin/users/:id - Get user detail
func (h *Handler) GetUserDetail(c *gin.Context) {
	var id uint
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid user id")
		return
	}

	detail, err := h.svc.GetUserDetail(id)
	if err != nil {
		if err.Error() == "user not found" {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, detail)
}

// PUT /api/admin/users/:id/level - Set user level
func (h *Handler) SetUserLevel(c *gin.Context) {
	var id uint
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid user id")
		return
	}

	var req struct {
		Level string `json:"level" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	if err := h.svc.SetUserLevel(id, req.Level); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "level updated"})
}
