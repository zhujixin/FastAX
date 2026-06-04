# FastAX 缺口闭合开发计划

> **基于**：PDD 验证报告 v2 (2026-06-02)  
> **当前状态**：✅ **全部完成**，PDD API 覆盖率 100% (118/118)，25 测试包全绿  
> **实际工时**：~2 人日 | 20 个 API 全部补齐

---

## 1. 缺口总览

```
已完成 98/118 API (83%)          缺口 20 API (17%)
├── ✅ 核心链路 100%              ├── ⚠️ 供应商自服务 3 API
├── ✅ Admin 核心 100%            ├── ⚠️ Admin 渠道+系统 7 API
├── ✅ Relay 代理 100%            ├── ⚠️ STT 语音转文本 1 API
├── ✅ 数据库 28 表 100%          ├── ⚠️ BYOK 增强 2 API
├── ✅ 测试 24 包全绿             ├── ⚠️ 成本优化缓存 2 API
└── ✅ 中间件 4/4 + i18n 5语言     ├── ⚠️ 护栏全局配置 1 API
                                  ├── ⚠️ 多协议模型变体 1 API
                                  └── ⚠️ Admin 供应商结算 3 API
```

---

## 2. 缺口分阶段计划

### 阶段 G1：P1 核心缺口闭合（6 个 task，4-6 人日）

> 目标：补齐生产环境必需的核心功能缺口

#### Task G1.1 — 供应商自服务商品管理（1-1.5d）

**缺失 API（3 个）**：
- `GET /api/vendor/products` — 供应商查看自己的商品列表
- `POST /api/vendor/products` — 供应商创建商品
- `PUT /api/vendor/products/{id}/price` — 供应商调价

**现状**：`vendor.Handler.CreateProduct`、`ListProducts` 方法已存在，但仅注册在 Admin 路由下。`UpdateProduct` 已注册在 vendor 自服务路由但缺少调价专用端点。

**实施方案**（`internal/router/router.go` + `internal/domain/vendor/handler.go`）：

```
// router.go — 在 vendorGroup 中新增：
vendorGroup.GET("/products", h.Vendor.ListMyProducts)     // 新建 handler
vendorGroup.POST("/products", h.Vendor.CreateMyProduct)    // 新建 handler（带 vendor_id 注入）
vendorGroup.PUT("/products/:id/price", h.Vendor.UpdateProductPrice) // 新建 handler
```

**新增 handler 方法**（`internal/domain/vendor/handler.go`）：
- `ListMyProducts` — 从 JWT 获取 user_id → 查 supplier_vendor → 列出该 vendor 的商品
- `CreateMyProduct` — 从 JWT 获取 user_id → 查 supplier_vendor → 创建商品（vendor_id 从上下文注入）
- `UpdateProductPrice` — 仅允许修改 price 字段，含平台限价校验

**测试**：`internal/domain/vendor/handler_test.go` 新增 3 个测试用例

| 文件 | 改动类型 | 说明 |
|------|---------|------|
| `internal/domain/vendor/handler.go` | 新增 3 方法 | ListMyProducts, CreateMyProduct, UpdateProductPrice |
| `internal/domain/vendor/service.go` | 新增 1 方法 | UpdateProductPrice（或复用 UpdateProduct 加校验） |
| `internal/router/router.go` | 新增 3 路由 | 供应商自服务商品路由 |
| `internal/domain/vendor/handler_test.go` | 新增测试 | 3 个 handler 测试 |

---

#### Task G1.2 — Admin 渠道管理 CRUD（1-1.5d）

**缺失 API（3 个）**：
- `GET /api/admin/channels` — 渠道列表
- `PUT /api/admin/channels/{id}/status` — 启用/禁用渠道
- `PUT /api/admin/channels/{id}/priority` — 调整优先级

**现状**：`Ability` 模型和 `Supplier` 模型已存在，但缺乏独立的 Channel 管理端点。当前通过 Supplier CRUD 间接管理。

**实施方案**：

渠道 = Supplier + Ability 的组合视图。在 vendor domain 中新增 Channel 管理接口：

```
// router.go — 在 admin 路由组中新增：
admin.GET("/channels", h.Vendor.ListChannels)
admin.PUT("/channels/:id/status", h.Vendor.SetChannelStatus)
admin.PUT("/channels/:id/priority", h.Vendor.SetChannelPriority)
```

**新增 handler 方法**（`internal/domain/vendor/handler.go`）：
- `ListChannels` — 联表查询 Supplier + Ability，返回渠道完整信息（含模型支持列表、健康状态）
- `SetChannelStatus` — 更新 Supplier.Status（0=禁用, 1=启用），同步刷新内存缓存
- `SetChannelPriority` — 更新 Supplier.Priority，同步刷新内存缓存

**Service 层**（`internal/domain/vendor/service.go`）：
- `ListChannels(ctx) ([]ChannelDTO, error)`
- `SetChannelStatus(ctx, id, status) error`
- `SetChannelPriority(ctx, id, priority) error`

| 文件 | 改动类型 | 说明 |
|------|---------|------|
| `internal/domain/vendor/handler.go` | 新增 3 方法 | ListChannels, SetChannelStatus, SetChannelPriority |
| `internal/domain/vendor/service.go` | 新增 3 方法 | 对应的 Service 接口方法 |
| `internal/domain/vendor/model.go` | 新增 DTO | ChannelDTO（含 supplier + ability 聚合字段） |
| `internal/router/router.go` | 新增 3 路由 | Admin 渠道管理路由 |
| `internal/domain/vendor/handler_test.go` | 新增测试 | 3 个 handler 测试 |

---

#### Task G1.3 — Admin 供应商佣金配置（0.5d）

**缺失 API（1 个）**：
- `PUT /api/admin/vendor-commission-rates/{id}` — 配置供应商佣金比例

**现状**：`SupplierVendor.CommissionRate` 字段已存在，但无独立更新端点。

**实施方案**：

```
// router.go — 在 admin 路由组中新增：
admin.PUT("/vendor-commission-rates/:id", h.Vendor.UpdateCommissionRate)
```

| 文件 | 改动类型 | 说明 |
|------|---------|------|
| `internal/domain/vendor/handler.go` | 新增 1 方法 | UpdateCommissionRate |
| `internal/domain/vendor/service.go` | 新增 1 方法 | UpdateCommissionRate |
| `internal/router/router.go` | 新增 1 路由 | 佣金配置路由 |

---

#### Task G1.4 — Admin 供应商结算记录查看（0.5d）

**缺失 API（1 个）**：
- `GET /api/admin/vendors/{id}/settlements` — 管理员查看某供应商的结算记录

**现状**：`Settlement` 模型已存在，供应商自服务已有 `GET /api/vendor/settlements`，但 Admin 端缺少按 vendor_id 的查询端点。

**实施方案**：

```
// router.go — 在 admin 路由组中新增：
admin.GET("/vendors/:id/settlements", h.Vendor.ListVendorSettlements)
```

| 文件 | 改动类型 | 说明 |
|------|---------|------|
| `internal/domain/vendor/handler.go` | 新增 1 方法 | ListVendorSettlements (admin 视角) |
| `internal/router/router.go` | 新增 1 路由 | Admin 供应商结算路由 |

---

#### Task G1.5 — STT 语音转文本端点（0.5-1d）

**缺失 API（1 个）**：
- `POST /v1/audio/transcriptions` — 语音转文本（Whisper 兼容）

**现状**：`POST /v1/audio/speech`（TTS）已实现，但 STT transcriptions 端点缺失。proxy handler 中有 `AudioSpeech` 方法但无 `AudioTranscriptions`。

**实施方案**：

```
// router.go — 在 v1 路由组中新增：
v1.POST("/audio/transcriptions", h.Proxy.AudioTranscriptions)
```

**新增**（`internal/domain/proxy/handler.go`）：
- `AudioTranscriptions` — 接受 multipart/form-data（音频文件 + model 参数），转发到供应商的 STT 端点

| 文件 | 改动类型 | 说明 |
|------|---------|------|
| `internal/domain/proxy/handler.go` | 新增 1 方法 | AudioTranscriptions |
| `internal/domain/proxy/service.go` | 新增 1 方法 | TranscribeAudio |
| `internal/router/router.go` | 新增 1 路由 | /v1/audio/transcriptions |

---

#### Task G1.6 — BYOK 用量统计（0.5d）

**缺失 API（1 个）**：
- `GET /api/byok/usage` — BYOK 用量统计

**现状**：`byok.Handler.ListKeys` 已返回 Key 列表含 `LastUsedAt`，但缺少聚合用量统计端点。`BYOKKey` 模型有关联使用记录的潜力（通过 `CallLog`）。

**实施方案**：

```
// router.go — 在 byokGroup 中新增：
byokGroup.GET("/usage", h.BYOK.GetUsage)
```

**新增**（`internal/domain/byok/handler.go` + `service.go`）：
- `GetUsage` — 聚合查询当前用户所有 BYOK Key 的总调用次数、Token 消耗、费用估算（从 CallLog 表按 supplier 类型聚合）

| 文件 | 改动类型 | 说明 |
|------|---------|------|
| `internal/domain/byok/handler.go` | 新增 1 方法 | GetUsage |
| `internal/domain/byok/service.go` | 新增 1 方法 | GetUsageStats |
| `internal/router/router.go` | 新增 1 路由 | BYOK 用量路由 |

---

### 阶段 G2：P2 增强缺口闭合（5 个 task，4-6 人日）

> 目标：补齐 P2 优先级的管理功能和配置端点

#### Task G2.1 — Admin 系统配置管理（1-1.5d）

**缺失 API（2 个）**：
- `GET /api/admin/system/config` — 查看系统配置
- `PUT /api/admin/system/config` — 更新系统配置

**现状**：系统配置通过 `config.yaml` 文件和 Viper 管理，无运行时 API 修改能力。

**实施方案**：

新增 `system_config` 表（或复用现有方案），支持运行时配置查询和热更新：

```sql
CREATE TABLE system_config (
  id INTEGER PRIMARY KEY,
  config_key TEXT UNIQUE NOT NULL,
  config_value TEXT NOT NULL,
  description TEXT,
  updated_at INTEGER NOT NULL
);
```

| 文件 | 改动类型 | 说明 |
|------|---------|------|
| `internal/shared/model/system_config.go` | 新增文件 | SystemConfig GORM 模型 |
| `internal/domain/system/handler.go` | 新增文件 | GetConfig, UpdateConfig handler |
| `internal/domain/system/service.go` | 新增文件 | SystemService interface + 实现 |
| `internal/router/router.go` | 新增 2 路由 + Handler 注入 | Admin 系统配置路由 |

**或者更简单的方案**：直接在现有的 admin 路由中用内联 handler 读写 `config.yaml`（通过 Viper），无需新建 domain。

---

#### Task G2.2 — Admin 管理员账号管理（1d）

**缺失 API（2 个）**：
- `GET /api/admin/system/admins` — 管理员列表
- `POST /api/admin/system/admins` — 添加管理员

**现状**：管理员通过 `user.role = 'admin'` 或 `'super_admin'` 标识，但无独立的管理员管理端点。

**实施方案**（复用现有 user domain）：

```
// router.go — 在 admin 路由组中新增：
admin.GET("/system/admins", h.User.ListAdmins)
admin.POST("/system/admins", h.User.CreateAdmin)
```

| 文件 | 改动类型 | 说明 |
|------|---------|------|
| `internal/domain/user/handler.go` | 新增 2 方法 | ListAdmins (筛选 role=admin/super_admin), CreateAdmin |
| `internal/domain/user/service.go` | 新增 2 方法 | 对应的 Service 方法 |
| `internal/router/router.go` | 新增 2 路由 | Admin 管理员路由 |

---

#### Task G2.3 — 护栏全局配置 + 模型变体详情（0.5-1d）

**缺失 API（2 个）**：
- `PUT /api/admin/guardrails/config` — 护栏全局配置（默认模式、全局开关）
- `GET /models/:variant` — 模型变体详情

**实施方案**：

**护栏全局配置**（`internal/domain/guardrail/handler.go`）：
- 新增 `UpdateConfig` handler，更新护栏全局运行模式（enforce/monitor/log）和全局开关
- 配置可存储在 `system_config` 表或 guardrail 专用配置中

**模型变体详情**（`internal/domain/proxy/handler.go` 或 `internal/domain/market/handler.go`）：
- 新增 `GetModelVariant` handler，解析 `:variant` 后缀，返回变体详情（价格系数、优先级、可用供应商）

```
// router.go：
admin.PUT("/guardrails/config", h.Guardrail.UpdateConfig)
v1.GET("/models/:variant", h.Proxy.GetModelVariant)  // 或注册为 /models/variants/:variant
```

| 文件 | 改动类型 | 说明 |
|------|---------|------|
| `internal/domain/guardrail/handler.go` | 新增 1 方法 | UpdateConfig |
| `internal/domain/guardrail/service.go` | 新增 1 方法 | UpdateGlobalConfig |
| `internal/domain/proxy/handler.go` 或 `market/handler.go` | 新增 1 方法 | GetModelVariant |
| `internal/router/router.go` | 新增 2 路由 | 护栏配置 + 模型变体 |

---

#### Task G2.4 — BYOK 路由偏好配置（0.5d）

**缺失 API（1 个）**：
- `PUT /api/byok/preference` — BYOK 路由优先级配置

**实施方案**：

在 `byok_key` 表或用户配置中新增 `routing_preference` 字段（JSON），定义 BYOK 使用策略：
```json
{
  "mode": "byok_first",      // byok_first | platform_only | byok_only
  "fallback_enabled": true,   // BYOK 不足时是否回退平台
  "max_platform_fee_pct": 5   // 最大平台费率
}
```

| 文件 | 改动类型 | 说明 |
|------|---------|------|
| `internal/domain/byok/handler.go` | 新增 1 方法 | SetPreference |
| `internal/domain/byok/service.go` | 新增 1 方法 | UpdatePreference |
| `internal/shared/model/byok.go` | 新增字段/表 | RoutingPreference |
| `internal/router/router.go` | 新增 1 路由 | BYOK 偏好路由 |

---

#### Task G2.5 — 语义缓存统计与配置（1-1.5d）

**缺失 API（2 个）**：
- `GET /api/cache/stats` — 语义缓存统计（命中率、缓存条目数、节省成本）
- `PUT /api/cache/config` — 缓存策略配置（相似度阈值、TTL、最大条目数）

**现状**：`SemanticCache` 模型已存在，`cost.Service` 有预算和告警功能，但缺少缓存统计和配置端点。

**实施方案**（`internal/domain/cost/handler.go` + `service.go`）：

```
// router.go — 在 admin 路由组中新增：
admin.GET("/cache/stats", h.Cost.GetCacheStats)
admin.PUT("/cache/config", h.Cost.UpdateCacheConfig)
```

- `GetCacheStats` — 查询 `semantic_caches` 表：总条目数、总命中次数、命中率、估算节省费用
- `UpdateCacheConfig` — 更新缓存策略参数（阈值、TTL 等），可存储在 `system_config` 表

| 文件 | 改动类型 | 说明 |
|------|---------|------|
| `internal/domain/cost/handler.go` | 新增 2 方法 | GetCacheStats, UpdateCacheConfig |
| `internal/domain/cost/service.go` | 新增 2 方法 | 对应的 Service 方法 |
| `internal/router/router.go` | 新增 2 路由 | Admin 缓存管理路由 |

---

### 阶段 G3：文档同步与收尾（1-2 人日）

#### Task G3.1 — PDD 文档回补（0.5-1d）

将额外实现的 20+ 个 API 补充到 PDD 文档中：
- 支付独立路由（5 个）
- 企业子账号管理（6 个）
- 佣金提现（3 个）
- 模型推荐/市场扩展（4 个）
- 护栏实时检测（1 个）
- 通知模板管理（3 个）

| 文件 | 说明 |
|------|------|
| `doc/FastAX-PDD/03-api.md` | 补充 §7.7 扩展 API 章节 |

#### Task G3.2 — 全量回归测试（0.5d）

```bash
go test ./... -count=1 -race
go build -o bin/fastax ./cmd/fastax
```

确保 G1-G2 新增代码不破坏现有测试。

#### Task G3.3 — 最终 PDD 验证（0.5d）

更新 `doc/PDD-verification-report.md`，确认 118/118 API 全部实现。

---

## 3. 工时汇总

| 阶段 | Task | 内容 | 人日 |
|------|------|------|------|
| **G1** | G1.1 | 供应商自服务商品管理 (3 API) | 1-1.5d |
| | G1.2 | Admin 渠道管理 CRUD (3 API) | 1-1.5d |
| | G1.3 | Admin 供应商佣金配置 (1 API) | 0.5d |
| | G1.4 | Admin 供应商结算查看 (1 API) | 0.5d |
| | G1.5 | STT 语音转文本 (1 API) | 0.5-1d |
| | G1.6 | BYOK 用量统计 (1 API) | 0.5d |
| **G2** | G2.1 | Admin 系统配置管理 (2 API) | 1-1.5d |
| | G2.2 | Admin 管理员账号管理 (2 API) | 1d |
| | G2.3 | 护栏全局配置 + 模型变体 (2 API) | 0.5-1d |
| | G2.4 | BYOK 路由偏好 (1 API) | 0.5d |
| | G2.5 | 语义缓存统计与配置 (2 API) | 1-1.5d |
| **G3** | G3.1 | PDD 文档回补 | 0.5-1d |
| | G3.2 | 全量回归测试 | 0.5d |
| | G3.3 | 最终 PDD 验证 | 0.5d |
| | | **总计** | **10-15d** |

---

## 4. 文件改动清单

| 文件 | G1.1 | G1.2 | G1.3 | G1.4 | G1.5 | G1.6 | G2.1 | G2.2 | G2.3 | G2.4 | G2.5 |
|------|:----:|:----:|:----:|:----:|:----:|:----:|:----:|:----:|:----:|:----:|:----:|
| `internal/router/router.go` | ✏️ | ✏️ | ✏️ | ✏️ | ✏️ | ✏️ | ✏️ | ✏️ | ✏️ | ✏️ | ✏️ |
| `internal/domain/vendor/handler.go` | ✏️ | ✏️ | ✏️ | ✏️ | | | | | | | |
| `internal/domain/vendor/service.go` | ✏️ | ✏️ | ✏️ | | | | | | | | |
| `internal/domain/vendor/model.go` | | ✏️ | | | | | | | | | |
| `internal/domain/proxy/handler.go` | | | | | ✏️ | | | | ✏️ | | |
| `internal/domain/proxy/service.go` | | | | | ✏️ | | | | | | |
| `internal/domain/byok/handler.go` | | | | | | ✏️ | | | | ✏️ | |
| `internal/domain/byok/service.go` | | | | | | ✏️ | | | | ✏️ | |
| `internal/shared/model/byok.go` | | | | | | | | | | ✏️ | |
| `internal/domain/user/handler.go` | | | | | | | | ✏️ | | | |
| `internal/domain/user/service.go` | | | | | | | | ✏️ | | | |
| `internal/domain/guardrail/handler.go` | | | | | | | | | ✏️ | | |
| `internal/domain/guardrail/service.go` | | | | | | | | | ✏️ | | |
| `internal/domain/cost/handler.go` | | | | | | | | | | | ✏️ |
| `internal/domain/cost/service.go` | | | | | | | | | | | ✏️ |
| `internal/shared/model/system_config.go` | | | | | | | 🆕 | | | | |
| `internal/domain/system/` | | | | | | | 🆕 | | | | |
| **测试文件** | ✏️ | ✏️ | ✏️ | ✏️ | ✏️ | ✏️ | ✏️ | ✏️ | ✏️ | ✏️ | ✏️ |

> ✏️ = 修改现有文件 | 🆕 = 新增文件

---

## 5. 依赖关系

```
G1.1 (vendor 商品) ──┐
G1.2 (渠道管理)   ──┤
G1.3 (佣金配置)   ──┤ 无相互依赖，可并行
G1.4 (结算查看)   ──┤
G1.5 (STT)        ──┤
G1.6 (BYOK 用量) ──┘
        │
        ▼
G2.1 (系统配置) ──┐
G2.2 (管理员)   ──┤ 依赖 G1 经验，可并行
G2.3 (护栏+变体)──┤
G2.4 (BYOK 偏好)──┤
G2.5 (缓存统计) ──┘
        │
        ▼
G3.1 (文档) ──→ G3.2 (回归测试) ──→ G3.3 (最终验证)
```

**G1 阶段 6 个 task 可完全并行执行**（操作不同 domain 文件，无冲突）。

---

## 6. 实施原则

1. **最小改动** — 优先复用现有 handler/service 方法，仅在路由层新增注册
2. **先路由后逻辑** — 先注册路由确认 API 签名，再实现 handler/service 逻辑
3. **测试先行** — 每个 task 先写测试用例（TDD），再实现功能
4. **保持全绿** — 任何时候 `go test ./...` 必须全部通过
5. **增量提交** — 每个 task 独立 commit，便于 review 和回滚

---

## 7. 关键注意事项

| 注意点 | 说明 |
|--------|------|
| **vendor 商品创建** | CreateMyProduct 必须从 JWT 注入 vendor_id，防止供应商冒用其他 vendor 身份 |
| **渠道管理** | SetChannelStatus/SetChannelPriority 修改后需触发 `SyncChannelCache` 刷新内存缓存 |
| **STT multipart** | AudioTranscriptions 接受 `multipart/form-data`，与 JSON API 的 handler 签名不同，需特殊处理 |
| **系统配置存储** | 可选择新建 `system_config` 表或继续用 Viper 热重载。MVP 阶段建议用 Viper + 文件写回方案 |
| **BYOK 偏好** | RoutingPreference 字段可先以 JSON TEXT 存储在 byok_key 表或 user_profile 扩展字段中 |
| **护栏全局配置** | UpdateConfig 需并发安全（sync.RWMutex 保护），避免读写竞争 |
