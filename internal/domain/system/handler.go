// 系统配置模块 HTTP 处理器
//
// 本文件实现了系统配置相关的 HTTP 接口：
//   - GetConfig: 获取所有系统配置
//   - UpdateConfig: 批量更新系统配置
package system

import (
	"net/http"

	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

// Handler 系统配置模块 HTTP 处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建系统配置处理器实例
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GetConfig 获取系统配置
// GET /api/admin/system/config
func (h *Handler) GetConfig(c *gin.Context) {
	entries, err := h.svc.GetAll()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"items": entries})
}

// UpdateConfig 更新系统配置
// PUT /api/admin/system/config
func (h *Handler) UpdateConfig(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	if err := h.svc.UpdateAll(req); err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "config updated"})
}
