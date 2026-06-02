// 模型市场模块 HTTP 处理器
//
// 本文件实现了模型市场相关的 HTTP 接口：
//
// 模型查询：
//   - ListModels: 模型列表（支持按供应商、类型筛选）
//   - CompareModels: 模型多维度对比
//   - GetBenchmarks: 获取基准测试数据
//   - RecommendModels: 模型推荐
//
// 供应商健康：
//   - ListProviders: 供应商健康面板（公开）
//   - GetProviderHealth: 单个供应商健康状态
package market

import (
	"net/http"
	"strconv"

	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

// Handler 模型市场模块 HTTP 处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建模型市场处理器实例
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListModels 获取模型列表
// GET /api/models
func (h *Handler) ListModels(c *gin.Context) {
	provider := c.Query("provider")
	modelType := c.Query("type")
	result, err := h.svc.ListModels(provider, modelType)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *Handler) CompareModels(c *gin.Context) {
	var req struct {
		Models []string `json:"models" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	result, err := h.svc.CompareModels(req.Models)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *Handler) GetProviderHealth(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid provider id")
		return
	}
	result, err := h.svc.GetProviderHealth(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *Handler) ListProviders(c *gin.Context) {
	result, err := h.svc.ListProviders()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}

// GET /api/models/benchmarks?model=gpt-4
func (h *Handler) GetBenchmarks(c *gin.Context) {
	model := c.Query("model")

	result, err := h.svc.GetBenchmarks(model)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}

// GET /api/models/recommend
func (h *Handler) RecommendModels(c *gin.Context) {
	userID, _ := c.Get("user_id")

	result, err := h.svc.RecommendModels(userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}
