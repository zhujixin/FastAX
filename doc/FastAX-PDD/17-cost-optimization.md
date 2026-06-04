> **Domain**: `domain/cost` — 成本优化 | **PRD**: FastAX-PRD/12-cost-optimization.md
### 5.14 成本优化引擎 (PRD §6.15 COST)

#### 5.14.1 语义缓存

```
语义缓存流程:

  请求 Prompt → 向量化 (ONNX/sentence-transformers)
      → 在 Redis Stack / pgvector 中检索相似向量
          ├── 命中 (相似度 > 阈值) → 返回缓存响应 (按缓存计费比率计费)
          └── 未命中 → 转发供应商 → 缓存响应+向量 → 返回

缓存计费:
  缓存命中按原价的 10% 计费 (可配置)
  缓存条目加密存储 (AES-256-GCM)
  用户可要求删除缓存数据 (合规要求)
```

#### 5.14.2 预算与告警 (DB 持久化)

```
预算封顶:
  - 按 月/周/日 维度设置成本上限
  - 超限自动熔断 (仅管理员可解除)
  - 阈值告警: 50%/80%/90%/100%
  - 数据存储: user_budgets 表 (SQLite 持久化)
  - 消费追踪: 内存计数器 + 定期同步到 DB

成本告警:
  - 多阈值配置 (如 [50, 80, 100] 表示 50%/80%/100% 时告警)
  - 数据存储: cost_alerts 表 (SQLite 持久化，thresholds 以 JSON 数组存储)
  - 实时检查: CheckAlerts(userID) 返回当前触发阈值列表

模型回退链:
  用户配置: gpt-4 → claude-3-haiku → deepseek-chat
  高成本模型不可用/超预算时自动沿回退链降级
```

#### 5.14.3 缓存管理 API

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/admin/cache/stats` | GET | Admin | 语义缓存统计 (条目数/命中率/预估节省) |
| `/api/admin/cache/config` | PUT | Admin | 缓存策略配置 (相似度阈值/TTL/最大条目数) |

#### 5.14.4 Service 方法

```go
// 预算管理
SetBudget(userID, period, limit) (*BudgetSetting, error)
GetBudget(userID) (*BudgetStatus, error)
CheckBudget(userID) (bool, float64, error)
RecordSpending(userID, amount)

// 告警管理
SetAlert(userID, thresholds) (*AlertSetting, error)
GetAlerts(userID) (*AlertSetting, error)
CheckAlerts(userID) ([]float64, error)

// 语义缓存
SetCache(req *CacheRequest) error
GetCache(promptHash, modelName) (*SemanticCache, error)
CleanExpired() (int64, error)

// 成本追踪
GetCostBySupplier(period) ([]CostRecord, error)
GetCostByModel(period) ([]CostRecord, error)

// 缓存管理
GetCacheStats() (*CacheStats, error)
UpdateCacheConfig(req) error
```

#### 5.14.5 数据表

| 表名 | 说明 |
|------|------|
| `semantic_caches` | 语义缓存条目 (prompt_hash + response_encrypted) |
| `user_budgets` | 用户预算设置 (user_id unique, period, limit, spent) |
| `cost_alerts` | 成本告警配置 (user_id unique, thresholds JSON) |
