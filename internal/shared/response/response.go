// 统一响应格式
//
// 本文件定义了 API 统一响应格式和辅助函数：
//   - APIResponse: 统一响应结构体，包含 code、message、data、trace_id
//   - PaginatedData: 分页数据结构体，包含 items、total、page、size
//   - Success: 成功响应（code=0）
//   - Error: 错误响应（自定义业务码）
//   - Paginate: 分页响应
//   - 错误码常量: 定义所有业务错误码
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse 统一响应结构体
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
}

// PaginatedData 分页数据结构体
type PaginatedData struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

const (
	CodeSuccess = 0

	CodeParamInvalid    = 1001
	CodeVerifyFailed    = 1002
	CodeTokenExpired    = 2001
	CodeAccountFrozen   = 2002
	CodePermissionDeny  = 3001
	CodeRateLimited     = 3002
	CodeNotFound        = 4001
	CodeDuplicateOp     = 5001
	CodeBalanceInsufficient = 6001
	CodeTokenExpiredOp  = 6002
	CodeOverLimit       = 6003
	CodeTooFrequent     = 7001
	CodeInternalError   = 9001
	CodeServiceUnavail  = 9002
)

var codeMessages = map[int]string{
	CodeSuccess:         "success",
	CodeParamInvalid:    "invalid parameters",
	CodeVerifyFailed:    "verification code expired or invalid",
	CodeTokenExpired:    "token expired or invalid",
	CodeAccountFrozen:   "account frozen",
	CodePermissionDeny:  "permission denied",
	CodeRateLimited:     "rate limit exceeded",
	CodeNotFound:        "resource not found",
	CodeDuplicateOp:     "duplicate operation",
	CodeBalanceInsufficient: "insufficient balance",
	CodeTokenExpiredOp:  "token expired",
	CodeOverLimit:       "purchase limit exceeded",
	CodeTooFrequent:     "too frequent",
	CodeInternalError:   "internal error",
	CodeServiceUnavail:  "service unavailable",
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

func SuccessPaginated(c *gin.Context, items interface{}, total int64, page, size int) {
	Success(c, PaginatedData{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

func Error(c *gin.Context, httpStatus, code int, msg ...string) {
	message := codeMessages[code]
	if len(msg) > 0 && msg[0] != "" {
		message = msg[0]
	}
	c.AbortWithStatusJSON(httpStatus, APIResponse{
		Code:    code,
		Message: message,
	})
}

func InternalError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, CodeInternalError)
}
