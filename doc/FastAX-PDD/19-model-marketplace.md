> **Domain**: `domain/market` — 模型市场 | **PRD**: FastAX-PRD/14-model-marketplace.md
### 5.16 模型市场与发现 (PRD §6.17 MKT)

| 功能 | 设计要点 | 优先级 |
|------|---------|--------|
| 模型对比工具 | 价格/延迟/上下文窗口/能力评分 多维度对比 | P1 |
| 供应商健康面板 | 每供应商: 历史可用率/平均延迟/P95延迟/错误率 | P0 |
| 模型基准测试 | 定期推理速度/质量评分测试, 生成性能报告 | P2 |
| 推荐引擎 | 基于用户使用历史推荐性价比最优组合 | P2 |
| 变更日志 | 模型新增/下架/价格变更时推送通知 | P1 |
| 模型变体 | base_model + suffix → provider 映射 (如 gpt-4 → 多个供应商) | P1 |

#### 5.16.1 API 端点

**公开 (无需认证)**：

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/models` | GET | 模型列表 (可选 provider/modelType 筛选) |
| `/api/models/benchmarks` | GET | 基准测试数据 (可选 modelName) |
| `/api/models/variants/:variant` | GET | 模型变体详情 (base+suffix → provider) |
| `/api/providers/health` | GET | 供应商健康总览 (全部供应商状态) |
| `/api/providers/:id/health` | GET | 供应商健康详情 (历史可用率/延迟/错误率) |

**用户端 (JWT)**：

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/models/compare` | POST | 模型多维度对比 (传入 modelNames[]) |
| `/api/models/recommend` | GET | 模型推荐 (基于用户历史) |

#### 5.16.2 Service 方法

```go
// 模型探索
ListModels(provider, modelType string) ([]ModelInfo, error)
CompareModels(modelNames []string) ([]ModelComparison, error)
GetBenchmarks(modelName string) ([]BenchmarkResult, error)
RecommendModels(userID uint) ([]ModelRecommendation, error)

// 供应商健康
GetProviderHealth(providerID uint) ([]ProviderHealthRecord, error)
ListProviders() ([]ProviderStatus, error)

// 模型变体
CreateVariant(req *VariantRequest) (*ModelVariant, error)
ListVariants(baseModel string) ([]ModelVariant, error)
FindVariant(modelName string) (*ModelVariant, error)
```

#### 5.16.3 数据表

| 表名 | 说明 |
|------|------|
| `model_variants` | 模型变体映射 (base_model + suffix + provider_id → price_coefficient) |
| `provider_healths` | 供应商健康历史记录 (period_start/period_end + status/avg_latency/error_rate) |
