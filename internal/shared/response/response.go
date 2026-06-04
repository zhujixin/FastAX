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

// 多语言错误消息
var codeMessagesI18n = map[int]map[string]string{
	CodeSuccess: {
		"zh-CN": "成功",
		"en":    "success",
		"ja":    "成功",
	},
	CodeParamInvalid: {
		"zh-CN": "参数无效",
		"en":    "invalid parameters",
		"ja":    "パラメータが無効です",
	},
	CodeVerifyFailed: {
		"zh-CN": "验证码已过期或无效",
		"en":    "verification code expired or invalid",
		"ja":    "確認コードの有効期限が切れているか無効です",
	},
	CodeTokenExpired: {
		"zh-CN": "Token 已过期或无效",
		"en":    "token expired or invalid",
		"ja":    "トークンの有効期限が切れているか無効です",
	},
	CodeAccountFrozen: {
		"zh-CN": "账号已被冻结",
		"en":    "account frozen",
		"ja":    "アカウントが凍結されています",
	},
	CodePermissionDeny: {
		"zh-CN": "权限不足",
		"en":    "permission denied",
		"ja":    "権限が不足しています",
	},
	CodeRateLimited: {
		"zh-CN": "请求过于频繁，请稍后再试",
		"en":    "rate limit exceeded",
		"ja":    "リクエストが多すぎます。しばらく待ってから再試行してください",
	},
	CodeNotFound: {
		"zh-CN": "资源不存在",
		"en":    "resource not found",
		"ja":    "リソースが見つかりません",
	},
	CodeDuplicateOp: {
		"zh-CN": "重复操作",
		"en":    "duplicate operation",
		"ja":    "重複操作です",
	},
	CodeBalanceInsufficient: {
		"zh-CN": "余额不足",
		"en":    "insufficient balance",
		"ja":    "残高が不足しています",
	},
	CodeTokenExpiredOp: {
		"zh-CN": "Token 已过期",
		"en":    "token expired",
		"ja":    "トークンの有効期限が切れています",
	},
	CodeOverLimit: {
		"zh-CN": "超出购买限额",
		"en":    "purchase limit exceeded",
		"ja":    "購入限度額を超えています",
	},
	CodeTooFrequent: {
		"zh-CN": "操作过于频繁",
		"en":    "too frequent",
		"ja":    "操作が頻繁すぎます",
	},
	CodeInternalError: {
		"zh-CN": "系统内部错误",
		"en":    "internal error",
		"ja":    "システム内部エラー",
	},
	CodeServiceUnavail: {
		"zh-CN": "服务暂不可用",
		"en":    "service unavailable",
		"ja":    "サービスは一時的に利用できません",
	},
}

// getMessage 根据语言获取错误消息
func getMessage(code int, lang string) string {
	if m, ok := codeMessagesI18n[code]; ok {
		if msg, ok := m[lang]; ok {
			return msg
		}
		if msg, ok := m["en"]; ok {
			return msg
		}
	}
	return codeMessages[code]
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
	// 优先使用传入的自定义消息
	if len(msg) > 0 && msg[0] != "" {
		c.AbortWithStatusJSON(httpStatus, APIResponse{
			Code:    code,
			Message: msg[0],
		})
		return
	}
	// 根据 Accept-Language 选择多语言消息
	lang := c.GetString("language")
	if lang == "" {
		lang = "en"
	}
	message := getMessage(code, lang)
	c.AbortWithStatusJSON(httpStatus, APIResponse{
		Code:    code,
		Message: message,
	})
}

// SanitizedError returns a sanitized error to the client.
// In release mode, only the predefined error message is shown.
// In debug mode, the internal error details are included for debugging.
func SanitizedError(c *gin.Context, httpStatus, code int, internalErr error) {
	if gin.Mode() == gin.ReleaseMode {
		Error(c, httpStatus, code) // Predefined message only
	} else {
		Error(c, httpStatus, code, internalErr.Error())
	}
}

func InternalError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, CodeInternalError)
}
