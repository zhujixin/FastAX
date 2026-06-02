// 日志模块 HTTP 处理器
//
// 本文件实现了日志相关的 HTTP 接口：
//
// 审计日志：
//   - ListAuditLogs: 审计日志列表（支持分页、筛选）
//   - ExportAuditLogs: 导出审计日志（CSV/Excel）
//
// 调用日志：
//   - ListCallLogs: API 调用日志列表（支持分页、筛选）
package log

import (
	"net/http"

	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

// Handler 日志模块 HTTP 处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建日志处理器实例
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListAuditLogs 获取审计日志列表
// GET /api/admin/audit/logs
func (h *Handler) ListAuditLogs(c *gin.Context) {
	var query AuditLogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	items, total, err := h.svc.ListAuditLogs(&query)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.SuccessPaginated(c, items, total, query.Page, query.PageSize)
}

// ExportAuditLogs exports audit logs matching the query as a CSV file download.
func (h *Handler) ExportAuditLogs(c *gin.Context) {
	var query AuditLogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	data, err := h.svc.Export(&query)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="audit_logs.csv"`)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

// ListCallLogs returns call logs with pagination and filtering.
func (h *Handler) ListCallLogs(c *gin.Context) {
	var query CallLogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	items, total, err := h.svc.ListCallLogs(&query)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.SuccessPaginated(c, items, total, query.Page, query.PageSize)
}
