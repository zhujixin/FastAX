package user

import (
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
