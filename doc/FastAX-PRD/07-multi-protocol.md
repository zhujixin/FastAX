> **Domain**: `domain/proxy` (adaptor/) — 多协议原生支持 | **PDD**: §5.9 | **实现参考**: one-api relay/adaptor/ + channeltype/

### 6.10 多协议原生支持（PROTO）

对标 OpenRouter、New-API、VoAPI 等主流平台的多协议原生支持能力。
参考 one-api 已验证的 Adaptor 模式：每个协议有独立适配器实现 `Adaptor` 接口，包括 `GetRequestURL`（构建上游 URL）、`SetupRequestHeader`（设置认证）、`ConvertRequest`（格式转换）、`DoResponse`（响应解析+计费）。

| 功能 | 需求描述 | 优先级 | 备注 |
|------|----------|--------|------|
| PROTO-01 | **Anthropic Messages API 原生支持**：直接支持 `/v1/messages` 端点，无需格式转换，支持 extended thinking、PDF 输入等 Claude 原生特性 | P0 | 参考 one-api relay/adaptor/anthropic/ — GetRequestURL 构建 `/v1/messages`，SetupRequestHeader 设 `x-api-key`+`anthropic-version` |
| PROTO-02 | **Gemini API 原生支持**：支持 Google Gemini 原生 API 协议，包括 grounding、caching、safety 参数透传 | P0 | 参考 one-api relay/adaptor/gemini/ |
| PROTO-03 | **OpenAI Realtime API 代理**：支持 WebRTC/WebSocket 实时语音会话代理，转发 beta.realtime 端点 | P1 | 需网关层特殊处理 WebSocket 升级 |
| PROTO-04 | **协议自动检测与转换**：根据请求路径自动检测目标协议（OpenAI/Anthropic/Gemini），路由到对应 Adaptor | P0 | 参考 one-api relaymode.GetByPath() + GetAdaptor(apiType) switch 分发 |
| PROTO-05 | **模型后缀变体（Model Suffix）**：支持 `model:variant` 语法 | P1 | 参考 OpenRouter 模型变体设计 |
| PROTO-06 | **模型自动发现**：通过 `/v1/models` 自动获取供应商可用模型列表 | P1 | — |
| PROTO-07 | **模型重命名/别名**：管理员可自定义模型显示名称和 API 请求名映射，支持 `gpt-4 → gpt-4-turbo` 等映射 | P1 | 便于渠道切换时保持用户端兼容 |
| PROTO-08 | **Rerank 模型支持**：支持 Cohere、Jina 等 Rerank API，统一 `POST /v1/rerank` 端点 | P2 | — |
| PROTO-09 | **Embeddings 多供应商路由**：支持 OpenAI、Cohere、Google 等多供应商 Embedding 接口聚合 | P1 | — |
| PROTO-10 | **MCP 协议支持**：支持 Model Context Protocol，AI Agent 可通过 MCP 发现和调用平台能力 | P1 | — |

