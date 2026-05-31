package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SupplierMeta holds supplier context passed to adaptors
type SupplierMeta struct {
	SupplierID   uint
	ChannelID    uint
	APIBaseURL   string
	APIKey       string
	APIType      APIType
	Model        string            // target model (may be remapped)
	ModelMapping map[string]string // upstream model -> real model
}

// Request represents an incoming API request
type Request struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`

	// Anthropic native fields (used when client sends Anthropic Messages API format)
	// RawMessages holds the original Anthropic-format messages as raw JSON.
	// When set, AnthropicAdaptor.ConvertRequest passes them through directly
	// instead of converting from OpenAI format.
	RawMessages json.RawMessage `json:"-"`
	SystemPrompt string         `json:"system,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Response represents an API response
type Response struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Choices []Choice  `json:"choices"`
	Usage   Usage     `json:"usage"`
	Error   *APIError `json:"error,omitempty"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error [%s/%s]: %s", e.Type, e.Code, e.Message)
}

// StreamChunk represents a single SSE chunk
type StreamChunk struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
	Usage   *Usage         `json:"usage,omitempty"`
	Error   *APIError      `json:"error,omitempty"`
}

type StreamChoice struct {
	Index        int          `json:"index"`
	Delta        StreamDelta  `json:"delta"`
	FinishReason *string      `json:"finish_reason"`
}

type StreamDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// Adaptor is the interface for supplier-specific API adapters.
// Follows the one-api pattern with 9 methods.
type Adaptor interface {
	// Init initializes the adaptor with supplier metadata
	Init(meta *SupplierMeta)

	// GetRequestURL returns the full request URL for the supplier API
	GetRequestURL(meta *SupplierMeta) (string, error)

	// SetupRequestHeader sets supplier-specific headers
	SetupRequestHeader(req *http.Request, meta *SupplierMeta) error

	// ConvertRequest converts a generic request to supplier-specific format
	ConvertRequest(req *Request) ([]byte, error)

	// ConvertImageRequest converts an image generation request
	ConvertImageRequest(req *ImageRequest) ([]byte, error)

	// DoRequest sends the HTTP request to the supplier API
	DoRequest(ctx context.Context, meta *SupplierMeta, body io.Reader) (*http.Response, error)

	// DoResponse processes the supplier response and returns usage info
	DoResponse(ctx context.Context, resp *http.Response, meta *SupplierMeta) (*Usage, *StreamChunk, error)

	// GetModelList returns available models for this adaptor
	GetModelList() []string

	// GetChannelName returns the adaptor/channel name
	GetChannelName() string
}

// ImageRequest represents an image generation request
type ImageRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	N      int    `json:"n,omitempty"`
	Size   string `json:"size,omitempty"`
}

// AudioRequest represents a text-to-speech request
type AudioRequest struct {
	Model  string `json:"model"`
	Input  string `json:"input"`
	Voice  string `json:"voice"`
	Speed  float64 `json:"speed,omitempty"`
	Format string `json:"response_format,omitempty"`
}

// APIType represents the type of API
type APIType string

const (
	APITypeOpenAI    APIType = "openai"
	APITypeAnthropic APIType = "anthropic"
	APITypeGemini    APIType = "gemini"
)

// GetAdaptor returns the appropriate adaptor for the given API type
func GetAdaptor(apiType APIType) Adaptor {
	switch apiType {
	case APITypeOpenAI:
		return &OpenAIAdaptor{}
	case APITypeAnthropic:
		return &AnthropicAdaptor{}
	case APITypeGemini:
		return &GeminiAdaptor{}
	default:
		return &OpenAIAdaptor{} // Default to OpenAI-compatible
	}
}

// --- OpenAI Adaptor ---

// OpenAIAdaptor implements the Adaptor interface for OpenAI-compatible APIs
type OpenAIAdaptor struct {
	meta *SupplierMeta
}

func (a *OpenAIAdaptor) Init(meta *SupplierMeta) {
	a.meta = meta
}

func (a *OpenAIAdaptor) GetChannelName() string { return "openai" }

func (a *OpenAIAdaptor) GetRequestURL(meta *SupplierMeta) (string, error) {
	return meta.APIBaseURL + "/v1/chat/completions", nil
}

func (a *OpenAIAdaptor) SetupRequestHeader(req *http.Request, meta *SupplierMeta) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)
	return nil
}

func (a *OpenAIAdaptor) ConvertRequest(req *Request) ([]byte, error) {
	// OpenAI format is the canonical format, pass through
	return jsonMarshal(req)
}

func (a *OpenAIAdaptor) ConvertImageRequest(req *ImageRequest) ([]byte, error) {
	return jsonMarshal(req)
}

func (a *OpenAIAdaptor) DoRequest(ctx context.Context, meta *SupplierMeta, body io.Reader) (*http.Response, error) {
	url, err := a.GetRequestURL(meta)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := a.SetupRequestHeader(httpReq, meta); err != nil {
		return nil, fmt.Errorf("setup header: %w", err)
	}
	client := &http.Client{Timeout: 120 * time.Second}
	return client.Do(httpReq)
}

func (a *OpenAIAdaptor) DoResponse(ctx context.Context, resp *http.Response, meta *SupplierMeta) (*Usage, *StreamChunk, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		apiErr := ParseUpstreamError(body, APITypeOpenAI)
		return nil, nil, apiErr
	}

	var openaiResp Response
	if err := jsonUnmarshal(body, &openaiResp); err != nil {
		return nil, nil, fmt.Errorf("unmarshal response: %w", err)
	}
	return &openaiResp.Usage, nil, nil
}

func (a *OpenAIAdaptor) GetModelList() []string {
	return []string{
		"gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-4", "gpt-3.5-turbo",
		"o1", "o1-mini", "o1-pro", "o3-mini",
	}
}

// --- Anthropic Adaptor ---

// AnthropicAdaptor implements the Adaptor interface for Anthropic API
type AnthropicAdaptor struct {
	meta *SupplierMeta
}

func (a *AnthropicAdaptor) Init(meta *SupplierMeta) {
	a.meta = meta
}

func (a *AnthropicAdaptor) GetChannelName() string { return "anthropic" }

func (a *AnthropicAdaptor) GetRequestURL(meta *SupplierMeta) (string, error) {
	return meta.APIBaseURL + "/v1/messages", nil
}

func (a *AnthropicAdaptor) SetupRequestHeader(req *http.Request, meta *SupplierMeta) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", meta.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	return nil
}

func (a *AnthropicAdaptor) ConvertRequest(req *Request) ([]byte, error) {
	// If raw Anthropic-format messages are provided (client sent native Anthropic request),
	// pass them through directly without re-conversion.
	if len(req.RawMessages) > 0 {
		anthropicReq := map[string]any{
			"model":      req.Model,
			"max_tokens": req.MaxTokens,
			"messages":   json.RawMessage(req.RawMessages),
		}
		if req.SystemPrompt != "" {
			anthropicReq["system"] = req.SystemPrompt
		}
		if req.Stream {
			anthropicReq["stream"] = true
		}
		if req.Temperature > 0 {
			anthropicReq["temperature"] = req.Temperature
		}
		if req.TopP > 0 {
			anthropicReq["top_p"] = req.TopP
		}
		return jsonMarshal(anthropicReq)
	}

	// Otherwise, convert OpenAI format to Anthropic format.
	// Anthropic separates system prompt from messages.
	var systemContent string
	var messages []map[string]any

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemContent = msg.Content
			continue
		}
		messages = append(messages, map[string]any{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	anthropicReq := map[string]any{
		"model":      req.Model,
		"max_tokens": req.MaxTokens,
		"messages":   messages,
	}
	if systemContent != "" {
		anthropicReq["system"] = systemContent
	}
	if req.Stream {
		anthropicReq["stream"] = true
	}
	if req.Temperature > 0 {
		anthropicReq["temperature"] = req.Temperature
	}
	if req.TopP > 0 {
		anthropicReq["top_p"] = req.TopP
	}

	return jsonMarshal(anthropicReq)
}

func (a *AnthropicAdaptor) ConvertImageRequest(req *ImageRequest) ([]byte, error) {
	return nil, fmt.Errorf("anthropic does not support image generation")
}

func (a *AnthropicAdaptor) DoRequest(ctx context.Context, meta *SupplierMeta, body io.Reader) (*http.Response, error) {
	url, err := a.GetRequestURL(meta)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := a.SetupRequestHeader(httpReq, meta); err != nil {
		return nil, fmt.Errorf("setup header: %w", err)
	}
	client := &http.Client{Timeout: 120 * time.Second}
	return client.Do(httpReq)
}

func (a *AnthropicAdaptor) DoResponse(ctx context.Context, resp *http.Response, meta *SupplierMeta) (*Usage, *StreamChunk, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		apiErr := ParseUpstreamError(body, APITypeAnthropic)
		return nil, nil, apiErr
	}

	// Parse Anthropic response
	var anthropicResp map[string]any
	if err := jsonUnmarshal(body, &anthropicResp); err != nil {
		return nil, nil, fmt.Errorf("unmarshal response: %w", err)
	}

	usage := &Usage{}
	if u, ok := anthropicResp["usage"].(map[string]any); ok {
		if pt, ok := u["input_tokens"].(float64); ok {
			usage.PromptTokens = int(pt)
		}
		if ct, ok := u["output_tokens"].(float64); ok {
			usage.CompletionTokens = int(ct)
		}
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	return usage, nil, nil
}

func (a *AnthropicAdaptor) GetModelList() []string {
	return []string{
		"claude-opus-4-20250514", "claude-sonnet-4-20250514",
		"claude-3-5-haiku-20241022", "claude-3-haiku-20240307",
	}
}

// --- Gemini Adaptor ---

// GeminiAdaptor implements the Adaptor interface for Google Gemini API
type GeminiAdaptor struct {
	meta *SupplierMeta
}

func (a *GeminiAdaptor) Init(meta *SupplierMeta) {
	a.meta = meta
}

func (a *GeminiAdaptor) GetChannelName() string { return "gemini" }

func (a *GeminiAdaptor) GetRequestURL(meta *SupplierMeta) (string, error) {
	// Gemini uses model in URL path
	return meta.APIBaseURL + "/v1beta/models/" + meta.Model + ":generateContent", nil
}

func (a *GeminiAdaptor) SetupRequestHeader(req *http.Request, meta *SupplierMeta) error {
	req.Header.Set("Content-Type", "application/json")
	// Gemini uses API key as query param, but we also support Bearer
	req.Header.Set("x-goog-api-key", meta.APIKey)
	return nil
}

func (a *GeminiAdaptor) ConvertRequest(req *Request) ([]byte, error) {
	// Convert OpenAI format to Gemini format
	var contents []map[string]any
	var systemInstruction map[string]any

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemInstruction = map[string]any{
				"parts": []map[string]any{{"text": msg.Content}},
			}
			continue
		}
		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}
		contents = append(contents, map[string]any{
			"role":  role,
			"parts": []map[string]any{{"text": msg.Content}},
		})
	}

	geminiReq := map[string]any{
		"contents": contents,
	}
	if systemInstruction != nil {
		geminiReq["systemInstruction"] = systemInstruction
	}

	// Generation config
	genConfig := map[string]any{}
	if req.MaxTokens > 0 {
		genConfig["maxOutputTokens"] = req.MaxTokens
	}
	if req.Temperature > 0 {
		genConfig["temperature"] = req.Temperature
	}
	if req.TopP > 0 {
		genConfig["topP"] = req.TopP
	}
	if len(genConfig) > 0 {
		geminiReq["generationConfig"] = genConfig
	}

	return jsonMarshal(geminiReq)
}

func (a *GeminiAdaptor) ConvertImageRequest(req *ImageRequest) ([]byte, error) {
	return nil, fmt.Errorf("gemini does not support image generation via this endpoint")
}

func (a *GeminiAdaptor) DoRequest(ctx context.Context, meta *SupplierMeta, body io.Reader) (*http.Response, error) {
	url, err := a.GetRequestURL(meta)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := a.SetupRequestHeader(httpReq, meta); err != nil {
		return nil, fmt.Errorf("setup header: %w", err)
	}
	client := &http.Client{Timeout: 120 * time.Second}
	return client.Do(httpReq)
}

func (a *GeminiAdaptor) DoResponse(ctx context.Context, resp *http.Response, meta *SupplierMeta) (*Usage, *StreamChunk, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		apiErr := ParseUpstreamError(body, APITypeGemini)
		return nil, nil, apiErr
	}

	// Parse Gemini response
	var geminiResp map[string]any
	if err := jsonUnmarshal(body, &geminiResp); err != nil {
		return nil, nil, fmt.Errorf("unmarshal response: %w", err)
	}

	usage := &Usage{}
	if um, ok := geminiResp["usageMetadata"].(map[string]any); ok {
		if pt, ok := um["promptTokenCount"].(float64); ok {
			usage.PromptTokens = int(pt)
		}
		if ct, ok := um["candidatesTokenCount"].(float64); ok {
			usage.CompletionTokens = int(ct)
		}
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	return usage, nil, nil
}

func (a *GeminiAdaptor) GetModelList() []string {
	return []string{
		"gemini-2.0-flash", "gemini-2.0-flash-lite", "gemini-1.5-pro", "gemini-1.5-flash",
	}
}

// --- Helper for streaming request building ---

// BuildStreamRequest creates an HTTP request for streaming relay
func BuildStreamRequest(ctx context.Context, adaptor Adaptor, meta *SupplierMeta, reqBody io.Reader) (*http.Request, error) {
	url, err := adaptor.GetRequestURL(meta)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := adaptor.SetupRequestHeader(httpReq, meta); err != nil {
		return nil, fmt.Errorf("setup header: %w", err)
	}
	return httpReq, nil
}

// ForwardSSE reads from src and writes SSE events to dst until [DONE] or EOF.
// Returns the total bytes written.
func ForwardSSE(dst io.Writer, src io.Reader) (int64, error) {
	buf := make([]byte, 4096)
	var total int64

	for {
		n, err := src.Read(buf)
		if n > 0 {
			data := buf[:n]
			// Check for [DONE] marker
			if bytes.Contains(data, []byte("data: [DONE]")) {
				// Write up to and including the [DONE]
				idx := bytes.Index(data, []byte("data: [DONE]"))
				end := idx + len("data: [DONE]")
				if end > len(data) {
					end = len(data)
				}
				written, wErr := dst.Write(data[:end])
				total += int64(written)
				if wErr != nil {
					return total, wErr
				}
				return total, nil
			}
			written, wErr := dst.Write(data)
			total += int64(written)
			if wErr != nil {
				return total, wErr
			}
		}
		if err != nil {
			if err == io.EOF {
				return total, nil
			}
			return total, err
		}
	}
}
