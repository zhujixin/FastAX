package relay

import (
	"fmt"
	"net/http"
)

// ParseUpstreamError parses error responses from different suppliers
// into a standardized APIError.
func ParseUpstreamError(body []byte, apiType APIType) *APIError {
	switch apiType {
	case APITypeAnthropic:
		return parseAnthropicError(body)
	case APITypeGemini:
		return parseGeminiError(body)
	default:
		return parseOpenAIError(body)
	}
}

// parseOpenAIError parses OpenAI-format errors:
// {"error": {"message": "...", "type": "...", "code": "..."}}
func parseOpenAIError(body []byte) *APIError {
	var errResp struct {
		Error *APIError `json:"error"`
	}
	if err := jsonUnmarshal(body, &errResp); err == nil && errResp.Error != nil {
		return errResp.Error
	}
	// Fallback: raw body as message
	return &APIError{
		Message: string(body),
		Type:    "upstream_error",
		Code:    "unknown",
	}
}

// parseAnthropicError parses Anthropic-format errors:
// {"type": "error", "error": {"type": "...", "message": "..."}}
func parseAnthropicError(body []byte) *APIError {
	var errResp struct {
		Type  string `json:"type"`
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := jsonUnmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
		return &APIError{
			Message: errResp.Error.Message,
			Type:    errResp.Error.Type,
			Code:    errResp.Type,
		}
	}
	return &APIError{
		Message: string(body),
		Type:    "upstream_error",
		Code:    "unknown",
	}
}

// parseGeminiError parses Gemini-format errors:
// {"error": {"code": 400, "message": "...", "status": "..."}}
func parseGeminiError(body []byte) *APIError {
	var errResp struct {
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}
	if err := jsonUnmarshal(body, &errResp); err == nil && errResp.Error != nil {
		return &APIError{
			Message: errResp.Error.Message,
			Type:    errResp.Error.Status,
			Code:    fmt.Sprintf("%d", errResp.Error.Code),
		}
	}
	return &APIError{
		Message: string(body),
		Type:    "upstream_error",
		Code:    "unknown",
	}
}

// NewAPIError creates an APIError with the given parameters
func NewAPIError(statusCode int, message string) *APIError {
	return &APIError{
		Message: message,
		Type:    "upstream_error",
		Code:    fmt.Sprintf("%d", statusCode),
	}
}

// ShouldRetryHTTP returns true if the HTTP status code warrants a retry
func ShouldRetryHTTP(statusCode int) bool {
	return statusCode == 429 || statusCode >= 500
}

// HTTPStatusFromAPIError maps an APIError to an appropriate HTTP status code
func HTTPStatusFromAPIError(err *APIError) int {
	if err == nil {
		return http.StatusOK
	}
	switch err.Code {
	case "401", "invalid_api_key", "authentication_error":
		return http.StatusUnauthorized
	case "403", "forbidden", "permission_error":
		return http.StatusForbidden
	case "404", "not_found":
		return http.StatusNotFound
	case "429", "rate_limit_exceeded":
		return http.StatusTooManyRequests
	case "insufficient_quota":
		return http.StatusPaymentRequired
	default:
		return http.StatusBadGateway
	}
}
