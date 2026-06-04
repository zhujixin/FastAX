---
domain: observability
pdd_section: "§5.18"
priority: P0
status: planned
depends_on: [proxy, log]
required_by: []
version: "3.1"
last_updated: "2026-06-04"
---
> **Domain**: `domain/observability` — OpenTelemetry 可观测性 | **PDD**: §5.18 | **新增于**: PRD v3.1

### 6.19 OpenTelemetry 全链路可观测性（OBSV）

参考 Portkey、LiteLLM、Cloudflare AI Gateway 等行业方案。OpenTelemetry 已成为 2025-2026 年 AI 遥测的事实标准，GenAI 语义约定（semconv 1.40+）标准化了 AI 请求的追踪属性。这是**企业客户选型的硬性指标**。

| 功能 | 需求描述 | 优先级 | 备注 |
|------|----------|--------|------|
| OBSV-01 | **请求级 Trace/Span**：在 Proxy 转发全链路创建 Span 树（请求接收 → 鉴权 → 路由决策 → 上游转发 → 响应返回 → 后处理），每个阶段独立 Span | P0 | 基于 OTel Go SDK，Span 属性包含模型名、供应商、token 用量、延迟等 |
| OBSV-02 | **GenAI 语义约定**：遵循 OpenTelemetry GenAI Semantic Convention v1.40+，标准化 `gen_ai.operation.name`、`gen_ai.input.messages`、`gen_ai.output.messages`、`gen_ai.usage.input_tokens`、`gen_ai.usage.output_tokens` 等属性 | P0 | 确保与 Grafana/Datadog/Jaeger 等后端自动兼容；参考 Portkey OTel Analytics Push |
| OBSV-03 | **结构化日志关联**：所有日志自动注入 `trace_id` 和 `span_id`，支持通过 trace_id 关联日志、指标、追踪 | P0 | 与 ROUTE-11 调用日志联动增强 |
| OBSV-04 | **W3C Traceparent 传播**：支持标准 `traceparent` Header 接收上游追踪上下文，注入 `tracestate` 传递到供应商端（如供应商支持） | P1 | 分布式追踪跨系统关联 |
| OBSV-05 | **Prometheus 指标导出**：暴露 `/metrics` 端点，包含 40+ 业务和技术指标（请求数、延迟 histogram、错误率、缓存命中率、token 消耗、费用），按模型/供应商/用户维度打标签 | P1 | 与现有 `domain/stats` 互补——Stats 面向业务报表，OTel 面向运维监控 |
| OBSV-06 | **预置 Grafana 仪表板**：提供 JSON 格式的预配置仪表板 —— LLM 请求总览（延迟/错误率/token 用量）、供应商健康面板、缓存命中率趋势、用户费用排行 | P2 | 开箱即用，降低运维门槛 |

**关键设计决策**：

| 决策 | 要点 | 参考 |
|------|------|------|
| **OTel 导出器** | 默认 OTLP/gRPC 导出到本地 Collector，可选 stdout（开发）、Jaeger（调试）、Prometheus（指标） | Portkey OTel Log Export |
| **采样策略** | 生产环境 10% 尾部采样（保留所有错误+高延迟请求 100%），开发环境 100% | — |
| **Span 粒度** | 每个转发阶段一个 Span，Span 属性包含渠道 ID、模型名、供应商平台、token 用量、计费金额 | LiteLLM callback 模式 |
| **结构化消息 Span** | 支持多模态 parts（text/tool_call/image/audio/file/reasoning）的 Span Event 记录 | Portkey Structured Message Spans |
| **不与业务耦合** | OTel 作为横切关注点通过 Gin 中间件 + Proxy 装饰器注入，不侵入业务逻辑 | — |

**Span 树示例**：

```
HTTP POST /v1/chat/completions (root span)
├── AuthMiddleware.JWTVerify (50μs)
├── Guardrail.BeforeCheck (3ms)
│   ├── pii_detect (1ms)
│   └── injection_detect (2ms)
├── ProxyService.Relay (total: 800ms)
│   ├── RouteDecision (2ms) — gen_ai.model=deepseek-chat, gen_ai.vendor=deepseek
│   ├── Billing.PreCharge (5ms)
│   └── UpstreamCall (790ms) — gen_ai.usage.input_tokens=1500, gen_ai.usage.output_tokens=800
├── Guardrail.AfterCheck (2ms)
└── Billing.PostCharge (3ms)
```

**实施路径**：

| 阶段 | 内容 | 预计人日 |
|------|------|---------|
| Phase 1 | OTel SDK 集成 + 基础 Span（OBSV-01/02/03） | 3-5d |
| Phase 2 | Prometheus 指标 + Traceparent 传播（OBSV-04/05） | 2-4d |
| Phase 3 | Grafana 仪表板 + 尾部采样（OBSV-06） | 2-3d |

---
## 相关模块

| 关系 | 模块 | 说明 |
|------|------|------|
| 依赖 | [代理模块](02-token-proxy-vendor.md) | OTel Span 覆盖代理全链路 |
| 依赖 | [日志模块](05-notify-stats-admin.md) | trace_id 注入调用日志 |
| 关联 | [语义缓存](22-semantic-cache.md) | 缓存命中率/延迟指标 |
| 关联 | [安全护栏](09-guardrails.md) | 护栏检测结果 Span Event |