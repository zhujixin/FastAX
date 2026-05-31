package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- OpenAI Adaptor Tests ---

func TestOpenAIAdaptor_ConvertRequest(t *testing.T) {
	a := &OpenAIAdaptor{}
	req := &Request{
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "hello"}},
		Stream:   false,
	}

	body, err := a.ConvertRequest(req)
	if err != nil {
		t.Fatalf("ConvertRequest error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if result["model"] != "gpt-4" {
		t.Errorf("expected model gpt-4, got %v", result["model"])
	}
}

func TestOpenAIAdaptor_GetRequestURL(t *testing.T) {
	a := &OpenAIAdaptor{}
	meta := &SupplierMeta{APIBaseURL: "https://api.openai.com"}
	url, err := a.GetRequestURL(meta)
	if err != nil {
		t.Fatalf("GetRequestURL error: %v", err)
	}
	expected := "https://api.openai.com/v1/chat/completions"
	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}

func TestOpenAIAdaptor_SetupRequestHeader(t *testing.T) {
	a := &OpenAIAdaptor{}
	meta := &SupplierMeta{APIKey: "test-key-123"}
	req, _ := http.NewRequest("POST", "http://example.com", nil)

	if err := a.SetupRequestHeader(req, meta); err != nil {
		t.Fatalf("SetupRequestHeader error: %v", err)
	}
	if req.Header.Get("Authorization") != "Bearer test-key-123" {
		t.Errorf("expected Bearer test-key-123, got %s", req.Header.Get("Authorization"))
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("expected application/json, got %s", req.Header.Get("Content-Type"))
	}
}

func TestOpenAIAdaptor_DoResponse_Success(t *testing.T) {
	a := &OpenAIAdaptor{}
	respBody := `{
		"id": "chatcmpl-123",
		"object": "chat.completion",
		"model": "gpt-4",
		"choices": [{"index": 0, "message": {"role": "assistant", "content": "hi"}, "finish_reason": "stop"}],
		"usage": {"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15}
	}`
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(respBody))),
	}

	usage, _, err := a.DoResponse(context.Background(), resp, &SupplierMeta{})
	if err != nil {
		t.Fatalf("DoResponse error: %v", err)
	}
	if usage.PromptTokens != 10 {
		t.Errorf("expected 10 prompt tokens, got %d", usage.PromptTokens)
	}
	if usage.CompletionTokens != 5 {
		t.Errorf("expected 5 completion tokens, got %d", usage.CompletionTokens)
	}
}

func TestOpenAIAdaptor_DoResponse_Error(t *testing.T) {
	a := &OpenAIAdaptor{}
	respBody := `{"error": {"message": "Invalid API key", "type": "invalid_request_error", "code": "invalid_api_key"}}`
	resp := &http.Response{
		StatusCode: 401,
		Body:       io.NopCloser(bytes.NewReader([]byte(respBody))),
	}

	_, _, err := a.DoResponse(context.Background(), resp, &SupplierMeta{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Code != "invalid_api_key" {
		t.Errorf("expected code invalid_api_key, got %s", apiErr.Code)
	}
}

func TestOpenAIAdaptor_GetModelList(t *testing.T) {
	a := &OpenAIAdaptor{}
	models := a.GetModelList()
	if len(models) == 0 {
		t.Error("expected non-empty model list")
	}
	found := false
	for _, m := range models {
		if m == "gpt-4o" {
			found = true
		}
	}
	if !found {
		t.Error("expected gpt-4o in model list")
	}
}

func TestOpenAIAdaptor_GetChannelName(t *testing.T) {
	a := &OpenAIAdaptor{}
	if a.GetChannelName() != "openai" {
		t.Errorf("expected openai, got %s", a.GetChannelName())
	}
}

// --- Anthropic Adaptor Tests ---

func TestAnthropicAdaptor_ConvertRequest(t *testing.T) {
	a := &AnthropicAdaptor{}
	req := &Request{
		Model:    "claude-sonnet-4-20250514",
		Messages: []Message{{Role: "user", Content: "hello"}},
		MaxTokens: 1024,
	}

	body, err := a.ConvertRequest(req)
	if err != nil {
		t.Fatalf("ConvertRequest error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if result["model"] != "claude-sonnet-4-20250514" {
		t.Errorf("expected model claude-sonnet-4-20250514, got %v", result["model"])
	}
	if result["max_tokens"] != float64(1024) {
		t.Errorf("expected max_tokens 1024, got %v", result["max_tokens"])
	}
}

func TestAnthropicAdaptor_ConvertRequest_RawMessages(t *testing.T) {
	a := &AnthropicAdaptor{}
	rawMsgs := json.RawMessage(`[{"role": "user", "content": [{"type": "text", "text": "hello"}]}]`)
	req := &Request{
		Model:       "claude-sonnet-4-20250514",
		MaxTokens:   1024,
		RawMessages: rawMsgs,
		SystemPrompt: "Be helpful",
	}

	body, err := a.ConvertRequest(req)
	if err != nil {
		t.Fatalf("ConvertRequest error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if result["model"] != "claude-sonnet-4-20250514" {
		t.Errorf("expected model claude-sonnet-4-20250514, got %v", result["model"])
	}
	if result["system"] != "Be helpful" {
		t.Errorf("expected system 'Be helpful', got %v", result["system"])
	}
	// Messages should be passed through as-is (content blocks preserved)
	msgs := result["messages"].([]any)
	if len(msgs) != 1 {
		t.Errorf("expected 1 message, got %d", len(msgs))
	}
	msg := msgs[0].(map[string]any)
	content := msg["content"].([]any)
	if len(content) != 1 {
		t.Errorf("expected 1 content block, got %d", len(content))
	}
	block := content[0].(map[string]any)
	if block["type"] != "text" || block["text"] != "hello" {
		t.Errorf("expected text block 'hello', got %v", block)
	}
}

func TestAnthropicAdaptor_ConvertRequest_RawMessages_Stream(t *testing.T) {
	a := &AnthropicAdaptor{}
	rawMsgs := json.RawMessage(`[{"role": "user", "content": "hi"}]`)
	req := &Request{
		Model:       "claude-sonnet-4-20250514",
		Stream:      true,
		MaxTokens:   512,
		RawMessages: rawMsgs,
	}

	body, err := a.ConvertRequest(req)
	if err != nil {
		t.Fatalf("ConvertRequest error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if result["stream"] != true {
		t.Errorf("expected stream true, got %v", result["stream"])
	}
	if result["max_tokens"] != float64(512) {
		t.Errorf("expected max_tokens 512, got %v", result["max_tokens"])
	}
}

func TestAnthropicAdaptor_ConvertRequest_WithSystem(t *testing.T) {
	a := &AnthropicAdaptor{}
	req := &Request{
		Model: "claude-sonnet-4-20250514",
		Messages: []Message{
			{Role: "system", Content: "You are helpful"},
			{Role: "user", Content: "hello"},
		},
		MaxTokens: 512,
	}

	body, err := a.ConvertRequest(req)
	if err != nil {
		t.Fatalf("ConvertRequest error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if result["system"] != "You are helpful" {
		t.Errorf("expected system prompt, got %v", result["system"])
	}
	msgs := result["messages"].([]any)
	if len(msgs) != 1 {
		t.Errorf("expected 1 message (system excluded), got %d", len(msgs))
	}
}

func TestAnthropicAdaptor_SetupRequestHeader(t *testing.T) {
	a := &AnthropicAdaptor{}
	meta := &SupplierMeta{APIKey: "sk-ant-test"}
	req, _ := http.NewRequest("POST", "http://example.com", nil)

	if err := a.SetupRequestHeader(req, meta); err != nil {
		t.Fatalf("SetupRequestHeader error: %v", err)
	}
	if req.Header.Get("x-api-key") != "sk-ant-test" {
		t.Errorf("expected sk-ant-test, got %s", req.Header.Get("x-api-key"))
	}
	if req.Header.Get("anthropic-version") != "2023-06-01" {
		t.Errorf("expected 2023-06-01, got %s", req.Header.Get("anthropic-version"))
	}
}

func TestAnthropicAdaptor_DoResponse_Success(t *testing.T) {
	a := &AnthropicAdaptor{}
	respBody := `{
		"id": "msg_123",
		"type": "message",
		"model": "claude-sonnet-4-20250514",
		"content": [{"type": "text", "text": "Hello!"}],
		"usage": {"input_tokens": 10, "output_tokens": 5}
	}`
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(respBody))),
	}

	usage, _, err := a.DoResponse(context.Background(), resp, &SupplierMeta{})
	if err != nil {
		t.Fatalf("DoResponse error: %v", err)
	}
	if usage.PromptTokens != 10 {
		t.Errorf("expected 10 prompt tokens, got %d", usage.PromptTokens)
	}
	if usage.CompletionTokens != 5 {
		t.Errorf("expected 5 completion tokens, got %d", usage.CompletionTokens)
	}
}

func TestAnthropicAdaptor_DoResponse_Error(t *testing.T) {
	a := &AnthropicAdaptor{}
	respBody := `{"type": "error", "error": {"type": "authentication_error", "message": "Invalid API key"}}`
	resp := &http.Response{
		StatusCode: 401,
		Body:       io.NopCloser(bytes.NewReader([]byte(respBody))),
	}

	_, _, err := a.DoResponse(context.Background(), resp, &SupplierMeta{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Type != "authentication_error" {
		t.Errorf("expected type authentication_error, got %s", apiErr.Type)
	}
}

func TestAnthropicAdaptor_GetRequestURL(t *testing.T) {
	a := &AnthropicAdaptor{}
	meta := &SupplierMeta{APIBaseURL: "https://api.anthropic.com"}
	url, err := a.GetRequestURL(meta)
	if err != nil {
		t.Fatalf("GetRequestURL error: %v", err)
	}
	expected := "https://api.anthropic.com/v1/messages"
	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}

// --- Gemini Adaptor Tests ---

func TestGeminiAdaptor_ConvertRequest(t *testing.T) {
	a := &GeminiAdaptor{}
	req := &Request{
		Model:    "gemini-2.0-flash",
		Messages: []Message{{Role: "user", Content: "hello"}},
		MaxTokens: 1024,
	}

	body, err := a.ConvertRequest(req)
	if err != nil {
		t.Fatalf("ConvertRequest error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	contents := result["contents"].([]any)
	if len(contents) != 1 {
		t.Errorf("expected 1 content, got %d", len(contents))
	}
	genConfig := result["generationConfig"].(map[string]any)
	if genConfig["maxOutputTokens"] != float64(1024) {
		t.Errorf("expected maxOutputTokens 1024, got %v", genConfig["maxOutputTokens"])
	}
}

func TestGeminiAdaptor_ConvertRequest_WithSystem(t *testing.T) {
	a := &GeminiAdaptor{}
	req := &Request{
		Model: "gemini-2.0-flash",
		Messages: []Message{
			{Role: "system", Content: "You are helpful"},
			{Role: "user", Content: "hello"},
		},
	}

	body, err := a.ConvertRequest(req)
	if err != nil {
		t.Fatalf("ConvertRequest error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	si := result["systemInstruction"].(map[string]any)
	parts := si["parts"].([]any)
	first := parts[0].(map[string]any)
	if first["text"] != "You are helpful" {
		t.Errorf("expected system prompt, got %v", first["text"])
	}
}

func TestGeminiAdaptor_GetRequestURL(t *testing.T) {
	a := &GeminiAdaptor{}
	meta := &SupplierMeta{APIBaseURL: "https://generativelanguage.googleapis.com", Model: "gemini-2.0-flash"}
	url, err := a.GetRequestURL(meta)
	if err != nil {
		t.Fatalf("GetRequestURL error: %v", err)
	}
	expected := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"
	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}

func TestGeminiAdaptor_DoResponse_Success(t *testing.T) {
	a := &GeminiAdaptor{}
	respBody := `{
		"candidates": [{"content": {"parts": [{"text": "Hello!"}]}}],
		"usageMetadata": {"promptTokenCount": 10, "candidatesTokenCount": 5, "totalTokenCount": 15}
	}`
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(respBody))),
	}

	usage, _, err := a.DoResponse(context.Background(), resp, &SupplierMeta{})
	if err != nil {
		t.Fatalf("DoResponse error: %v", err)
	}
	if usage.PromptTokens != 10 {
		t.Errorf("expected 10 prompt tokens, got %d", usage.PromptTokens)
	}
	if usage.CompletionTokens != 5 {
		t.Errorf("expected 5 completion tokens, got %d", usage.CompletionTokens)
	}
}

func TestGeminiAdaptor_DoResponse_Error(t *testing.T) {
	a := &GeminiAdaptor{}
	respBody := `{"error": {"code": 400, "message": "Invalid request", "status": "INVALID_ARGUMENT"}}`
	resp := &http.Response{
		StatusCode: 400,
		Body:       io.NopCloser(bytes.NewReader([]byte(respBody))),
	}

	_, _, err := a.DoResponse(context.Background(), resp, &SupplierMeta{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Type != "INVALID_ARGUMENT" {
		t.Errorf("expected type INVALID_ARGUMENT, got %s", apiErr.Type)
	}
}

func TestGeminiAdaptor_GetChannelName(t *testing.T) {
	a := &GeminiAdaptor{}
	if a.GetChannelName() != "gemini" {
		t.Errorf("expected gemini, got %s", a.GetChannelName())
	}
}

// --- GetAdaptor Factory Tests ---

func TestGetAdaptor_OpenAI(t *testing.T) {
	a := GetAdaptor(APITypeOpenAI)
	if a.GetChannelName() != "openai" {
		t.Errorf("expected openai, got %s", a.GetChannelName())
	}
}

func TestGetAdaptor_Anthropic(t *testing.T) {
	a := GetAdaptor(APITypeAnthropic)
	if a.GetChannelName() != "anthropic" {
		t.Errorf("expected anthropic, got %s", a.GetChannelName())
	}
}

func TestGetAdaptor_Gemini(t *testing.T) {
	a := GetAdaptor(APITypeGemini)
	if a.GetChannelName() != "gemini" {
		t.Errorf("expected gemini, got %s", a.GetChannelName())
	}
}

func TestGetAdaptor_Unknown_DefaultsOpenAI(t *testing.T) {
	a := GetAdaptor(APIType("unknown"))
	if a.GetChannelName() != "openai" {
		t.Errorf("expected openai for unknown type, got %s", a.GetChannelName())
	}
}

// --- DoRequest Integration Test (using httptest) ---

func TestOpenAIAdaptor_DoRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Bearer test-key, got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json, got %s", r.Header.Get("Content-Type"))
		}
		// Return a valid response
		w.WriteHeader(200)
		w.Write([]byte(`{"id":"test","choices":[],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
	}))
	defer server.Close()

	a := &OpenAIAdaptor{}
	meta := &SupplierMeta{
		APIBaseURL: server.URL,
		APIKey:     "test-key",
	}

	body := bytes.NewReader([]byte(`{"model":"gpt-4","messages":[]}`))
	resp, err := a.DoRequest(context.Background(), meta, body)
	if err != nil {
		t.Fatalf("DoRequest error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAnthropicAdaptor_DoRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "sk-ant-test" {
			t.Errorf("expected sk-ant-test, got %s", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("expected 2023-06-01, got %s", r.Header.Get("anthropic-version"))
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"type":"message","content":[],"usage":{"input_tokens":1,"output_tokens":1}}`))
	}))
	defer server.Close()

	a := &AnthropicAdaptor{}
	meta := &SupplierMeta{
		APIBaseURL: server.URL,
		APIKey:     "sk-ant-test",
	}

	body := bytes.NewReader([]byte(`{"model":"claude","messages":[]}`))
	resp, err := a.DoRequest(context.Background(), meta, body)
	if err != nil {
		t.Fatalf("DoRequest error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// --- ForwardSSE Tests ---

func TestForwardSSE_NormalFlow(t *testing.T) {
	input := "data: {\"choices\":[{\"delta\":{\"content\":\"Hi\"}}]}\n\ndata: {\"choices\":[{\"delta\":{\"content\":\" there\"}}]}\n\ndata: [DONE]\n\n"
	src := bytes.NewReader([]byte(input))
	dst := &bytes.Buffer{}

	n, err := ForwardSSE(dst, src)
	if err != nil {
		t.Fatalf("ForwardSSE error: %v", err)
	}
	if n == 0 {
		t.Error("expected non-zero bytes written")
	}
	result := dst.String()
	if !bytes.Contains([]byte(result), []byte("data: [DONE]")) {
		t.Error("expected [DONE] in output")
	}
}

func TestForwardSSE_EmptyInput(t *testing.T) {
	src := bytes.NewReader([]byte(""))
	dst := &bytes.Buffer{}

	_, err := ForwardSSE(dst, src)
	if err != nil {
		t.Fatalf("ForwardSSE error: %v", err)
	}
	if dst.Len() != 0 {
		t.Errorf("expected empty output, got %d bytes", dst.Len())
	}
}

// --- Error Parsing Tests ---

func TestParseUpstreamError_OpenAI(t *testing.T) {
	body := []byte(`{"error": {"message": "Invalid key", "type": "invalid_request_error", "code": "invalid_api_key"}}`)
	err := ParseUpstreamError(body, APITypeOpenAI)
	if err.Message != "Invalid key" {
		t.Errorf("expected 'Invalid key', got '%s'", err.Message)
	}
	if err.Code != "invalid_api_key" {
		t.Errorf("expected 'invalid_api_key', got '%s'", err.Code)
	}
}

func TestParseUpstreamError_Anthropic(t *testing.T) {
	body := []byte(`{"type": "error", "error": {"type": "authentication_error", "message": "Invalid API key"}}`)
	err := ParseUpstreamError(body, APITypeAnthropic)
	if err.Type != "authentication_error" {
		t.Errorf("expected 'authentication_error', got '%s'", err.Type)
	}
}

func TestParseUpstreamError_Gemini(t *testing.T) {
	body := []byte(`{"error": {"code": 400, "message": "Bad request", "status": "INVALID_ARGUMENT"}}`)
	err := ParseUpstreamError(body, APITypeGemini)
	if err.Code != "400" {
		t.Errorf("expected '400', got '%s'", err.Code)
	}
	if err.Type != "INVALID_ARGUMENT" {
		t.Errorf("expected 'INVALID_ARGUMENT', got '%s'", err.Type)
	}
}

func TestShouldRetryHTTP(t *testing.T) {
	tests := []struct {
		code     int
		expected bool
	}{
		{200, false},
		{400, false},
		{401, false},
		{403, false},
		{429, true},
		{500, true},
		{502, true},
		{503, true},
	}
	for _, tt := range tests {
		if got := ShouldRetryHTTP(tt.code); got != tt.expected {
			t.Errorf("ShouldRetryHTTP(%d) = %v, want %v", tt.code, got, tt.expected)
		}
	}
}

func TestHTTPStatusFromAPIError(t *testing.T) {
	tests := []struct {
		code     string
		expected int
	}{
		{"401", 401},
		{"invalid_api_key", 401},
		{"429", 429},
		{"rate_limit_exceeded", 429},
		{"insufficient_quota", 402},
		{"unknown", 502},
	}
	for _, tt := range tests {
		err := &APIError{Code: tt.code}
		if got := HTTPStatusFromAPIError(err); got != tt.expected {
			t.Errorf("HTTPStatusFromAPIError(%s) = %d, want %d", tt.code, got, tt.expected)
		}
	}
}

func TestAPIError_Error(t *testing.T) {
	e := &APIError{Message: "test", Type: "err", Code: "42"}
	s := e.Error()
	if s == "" {
		t.Error("expected non-empty error string")
	}
}
