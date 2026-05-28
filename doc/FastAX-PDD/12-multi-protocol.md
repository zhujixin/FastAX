> **Domain**: `domain/proxy` (adaptor) — 多协议原生支持 | **PRD**: FastAX-PRD/07-multi-protocol.md | **参考**: one-api relay/adaptor/
### 5.9 多协议原生支持 (PRD §6.10 PROTO)

#### 5.9.1 协议适配架构

```
请求进入 (单一端口)
    │
    ▼
┌──────────────────────┐
│  协议检测 (GetByPath) │  ── 根据请求路径自动识别协议
│  /v1/chat/completions│      OpenAI
│  /v1/messages        │      Anthropic
│  /v1/models          │      OpenAI
│  /v1/rerank          │      Cohere
└──────┬───────────────┘
       │
       ▼
┌──────────────────────┐
│  Adaptor 分发 (switch) │  ── GetAdaptor(apiType) 路由到对应 adaptor
│  ┌──────┐ ┌──────┐   │
│  │OpenAI│ │Anthr.│   │      每个 Adaptor 独立实现:
│  │      │ │opic  │   │      GetRequestURL / SetupRequestHeader
│  └──────┘ └──────┘   │      ConvertRequest / DoResponse
│  ┌──────┐ ┌──────┐   │      GetModelList / GetChannelName
│  │Gemini│ │Other │   │
│  └──────┘ └──────┘   │
└──────────────────────┘
```

#### 5.9.2 协议适配器注册表

| 协议 | 适配器包 | 请求路径 | 认证 Header | 参考 one-api |
|------|---------|----------|------------|-------------|
| OpenAI | adaptor/openai | `/v1/chat/completions` | `Bearer <key>` | relay/adaptor/openai/ |
| Anthropic | adaptor/anthropic | `/v1/messages` | `x-api-key` + `anthropic-version` | relay/adaptor/anthropic/ |
| Gemini | adaptor/gemini | 原生协议端点 | `Bearer` / API Key | relay/adaptor/gemini/ |
| Azure OpenAI | adaptor/openai (泛化) | `{deployment}/chat/completions` | `api-key` header | 复用 OpenAI Adaptor 重写 GetRequestURL |

#### 5.9.3 模型后缀变体 (PRD PROTO-05)

```
model:variant 语法解析示例:

  gpt-4:nitro     → 选择 gpt-4 供应商中延迟最低的
  claude-3:floor  → 选择 Claude-3 供应商中成本最低的
  deepseek:thinking → 选择支持 reasoning 的 DeepSeek 渠道

实现: 路由引擎解析 suffix 后，在内存缓存中附加筛选条件
  - nitro   → 按 response_time 升序
  - floor   → 按 price 升序 (需 provider_health 表)
  - free    → 仅选 price=0 的渠道
```

