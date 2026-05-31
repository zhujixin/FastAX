package proxy

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fastax/fastax-server/internal/domain/plugin"
	"github.com/fastax/fastax-server/internal/domain/proxy/relay"
	"github.com/fastax/fastax-server/internal/shared/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Supplier{}, &model.Ability{},
		&model.TokenProduct{}, &model.ProviderHealth{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestCircuitBreaker_ClosedToOpen(t *testing.T) {
	cb := NewCircuitBreaker(3, 2, 5*time.Second)

	// Should start closed
	if !cb.Allow() {
		t.Fatal("should allow in closed state")
	}

	// Record 3 failures → open
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.Allow() {
		t.Fatal("should not allow in open state")
	}
	if cb.GetState() != CircuitOpen {
		t.Errorf("state = %v, want Open", cb.GetState())
	}
}

func TestCircuitBreaker_HalfOpenRecovery(t *testing.T) {
	cb := NewCircuitBreaker(2, 2, 50*time.Millisecond)

	cb.RecordFailure()
	cb.RecordFailure()
	if cb.GetState() != CircuitOpen {
		t.Fatal("should be open")
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	if !cb.Allow() {
		t.Fatal("should allow in half-open state")
	}
	if cb.GetState() != CircuitHalfOpen {
		t.Errorf("state = %v, want HalfOpen", cb.GetState())
	}

	// Two successes → closed
	cb.RecordSuccess()
	cb.RecordSuccess()
	if cb.GetState() != CircuitClosed {
		t.Errorf("state = %v, want Closed", cb.GetState())
	}
}

func TestCircuitBreaker_HalfOpenFailure(t *testing.T) {
	cb := NewCircuitBreaker(2, 2, 50*time.Millisecond)

	cb.RecordFailure()
	cb.RecordFailure()
	time.Sleep(60 * time.Millisecond)

	cb.Allow() // → half-open
	cb.RecordFailure()
	if cb.GetState() != CircuitOpen {
		t.Errorf("state = %v, want Open", cb.GetState())
	}
}

func TestShouldDisableChannel(t *testing.T) {
	tests := []struct {
		code     int
		expected bool
	}{
		{200, false},
		{400, false},
		{401, false},
		{403, false},
		{429, false},
		{500, true},
		{502, true},
		{503, true},
	}
	for _, tt := range tests {
		if got := ShouldDisableChannel(tt.code); got != tt.expected {
			t.Errorf("ShouldDisableChannel(%d) = %v, want %v", tt.code, got, tt.expected)
		}
	}
}

// --- Router Tests ---

func TestRouter_SelectChannel(t *testing.T) {
	db := setupTestDB(t)
	router := NewRouter(db)

	// Create suppliers
	db.Create(&model.Supplier{Code: "s1", Name: "S1", APIBaseURL: "https://s1.com", APIKeyEncrypted: "k", Priority: 10, Weight: 10, Status: 1})
	db.Create(&model.Supplier{Code: "s2", Name: "S2", APIBaseURL: "https://s2.com", APIKeyEncrypted: "k", Priority: 5, Weight: 10, Status: 1})

	// Create abilities
	db.Create(&model.Ability{Group: "default", Model: "gpt-4", ChannelID: 1})
	db.Create(&model.Ability{Group: "default", Model: "gpt-4", ChannelID: 2})

	router.LoadChannels()

	ch, err := router.SelectChannel("default", "gpt-4", nil)
	if err != nil {
		t.Fatalf("SelectChannel() error = %v", err)
	}
	// Should prefer higher priority
	if ch.Priority != 10 {
		t.Errorf("priority = %v, want 10", ch.Priority)
	}
}

func TestRouter_SelectChannel_NoMatch(t *testing.T) {
	db := setupTestDB(t)
	router := NewRouter(db)

	_, err := router.SelectChannel("default", "nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for no matching channel")
	}
}

func TestRouter_SelectChannel_Disabled(t *testing.T) {
	db := setupTestDB(t)
	router := NewRouter(db)

	db.Create(&model.Supplier{Code: "s1", Name: "S1", APIBaseURL: "https://s1.com", APIKeyEncrypted: "k", Priority: 10, Weight: 10, Status: 1})
	db.Create(&model.Ability{Group: "default", Model: "gpt-4", ChannelID: 1})
	router.LoadChannels()

	disabled := map[uint]bool{1: true}
	_, err := router.SelectChannel("default", "gpt-4", disabled)
	if err == nil {
		t.Fatal("expected error when all channels disabled")
	}
}

func TestRouter_SelectChannel_DisabledSupplier(t *testing.T) {
	db := setupTestDB(t)
	router := NewRouter(db)

	// Create as enabled, then disable (GORM default:1 overrides zero value)
	supplier := model.Supplier{Code: "s1", Name: "S1", APIBaseURL: "https://s1.com", APIKeyEncrypted: "k", Priority: 10, Weight: 10, Status: 1}
	db.Create(&supplier)
	db.Model(&supplier).Update("status", 0)
	db.Create(&model.Ability{Group: "default", Model: "gpt-4", ChannelID: 1})
	router.LoadChannels()

	_, err := router.SelectChannel("default", "gpt-4", nil)
	if err == nil {
		t.Fatal("expected error for disabled supplier")
	}
}

func TestRouter_WeightedRandom(t *testing.T) {
	db := setupTestDB(t)
	router := NewRouter(db)

	db.Create(&model.Supplier{Code: "s1", Name: "S1", APIBaseURL: "https://s1.com", APIKeyEncrypted: "k", Priority: 0, Weight: 100, Status: 1})
	db.Create(&model.Supplier{Code: "s2", Name: "S2", APIBaseURL: "https://s2.com", APIKeyEncrypted: "k", Priority: 0, Weight: 1, Status: 1})
	db.Create(&model.Ability{Group: "default", Model: "gpt-4", ChannelID: 1})
	db.Create(&model.Ability{Group: "default", Model: "gpt-4", ChannelID: 2})
	router.LoadChannels()

	// With high weight difference, channel 1 should be selected most of the time
	count1 := 0
	for range 100 {
		ch, _ := router.SelectChannel("default", "gpt-4", nil)
		if ch.ChannelID == 1 {
			count1++
		}
	}
	if count1 < 50 {
		t.Errorf("channel 1 selected %d/100 times, expected >50", count1)
	}
}

// --- Service Integration Tests ---

func setupTestService(t *testing.T, upstreamURL string) (*Service, *gorm.DB) {
	t.Helper()
	db := setupTestDB(t)

	// Create a supplier pointing to the test upstream
	db.Create(&model.Supplier{
		Code:            "openai",
		Name:            "OpenAI",
		APIBaseURL:      upstreamURL,
		APIKeyEncrypted: "test-key",
		Priority:        10,
		Weight:          10,
		Status:          1,
	})
	db.Create(&model.Ability{Group: "default", Model: "gpt-4", ChannelID: 1})

	svc := NewService(db)
	return svc, db
}

func TestService_Relay_Success(t *testing.T) {
	// Mock upstream server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if !strings.Contains(r.Header.Get("Authorization"), "test-key") {
			t.Errorf("expected auth header with test-key")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
			"id": "chatcmpl-123",
			"object": "chat.completion",
			"model": "gpt-4",
			"choices": [{"index": 0, "message": {"role": "assistant", "content": "Hello!"}, "finish_reason": "stop"}],
			"usage": {"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15}
		}`))
	}))
	defer server.Close()

	svc, _ := setupTestService(t, server.URL)

	resp, err := svc.Relay(context.Background(), 1, &RelayRequest{
		Model:    "gpt-4",
		Messages: []relay.Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("Relay error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(string(resp.Body), "Hello!") {
		t.Error("expected 'Hello!' in response body")
	}
}

func TestService_Relay_RetryOn5xx(t *testing.T) {
	// Create two upstream servers: first returns 500, second succeeds
	callCount1 := 0
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount1++
		w.WriteHeader(500)
		w.Write([]byte(`{"error": {"message": "internal error"}}`))
	}))
	defer server1.Close()

	callCount2 := 0
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount2++
		w.WriteHeader(200)
		w.Write([]byte(`{"id":"test","choices":[{"message":{"role":"assistant","content":"ok"}}],"usage":{}}`))
	}))
	defer server2.Close()

	db := setupTestDB(t)
	// server1 has higher priority so it's always selected first
	db.Create(&model.Supplier{Code: "s1", Name: "S1", APIBaseURL: server1.URL, APIKeyEncrypted: "k1", Priority: 20, Weight: 10, Status: 1})
	db.Create(&model.Supplier{Code: "s2", Name: "S2", APIBaseURL: server2.URL, APIKeyEncrypted: "k2", Priority: 10, Weight: 10, Status: 1})
	db.Create(&model.Ability{Group: "default", Model: "gpt-4", ChannelID: 1})
	db.Create(&model.Ability{Group: "default", Model: "gpt-4", ChannelID: 2})

	svc := NewService(db)

	resp, err := svc.Relay(context.Background(), 1, &RelayRequest{
		Model:    "gpt-4",
		Messages: []relay.Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("Relay error after retries: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	// First channel should have been tried and failed
	if callCount1 != 1 {
		t.Errorf("expected 1 call to server1, got %d", callCount1)
	}
	// Second channel should have succeeded
	if callCount2 != 1 {
		t.Errorf("expected 1 call to server2, got %d", callCount2)
	}
}

func TestService_Relay_AllRetriesFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(502)
		w.Write([]byte(`{"error": {"message": "bad gateway"}}`))
	}))
	defer server.Close()

	db := setupTestDB(t)
	// Single supplier - all retries hit the same server
	db.Create(&model.Supplier{Code: "s1", Name: "S1", APIBaseURL: server.URL, APIKeyEncrypted: "k", Priority: 10, Weight: 10, Status: 1})
	db.Create(&model.Ability{Group: "default", Model: "gpt-4", ChannelID: 1})

	svc := NewService(db)

	_, err := svc.Relay(context.Background(), 1, &RelayRequest{
		Model:    "gpt-4",
		Messages: []relay.Message{{Role: "user", Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error after all retries fail")
	}
	// With single channel, after first failure it's added to disabled set,
	// so subsequent retries get "no available channel"
	if !strings.Contains(err.Error(), "no available channel") && !strings.Contains(err.Error(), "all retries exhausted") {
		t.Errorf("expected retry error, got: %v", err)
	}
}

func TestService_RelayStream_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"Hi\"}}]}\n\n"))
		w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\" there\"}}]}\n\n"))
		w.(http.Flusher).Flush()
		w.Write([]byte("data: [DONE]\n\n"))
		w.(http.Flusher).Flush()
	}))
	defer server.Close()

	svc, _ := setupTestService(t, server.URL)

	resp, err := svc.RelayStream(context.Background(), 1, &RelayRequest{
		Model:    "gpt-4",
		Messages: []relay.Message{{Role: "user", Content: "hi"}},
		Stream:   true,
	})
	if err != nil {
		t.Fatalf("RelayStream error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "data:") {
		t.Error("expected SSE data in response")
	}
	if !strings.Contains(bodyStr, "[DONE]") {
		t.Error("expected [DONE] marker in response")
	}
}

// --- Anthropic Messages API Tests ---

func setupAnthropicTestService(t *testing.T, upstreamURL string, code string) (*Service, *gorm.DB) {
	t.Helper()
	db := setupTestDB(t)

	db.Create(&model.Supplier{
		Code:            code,
		Name:            "Test Supplier",
		APIBaseURL:      upstreamURL,
		APIKeyEncrypted: "test-ant-key",
		Priority:        10,
		Weight:          10,
		Status:          1,
	})
	db.Create(&model.Ability{Group: "default", Model: "claude-sonnet-4-20250514", ChannelID: 1})

	svc := NewService(db)
	return svc, db
}

func TestService_Relay_AnthropicNativeRequest(t *testing.T) {
	// Mock Anthropic upstream
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Anthropic headers
		if r.Header.Get("x-api-key") != "test-ant-key" {
			t.Errorf("expected x-api-key test-ant-key, got %s", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("expected anthropic-version 2023-06-01, got %s", r.Header.Get("anthropic-version"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
			"id": "msg_123",
			"type": "message",
			"role": "assistant",
			"model": "claude-sonnet-4-20250514",
			"content": [{"type": "text", "text": "Hello from Claude!"}],
			"stop_reason": "end_turn",
			"usage": {"input_tokens": 10, "output_tokens": 5}
		}`))
	}))
	defer server.Close()

	svc, _ := setupAnthropicTestService(t, server.URL, "anthropic")

	// Simulate what ChatMessages handler does: build RelayRequest with raw messages
	rawMsgs := json.RawMessage(`[{"role": "user", "content": "Hi Claude"}]`)
	req := &RelayRequest{
		Model:       "claude-sonnet-4-20250514",
		Stream:      false,
		MaxTokens:   1024,
		RawMessages: rawMsgs,
		Messages:    extractAnthropicMessages(rawMsgs),
	}

	resp, err := svc.Relay(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("Relay error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(string(resp.Body), "Hello from Claude!") {
		t.Error("expected 'Hello from Claude!' in response body")
	}
}

func TestService_Relay_AnthropicWithSystemPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read and verify the request body has system prompt
		body, _ := io.ReadAll(r.Body)
		bodyStr := string(body)
		if !strings.Contains(bodyStr, `"system"`) {
			t.Error("expected system field in request body")
		}
		if !strings.Contains(bodyStr, "You are helpful") {
			t.Error("expected system prompt text in request body")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
			"id": "msg_456",
			"type": "message",
			"role": "assistant",
			"content": [{"type": "text", "text": "Sure!"}],
			"usage": {"input_tokens": 20, "output_tokens": 3}
		}`))
	}))
	defer server.Close()

	svc, _ := setupAnthropicTestService(t, server.URL, "anthropic")

	rawMsgs := json.RawMessage(`[{"role": "user", "content": "Hello"}]`)
	req := &RelayRequest{
		Model:        "claude-sonnet-4-20250514",
		MaxTokens:    512,
		RawMessages:  rawMsgs,
		SystemPrompt: "You are helpful",
		Messages:     extractAnthropicMessages(rawMsgs),
	}

	resp, err := svc.Relay(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("Relay error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestService_RelayStream_AnthropicSSE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		// Anthropic SSE format
		w.Write([]byte("event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"msg_1\",\"type\":\"message\",\"role\":\"assistant\",\"content\":[]}}\n\n"))
		w.(http.Flusher).Flush()
		w.Write([]byte("event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"Hi\"}}\n\n"))
		w.(http.Flusher).Flush()
		w.Write([]byte("event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\" there\"}}\n\n"))
		w.(http.Flusher).Flush()
		w.Write([]byte("event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n"))
		w.(http.Flusher).Flush()
	}))
	defer server.Close()

	svc, _ := setupAnthropicTestService(t, server.URL, "anthropic")

	rawMsgs := json.RawMessage(`[{"role": "user", "content": "Hi"}]`)
	req := &RelayRequest{
		Model:       "claude-sonnet-4-20250514",
		Stream:      true,
		MaxTokens:   1024,
		RawMessages: rawMsgs,
		Messages:    extractAnthropicMessages(rawMsgs),
	}

	resp, err := svc.RelayStream(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("RelayStream error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "message_start") {
		t.Error("expected message_start event in response")
	}
	if !strings.Contains(bodyStr, "content_block_delta") {
		t.Error("expected content_block_delta event in response")
	}
	if !strings.Contains(bodyStr, "\"Hi\"") {
		t.Error("expected 'Hi' in streamed content")
	}
	if !strings.Contains(bodyStr, "\" there\"") {
		t.Error("expected ' there' in streamed content")
	}
	if !strings.Contains(bodyStr, "message_stop") {
		t.Error("expected message_stop event in response")
	}
}

func TestService_Relay_AnthropicContentBlocks(t *testing.T) {
	// Test with Anthropic content block format (array of content objects)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		// Verify messages are passed through as content blocks
		if !strings.Contains(string(body), `"type":"text"`) {
			t.Error("expected content block type:text in request")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"id":"msg_1","type":"message","content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":5,"output_tokens":2}}`))
	}))
	defer server.Close()

	svc, _ := setupAnthropicTestService(t, server.URL, "anthropic")

	// Content block format
	rawMsgs := json.RawMessage(`[{"role": "user", "content": [{"type": "text", "text": "Describe this"}]}]`)
	req := &RelayRequest{
		Model:       "claude-sonnet-4-20250514",
		MaxTokens:   1024,
		RawMessages: rawMsgs,
		Messages:    extractAnthropicMessages(rawMsgs),
	}

	resp, err := svc.Relay(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("Relay error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestService_Relay_AnthropicMultiTurn(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		bodyStr := string(body)
		// Should have both user and assistant messages
		if !strings.Contains(bodyStr, `"user"`) || !strings.Contains(bodyStr, `"assistant"`) {
			t.Error("expected both user and assistant roles in messages")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"id":"msg_1","type":"message","content":[{"type":"text","text":"done"}],"usage":{"input_tokens":20,"output_tokens":5}}`))
	}))
	defer server.Close()

	svc, _ := setupAnthropicTestService(t, server.URL, "anthropic")

	rawMsgs := json.RawMessage(`[
		{"role": "user", "content": "Hi"},
		{"role": "assistant", "content": "Hello!"},
		{"role": "user", "content": "How are you?"}
	]`)
	req := &RelayRequest{
		Model:       "claude-sonnet-4-20250514",
		MaxTokens:   1024,
		RawMessages: rawMsgs,
		Messages:    extractAnthropicMessages(rawMsgs),
	}

	resp, err := svc.Relay(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("Relay error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestService_Relay_CircuitBreakerTrips(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(500)
		w.Write([]byte(`{"error": {"message": "server error"}}`))
	}))
	defer server.Close()

	svc, _ := setupTestService(t, server.URL)

	// First relay: 3 retries (all fail), breaker records 3 failures
	svc.Relay(context.Background(), 1, &RelayRequest{
		Model:    "gpt-4",
		Messages: []relay.Message{{Role: "user", Content: "hi"}},
	})

	// Second relay: 3 more retries (all fail), breaker should trip
	svc.Relay(context.Background(), 1, &RelayRequest{
		Model:    "gpt-4",
		Messages: []relay.Message{{Role: "user", Content: "hi"}},
	})

	// After 6 failures (threshold is 5), breaker should be open
	// Third relay should fail immediately with "no available channel"
	_, err := svc.Relay(context.Background(), 1, &RelayRequest{
		Model:    "gpt-4",
		Messages: []relay.Message{{Role: "user", Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error when circuit breaker is open")
	}
}

func TestService_ListModels(t *testing.T) {
	db := setupTestDB(t)
	db.Create(&model.TokenProduct{Model: "gpt-4", Status: 1})
	db.Create(&model.TokenProduct{Model: "gpt-4o", Status: 1})
	// Create with status=1 first, then disable (GORM default:1 overrides zero value)
	p3 := model.TokenProduct{Model: "disabled", Status: 1}
	db.Create(&p3)
	db.Model(&p3).Update("status", 0)

	svc := NewService(db)
	models, err := svc.ListModels()
	if err != nil {
		t.Fatalf("ListModels error: %v", err)
	}
	if len(models) != 2 {
		t.Errorf("expected 2 models, got %d: %v", len(models), models)
	}
}

func TestService_GetSupplier(t *testing.T) {
	db := setupTestDB(t)
	db.Create(&model.Supplier{Code: "s1", Name: "S1", APIBaseURL: "https://s1.com", APIKeyEncrypted: "k", Status: 1})

	svc := NewService(db)
	supplier, err := svc.GetSupplier(1)
	if err != nil {
		t.Fatalf("GetSupplier error: %v", err)
	}
	if supplier.Code != "s1" {
		t.Errorf("expected s1, got %s", supplier.Code)
	}
}

func TestService_GetSupplier_NotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	_, err := svc.GetSupplier(999)
	if err == nil {
		t.Fatal("expected error for non-existent supplier")
	}
}

func TestGetAPIType(t *testing.T) {
	tests := []struct {
		code     string
		expected relay.APIType
	}{
		{"openai", relay.APITypeOpenAI},
		{"anthropic", relay.APITypeAnthropic},
		{"claude", relay.APITypeAnthropic},
		{"gemini", relay.APITypeGemini},
		{"google", relay.APITypeGemini},
		{"azure", relay.APITypeOpenAI},
		{"", relay.APITypeOpenAI},
	}
	for _, tt := range tests {
		s := &model.Supplier{Code: tt.code}
		got := getAPIType(s)
		if got != tt.expected {
			t.Errorf("getAPIType(%q) = %v, want %v", tt.code, got, tt.expected)
		}
	}
}

// --- Health Check Integration Tests ---

func TestRouter_HealthCheck_FiltersUnhealthy(t *testing.T) {
	db := setupTestDB(t)
	router := NewRouter(db)

	// Create two suppliers with same priority
	db.Create(&model.Supplier{Code: "s1", Name: "S1", APIBaseURL: "https://s1.com", APIKeyEncrypted: "k", Priority: 10, Weight: 10, Status: 1})
	db.Create(&model.Supplier{Code: "s2", Name: "S2", APIBaseURL: "https://s2.com", APIKeyEncrypted: "k", Priority: 10, Weight: 10, Status: 1})
	db.Create(&model.Ability{Group: "default", Model: "gpt-4", ChannelID: 1})
	db.Create(&model.Ability{Group: "default", Model: "gpt-4", ChannelID: 2})

	router.LoadChannels()

	// Set up health checker
	hc := NewHealthChecker(db, time.Minute)
	hc.SetStatus(1, "unhealthy")
	hc.SetStatus(2, "healthy")
	router.SetHealthChecker(hc)

	// Should always select channel 2 (channel 1 is unhealthy)
	for range 20 {
		ch, err := router.SelectChannel("default", "gpt-4", nil)
		if err != nil {
			t.Fatalf("SelectChannel error: %v", err)
		}
		if ch.ChannelID != 2 {
			t.Errorf("expected channel 2, got %d", ch.ChannelID)
		}
	}
}

func TestRouter_HealthCheck_AllUnhealthy(t *testing.T) {
	db := setupTestDB(t)
	router := NewRouter(db)

	db.Create(&model.Supplier{Code: "s1", Name: "S1", APIBaseURL: "https://s1.com", APIKeyEncrypted: "k", Priority: 10, Weight: 10, Status: 1})
	db.Create(&model.Ability{Group: "default", Model: "gpt-4", ChannelID: 1})
	router.LoadChannels()

	hc := NewHealthChecker(db, time.Minute)
	hc.SetStatus(1, "unhealthy")
	router.SetHealthChecker(hc)

	_, err := router.SelectChannel("default", "gpt-4", nil)
	if err == nil {
		t.Fatal("expected error when all channels are unhealthy")
	}
}

func TestRouter_HealthCheck_UnknownIsAllowed(t *testing.T) {
	db := setupTestDB(t)
	router := NewRouter(db)

	db.Create(&model.Supplier{Code: "s1", Name: "S1", APIBaseURL: "https://s1.com", APIKeyEncrypted: "k", Priority: 10, Weight: 10, Status: 1})
	db.Create(&model.Ability{Group: "default", Model: "gpt-4", ChannelID: 1})
	router.LoadChannels()

	hc := NewHealthChecker(db, time.Minute)
	// Don't set any status → "unknown"
	router.SetHealthChecker(hc)

	// Unknown status should be allowed (not filtered out)
	ch, err := router.SelectChannel("default", "gpt-4", nil)
	if err != nil {
		t.Fatalf("SelectChannel error: %v", err)
	}
	if ch.ChannelID != 1 {
		t.Errorf("expected channel 1, got %d", ch.ChannelID)
	}
}

func TestHealthChecker_SetGetStatus(t *testing.T) {
	db := setupTestDB(t)
	hc := NewHealthChecker(db, time.Minute)

	// Default is "unknown"
	if s := hc.GetStatus(1); s != "unknown" {
		t.Errorf("expected unknown, got %s", s)
	}

	hc.SetStatus(1, "healthy")
	if s := hc.GetStatus(1); s != "healthy" {
		t.Errorf("expected healthy, got %s", s)
	}

	hc.SetStatus(1, "unhealthy")
	if s := hc.GetStatus(1); s != "unhealthy" {
		t.Errorf("expected unhealthy, got %s", s)
	}
}

func TestService_Stop(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)
	// Should not panic
	svc.Stop()
}

func TestRouter_AutoRefresh(t *testing.T) {
	db := setupTestDB(t)
	router := NewRouter(db)

	db.Create(&model.Supplier{Code: "s1", Name: "S1", APIBaseURL: "https://s1.com", APIKeyEncrypted: "k", Priority: 10, Weight: 10, Status: 1})
	db.Create(&model.Ability{Group: "default", Model: "gpt-4", ChannelID: 1})
	router.LoadChannels()

	// Start auto-refresh with very short interval
	router.StartAutoRefresh(50 * time.Millisecond)
	defer router.Stop()

	// Add a new supplier while auto-refresh is running
	db.Create(&model.Supplier{Code: "s2", Name: "S2", APIBaseURL: "https://s2.com", APIKeyEncrypted: "k", Priority: 10, Weight: 10, Status: 1})
	db.Create(&model.Ability{Group: "default", Model: "gpt-4", ChannelID: 2})

	// Wait for at least one refresh cycle
	time.Sleep(100 * time.Millisecond)

	// Now there should be 2 channels for gpt-4
	channels := router.GetChannels()
	count := 0
	for _, ch := range channels {
		if ch.Model == "gpt-4" {
			count++
		}
	}
	if count != 2 {
		t.Errorf("expected 2 channels after refresh, got %d", count)
	}
}

// --- Plugin Integration Tests ---

// proxyTestPlugin is a test plugin that tracks calls for proxy integration tests.
type proxyTestPlugin struct {
	name        string
	reqCalled   bool
	respCalled  bool
	reqModel    string
	respStatus  int
	blockReq    bool // if true, OnRequest returns error
	reqDelay    time.Duration
}

func (p *proxyTestPlugin) Name() string                       { return p.name }
func (p *proxyTestPlugin) Init(config map[string]string) error { return nil }
func (p *proxyTestPlugin) OnRequest(req *plugin.RequestContext) error {
	p.reqCalled = true
	p.reqModel = req.Model
	if p.reqDelay > 0 {
		time.Sleep(p.reqDelay)
	}
	if p.blockReq {
		return nil // return nil since proxy ignores plugin errors
	}
	return nil
}
func (p *proxyTestPlugin) OnResponse(resp *plugin.ResponseContext) error {
	p.respCalled = true
	p.respStatus = resp.StatusCode
	return nil
}
func (p *proxyTestPlugin) OnError(errCtx *plugin.ErrorContext) {}

// proxyPanicPlugin panics in OnRequest to test fault isolation.
type proxyPanicPlugin struct {
	name string
}

func (p *proxyPanicPlugin) Name() string                        { return p.name }
func (p *proxyPanicPlugin) Init(config map[string]string) error  { return nil }
func (p *proxyPanicPlugin) OnRequest(req *plugin.RequestContext) error  { panic("plugin boom") }
func (p *proxyPanicPlugin) OnResponse(resp *plugin.ResponseContext) error { return nil }
func (p *proxyPanicPlugin) OnError(errCtx *plugin.ErrorContext)          {}

func TestService_Relay_WithPlugin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"id":"test","choices":[{"message":{"role":"assistant","content":"ok"}}],"usage":{}}`))
	}))
	defer server.Close()

	svc, _ := setupTestService(t, server.URL)

	p := &proxyTestPlugin{name: "tracker"}
	pm := plugin.NewPluginManager()
	pm.Register("tracker", p, nil)
	svc.SetPluginManager(pm)

	resp, err := svc.Relay(context.Background(), 1, &RelayRequest{
		Model:    "gpt-4",
		Messages: []relay.Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("Relay error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if !p.reqCalled {
		t.Error("plugin OnRequest not called")
	}
	if p.reqModel != "gpt-4" {
		t.Errorf("expected model gpt-4, got %s", p.reqModel)
	}
	if !p.respCalled {
		t.Error("plugin OnResponse not called")
	}
	if p.respStatus != 200 {
		t.Errorf("expected response status 200, got %d", p.respStatus)
	}
}

func TestService_Relay_PluginTimeoutDoesNotBlock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"id":"test","choices":[{"message":{"role":"assistant","content":"ok"}}],"usage":{}}`))
	}))
	defer server.Close()

	svc, _ := setupTestService(t, server.URL)

	// Plugin that takes 2s (> 500ms timeout)
	p := &proxyTestPlugin{name: "slow", reqDelay: 2 * time.Second}
	pm := plugin.NewPluginManager()
	pm.Register("slow", p, nil)
	svc.SetPluginManager(pm)

	start := time.Now()
	resp, err := svc.Relay(context.Background(), 1, &RelayRequest{
		Model:    "gpt-4",
		Messages: []relay.Message{{Role: "user", Content: "hi"}},
	})
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Relay should succeed despite plugin timeout: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	// Should complete in ~500ms (plugin timeout), not 2s
	if elapsed > 2*time.Second {
		t.Errorf("relay took too long (%v), plugin timeout not working", elapsed)
	}
}

func TestService_Relay_PluginPanicDoesNotBlock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"id":"test","choices":[{"message":{"role":"assistant","content":"ok"}}],"usage":{}}`))
	}))
	defer server.Close()

	svc, _ := setupTestService(t, server.URL)

	p := &proxyPanicPlugin{name: "panicker"}
	pm := plugin.NewPluginManager()
	pm.Register("panicker", p, nil)
	svc.SetPluginManager(pm)

	resp, err := svc.Relay(context.Background(), 1, &RelayRequest{
		Model:    "gpt-4",
		Messages: []relay.Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("Relay should succeed despite plugin panic: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestService_RelayStream_WithPlugin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"Hi\"}}]}\n\n"))
		w.(http.Flusher).Flush()
		w.Write([]byte("data: [DONE]\n\n"))
		w.(http.Flusher).Flush()
	}))
	defer server.Close()

	svc, _ := setupTestService(t, server.URL)

	p := &proxyTestPlugin{name: "tracker"}
	pm := plugin.NewPluginManager()
	pm.Register("tracker", p, nil)
	svc.SetPluginManager(pm)

	resp, err := svc.RelayStream(context.Background(), 1, &RelayRequest{
		Model:    "gpt-4",
		Messages: []relay.Message{{Role: "user", Content: "hi"}},
		Stream:   true,
	})
	if err != nil {
		t.Fatalf("RelayStream error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if !p.reqCalled {
		t.Error("plugin OnRequest not called for stream")
	}
	if p.reqModel != "gpt-4" {
		t.Errorf("expected model gpt-4, got %s", p.reqModel)
	}
}

func TestService_Relay_NoPluginManager(t *testing.T) {
	// Verify relay works fine without any plugin manager set
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"id":"test","choices":[{"message":{"role":"assistant","content":"ok"}}],"usage":{}}`))
	}))
	defer server.Close()

	svc, _ := setupTestService(t, server.URL)
	// Do NOT set plugin manager — should be nil-safe

	resp, err := svc.Relay(context.Background(), 1, &RelayRequest{
		Model:    "gpt-4",
		Messages: []relay.Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("Relay error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// --- Image Generation Tests ---

func setupImageTestService(t *testing.T, upstreamURL string) (*Service, *gorm.DB) {
	t.Helper()
	db := setupTestDB(t)

	db.Create(&model.Supplier{
		Code:            "openai",
		Name:            "OpenAI",
		APIBaseURL:      upstreamURL,
		APIKeyEncrypted: "test-key",
		Priority:        10,
		Weight:          10,
		Status:          1,
	})
	db.Create(&model.Ability{Group: "default", Model: "dall-e-3", ChannelID: 1})

	svc := NewService(db)
	return svc, db
}

func TestService_ImageRelay_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request path
		if r.URL.Path != "/v1/images/generations" {
			t.Errorf("expected path /v1/images/generations, got %s", r.URL.Path)
		}
		// Verify auth header
		if !strings.Contains(r.Header.Get("Authorization"), "test-key") {
			t.Error("expected auth header with test-key")
		}
		// Verify request body
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "dall-e-3") {
			t.Error("expected model dall-e-3 in request body")
		}
		if !strings.Contains(string(body), "a white cat") {
			t.Error("expected prompt in request body")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
			"created": 1715368000,
			"data": [{"url": "https://example.com/image.png", "revised_prompt": "a white cat"}]
		}`))
	}))
	defer server.Close()

	svc, _ := setupImageTestService(t, server.URL)

	resp, err := svc.ImageRelay(context.Background(), 1, &relay.ImageRequest{
		Model:  "dall-e-3",
		Prompt: "a white cat",
		N:      1,
		Size:   "1024x1024",
	})
	if err != nil {
		t.Fatalf("ImageRelay error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(string(resp.Body), "image.png") {
		t.Error("expected image URL in response body")
	}
}

func TestService_ImageRelay_RetryOn5xx(t *testing.T) {
	callCount1 := 0
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount1++
		w.WriteHeader(500)
		w.Write([]byte(`{"error": {"message": "internal error"}}`))
	}))
	defer server1.Close()

	callCount2 := 0
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount2++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"created": 1, "data": [{"url": "https://example.com/img.png"}]}`))
	}))
	defer server2.Close()

	db := setupTestDB(t)
	db.Create(&model.Supplier{Code: "s1", Name: "S1", APIBaseURL: server1.URL, APIKeyEncrypted: "k1", Priority: 20, Weight: 10, Status: 1})
	db.Create(&model.Supplier{Code: "s2", Name: "S2", APIBaseURL: server2.URL, APIKeyEncrypted: "k2", Priority: 10, Weight: 10, Status: 1})
	db.Create(&model.Ability{Group: "default", Model: "dall-e-3", ChannelID: 1})
	db.Create(&model.Ability{Group: "default", Model: "dall-e-3", ChannelID: 2})

	svc := NewService(db)

	resp, err := svc.ImageRelay(context.Background(), 1, &relay.ImageRequest{
		Model:  "dall-e-3",
		Prompt: "test",
	})
	if err != nil {
		t.Fatalf("ImageRelay error after retries: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if callCount1 != 1 {
		t.Errorf("expected 1 call to server1, got %d", callCount1)
	}
	if callCount2 != 1 {
		t.Errorf("expected 1 call to server2, got %d", callCount2)
	}
}

func TestService_ImageRelay_NoChannel(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	_, err := svc.ImageRelay(context.Background(), 1, &relay.ImageRequest{
		Model:  "nonexistent",
		Prompt: "test",
	})
	if err == nil {
		t.Fatal("expected error for no available channel")
	}
}

func TestService_ImageRelay_AnthropicUnsupported(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte(`{"error": {"message": "not supported"}}`))
	}))
	defer server.Close()

	db := setupTestDB(t)
	db.Create(&model.Supplier{Code: "anthropic", Name: "Claude", APIBaseURL: server.URL, APIKeyEncrypted: "k", Priority: 10, Weight: 10, Status: 1})
	db.Create(&model.Ability{Group: "default", Model: "claude-sonnet-4-20250514", ChannelID: 1})

	svc := NewService(db)

	// Anthropic adaptor's ConvertImageRequest returns error, so this should fail
	_, err := svc.ImageRelay(context.Background(), 1, &relay.ImageRequest{
		Model:  "claude-sonnet-4-20250514",
		Prompt: "test",
	})
	if err == nil {
		t.Fatal("expected error for unsupported adaptor")
	}
}

// --- Audio Speech Tests ---

func setupAudioTestService(t *testing.T, upstreamURL string) (*Service, *gorm.DB) {
	t.Helper()
	db := setupTestDB(t)

	db.Create(&model.Supplier{
		Code:            "openai",
		Name:            "OpenAI",
		APIBaseURL:      upstreamURL,
		APIKeyEncrypted: "test-key",
		Priority:        10,
		Weight:          10,
		Status:          1,
	})
	db.Create(&model.Ability{Group: "default", Model: "tts-1", ChannelID: 1})

	svc := NewService(db)
	return svc, db
}

func TestService_AudioRelay_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request path
		if r.URL.Path != "/v1/audio/speech" {
			t.Errorf("expected path /v1/audio/speech, got %s", r.URL.Path)
		}
		// Verify auth header
		if !strings.Contains(r.Header.Get("Authorization"), "test-key") {
			t.Error("expected auth header with test-key")
		}
		// Verify request body
		body, _ := io.ReadAll(r.Body)
		bodyStr := string(body)
		if !strings.Contains(bodyStr, "tts-1") {
			t.Error("expected model tts-1 in request body")
		}
		if !strings.Contains(bodyStr, "Hello world") {
			t.Error("expected input text in request body")
		}
		if !strings.Contains(bodyStr, "alloy") {
			t.Error("expected voice in request body")
		}

		w.Header().Set("Content-Type", "audio/mpeg")
		w.WriteHeader(200)
		// Return fake audio data
		w.Write([]byte{0xFF, 0xFB, 0x90, 0x00}) // MP3 frame sync
	}))
	defer server.Close()

	svc, _ := setupAudioTestService(t, server.URL)

	resp, err := svc.AudioRelay(context.Background(), 1, &relay.AudioRequest{
		Model: "tts-1",
		Input: "Hello world",
		Voice: "alloy",
	})
	if err != nil {
		t.Fatalf("AudioRelay error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if len(resp.Body) == 0 {
		t.Error("expected non-empty audio response body")
	}
}

func TestService_AudioRelay_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		bodyStr := string(body)
		if !strings.Contains(bodyStr, `"speed":1.5`) {
			t.Error("expected speed 1.5 in request body")
		}
		if !strings.Contains(bodyStr, `"response_format":"opus"`) {
			t.Error("expected response_format opus in request body")
		}

		w.Header().Set("Content-Type", "audio/ogg")
		w.WriteHeader(200)
		w.Write([]byte{0x4F, 0x67, 0x67, 0x53}) // OGG magic bytes
	}))
	defer server.Close()

	svc, _ := setupAudioTestService(t, server.URL)

	resp, err := svc.AudioRelay(context.Background(), 1, &relay.AudioRequest{
		Model:  "tts-1",
		Input:  "Hello with options",
		Voice:  "nova",
		Speed:  1.5,
		Format: "opus",
	})
	if err != nil {
		t.Fatalf("AudioRelay error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestService_AudioRelay_RetryOn5xx(t *testing.T) {
	callCount1 := 0
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount1++
		w.WriteHeader(500)
		w.Write([]byte(`{"error": {"message": "internal error"}}`))
	}))
	defer server1.Close()

	callCount2 := 0
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount2++
		w.Header().Set("Content-Type", "audio/mpeg")
		w.WriteHeader(200)
		w.Write([]byte{0xFF, 0xFB})
	}))
	defer server2.Close()

	db := setupTestDB(t)
	db.Create(&model.Supplier{Code: "s1", Name: "S1", APIBaseURL: server1.URL, APIKeyEncrypted: "k1", Priority: 20, Weight: 10, Status: 1})
	db.Create(&model.Supplier{Code: "s2", Name: "S2", APIBaseURL: server2.URL, APIKeyEncrypted: "k2", Priority: 10, Weight: 10, Status: 1})
	db.Create(&model.Ability{Group: "default", Model: "tts-1", ChannelID: 1})
	db.Create(&model.Ability{Group: "default", Model: "tts-1", ChannelID: 2})

	svc := NewService(db)

	resp, err := svc.AudioRelay(context.Background(), 1, &relay.AudioRequest{
		Model: "tts-1",
		Input: "test",
		Voice: "alloy",
	})
	if err != nil {
		t.Fatalf("AudioRelay error after retries: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if callCount1 != 1 {
		t.Errorf("expected 1 call to server1, got %d", callCount1)
	}
	if callCount2 != 1 {
		t.Errorf("expected 1 call to server2, got %d", callCount2)
	}
}

func TestService_AudioRelay_NoChannel(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db)

	_, err := svc.AudioRelay(context.Background(), 1, &relay.AudioRequest{
		Model: "nonexistent",
		Input: "test",
		Voice: "alloy",
	})
	if err == nil {
		t.Fatal("expected error for no available channel")
	}
}
