package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/fastax/fastax-server/internal/domain/plugin"
	"github.com/fastax/fastax-server/internal/domain/proxy/relay"
	"github.com/fastax/fastax-server/internal/shared/model"
	"gorm.io/gorm"
)

type Service struct {
	db            *gorm.DB
	router        *Router
	healthChecker *HealthChecker
	breakers      map[uint]*CircuitBreaker // channelID → circuit breaker
	client        *http.Client
	pluginManager *plugin.PluginManager
}

func NewService(db *gorm.DB) *Service {
	router := NewRouter(db)

	// Create health checker and wire it to the router
	hc := NewHealthChecker(db, 5*time.Minute)
	router.SetHealthChecker(hc)

	svc := &Service{
		db:            db,
		router:        router,
		healthChecker: hc,
		breakers:      make(map[uint]*CircuitBreaker),
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
	svc.router.LoadChannels()

	// Start background services
	hc.Start()                        // Health checks every 5 min
	router.StartAutoRefresh(60 * time.Second) // Channel cache refresh every 60s

	return svc
}

// Stop stops background goroutines (health checker, cache refresh)
func (s *Service) Stop() {
	s.healthChecker.Stop()
	s.router.Stop()
}

// SetPluginManager injects the plugin manager into the proxy service.
// Must be called before any Relay/RelayStream calls.
func (s *Service) SetPluginManager(pm *plugin.PluginManager) {
	s.pluginManager = pm
}

type RelayRequest struct {
	Model    string          `json:"model"`
	Messages []relay.Message `json:"messages"`
	Stream   bool            `json:"stream"`
	MaxTokens int            `json:"max_tokens,omitempty"`
	Temperature float64      `json:"temperature,omitempty"`
	TopP       float64       `json:"top_p,omitempty"`

	// Anthropic native format fields
	// RawMessages stores the original Anthropic-format messages JSON (content blocks etc.)
	RawMessages  json.RawMessage `json:"-"`
	SystemPrompt string          `json:"system,omitempty"`
}

type RelayResponse struct {
	StatusCode int             `json:"-"`
	Body       []byte          `json:"-"`
	Resp       *relay.Response `json:"-"`
}

// RelayStreamResponse holds the streaming relay result
type RelayStreamResponse struct {
	StatusCode int
	Header     http.Header
	Body       io.ReadCloser
}

// Relay handles the main non-streaming request forwarding logic
func (s *Service) Relay(ctx context.Context, userID uint, req *RelayRequest) (*RelayResponse, error) {
	// Execute request plugins before forwarding (best-effort, never blocks main request)
	if s.pluginManager != nil {
		reqCtx := &plugin.RequestContext{
			UserID: userID,
			Model:  req.Model,
		}
		_ = s.pluginManager.ExecuteRequest(ctx, reqCtx)
	}

	disabled := s.collectDisabledChannels()
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		channel, err := s.router.SelectChannel("", req.Model, disabled)
		if err != nil {
			return nil, fmt.Errorf("no available channel: %w", err)
		}

		resp, err := s.doRelay(ctx, channel, req)
		if err != nil {
			s.recordFailure(channel.ChannelID)
			disabled[channel.ChannelID] = true
			lastErr = err
			continue
		}

		s.recordSuccess(channel.ChannelID)

		// Execute response plugins after successful forwarding (best-effort)
		if s.pluginManager != nil {
			respCtx := &plugin.ResponseContext{
				StatusCode: resp.StatusCode,
				Body:       resp.Body,
			}
			_ = s.pluginManager.ExecuteResponse(ctx, respCtx)
		}

		return resp, nil
	}

	return nil, fmt.Errorf("all retries exhausted: %w", lastErr)
}

// RelayStream handles streaming request forwarding with SSE
func (s *Service) RelayStream(ctx context.Context, userID uint, req *RelayRequest) (*RelayStreamResponse, error) {
	// Execute request plugins before forwarding (best-effort, never blocks main request)
	if s.pluginManager != nil {
		reqCtx := &plugin.RequestContext{
			UserID: userID,
			Model:  req.Model,
		}
		_ = s.pluginManager.ExecuteRequest(ctx, reqCtx)
	}

	disabled := s.collectDisabledChannels()
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		channel, err := s.router.SelectChannel("", req.Model, disabled)
		if err != nil {
			return nil, fmt.Errorf("no available channel: %w", err)
		}

		resp, err := s.doRelayStream(ctx, channel, req)
		if err != nil {
			s.recordFailure(channel.ChannelID)
			disabled[channel.ChannelID] = true
			lastErr = err
			continue
		}

		s.recordSuccess(channel.ChannelID)

		// Note: response plugins are NOT called for streaming responses
		// because the response body is consumed by the caller via SSE.

		return resp, nil
	}

	return nil, fmt.Errorf("all retries exhausted: %w", lastErr)
}

// doRelay executes a non-streaming relay to a specific channel
func (s *Service) doRelay(ctx context.Context, channel *ChannelEntry, req *RelayRequest) (*RelayResponse, error) {
	supplier, adaptor, meta, err := s.prepareRelay(channel, req)
	if err != nil {
		return nil, err
	}

	// Convert request
	relayReq := &relay.Request{
		Model:       meta.Model,
		Messages:    req.Messages,
		Stream:      false,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}
	if len(req.RawMessages) > 0 {
		relayReq.RawMessages = json.RawMessage(req.RawMessages)
	}
	if req.SystemPrompt != "" {
		relayReq.SystemPrompt = req.SystemPrompt
	}
	reqBody, err := adaptor.ConvertRequest(relayReq)
	if err != nil {
		return nil, fmt.Errorf("convert request: %w", err)
	}

	// Execute via adaptor
	httpResp, err := adaptor.DoRequest(ctx, meta, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer httpResp.Body.Close()

	// Check for retryable errors
	if relay.ShouldRetryHTTP(httpResp.StatusCode) {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("upstream error %d: %s", httpResp.StatusCode, string(body))
	}

	// Read full response for non-streaming
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Parse response via adaptor
	usage, _, err := adaptor.DoResponse(ctx, httpResp, meta)
	if err != nil {
		// Still return the raw body even if parsing fails
		return &RelayResponse{
			StatusCode: httpResp.StatusCode,
			Body:       body,
		}, nil
	}

	_ = usage   // TODO: use for billing
	_ = supplier // TODO: use for logging

	return &RelayResponse{
		StatusCode: httpResp.StatusCode,
		Body:       body,
	}, nil
}

// doRelayStream executes a streaming relay to a specific channel
func (s *Service) doRelayStream(ctx context.Context, channel *ChannelEntry, req *RelayRequest) (*RelayStreamResponse, error) {
	_, adaptor, meta, err := s.prepareRelay(channel, req)
	if err != nil {
		return nil, err
	}

	// Convert request with stream=true
	relayReq := &relay.Request{
		Model:       meta.Model,
		Messages:    req.Messages,
		Stream:      true,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}
	if len(req.RawMessages) > 0 {
		relayReq.RawMessages = json.RawMessage(req.RawMessages)
	}
	if req.SystemPrompt != "" {
		relayReq.SystemPrompt = req.SystemPrompt
	}
	reqBody, err := adaptor.ConvertRequest(relayReq)
	if err != nil {
		return nil, fmt.Errorf("convert request: %w", err)
	}

	// Execute via adaptor
	httpResp, err := adaptor.DoRequest(ctx, meta, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	// Check for retryable errors before streaming
	if relay.ShouldRetryHTTP(httpResp.StatusCode) {
		body, _ := io.ReadAll(httpResp.Body)
		httpResp.Body.Close()
		return nil, fmt.Errorf("upstream error %d: %s", httpResp.StatusCode, string(body))
	}

	return &RelayStreamResponse{
		StatusCode: httpResp.StatusCode,
		Header:     httpResp.Header,
		Body:       httpResp.Body, // caller is responsible for closing
	}, nil
}

// prepareRelay builds the supplier, adaptor, and meta for a relay request
func (s *Service) prepareRelay(channel *ChannelEntry, req *RelayRequest) (*model.Supplier, relay.Adaptor, *relay.SupplierMeta, error) {
	var supplier model.Supplier
	if err := s.db.First(&supplier, channel.ChannelID).Error; err != nil {
		return nil, nil, nil, fmt.Errorf("supplier not found: %w", err)
	}

	adaptor := s.getAdaptor(&supplier)
	meta := &relay.SupplierMeta{
		SupplierID: supplier.ID,
		ChannelID:  channel.ChannelID,
		APIBaseURL: supplier.APIBaseURL,
		APIKey:     supplier.APIKeyEncrypted,
		APIType:    getAPIType(&supplier),
		Model:      req.Model,
	}
	adaptor.Init(meta)

	return &supplier, adaptor, meta, nil
}

func (s *Service) collectDisabledChannels() map[uint]bool {
	disabled := make(map[uint]bool)
	for chID, cb := range s.breakers {
		if !cb.Allow() {
			disabled[chID] = true
		}
	}
	return disabled
}

func (s *Service) recordFailure(channelID uint) {
	cb, ok := s.breakers[channelID]
	if !ok {
		cb = NewCircuitBreaker(5, 3, 5*time.Minute)
		s.breakers[channelID] = cb
	}
	cb.RecordFailure()
}

func (s *Service) recordSuccess(channelID uint) {
	cb, ok := s.breakers[channelID]
	if !ok {
		return
	}
	cb.RecordSuccess()
}

// getAdaptor returns the appropriate adaptor based on supplier config
func (s *Service) getAdaptor(supplier *model.Supplier) relay.Adaptor {
	return relay.GetAdaptor(getAPIType(supplier))
}

// getAPIType extracts the API type from supplier configuration
func getAPIType(supplier *model.Supplier) relay.APIType {
	// Determine from supplier code or models field
	code := supplier.Code
	switch {
	case code == "anthropic" || code == "claude":
		return relay.APITypeAnthropic
	case code == "gemini" || code == "google":
		return relay.APITypeGemini
	default:
		return relay.APITypeOpenAI
	}
}

// GetRouter returns the router (for testing)
func (s *Service) GetRouter() *Router {
	return s.router
}

// ImageRelay handles image generation request forwarding
func (s *Service) ImageRelay(ctx context.Context, userID uint, req *relay.ImageRequest) (*RelayResponse, error) {
	disabled := s.collectDisabledChannels()
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		channel, err := s.router.SelectChannel("", req.Model, disabled)
		if err != nil {
			return nil, fmt.Errorf("no available channel: %w", err)
		}

		resp, err := s.doImageRelay(ctx, channel, req)
		if err != nil {
			s.recordFailure(channel.ChannelID)
			disabled[channel.ChannelID] = true
			lastErr = err
			continue
		}

		s.recordSuccess(channel.ChannelID)
		return resp, nil
	}

	return nil, fmt.Errorf("all retries exhausted: %w", lastErr)
}

// doImageRelay executes an image generation relay to a specific channel
func (s *Service) doImageRelay(ctx context.Context, channel *ChannelEntry, req *relay.ImageRequest) (*RelayResponse, error) {
	var supplier model.Supplier
	if err := s.db.First(&supplier, channel.ChannelID).Error; err != nil {
		return nil, fmt.Errorf("supplier not found: %w", err)
	}

	adaptor := relay.GetAdaptor(getAPIType(&supplier))
	meta := &relay.SupplierMeta{
		SupplierID: supplier.ID,
		ChannelID:  channel.ChannelID,
		APIBaseURL: supplier.APIBaseURL,
		APIKey:     supplier.APIKeyEncrypted,
		APIType:    getAPIType(&supplier),
		Model:      req.Model,
	}
	adaptor.Init(meta)

	reqBody, err := adaptor.ConvertImageRequest(req)
	if err != nil {
		return nil, fmt.Errorf("convert image request: %w", err)
	}

	// Build request with image endpoint URL
	url := meta.APIBaseURL + "/v1/images/generations"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := adaptor.SetupRequestHeader(httpReq, meta); err != nil {
		return nil, fmt.Errorf("setup header: %w", err)
	}

	httpResp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer httpResp.Body.Close()

	if relay.ShouldRetryHTTP(httpResp.StatusCode) {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("upstream error %d: %s", httpResp.StatusCode, string(body))
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return &RelayResponse{
		StatusCode: httpResp.StatusCode,
		Body:       body,
	}, nil
}

// AudioRelay handles text-to-speech request forwarding
func (s *Service) AudioRelay(ctx context.Context, userID uint, req *relay.AudioRequest) (*RelayResponse, error) {
	disabled := s.collectDisabledChannels()
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		channel, err := s.router.SelectChannel("", req.Model, disabled)
		if err != nil {
			return nil, fmt.Errorf("no available channel: %w", err)
		}

		resp, err := s.doAudioRelay(ctx, channel, req)
		if err != nil {
			s.recordFailure(channel.ChannelID)
			disabled[channel.ChannelID] = true
			lastErr = err
			continue
		}

		s.recordSuccess(channel.ChannelID)
		return resp, nil
	}

	return nil, fmt.Errorf("all retries exhausted: %w", lastErr)
}

// doAudioRelay executes a text-to-speech relay to a specific channel
func (s *Service) doAudioRelay(ctx context.Context, channel *ChannelEntry, req *relay.AudioRequest) (*RelayResponse, error) {
	var supplier model.Supplier
	if err := s.db.First(&supplier, channel.ChannelID).Error; err != nil {
		return nil, fmt.Errorf("supplier not found: %w", err)
	}

	adaptor := relay.GetAdaptor(getAPIType(&supplier))
	meta := &relay.SupplierMeta{
		SupplierID: supplier.ID,
		ChannelID:  channel.ChannelID,
		APIBaseURL: supplier.APIBaseURL,
		APIKey:     supplier.APIKeyEncrypted,
		APIType:    getAPIType(&supplier),
		Model:      req.Model,
	}
	adaptor.Init(meta)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal audio request: %w", err)
	}

	// Build request with audio endpoint URL
	url := meta.APIBaseURL + "/v1/audio/speech"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := adaptor.SetupRequestHeader(httpReq, meta); err != nil {
		return nil, fmt.Errorf("setup header: %w", err)
	}

	httpResp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer httpResp.Body.Close()

	if relay.ShouldRetryHTTP(httpResp.StatusCode) {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("upstream error %d: %s", httpResp.StatusCode, string(body))
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return &RelayResponse{
		StatusCode: httpResp.StatusCode,
		Body:       body,
	}, nil
}

// GetSupplier returns a supplier by ID
func (s *Service) GetSupplier(id uint) (*model.Supplier, error) {
	var supplier model.Supplier
	if err := s.db.First(&supplier, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("supplier not found")
		}
		return nil, err
	}
	return &supplier, nil
}

// ListModels returns all available models
func (s *Service) ListModels() ([]string, error) {
	var products []model.TokenProduct
	if err := s.db.Where("status = ?", 1).Find(&products).Error; err != nil {
		return nil, err
	}

	modelSet := make(map[string]bool)
	for _, p := range products {
		if p.Model != "" {
			modelSet[p.Model] = true
		}
	}

	models := make([]string, 0, len(modelSet))
	for m := range modelSet {
		models = append(models, m)
	}
	return models, nil
}

// ---------- Rerank ----------

type RerankRequest struct {
	Model     string   `json:"model" binding:"required"`
	Query     string   `json:"query" binding:"required"`
	Documents []string `json:"documents" binding:"required,min=1"`
	TopN      int      `json:"top_n,omitempty"`
}

type RerankResult struct {
	Index          int     `json:"index"`
	Document       string  `json:"document"`
	RelevanceScore float64 `json:"relevance_score"`
}

type RerankResponse struct {
	Model   string         `json:"model"`
	Results []RerankResult `json:"results"`
	Usage   relay.Usage    `json:"usage"`
}

// Rerank performs document reranking using the relay engine.
func (s *Service) Rerank(ctx context.Context, userID uint, req *RerankRequest) (*RerankResponse, error) {
	// For MVP, use a simple scoring algorithm
	// In production, this would call an actual rerank API (e.g., Cohere, Jina)

	results := make([]RerankResult, len(req.Documents))
	queryLower := strings.ToLower(req.Query)

	for i, doc := range req.Documents {
		docLower := strings.ToLower(doc)

		// Simple relevance scoring based on word overlap
		score := calculateRelevance(queryLower, docLower)

		results[i] = RerankResult{
			Index:          i,
			Document:       doc,
			RelevanceScore: score,
		}
	}

	// Sort by relevance score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].RelevanceScore > results[j].RelevanceScore
	})

	// Apply top_n limit
	topN := req.TopN
	if topN <= 0 || topN > len(results) {
		topN = len(results)
	}
	results = results[:topN]

	return &RerankResponse{
		Model:   req.Model,
		Results: results,
		Usage: relay.Usage{
			PromptTokens: len(strings.Fields(req.Query)),
			TotalTokens:  len(strings.Fields(req.Query)),
		},
	}, nil
}

func calculateRelevance(query, doc string) float64 {
	queryWords := strings.Fields(query)
	docWords := strings.Fields(doc)

	if len(queryWords) == 0 || len(docWords) == 0 {
		return 0
	}

	matchCount := 0
	for _, qw := range queryWords {
		for _, dw := range docWords {
			if strings.Contains(dw, qw) || strings.Contains(qw, dw) {
				matchCount++
				break
			}
		}
	}

	return float64(matchCount) / float64(len(queryWords))
}

// ---------- Video Generation ----------

type VideoRequest struct {
	Model  string `json:"model" binding:"required"`
	Prompt string `json:"prompt" binding:"required"`
	N      int    `json:"n,omitempty"`
	Size   string `json:"size,omitempty"`
}

// VideoRelay handles video generation request forwarding
func (s *Service) VideoRelay(ctx context.Context, userID uint, req *VideoRequest) (*RelayResponse, error) {
	disabled := s.collectDisabledChannels()
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		channel, err := s.router.SelectChannel("", req.Model, disabled)
		if err != nil {
			return nil, fmt.Errorf("no available channel: %w", err)
		}

		resp, err := s.doVideoRelay(ctx, channel, req)
		if err != nil {
			s.recordFailure(channel.ChannelID)
			disabled[channel.ChannelID] = true
			lastErr = err
			continue
		}

		s.recordSuccess(channel.ChannelID)
		return resp, nil
	}

	return nil, fmt.Errorf("all retries exhausted: %w", lastErr)
}

// doVideoRelay executes a video generation relay to a specific channel
func (s *Service) doVideoRelay(ctx context.Context, channel *ChannelEntry, req *VideoRequest) (*RelayResponse, error) {
	var supplier model.Supplier
	if err := s.db.First(&supplier, channel.ChannelID).Error; err != nil {
		return nil, fmt.Errorf("supplier not found: %w", err)
	}

	adaptor := relay.GetAdaptor(getAPIType(&supplier))
	meta := &relay.SupplierMeta{
		SupplierID: supplier.ID,
		ChannelID:  channel.ChannelID,
		APIBaseURL: supplier.APIBaseURL,
		APIKey:     supplier.APIKeyEncrypted,
		APIType:    getAPIType(&supplier),
		Model:      req.Model,
	}
	adaptor.Init(meta)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal video request: %w", err)
	}

	// Build request with video endpoint URL
	url := meta.APIBaseURL + "/v1/video/generations"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := adaptor.SetupRequestHeader(httpReq, meta); err != nil {
		return nil, fmt.Errorf("setup header: %w", err)
	}

	httpResp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer httpResp.Body.Close()

	if relay.ShouldRetryHTTP(httpResp.StatusCode) {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("upstream error %d: %s", httpResp.StatusCode, string(body))
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return &RelayResponse{
		StatusCode: httpResp.StatusCode,
		Body:       body,
	}, nil
}
