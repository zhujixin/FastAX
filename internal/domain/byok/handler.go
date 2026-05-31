package byok

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

// ListKeys returns all BYOK keys for the authenticated user.
func (h *Handler) ListKeys(c *gin.Context) {
	userID, _ := c.Get("user_id")

	keys, err := h.svc.ListKeys(userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, keys)
}

// AddKey creates a new BYOK key for the authenticated user.
func (h *Handler) AddKey(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req AddKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	key, err := h.svc.AddKey(userID.(uint), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, key)
}

// DeleteKey deletes a BYOK key by ID for the authenticated user.
func (h *Handler) DeleteKey(c *gin.Context) {
	userID, _ := c.Get("user_id")

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid key id")
		return
	}

	if err := h.svc.DeleteKey(uint(id), userID.(uint)); err != nil {
		if err.Error() == "key not found" {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "key deleted"})
}

// SetKeyStatus updates the status of a BYOK key.
func (h *Handler) SetKeyStatus(c *gin.Context) {
	userID, _ := c.Get("user_id")

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "invalid key id")
		return
	}

	var body struct {
		Status *int `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}
	if body.Status == nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "status is required")
		return
	}

	if err := h.svc.SetKeyStatus(uint(id), userID.(uint), *body.Status); err != nil {
		if err.Error() == "key not found" {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "status updated"})
}
