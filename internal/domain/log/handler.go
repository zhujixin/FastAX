package log

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

// ListAuditLogs returns audit logs with pagination and filtering.
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
