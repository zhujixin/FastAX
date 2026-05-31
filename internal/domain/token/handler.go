package token

import (
	"fmt"
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

// GET /api/tokens/products/:id
func (h *Handler) GetProduct(c *gin.Context) {
	var id uint
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid product id")
		return
	}

	product, err := h.svc.GetProduct(id)
	if err != nil {
		if err.Error() == "product not found" {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, product)
}

// GET /api/tokens/my/usage
func (h *Handler) GetUsageHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")

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

	items, total, err := h.svc.GetUsageHistory(userID.(uint), page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.SuccessPaginated(c, items, total, page, pageSize)
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
