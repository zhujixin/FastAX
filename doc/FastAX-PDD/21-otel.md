> **Domain**: `domain/observability` — OpenTelemetry 可观测性 | **PRD**: FastAX-PRD/23-otel-observability.md
### 5.18 OpenTelemetry 全链路可观测性 (PRD §6.19 OBSV)

#### 5.18.1 OTel 集成架构

```
┌─────────────────────────────────────────────────────┐
│                   FastAX Server                      │
│                                                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────────────┐  │
│  │ Gin      │→ │ OTel     │→ │ ProxyService     │  │
│  │ Router   │  │ Middleware│  │ (装饰器注入 Span) │  │
│  └──────────┘  └──────────┘  └──────────────────┘  │
│                                                      │
│  OTel Go SDK (go.opentelemetry.io/otel)              │
│  ├── TracerProvider (OTLP/gRPC Exporter)             │
│  ├── MeterProvider  (Prometheus Exporter)            │
│  └── LogProvider    (trace_id + span_id 注入)        │
│                                                      │
└──────────────────────┬──────────────────────────────┘
                       │ OTLP/gRPC
                       ▼
              ┌─────────────────┐
              │ OTel Collector  │
              │ ├── Jaeger      │ (Trace 查询)
              │ ├── Prometheus  │ (指标抓取)
              │ └── Grafana     │ (仪表板)
              └─────────────────┘
```

#### 5.18.2 Span 树设计

```
POST /v1/chat/completions (root span, name: "chat.completions")
├── Span: auth.jwt_verify         (50μs)
│   属性: auth.method=jwt, user.id=123
├── Span: guardrail.before_check  (3ms)
│   ├── Event: pii_detect         (1ms, findings=0)
│   └── Event: injection_detect   (2ms, findings=0)
├── Span: proxy.relay             (800ms)
│   ├── Span: route.decision      (2ms)
│   │   属性: gen_ai.model=deepseek-chat, gen_ai.vendor=deepseek, gen_ai.system=deepseek
│   ├── Span: billing.pre_charge  (5ms)
│   │   属性: cost.estimated=$0.012, balance.remaining=$5.23
│   └── Span: upstream.call       (790ms)
│       属性: gen_ai.usage.input_tokens=1500, gen_ai.usage.output_tokens=800
│             http.status_code=200, http.url=api.deepseek.com/v1/chat/completions
├── Span: guardrail.after_check   (2ms)
│   └── Event: content_moderation (2ms, passed=true)
└── Span: billing.post_charge     (3ms)
    属性: cost.actual=$0.014, cost.cached=false
```

#### 5.18.3 GenAI 语义约定 (semconv 1.40+)

| OTel 属性 | 值示例 | 来源 |
|-----------|--------|------|
| `gen_ai.operation.name` | `chat` / `embedding` / `rerank` | 端点类型 |
| `gen_ai.system` | `deepseek` / `openai` / `anthropic` | 供应商平台 |
| `gen_ai.request.model` | `deepseek-chat` | 请求中的 model 参数 |
| `gen_ai.response.model` | `deepseek-chat` | 响应中的 model 字段 |
| `gen_ai.usage.input_tokens` | `1500` | 响应 usage |
| `gen_ai.usage.output_tokens` | `800` | 响应 usage |
| `gen_ai.response.id` | `chatcmpl-xxx` | LLM 响应 ID |
| `gen_ai.response.finish_reasons` | `["stop"]` | 完成原因 |

#### 5.18.4 采样策略

| 环境 | 策略 | 配置 |
|------|------|------|
| 开发 | AlwaysOn (100%) | `sampler=always_on` |
| 生产 | TailSampling (10%) | 保留所有错误 (status=ERROR) + P95 延迟请求 100% |

#### 5.18.5 Prometheus 指标

| 指标名 | 类型 | 标签 | 说明 |
|--------|------|------|------|
| `fastax_requests_total` | Counter | model, vendor, status | 请求总数 |
| `fastax_request_duration_ms` | Histogram | model, vendor, stage | 各阶段延迟 (buckets: 10,50,100,500,1000,5000,10000) |
| `fastax_tokens_total` | Counter | model, vendor, type(input/output) | Token 消耗量 |
| `fastax_cost_usd_total` | Counter | model, vendor | 累计费用 |
| `fastax_cache_hit_ratio` | Gauge | cache_type(exact/semantic) | 缓存命中率 |
| `fastax_errors_total` | Counter | model, vendor, error_type | 错误计数 |
| `fastax_active_channels` | Gauge | vendor | 活跃渠道数 |

#### 5.18.6 API 端点

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/metrics` | GET | 无 (内部) | Prometheus 指标导出端点 |
| `/api/admin/traces` | GET | Admin | 查询最近 Trace 列表 (按时间/user_id/model 筛选) |
| `/api/admin/traces/:trace_id` | GET | Admin | 查询指定 Trace 的完整 Span 树 |
| `/api/admin/observability/config` | PUT | Admin | 配置采样率/导出器端点 |

#### 5.18.7 Service 方法

```go
// ObservabilityService — OTel 可观测性服务
type ObservabilityService struct {
    tracer trace.Tracer
    meter  metric.Meter
}

// Span 创建 (在 Proxy/Guardrail/Billing 中装饰器调用)
func (s *ObservabilityService) StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span)
func (s *ObservabilityService) AddEvent(span trace.Span, name string, attrs ...attribute.KeyValue)

// 指标记录
func (s *ObservabilityService) RecordRequest(model, vendor, status string, durationMs float64)
func (s *ObservabilityService) RecordTokens(model, vendor string, inputTokens, outputTokens int64)
func (s *ObservabilityService) RecordCost(model, vendor string, costUSD float64)
func (s *ObservabilityService) RecordCacheHit(cacheType string)

// Trace 查询
func (s *ObservabilityService) ListTraces(ctx context.Context, filter *TraceFilter) ([]*TraceSummary, error)
func (s *ObservabilityService) GetTrace(ctx context.Context, traceID string) (*TraceDetail, error)

// 配置
func (s *ObservabilityService) UpdateConfig(ctx context.Context, cfg *OtelConfig) error
```

#### 5.18.8 数据表

| 表名 | 说明 | 关键字段 |
|------|------|---------|
| `otel_export_config` | OTel 导出器配置 | exporter_type (otlp/jaeger/prometheus), endpoint_url, sample_rate, enabled |
| `trace_spans` | Span 持久化 (可选，默认依赖外部 Collector) | trace_id, span_id, parent_span_id, operation_name, start_time, duration_ms, attributes (JSON), status |

> 注：Traces 默认由外部 OTel Collector 管理，`trace_spans` 表仅在未部署 Collector 时作为本地回退。

---
## 相关模块

| 关系 | 模块 | 说明 |
|------|------|------|
| 依赖 | [代理模块](06-token-proxy-vendor.md) | OTel Span 覆盖代理全链路 |
| 关联 | [语义缓存](20-cache.md) | 缓存命中率/节省金额指标 |
| 关联 | [安全护栏](14-guardrails.md) | 护栏检测结果 Span Event |
| 关联 | [MCP 网关](22-mcp.md) | MCP 事件追踪 |
