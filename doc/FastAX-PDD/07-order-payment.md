> **Domain**: `domain/order` + `domain/payment` — 订单/支付 | **PRD**: FastAX-PRD/03-order-payment.md
### 5.3 交易模块 (Order + Payment Service)

#### 5.3.1 订单状态机

```
                    ┌──────┐
                    │ DRAFT│  (预留订单)
                    └──┬───┘
                       │ 提交
                       ▼
                  ┌─────────┐
         ┌───────→│ PENDING │←────── 超时恢复
         │        └────┬────┘
         │             │ 支付超时 (30min)
         │             ▼              ┌──────────┐
         │        ┌──────────┐        │ CANCELLED │
         │        │ TIMEOUT  │────────→ (系统取消) │
         │        └──────────┘        └──────────┘
         │
    ┌────┴─────┐        ┌──────────┐
    │ PAID     │────────→│ REFUNDING│──→ 管理员审核
    └────┬─────┘        └──────────┘       │
         │ 发放 Token                      │ 审核通过
         ▼                                 ▼
    ┌─────────┐                       ┌──────────┐
    │COMPLETED│                       │ REFUNDED │
    └─────────┘                       └──────────┘
```

#### 5.3.2 Token 配额计费模型 (参考 one-api quota)

**两阶段配额消费** (PRD R-BIZ-15):

```
请求处理流程中的配额操作:

  1. 请求到达 Proxy Service
     │
     ▼
  2. 预扣配额 (PreConsume)
     - 根据输入 tokens 数估算
     - 预扣额度: min(estimated_tokens × price, PreConsumedQuota)
     - PreConsumedQuota 默认 100 tokens (可配置)
     - Redis DECRBY 操作，毫秒级
     │
     ▼
  3. 转发请求到供应商
     │
     ├── 成功 ──→ 4. 实扣配额 (PostConsume)
     │               - 从 response 获取实际 usage
     │               - 计算: actual = prompt + completion tokens
     │               - 多退少补: 退回 pre - actual 差值
     │
     └── 失败 ──→ 4. 退回全部预扣配额
```

**配额计算公式** (PRD R-BIZ-16):

```
实际扣费 = 分组倍率 × 模型倍率 × (PromptToken数 + CompletionToken数 × 补全倍率)

  分组倍率:   default=1.0, vip=0.8, enterprise=0.7
  模型倍率:   gpt-4=30, gpt-3.5=1, claude-3=15, deepseek-chat=0.5
  补全倍率:   gpt-3.5-turbo=1.33, gpt-4=2.0, 默认为 1.0
```

**批量更新策略** (PRD R-BIZ-17, 参考 one-api BatchUpdateEnabled):

```
小额配额更新不直接写 DB，先缓存后批量刷入:

  请求处理 (高频) → Redis (CacheDecreaseUserQuota)
                       │
                       │ 定时/定量刷入
                       ├── 每 10s (batchUpdateInterval)
                       └── 满 100 条 (batchUpdateThreshold)
                       ↓
                    SQLite (批量写入)

配置参数:
  batchUpdateEnabled:    true
  batchUpdateInterval:   10 (秒)
  batchUpdateThreshold:  100 (条)
  preConsumedQuota:     100 (预扣额度, tokens)
```

#### 5.3.3 支付对账方案

```
每日对账流程:

  1. T+1 日 02:00 自动触发对账
  2. 拉取我方支付记录 (payment 表, 日期=T)
  3. 拉取微信/支付宝/Stripe 的交易流水 (API)
  4. 逐笔匹配: order_id, amount, status
  5. 差异处理:
     - 我方有、网关无 → 标记为 "疑似未支付", 人工复核
     - 网关有、我方无 → 标记为 "网关异常", 补入账
     - 金额不一致 → 标记为 "金额差异", 人工复核
  6. 生成对账报告，通知财务
```

