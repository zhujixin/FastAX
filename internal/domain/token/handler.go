package token

import (
	"net/http"

	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetProducts(c *gin.Context) {
	products, err := h.svc.GetProducts()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, products)
}

func (h *Handler) GetMyTokens(c *gin.Context) {
	userID, _ := c.Get("user_id")
	tokens, err := h.svc.GetUserTokens(userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, tokens)
}

func (h *Handler) Buy(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req BuyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	resp, err := h.svc.Buy(userID.(uint), &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *Handler) Transfer(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	resp, err := h.svc.Transfer(userID.(uint), &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *Handler) Extract(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req ExtractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	resp, err := h.svc.Extract(userID.(uint), &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	response.Success(c, resp)
}
