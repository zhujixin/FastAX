package notify

import (
	"net/http"
	"strconv"

	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

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
