// 智谱 GLM 适配器
//
// 智谱 GLM 使用与 OpenAI 不同的消息格式：
//   - 请求格式: { "model": "...", "messages": [...], "stream": false }
//   - 响应格式: { "choices": [{"message": {...}}], "usage": {...} }
//   - GLM-4 系列已兼容 OpenAI 格式，可直接使用 OpenAIAdaptor
//   - 旧版 GLM 需要 prompt→messages 转换
//
// 参考 one-api relay/adaptor/zhipu/
package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GLMAdaptor 智谱 GLM 适配器（兼容 OpenAI 格式 + GLM prompt 格式）
type GLMAdaptor struct {
	meta *SupplierMeta
}

func (a *GLMAdaptor) Init(meta *SupplierMeta) {
	a.meta = meta
}

func (a *GLMAdaptor) GetRequestURL(meta *SupplierMeta) (string, error) {
	return meta.APIBaseURL + "/v4/chat/completions", nil
}

func (a *GLMAdaptor) SetupRequestHeader(req *http.Request, meta *SupplierMeta) error {
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)
	req.Header.Set("Content-Type", "application/json")
	return nil
}

// ConvertRequest GLM-4 兼容 OpenAI 格式，直接传递
func (a *GLMAdaptor) ConvertRequest(req *Request) ([]byte, error) {
	// GLM-4 兼容 OpenAI Chat Completions 格式
	glmReq := map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"stream":      req.Stream,
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
		"top_p":       req.TopP,
	}
	return json.Marshal(glmReq)
}

func (a *GLMAdaptor) ConvertImageRequest(req *ImageRequest) ([]byte, error) {
	return json.Marshal(req)
}

func (a *GLMAdaptor) DoRequest(ctx context.Context, meta *SupplierMeta, body io.Reader) (*http.Response, error) {
	url, err := a.GetRequestURL(meta)
	if err != nil {
		return nil, fmt.Errorf("build glm request url: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("create glm request: %w", err)
	}
	if err := a.SetupRequestHeader(httpReq, meta); err != nil {
		return nil, fmt.Errorf("setup glm request header: %w", err)
	}
	return http.DefaultClient.Do(httpReq)
}

type glmResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func (a *GLMAdaptor) DoResponse(ctx context.Context, resp *http.Response, meta *SupplierMeta) (*Usage, *StreamChunk, error) {
	// GLM-4 response is already OpenAI-compatible, pass through
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read glm response: %w", err)
	}

	var glmResp glmResponse
	json.Unmarshal(body, &glmResp)

	usage := &Usage{
		PromptTokens:     glmResp.Usage.PromptTokens,
		CompletionTokens: glmResp.Usage.CompletionTokens,
		TotalTokens:      glmResp.Usage.TotalTokens,
	}

	// Restore body
	resp.Body = io.NopCloser(bytes.NewReader(body))
	return usage, nil, nil
}

func (a *GLMAdaptor) GetModelList() []string {
	return []string{"glm-4", "glm-4-flash", "glm-3-turbo", "cogview-3"}
}

func (a *GLMAdaptor) GetChannelName() string {
	return "glm"
}
