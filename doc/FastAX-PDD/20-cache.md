> **Domain**: `domain/cache` — 语义缓存引擎 | **PRD**: FastAX-PRD/22-semantic-cache.md
### 5.17 语义缓存引擎 (PRD §6.18 CACHE)

#### 5.17.1 双层缓存架构

```
请求到达 → [CacheMiddleware]
              ├── L1: SHA256 精确匹配 (Redis, <1ms)
              │       ├── 命中 → 返回缓存响应 + Header: x-fastax-cache-hit=exact
              │       └── 未命中 → L2
              ├── L2: 语义向量搜索 (chromem-go + ONNX all-MiniLM-L6-v2, 5-15ms)
              │       ├── 余弦相似度 ≥ 0.90 → 返回 + Header: x-fastax-cache-hit=semantic
              │       ├── 0.80 ≤ 相似度 < 0.90 → L3 (灰度区验证)
              │       └── 相似度 < 0.80 → MISS → 转发 LLM
              ├── L3: 灰度区验证 (廉价 LLM gpt-4o-mini, 5-10ms)
              │       ├── 验证为同义 → 返回 + Header: x-fastax-cache-hit=verified
              │       └── 验证为不同 → MISS → 转发 LLM
              └── 转发 LLM → 响应写入 L1+L2 缓存 → 返回 + Header: x-fastax-cache-hit=miss
```

| 层级 | 存储 | 延迟 | 命中率预期 | 误命中风险 |
|------|------|------|-----------|-----------|
| L1 精确匹配 | Redis String (SHA256 key) | <1ms | 15-25% | 0% (精确匹配) |
| L2 语义向量 | chromem-go 内存索引 + Redis 持久化 | 5-15ms | 30-50% | <2% (余弦阈值控制) |
| L3 灰度区验证 | LLM (gpt-4o-mini) 判定 | +5-10ms | 额外 3-5% | <0.5% (经 LLM 确认) |

#### 5.17.2 缓存控制 Header 规范

| Header | 方向 | 类型 | 说明 |
|--------|------|------|------|
| `x-fastax-cache-ttl` | 请求 | int (seconds) | 自定义缓存 TTL，覆盖默认值 |
| `x-fastax-cache-skip` | 请求 | "true"/"false" | 跳过缓存，强制执行 LLM 调用 |
| `x-fastax-cache-key` | 请求 | string | 自定义缓存键（替代 SHA256） |
| `x-fastax-cache-namespace` | 请求 | string | 命名空间隔离（如 `team:123`） |
| `x-fastax-cache-hit` | 响应 | "exact"/"semantic"/"verified"/"miss" | 命中类型标识 |
| `x-fastax-cache-savings` | 响应 | string | 本次节省金额（如 `$0.015`） |

#### 5.17.3 缓存命名空间隔离

```
Redis Key: cache:{namespace}:{model}:{sha256}
  - 普通用户: cache:user:{user_id}:{model}:{sha256}
  - 团队用户: cache:team:{team_id}:{model}:{sha256}
  - 全局共享: cache:global:{model}:{sha256}  (仅 FAQ/文档类问答)
```

#### 5.17.4 API 端点

| 接口 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/admin/cache/stats` | GET | Admin | 缓存统计（命中率/条目数/节省金额） |
| `/api/admin/cache/config` | PUT | Admin | 缓存策略配置（阈值/TTL/最大条目数） |
| `/api/admin/cache/warm` | POST | Admin | 缓存预热（批量导入历史 prompt-response） |
| `/api/admin/cache/clear` | DELETE | Admin | 清除指定 namespace/model 的缓存 |
| `/api/admin/cache/namespaces` | GET | Admin | 命名空间列表及各自统计 |
| `/api/user/cache/stats` | GET | JWT | 当前用户的缓存命中统计 |

#### 5.17.5 与 Proxy 集成点

```
ProxyService.Relay() 入口:
  1. CacheMiddleware.Get(ctx, req)  → 命中则直接返回
  2. 继续现有路由+转发流水线
  3. CacheMiddleware.Set(ctx, req, resp) → 异步写入 L1+L2

集成方式: Gin 中间件 + ProxyService 装饰器，不侵入现有业务逻辑。
```

#### 5.17.6 Service 方法

```go
// CacheService — 语义缓存核心服务
type CacheService struct {
    exact    *ExactCache      // L1: Redis SHA256 精确匹配
    semantic *SemanticCache   // L2: chromem-go 向量搜索
    verifier *LLMVerifier     // L3: 灰度区 LLM 验证
}

// L1: 精确匹配
func (s *CacheService) ExactGet(ctx context.Context, req *ChatRequest) (*CachedResponse, error)
func (s *CacheService) ExactSet(ctx context.Context, req *ChatRequest, resp *ChatResponse, ttl time.Duration) error

// L2: 语义搜索
func (s *CacheService) SemanticSearch(ctx context.Context, prompt string, namespace string) (*CacheHit, error)
func (s *CacheService) SemanticStore(ctx context.Context, entry *SemanticEntry) error

// L3: 灰度区验证
func (s *CacheService) VerifyIntent(ctx context.Context, prompt1, prompt2 string) (bool, error)

// 统一入口
func (s *CacheService) Get(ctx context.Context, req *ChatRequest) (*CachedResponse, error)
func (s *CacheService) Set(ctx context.Context, req *ChatRequest, resp *ChatResponse) error

// 管理操作
func (s *CacheService) GetStats(namespace string) (*CacheStats, error)
func (s *CacheService) WarmCache(ctx context.Context, entries []*SemanticEntry) (int, error)
func (s *CacheService) ClearNamespace(ctx context.Context, namespace string) (int64, error)
func (s *CacheService) UpdateConfig(ctx context.Context, cfg *CacheConfig) error
```

#### 5.17.7 数据表

| 表名 | 说明 | 关键字段 |
|------|------|---------|
| `exact_caches` | L1 精确匹配缓存元数据 | sha256_hash, namespace, model, ttl, created_at, hit_count |
| `semantic_cache_entries` | L2 语义缓存条目 | prompt_hash, prompt_text, response_encrypted (AES-256-GCM), embedding (BLOB 384维), model, namespace, similarity_threshold |
| `cache_namespaces` | 缓存命名空间配置 | namespace_key, isolation_mode (user/team/global), max_entries, default_ttl |
| `cache_stats` | 缓存统计聚合 | namespace, hit_count, miss_count, exact_hits, semantic_hits, verified_hits, cost_saved_usd, updated_at |

#### 5.17.8 配置项

```yaml
cache:
  enabled: true
  exact:
    ttl: 1h
    max_entries: 100000
  semantic:
    embedding_model: "all-MiniLM-L6-v2"  # ONNX 本地推理
    dimensions: 384
    high_threshold: 0.90
    gray_threshold: 0.80
    max_vectors: 1000000
  verifier:
    enabled: false
    model: "gpt-4o-mini"
  stream:
    enabled: false
    max_events: 10000
```

---
## 相关模块

| 关系 | 模块 | 说明 |
|------|------|------|
| 依赖 | [代理模块](06-token-proxy-vendor.md) | 缓存引擎拦截代理转发请求 |
| 依赖 | [成本优化](17-cost-optimization.md) | 缓存计费比率由成本策略驱动 |
| 被依赖 | [可观测性](21-otel.md) | 缓存命中率/延迟写入 OTel Span |
