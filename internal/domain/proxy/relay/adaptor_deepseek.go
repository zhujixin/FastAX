// DeepSeek 适配器
//
// DeepSeek 兼容 OpenAI API 格式，但在请求/响应格式上有细微差异：
//   - 请求：OpenAI 格式 (无需转换)
//   - 响应：OpenAI 格式 (无需转换)
//   - 认证：Bearer Token
//   - 特殊：支持 deepseek-reasoner 模型 (thinking chain)
//
// 参考 one-api relay/adaptor/deepseek/
package relay

// DeepSeekAdaptor 复用 OpenAIAdaptor（DeepSeek 完全兼容 OpenAI API）
// 使用 OpenAIAdaptor 作为基础，在 getAPIType 中通过 code="deepseek" 路由
type DeepSeekAdaptor struct {
	OpenAIAdaptor
}

// GetChannelName returns the adaptor channel name
func (a *DeepSeekAdaptor) GetChannelName() string {
	return "deepseek"
}
