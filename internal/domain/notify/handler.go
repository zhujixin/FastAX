// 通知模块 HTTP 处理器
//
// 本文件实现了通知相关的 HTTP 接口：
//
// 用户通知：
//   - List: 通知列表（支持分页、筛选）
//   - UnreadCount: 未读通知数量
//   - MarkRead: 标记单条通知为已读
//   - MarkAllRead: 标记所有通知为已读
//
// 通知模板管理（管理员）：
//   - ListTemplates: 模板列表
//   - CreateTemplate: 创建模板
//   - UpdateTemplate: 更新模板
package notify

import (
	"net/http"
	"strconv"

	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

// Handler 通知模块 HTTP 处理器
type Handler struct {
	svc *Service
}

// NewHandler 创建通知处理器实例
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List 获取通知列表
// GET /api/notifications
func (h *Handler) List(c *gin.Context) {
	userID, _ := c.Get("user_id")

	notifType := c.Query("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var isRead *bool
	if v := c.Query("is_read"); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			isRead = &b
		}
	}

	notifs, total, err := h.svc.ListByUser(userID.(uint), notifType, isRead, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.SuccessPaginated(c, notifs, total, page, pageSize)
}

func (h *Handler) UnreadCount(c *gin.Context) {
	userID, _ := c.Get("user_id")

	count, err := h.svc.GetUnreadCount(userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"count": count})
}

func (h *Handler) MarkRead(c *gin.Context) {
	userID, _ := c.Get("user_id")

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid notification id")
		return
	}

	if err := h.svc.MarkRead(uint(id), userID.(uint)); err != nil {
		if err.Error() == "notification not found" {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "marked as read"})
}

func (h *Handler) MarkAllRead(c *gin.Context) {
	userID, _ := c.Get("user_id")

	if err := h.svc.MarkAllRead(userID.(uint)); err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "all marked as read"})
}

func (h *Handler) ListTemplates(c *gin.Context) {
	channel := c.Query("channel")
	language := c.Query("language")

	templates, err := h.svc.ListTemplates(channel, language)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, templates)
}

func (h *Handler) CreateTemplate(c *gin.Context) {
	var req TemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	tmpl, err := h.svc.CreateTemplate(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, tmpl)
}

func (h *Handler) UpdateTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid template id")
		return
	}

	var req TemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	tmpl, err := h.svc.UpdateTemplate(uint(id), &req)
	if err != nil {
		if err.Error() == "template not found" {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, tmpl)
}
