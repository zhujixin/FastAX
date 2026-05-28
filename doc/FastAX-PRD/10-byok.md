> **Domain**: `domain/byok` — BYOK 自带 Key | **PDD**: §5.12

### 6.13 BYOK（自带 Key，Bring Your Own Key）

参考 OpenRouter BYOK、Segmind BYOK、TheRouter、Dymo 等平台设计。

| 功能 | 需求描述 | 优先级 | 备注 |
|------|----------|--------|------|
| BYOK-01 | **用户管理自有 Key**：用户可在个人中心添加自有供应商 API Key（OpenAI、Anthropic、Google 等），加密存储在平台 | P1 | AES-256-GCM 加密，仅用于转发 |
| BYOK-02 | **BYOK + 平台 Token 混合路由**：支持同一请求先消耗 BYOK 配额，不足时自动回退到平台 Token | P1 | 用户可配置路由优先级 |
| BYOK-03 | **BYOK 用量看板**：用户可查看自有 Key 的调用量、消耗金额、剩余额度 | P1 | 聚合多供应商统计 |
| BYOK-04 | **Key 轮换与过期管理**：自动检测 Key 过期/失效，支持 Key 轮换配置，失效时发通知 | P1 | — |
| BYOK-05 | **团队 Key 共享**：企业用户可将自有 Key 共享给子账号使用，支持额度分配 | P2 | — |
| BYOK-06 | **BYOK 模型限制**：用户可为自有 Key 设置可调用的模型白名单/黑名单 | P1 | — |
| BYOK-07 | **BYOK 平台费用**：BYOK 模式收取 ≤ 5% 的平台服务费或零加成（视套餐而定） | P1 | 参考 OpenRouter 5% 费率 |

