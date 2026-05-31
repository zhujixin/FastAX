package proxy

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/fastax/fastax-server/internal/domain/proxy/relay"
	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// AnthropicMessageRequest represents the Anthropic Messages API request format.
type AnthropicMessageRequest struct {
	Model       string          `json:"model"`
	Messages    json.RawMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Stream      bool            `json:"stream"`
	System      string          `json:"system,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	TopP        float64         `json:"top_p,omitempty"`
}

// ChatCompletions handles /v1/chat/completions
func (h *Handler) ChatCompletions(c *gin.Context) {
	var req RelayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	if req.Model == "" {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "model is required")
		return
	}

	userID, _ := c.Get("user_id")
	ctx := c.Request.Context()

	if req.Stream {
		h.handleStream(c, ctx, userID.(uint), &req)
		return
	}

	h.handleNonStream(c, ctx, userID.(uint), &req)
}

// ChatMessages handles /v1/messages (Anthropic Messages API)
func (h *Handler) ChatMessages(c *gin.Context) {
	var anthropicReq AnthropicMessageRequest
	if err := c.ShouldBindJSON(&anthropicReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": err.Error(),
			},
		})
		return
	}

	// Validate required fields per Anthropic spec
	if anthropicReq.Model == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": "model is required",
			},
		})
		return
	}
	if len(anthropicReq.Messages) == 0 || string(anthropicReq.Messages) == "null" {
		c.JSON(http.StatusBadRequest, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": "messages is required",
			},
		})
		return
	}
	if anthropicReq.MaxTokens <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": "max_tokens is required and must be > 0",
			},
		})
		return
	}

	// Convert Anthropic request to internal RelayRequest
	// Messages are kept as raw JSON to preserve Anthropic content block format
	req := &RelayRequest{
		Model:        anthropicReq.Model,
		Stream:       anthropicReq.Stream,
		MaxTokens:    anthropicReq.MaxTokens,
		Temperature:  anthropicReq.Temperature,
		TopP:         anthropicReq.TopP,
		RawMessages:  anthropicReq.Messages,
		SystemPrompt: anthropicReq.System,
	}

	// Also extract simplified messages for routing/fallback
	req.Messages = extractAnthropicMessages(anthropicReq.Messages)

	userID, _ := c.Get("user_id")
	ctx := c.Request.Context()

	if req.Stream {
		h.handleStream(c, ctx, userID.(uint), req)
		return
	}

	h.handleNonStream(c, ctx, userID.(uint), req)
}

// handleNonStream handles non-streaming relay requests
func (h *Handler) handleNonStream(c *gin.Context, ctx interface{}, userID uint, req *RelayRequest) {
	resp, err := h.svc.Relay(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, http.StatusBadGateway, response.CodeServiceUnavail, err.Error())
		return
	}

	c.Data(resp.StatusCode, "application/json", resp.Body)
}

// handleStream handles streaming relay requests with SSE
func (h *Handler) handleStream(c *gin.Context, ctx interface{}, userID uint, req *RelayRequest) {
	streamResp, err := h.svc.RelayStream(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, http.StatusBadGateway, response.CodeServiceUnavail, err.Error())
		return
	}
	defer streamResp.Body.Close()

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

	// Stream response status
	c.Status(streamResp.StatusCode)

	// Flush headers
	c.Writer.Flush()

	// Forward SSE chunks in real-time
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		// Fallback: read all and send
		io.Copy(c.Writer, streamResp.Body)
		return
	}

	buf := make([]byte, 4096)
	for {
		n, err := streamResp.Body.Read(buf)
		if n > 0 {
			c.Writer.Write(buf[:n])
			flusher.Flush()
		}
		if err != nil {
			if err != io.EOF {
				// Log error but don't send error to client (stream already started)
			}
			break
		}
	}
}

// ImageGenerations handles POST /v1/images/generations
func (h *Handler) ImageGenerations(c *gin.Context) {
	var req relay.ImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	if req.Model == "" {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "model is required")
		return
	}
	if req.Prompt == "" {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "prompt is required")
		return
	}

	userID, _ := c.Get("user_id")
	ctx := c.Request.Context()

	resp, err := h.svc.ImageRelay(ctx, userID.(uint), &req)
	if err != nil {
		response.Error(c, http.StatusBadGateway, response.CodeServiceUnavail, err.Error())
		return
	}

	c.Data(resp.StatusCode, "application/json", resp.Body)
}

// AudioSpeech handles POST /v1/audio/speech
func (h *Handler) AudioSpeech(c *gin.Context) {
	var req relay.AudioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	if req.Model == "" {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "model is required")
		return
	}
	if req.Input == "" {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "input is required")
		return
	}
	if req.Voice == "" {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "voice is required")
		return
	}

	userID, _ := c.Get("user_id")
	ctx := c.Request.Context()

	resp, err := h.svc.AudioRelay(ctx, userID.(uint), &req)
	if err != nil {
		response.Error(c, http.StatusBadGateway, response.CodeServiceUnavail, err.Error())
		return
	}

	// Audio response may be binary (mp3/opus/etc), pass content-type from upstream
	c.Data(resp.StatusCode, "application/octet-stream", resp.Body)
}

// ListModels handles GET /v1/models
func (h *Handler) ListModels(c *gin.Context) {
	models, err := h.svc.ListModels()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	type modelInfo struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
	}
	data := make([]modelInfo, len(models))
	for i, m := range models {
		data[i] = modelInfo{
			ID:      m,
			Object:  "model",
			Created: time.Now().Unix(),
		}
	}

	c.JSON(200, gin.H{
		"object": "list",
		"data":   data,
	})
}

// Rerank handles POST /v1/rerank
func (h *Handler) Rerank(c *gin.Context) {
	var req RerankRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	if req.Model == "" {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "model is required")
		return
	}
	if req.Query == "" {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "query is required")
		return
	}
	if len(req.Documents) == 0 {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "documents is required")
		return
	}

	userID, _ := c.Get("user_id")
	ctx := c.Request.Context()

	resp, err := h.svc.Rerank(ctx, userID.(uint), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	response.Success(c, resp)
}

// VideoGenerations handles POST /v1/video/generations
func (h *Handler) VideoGenerations(c *gin.Context) {
	var req VideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, err.Error())
		return
	}

	if req.Model == "" {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "model is required")
		return
	}
	if req.Prompt == "" {
		response.Error(c, http.StatusBadRequest, response.CodeParamInvalid, "prompt is required")
		return
	}

	userID, _ := c.Get("user_id")
	ctx := c.Request.Context()

	resp, err := h.svc.VideoRelay(ctx, userID.(uint), &req)
	if err != nil {
		response.Error(c, http.StatusBadGateway, response.CodeServiceUnavail, err.Error())
		return
	}

	c.Data(resp.StatusCode, "application/json", resp.Body)
}

// GetRelayResponse converts relay response for testing
func GetRelayResponse(resp *RelayResponse) *relay.Response {
	return resp.Resp
}

// extractAnthropicMessages parses Anthropic-format messages JSON and extracts
// simplified relay.Message entries (role + text content). This is used for
// compatibility with code that expects []relay.Message, while the raw JSON
// is preserved in RawMessages for direct forwarding to Anthropic upstreams.
func extractAnthropicMessages(raw json.RawMessage) []relay.Message {
	// Anthropic messages format:
	// [{"role": "user", "content": "hello"}]
	// or [{"role": "user", "content": [{"type": "text", "text": "hello"}]}]
	var msgs []struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(raw, &msgs); err != nil {
		return nil
	}

	result := make([]relay.Message, 0, len(msgs))
	for _, m := range msgs {
		text := extractTextFromContent(m.Content)
		result = append(result, relay.Message{
			Role:    m.Role,
			Content: text,
		})
	}
	return result
}

// extractTextFromContent extracts plain text from an Anthropic content field.
// Content can be a string or an array of content blocks.
func extractTextFromContent(content json.RawMessage) string {
	// Try as string first
	var s string
	if err := json.Unmarshal(content, &s); err == nil {
		return s
	}

	// Try as array of content blocks
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(content, &blocks); err == nil {
		for _, b := range blocks {
			if b.Type == "text" && b.Text != "" {
				return b.Text
			}
		}
	}
	return ""
}
