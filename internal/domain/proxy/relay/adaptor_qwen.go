// 通义千问 (Qwen) 适配器
//
// 通义千问使用与 OpenAI 不同的消息格式：
//   - 请求格式: { "model": "...", "input": { "messages": [...] }, "parameters": {...} }
//   - 响应格式: { "output": { "choices": [...] }, "usage": {...} }
//   - 需要 messages ↔ input.messages 转换
//
// 参考 one-api relay/adaptor/ali/ (Tongyi/Qwen)
package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// QwenAdaptor 通义千问适配器
type QwenAdaptor struct {
	meta *SupplierMeta
}

func (a *QwenAdaptor) Init(meta *SupplierMeta) {
	a.meta = meta
}

func (a *QwenAdaptor) GetRequestURL(meta *SupplierMeta) (string, error) {
	return meta.APIBaseURL + "/services/aigc/text-generation/generation", nil
}

func (a *QwenAdaptor) SetupRequestHeader(req *http.Request, meta *SupplierMeta) error {
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)
	req.Header.Set("Content-Type", "application/json")
	return nil
}

// qwenRequest 通义千问请求格式
type qwenRequest struct {
	Model      string         `json:"model"`
	Input      qwenInput      `json:"input"`
	Parameters qwenParameters `json:"parameters"`
}

type qwenInput struct {
	Messages []qwenMessage `json:"messages"`
}

type qwenMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type qwenParameters struct {
	MaxTokens     int     `json:"max_tokens,omitempty"`
	Temperature   float64 `json:"temperature,omitempty"`
	TopP          float64 `json:"top_p,omitempty"`
	ResultFormat  string  `json:"result_format"`
}

// ConvertRequest 将内部格式转换为通义千问格式
func (a *QwenAdaptor) ConvertRequest(req *Request) ([]byte, error) {
	qMsgs := make([]qwenMessage, len(req.Messages))
	for i, m := range req.Messages {
		role := m.Role
		if role == "system" {
			role = "system"
		}
		qMsgs[i] = qwenMessage{Role: role, Content: m.Content}
	}

	qReq := qwenRequest{
		Model: req.Model,
		Input: qwenInput{Messages: qMsgs},
		Parameters: qwenParameters{
			MaxTokens:    req.MaxTokens,
			Temperature:  req.Temperature,
			TopP:         req.TopP,
			ResultFormat: "message",
		},
	}
	return json.Marshal(qReq)
}

func (a *QwenAdaptor) ConvertImageRequest(req *ImageRequest) ([]byte, error) {
	return json.Marshal(req)
}

func (a *QwenAdaptor) DoRequest(ctx context.Context, meta *SupplierMeta, body io.Reader) (*http.Response, error) {
	url, err := a.GetRequestURL(meta)
	if err != nil {
		return nil, fmt.Errorf("build qwen request url: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("create qwen request: %w", err)
	}
	if err := a.SetupRequestHeader(httpReq, meta); err != nil {
		return nil, fmt.Errorf("setup qwen request header: %w", err)
	}
	return http.DefaultClient.Do(httpReq)
}

// qwenResponse 通义千问响应格式
type qwenResponse struct {
	Output struct {
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	} `json:"output"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// DoResponse 将通义千问响应转换为内部格式
func (a *QwenAdaptor) DoResponse(ctx context.Context, resp *http.Response, meta *SupplierMeta) (*Usage, *StreamChunk, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read qwen response: %w", err)
	}

	var qResp qwenResponse
	if err := json.Unmarshal(body, &qResp); err != nil {
		return nil, nil, fmt.Errorf("parse qwen response: %w", err)
	}

	usage := &Usage{
		PromptTokens:    qResp.Usage.InputTokens,
		CompletionTokens: qResp.Usage.OutputTokens,
		TotalTokens:     qResp.Usage.TotalTokens,
	}

	// Write converted response back to body for caller
	if len(qResp.Output.Choices) > 0 {
		converted := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"role":    qResp.Output.Choices[0].Message.Role,
						"content": qResp.Output.Choices[0].Message.Content,
					},
					"finish_reason": qResp.Output.Choices[0].FinishReason,
				},
			},
			"usage": usage,
		}
		newBody, _ := json.Marshal(converted)
		resp.Body = io.NopCloser(bytes.NewReader(newBody))
	}

	return usage, nil, nil
}

func (a *QwenAdaptor) GetModelList() []string {
	return []string{"qwen-turbo", "qwen-plus", "qwen-max", "qwen-max-longcontext"}
}

func (a *QwenAdaptor) GetChannelName() string {
	return "qwen"
}
