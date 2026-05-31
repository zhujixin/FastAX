package guardrail

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

func (h *Handler) ListRules(c *gin.Context) {
	stage := c.Query("stage")
	rules, err := h.svc.ListRules(stage)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, rules)
}

func (h *Handler) CreateRule(c *gin.Context) {
	var req RuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	rule, err := h.svc.CreateRule(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, rule)
}

func (h *Handler) SetRuleEnabled(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid rule id")
		return
	}
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	if err := h.svc.SetRuleEnabled(uint(id), req.Enabled); err != nil {
		if err.Error() == "rule not found" {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "rule enabled status updated"})
}

func (h *Handler) ListLogs(c *gin.Context) {
	traceID := c.Query("trace_id")
	stage := c.Query("stage")
	var userID uint
	if v := c.Query("user_id"); v != "" {
		uid, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid user_id")
			return
		}
		userID = uint(uid)
	}
	logs, err := h.svc.ListLogs(traceID, userID, stage)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, logs)
}

func (h *Handler) Detect(c *gin.Context) {
	var req DetectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	result, err := h.svc.Detect(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, result)
}
