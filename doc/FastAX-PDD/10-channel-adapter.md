> **Domain**: `shared/relay` — 渠道适配器模式 | **参考**: one-api relay/adaptor/
### 5.6 渠道代理 (Proxy Service 子模块)

#### 5.6.1 供应商适配器模式

```
┌──────────────────────────────────────┐
│        TokenSupplier (接口)            │
│  + chat(request): Response             │
│  + models(): List<Model>               │
│  + balance(): Balance                  │
│  + health(): HealthStatus              │
└────────┬────────┬────────┬────────────┘
         │        │        │
  ┌──────┴──┐ ┌──┴──────┐ ┌┴────────┐
  │ OpenAI  │ │ Claude  │ │ DeepSeek│
  │ Adapter │ │ Adapter │ │ Adapter │
  └─────────┘ └─────────┘ └─────────┘
  ┌─────────┐ ┌─────────┐ ┌─────────┐
  │ GLM     │ │ Qwen    │ │ Vendor  │
  │ Adapter │ │ Adapter │ │ Adapter │
  └─────────┘ └─────────┘ └─────────┘
         │         │          │
    ┌────┴─────────┴──────────┴───────┐
    │     上游 HTTP API 调用           │
    │    (各供应商原生协议)             │
    └────────────────────────────────┘
```

**适配器职责**:
1. 请求格式转换（平台统一请求 → 供应商原生格式）
2. 响应格式转换（供应商原生响应 → 平台统一格式）
3. 认证处理（不同供应商的 API Key 管理）
4. 错误处理（供应商特定错误码 → 统一错误码）
5. **Vendor Adapter** (通用适配器): 供入驻供应商使用，支持动态配置 API 端点和认证方式

#### 5.6.2 国内模型 API 兼容适配 (PRD F-PXY-09)

```
国内模型 → OpenAI 兼容适配层:

  DeepSeek:  原生 API 已兼容 OpenAI 格式，无需转换
  Qwen:      需转换请求格式 (messages → input/history)
             需转换响应格式 (output.choices → choices)
  GLM:       需转换请求格式 (prompt → messages)
             需转换响应格式 (response.choices → choices)
  文心一言:   需转换认证方式 (access_token → API Key)
             需转换请求/响应格式

  统一输出:   OpenAI 兼容格式，海外用户无需额外适配
```

