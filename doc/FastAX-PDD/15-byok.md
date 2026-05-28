> **Domain**: `domain/byok` — BYOK 自带 Key | **PRD**: FastAX-PRD/10-byok.md
### 5.12 BYOK (自带 Key, PRD §6.13)

#### 5.12.1 BYOK 架构

```
用户自有 Key 管理流程:

  1. 用户添加 Key
     Provider: OpenAI/Anthropic/Google/...
     Key: 加密存储 (AES-256-GCM, 每 Key 独立 IV)
     模型白名单: 可选限制可用模型

  2. 路由优先级配置
     优先使用 BYOK → 余额不足时回退到平台 Token
     或: 仅使用 BYOK (不消耗平台配额)

  3. 调用链路
     请求 → 路由决策
            ├── 用户 BYOK 有余额 → 用 BYOK 转发
            └── BYOK 余额不足 → 回退到平台 Token (需用户授权)
```

| 功能 | 设计 |
|------|------|
| Key 加密存储 | `byok_key` 表 + AES-256-GCM + 应用层解密 |
| 混合路由 | 同一请求内 BYOK→平台 Token 自动回退 |
| 用量看板 | 调用量/消耗金额/剩余额度 (聚合多供应商) |
| Key 轮换 | 定时检测失效 Key + 通知 + 自动轮换 |
| 平台费用 | BYOK 调用收取 ≤ 5% 服务费或零加成 |

