// Token 商品模块 HTTP 处理器
//
// 本文件实现了 Token 商品相关的 HTTP 接口：
//   - GetProducts: 获取商品列表（公开接口）
//   - GetProduct: 获取商品详情（公开接口）
//   - GetMyTokens: 获取用户持有的 Token（需登录）
//   - GetUsageHistory: 获取 Token 使用记录（需登录）
//   - Buy: 购买 Token（需登录）
//   - Transfer: 转让 Token（需登录）
//   - Extract: 提取 Token（需登录）
//   - CreateProduct: 创建商品（管理员）
//   - UpdateProduct: 更新商品（管理员）
package token

import (
	"fmt"
	"net/http"

	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

// Handler Token 商品模块 HTTP 处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建 Token 处理器实例
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GetProducts 获取商品列表（公开接口）
// GET /api/tokens/products
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

// POST /api/admin/products - Create product (admin)
func (h *Handler) CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	resp, err := h.svc.CreateProduct(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, resp)
}

// PUT /api/admin/products/:id - Update product (admin)
func (h *Handler) UpdateProduct(c *gin.Context) {
	var id uint
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &id); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid product id")
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	if err := h.svc.UpdateProduct(id, &req); err != nil {
		if err.Error() == "product not found" {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "product updated"})
}
