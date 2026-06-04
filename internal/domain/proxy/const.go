// Proxy 模块命名常量
//
// 本文件定义 proxy 模块中可配置的业务常量，消除魔术数字：
//   - 重试次数、熔断器参数
//   - HTTP 客户端超时
//   - 计费估算参数
//   - 音频并发限制、SSE 缓冲区大小
package proxy

import "time"

// ─── 重试 ───

const DefaultMaxRetries = 3

// ─── 熔断器 ───

const (
	DefaultCBFailureThreshold = 5
	DefaultCBSuccessThreshold = 3
	DefaultCBTimeout          = 5 * time.Minute
)

// ─── HTTP 客户端超时 ───

const (
	DefaultProxyHTTPTimeout = 60 * time.Second
	DefaultAdaptorTimeout   = 120 * time.Second
	DefaultHealthTimeout    = 10 * time.Second
)

// ─── 渠道缓存 ───

const DefaultChannelCacheRefresh = 60 * time.Second

// ─── 计费估算 ───

const (
	MinTokenEstimate          = 100
	TokensPerMessage          = 50
	StreamEstimateMultiplier  = 0.8
	MinStreamTokenEstimate    = 100
)

// ─── 计费批量刷入 ───

const (
	DefaultBillingBatchSize  = 100
	DefaultBillingFlushInterval = 10 * time.Second
)

// ─── 音频处理 ───

const (
	MaxConcurrentAudioUploads = 5
	SSEBufferSize             = 4096
)

// ─── 健康检测 ───

const DefaultHealthCheckInterval = 5 * time.Minute

// ─── 供应商代码映射 ───
// 数据库中的 supplier.code 值 → API 类型适配器

const (
	SupplierCodeAnthropic = "anthropic"
	SupplierCodeClaude    = "claude"
	SupplierCodeGemini    = "gemini"
	SupplierCodeGoogle    = "google"
	SupplierCodeDeepSeek  = "deepseek"
	SupplierCodeQwen      = "qwen"
	SupplierCodeTongyi    = "tongyi"
	SupplierCodeAli       = "ali"
	SupplierCodeGLM       = "glm"
	SupplierCodeZhipu     = "zhipu"
)
