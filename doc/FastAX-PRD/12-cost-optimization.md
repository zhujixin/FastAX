---
domain: cost
pdd_section: "§5.14"
priority: P1
status: completed
depends_on: [proxy]
required_by: [enterprise]
version: "3.1"
last_updated: "2026-06-04"
---
> **Domain**: `domain/cost` — 成本优化 | **PDD**: §5.14

### 6.15 成本优化引擎（COST）

参考 Portkey、LiteLLM、OpenRouter 等平台的成本优化策略。

| 功能 | 需求描述 | 优先级 | 备注 |
|------|----------|--------|------|
| COST-01 | **上游 Prompt Caching**：利用供应商的 Prompt Caching 能力（如 Claude、GPT-4o），自动标记可缓存内容，降低 50-90% 输入成本 | P1 | 参考各供应商 caching 实现 |
| COST-02 | **语义缓存（Semantic Cache）**：已由独立模块 `domain/cache` 承接，详见 [22-semantic-cache.md](22-semantic-cache.md) CACHE-01~07 | P2 → 已独立 | 本条目保留作历史引用，实现以 CACHE 模块为准 |
| COST-03 | **缓存计费比率**：缓存命中的请求按比率计费（如缓存命中按原价 10% 计费），管理员可配置比率 | P1 | — |
| COST-04 | **预算封顶**：按月/日/用户维度设置成本上限，超限自动熔断或告警 | P0 | 防止预算超支 |
| COST-05 | **成本告警 Webhook**：成本达到阈值（50%/80%/90%/100%）时发送告警通知 | P0 | — |
| COST-06 | **成本感知路由策略层**：定义成本优化策略（延迟约束、成本权重、时段偏好），由 ROUTE-20（最低成本路由）和 ROUTE-18（Token 长度路由）具体执行 | P1 | 策略定义层，执行层在 ROUTE-20；ROUTE-07 已被 ROUTE-20 替代 |
| COST-07 | **模型回退链**：配置模型降级链（如 `gpt-4 → claude-3-haiku → deepseek-chat`），高成本模型不可用时自动降级 | P1 | 用户可配置 |
| COST-08 | **Token 压缩**：对支持压缩的模型启用上下文压缩（如 RTK + Caveman 压缩），降低 15-95% Token 消耗 | P2 | — |
| COST-09 | **上下文压缩网关**：当输入 token 超过阈值（如 8,000）时，自动使用廉价模型（如 gpt-4o-mini）进行上下文压缩，保留核心语义，降低 40-70% token 消耗后转发至目标模型 | P2 | 参考 LiteLLM Context Compression + SmartContext Proxy |
| COST-10 | **分层上下文策略**：实现 T0-T3 四层上下文结构 —— T0（系统提示词稳定保留）、T1（最近 3 轮对话原文保留）、T2（语义检索相关历史片段）、T3（预计算摘要替代长历史），按层级差异化压缩 | P2 | 参考 SmartContext Proxy T0-T3 分层架构 |
| COST-11 | **碳感知路由**：接入实时电网碳强度数据（如 Electricity Maps API），在满足延迟和成本约束的前提下，优先选择低碳数据中心的供应商，满足企业 ESG 目标 | P2 | 行业前沿趋势，企业客户 ESG 需求 |

---
## 相关模块

| 关系 | 模块 | 说明 |
|------|------|------|
| 依赖 | [代理模块](02-token-proxy-vendor.md) | 成本策略通过路由引擎执行 |
| 被依赖 | [语义缓存](22-semantic-cache.md) | COST-03 计费策略驱动缓存计费比率 |
| 被依赖 | [企业功能](13-enterprise.md) | ENT-04 预算控制依赖成本策略 |
| 关联 | [模型市场](14-model-marketplace.md) | 模型比价数据用于成本路由决策 |